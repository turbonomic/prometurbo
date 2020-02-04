package mediationcontainer

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	Closed           TransportStatus = "closed"
	Ready            TransportStatus = "ready"
	handshakeTimeout                 = 60 * time.Second
	wsReadLimit                      = 33554432 // 32 MB
	writeWaitTimeout                 = 120 * time.Second
	pingPeriod                       = 60 * time.Second
)

type TransportStatus string

type WebSocketConnectionConfig MediationContainerConfig

func CreateWebSocketConnectionConfig(connConfig *MediationContainerConfig) (*WebSocketConnectionConfig, error) {
	_, err := url.ParseRequestURI(connConfig.LocalAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse local URL from %s when create WebSocketConnectionConfig.", connConfig.LocalAddress)
	}
	// Change URL scheme from ws to http or wss to https.
	serverURL, err := url.ParseRequestURI(connConfig.TurboServer)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse turboServer URL from %s when create WebSocketConnectionConfig.", connConfig.TurboServer)
	}
	switch serverURL.Scheme {
	case "http":
		serverURL.Scheme = "ws"
	case "https":
		serverURL.Scheme = "wss"
	}
	wsConfig := WebSocketConnectionConfig(*connConfig)
	wsConfig.TurboServer = serverURL.String()
	// If the path specified for the platform use it instead of assuming the default /vmturbo/remoteMediation
	if serverURL.Path != "" && serverURL.Path != "/" {
		wsConfig.WebSocketPath = ""
	}
	return &wsConfig, nil
}

// ============================ ClientWebSocketTransport - Start, Connect, Close ======================================
// Implementation of the ITransport for WebSocket communication to send and receive serialized protobuf message bytes
type ClientWebSocketTransport struct {
	status                   TransportStatus // current status of the transport layer.
	wsMux                    sync.Mutex      // protect ws from concurrent writing (ws.WriteMessage())
	ws                       *websocket.Conn // created during Connect()
	connConfig               *WebSocketConnectionConfig
	inputStreamCh            chan []byte // unbuffered channel
	closeRequested           bool
	stopListenerCh           chan bool // buffered channel
	connClosedNotificationCh chan bool // channel where the transport connection error will be notified
}

// Instantiate a new ClientWebSocketTransport endpoint for the client
func CreateClientWebSocketTransport(connConfig *WebSocketConnectionConfig) *ClientWebSocketTransport {
	transport := &ClientWebSocketTransport{
		connConfig:               connConfig,
		connClosedNotificationCh: make(chan bool),
	}
	return transport
}

// WebSocket connection is established with the server
func (clientTransport *ClientWebSocketTransport) Connect() error {
	// Close any previous connected WebSocket connection and set current connection to nil.
	clientTransport.closeAndResetWebSocket()

	// loop till server is up or close received
	clientTransport.closeRequested = false
	// TODO: give an optional timeout to wait for server in performWebSocketConnection()
	err := clientTransport.performWebSocketConnection() // Blocks or till transport is closed
	if err != nil {
		return fmt.Errorf("Cannot connect with server websocket %s", err)
	}

	glog.V(4).Infof("[Connect] Connected to server " + clientTransport.GetConnectionId())

	clientTransport.stopListenerCh = make(chan bool, 1) // Channel to stop the routine that listens for messages
	clientTransport.inputStreamCh = make(chan []byte)   // Message Queue
	// Message handler for received messages
	go clientTransport.ListenForMessages()
	go clientTransport.startPing()
	return nil
}

func (clientTransport *ClientWebSocketTransport) NotifyClosed() chan bool {
	return clientTransport.connClosedNotificationCh
}

func (clientTransport *ClientWebSocketTransport) GetConnectionId() string {
	if clientTransport.status == Closed {
		return ""
	}
	return clientTransport.ws.RemoteAddr().String() + "::" + clientTransport.ws.LocalAddr().String()
}

// Close the WebSocket Transport point: this is called by upper module (remoteMediationClient)
func (clientTransport *ClientWebSocketTransport) CloseTransportPoint() {
	glog.V(4).Infof("[CloseTransportPoint] closing transport endpoint and listener routine")
	clientTransport.closeRequested = true
	// close listener
	clientTransport.stopListenForMessages()
	clientTransport.closeAndResetWebSocket()
}

// Close current WebSocket connection and set it to nil.
func (clientTransport *ClientWebSocketTransport) closeAndResetWebSocket() {
	if clientTransport.status == Closed {
		return
	}

	// close WebSocket
	if clientTransport.ws != nil {
		glog.V(1).Infof("Begin to send websocket Close frame.")
		clientTransport.write(websocket.CloseMessage, []byte{})
		clientTransport.ws.Close()
		clientTransport.ws = nil
	}
	clientTransport.status = Closed
}

func (ws *ClientWebSocketTransport) write(mtype int, payload []byte) error {
	ws.wsMux.Lock()
	defer ws.wsMux.Unlock()
	ws.ws.SetWriteDeadline(time.Now().Add(writeWaitTimeout))
	return ws.ws.WriteMessage(mtype, payload)
}

// keep sending Ping msg to make sure the websocket connection is alive
// If don't send Ping msg, *some times* the ws.ReadMessage() won't be able to
//    know that the connection has gone.
func (ws *ClientWebSocketTransport) startPing() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ws.stopListenerCh:
			return
		case <-ticker.C:
			glog.V(3).Infof("begin to send Ping message")
			if err := ws.write(websocket.PingMessage, []byte{}); err != nil {
				glog.Errorf("Failed to send PingMessage to server:%v", err)
				return
			}
			glog.V(4).Infof("Sent ping message success")
		}
	}
}

// ================================================= Message Listener =============================================
//TODO: avoid close a closed channel
func (clientTransport *ClientWebSocketTransport) stopListenForMessages() {
	if clientTransport.stopListenerCh != nil {
		glog.V(4).Infof("[StopListenForMessages] closing stopListenerCh %+v", clientTransport.stopListenerCh)
		clientTransport.stopListenerCh <- true
		close(clientTransport.stopListenerCh)
		glog.V(4).Infof("[StopListenForMessages] closed stopListenerCh %+v", clientTransport.stopListenerCh)
	}
}

// Routine to listen for messages on the websocket.
// The websocket is continuously checked for messages and queued on the clientTransport.inputStream channel
// Routine exits when a message is sent on clientTransport.stopListenerCh.
//
func (clientTransport *ClientWebSocketTransport) ListenForMessages() {
	glog.V(3).Infof("[ListenForMessages] %s : ENTER  ", time.Now())
	defer close(clientTransport.inputStreamCh) //notify the receiver that websocket stop feeding data

	for {
		glog.V(4).Info("[ListenForMessages] waiting for messages on websocket transport")
		glog.V(4).Infof("[ListenForMessages] waiting for messages on websocket transport : %++v", clientTransport)
		select {
		case <-clientTransport.stopListenerCh:
			glog.V(1).Info("[ListenForMessages] stop listening for message")
			return
		default:
			if clientTransport.status != Ready {
				glog.Errorf("WebSocket transport layer status is %s", clientTransport.status)
				glog.Errorf("WebSocket is not ready.")
				continue
			}
			glog.V(3).Infof("[ListenForMessages]: connected, waiting for server response ...")

			msgType, data, err := clientTransport.ws.ReadMessage()
			glog.V(3).Infof("Received websocket message of type %d and size %d", msgType, len(data))

			if clientTransport.closeRequested {
				glog.V(1).Infof("stop listening for message because of requested")
				return
			}

			// Notify errors and break
			if err != nil {
				// destroy its websocket connection whenever there is an error
				if err == io.EOF {
					glog.Errorf("[ListenForMessages] received EOF on websocket %s", err)
				}

				glog.Errorf("[ListenForMessages] error during receive %v", err)
				// close current WebSocket connection.
				clientTransport.closeAndResetWebSocket()
				clientTransport.stopListenForMessages()

				//notify upper module that this connection is closed
				clientTransport.connClosedNotificationCh <- true // Note: this will block till the message is received

				glog.V(1).Infof("[ListenForMessages] websocket error notified, stop lisening for messages.")
				return
			}
			// write the message on the channel
			glog.V(3).Infof("[ListenForMessages] received message on websocket of size %d", len(data))

			clientTransport.queueRawMessage(data) // Note: this will block till the message is read
			glog.V(4).Infof("[ListenForMessages] delivered websocket message, continue listening for server messages...")
		} //end select
	} //end for
}

func (clientTransport *ClientWebSocketTransport) queueRawMessage(data []byte) {
	//TODO: this read should be accumulative - see onMessageReceived in AbstractWebsocketTransport
	clientTransport.inputStreamCh <- data
	// TODO: how to ensure that the channel is open to write
}

func (clientTransport *ClientWebSocketTransport) RawMessageReceiver() chan []byte {
	return clientTransport.inputStreamCh
}

// ==================================================== Message Sender ===============================================
// Send serialized protobuf message bytes
func (clientTransport *ClientWebSocketTransport) Send(messageToSend *TransportMessage) error {
	if clientTransport.closeRequested {
		glog.Errorf("Cannot send message : transport endpoint is closed")
		return errors.New("Cannot send message: transport endpoint is closed")
	}

	if clientTransport.status != Ready {
		glog.Errorf("WebSocket transport layer status is %s", clientTransport.status)
		return errors.New("Cannot send message: web socket is not ready")
	}

	if messageToSend == nil { //.RawMsg == nil {
		glog.Errorf("Cannot send message : marshalled msg is nil")
		return errors.New("Cannot send message: marshalled msg is nil")
	}

	err := clientTransport.write(websocket.BinaryMessage, messageToSend.RawMsg)
	if err != nil {
		glog.Errorf("Error sending message on client transport: %s", err)
		return fmt.Errorf("Error sending message on client transport: %s", err)
	}
	glog.V(4).Infof("Successfully sent message on client transport")
	return nil
}

// ====================================== Websocket Connection =========================================================
// Establish connection to server websocket until connected or until the transport endpoint is closed
func (clientTransport *ClientWebSocketTransport) performWebSocketConnection() error {
	connRetryIntervalSeconds := time.Second * 30 // TODO: use ConnectionRetry parameter from the connConfig or default
	connConfig := clientTransport.connConfig
	// WebSocket URL
	vmtServerUrl := connConfig.TurboServer + connConfig.WebSocketPath
	glog.Infof("[performWebSocketConnection]: %s", vmtServerUrl)

	for !clientTransport.closeRequested { // only set when CloseTransportPoint() is called
		ws, err := openWebSocketConn(connConfig, vmtServerUrl)

		if err != nil {
			// print at debug level after some time
			glog.V(3).Infof("[performWebSocketConnection] %v : unable to connect to %s. Retrying in %v\n", time.Now(), vmtServerUrl, connRetryIntervalSeconds)

			time.Sleep(connRetryIntervalSeconds)
		} else {
			setupPingPong(ws)
			clientTransport.ws = ws
			clientTransport.status = Ready

			glog.V(2).Infof("[performWebSocketConnection]*********** Connected to server " + clientTransport.GetConnectionId())
			glog.V(2).Infof("WebSocket transport layer is ready.")

			return nil
		}
	}
	glog.V(4).Infof("[performWebSocketConnection] exit connect routine, close = %v ", clientTransport.closeRequested)
	return errors.New("Abort client socket connect, transport is closed")
}

// set up websocket Ping-Pong protocol handlers
func setupPingPong(ws *websocket.Conn) {
	h := func(message string) error {
		glog.V(3).Infof("Recevied ping msg")
		err := ws.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(writeWaitTimeout))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}

		if err != nil {
			glog.Errorf("Failed to send PongMessage: %v", err)
		}
		return err
	}

	ws.SetPingHandler(h)

	h2 := func(message string) error {
		glog.V(3).Infof("Received pong msg")
		return nil
	}
	ws.SetPongHandler(h2)
	return
}

func openWebSocketConn(connConfig *WebSocketConnectionConfig, vmtServerUrl string) (*websocket.Conn, error) {
	//1. set up dialer
	d := &websocket.Dialer{
		HandshakeTimeout: handshakeTimeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
	}

	//2. auth header
	header := getAuthHeader(connConfig.WebSocketUsername, connConfig.WebSocketPassword)

	//3. connect it
	c, _, err := d.Dial(vmtServerUrl, header)
	if err != nil {
		glog.Errorf("Failed to connect to server(%s): %v", vmtServerUrl, err)
		return nil, err
	}

	c.SetReadLimit(wsReadLimit)

	return c, nil
}

func getAuthHeader(user, password string) http.Header {
	dat := []byte(fmt.Sprintf("%s:%s", user, password))
	header := http.Header{
		"Authorization": {"Basic " + base64.StdEncoding.EncodeToString(dat)},
	}

	return header
}
