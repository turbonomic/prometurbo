package mediationcontainer

import (
	"time"

	"github.com/golang/glog"
	goproto "github.com/golang/protobuf/proto"
)

// =====================================================================================================
// Implementation of ProtobufEndpoint to handle all the server protobuf messages sent to the client
type ClientProtobufEndpoint struct {
	Name              string
	singleMessageMode bool
	// Transport used to send and receive messages
	transport ITransport
	// Parser for the message - this will vary with the type of message communication the endpoint is being used for
	messageHandler ProtobufMessage
	// Channel where the endpoint will send the parsed messages
	ParsedMessageChannel chan *ParsedMessage // unbuffered channel
	// TODO: add message waiting policy
	stopMsgWaitCh chan bool // buffered channel
}

// Create a new instance of the ClientProtobufEndpoint that handles communication
// for a specific message type using the given transport point
func CreateClientProtoBufEndpoint(name string, transport ITransport, messageHandler ProtobufMessage, singleMessageMode bool) ProtobufEndpoint {
	endpoint := &ClientProtobufEndpoint{
		Name:                 name,
		transport:            transport, // the transport
		ParsedMessageChannel: make(chan *ParsedMessage),
		messageHandler:       messageHandler, // the message parser
		singleMessageMode:    singleMessageMode,
	}

	glog.V(3).Infof("Created Protobuf Endpoint " + endpoint.GetName())
	// Start a Message Handling routine to wait for messages arriving on the transport point
	if singleMessageMode {
		endpoint.ParsedMessageChannel = make(chan *ParsedMessage, 1)
		endpoint.waitForSingleServerMessage() // TODO: redo using MessageWaiting policy
	} else {
		endpoint.ParsedMessageChannel = make(chan *ParsedMessage)
		endpoint.stopMsgWaitCh = make(chan bool, 1)
		endpoint.waitForServerMessage()
	}

	return endpoint
}

func (endpoint *ClientProtobufEndpoint) GetName() string {
	return endpoint.Name
}

func (endpoint *ClientProtobufEndpoint) GetTransport() ITransport {
	return endpoint.transport
}

func (endpoint *ClientProtobufEndpoint) MessageReceiver() chan *ParsedMessage {
	return endpoint.ParsedMessageChannel
}

func (endpoint *ClientProtobufEndpoint) GetMessageHandler() ProtobufMessage {
	return endpoint.messageHandler
}

func (endpoint *ClientProtobufEndpoint) CloseEndpoint() {
	glog.V(4).Infof("[" + endpoint.Name + "] : closing endpoint and listener routine")
	// Send close to the listener routine
	if endpoint.stopMsgWaitCh != nil {
		glog.V(4).Infof("["+endpoint.Name+"] closing stopMsgWaitCh %+v", endpoint.stopMsgWaitCh)
		endpoint.stopMsgWaitCh <- true
		close(endpoint.stopMsgWaitCh)
		glog.V(4).Infof("["+endpoint.Name+"] closed stopMsgWaitCh %+v", endpoint.stopMsgWaitCh)
	}
}

func (endpoint *ClientProtobufEndpoint) Send(messageToSend *EndpointMessage) {
	glog.V(4).Infof("[%s] : Sending protobuf message", endpoint.Name) // %s", messageToSend.ProtobufMessage)
	// Marshal protobuf message to raw bytes
	msgMarshalled, err := goproto.Marshal(messageToSend.ProtobufMessage) // marshal to byte array
	if err != nil {
		glog.Errorf("[ClientProtobufEndpoint] during send - marshaling error: %s", err)
		return
	}
	// Send using the underlying transport
	tmsg := &TransportMessage{
		RawMsg: msgMarshalled,
	}
	err = endpoint.transport.Send(tmsg)
	if err != nil {
		glog.Errorf("[ClientProtobufEndpoint] during send - transport error: %s", err)
		return
	}
}

func (endpoint *ClientProtobufEndpoint) waitForServerMessage() {

	logPrefix := "[" + endpoint.Name + "][[waitForServerMessage] : "
	glog.V(4).Infof(logPrefix+" %s: ENTER  ", time.Now())

	go func() {
		// main loop for listening server message until its message receiver channel is closed.
		for {
			glog.V(4).Infof("["+endpoint.Name+"][waitForServerMessage] : waiting for server request at endpoint %v", endpoint)
			select {
			case <-endpoint.stopMsgWaitCh:
				glog.V(4).Infof(logPrefix+" closing MessageChannel %+v", endpoint.ParsedMessageChannel)
				close(endpoint.ParsedMessageChannel) // This listener routine is the writer for this channel
				glog.V(4).Infof(logPrefix+" closed MessageChannel %+v", endpoint.ParsedMessageChannel)
				return
			//default:
			case rawBytes, ok := <-endpoint.transport.RawMessageReceiver(): // block till  the message bytes from the transport channel,
				if !ok {
					glog.Errorf(logPrefix + "transport message channel is closed")
					return
				}
				// Parse the input stream using the registered message handler
				messageHandler := endpoint.GetMessageHandler()
				parsedMsg, err := messageHandler.parse(rawBytes)

				if err != nil {
					glog.Errorf(logPrefix + "received null message, dropping it")
					continue
				}

				glog.V(3).Infof(logPrefix+"received message is: %++v\n", parsedMsg.ServerMsg.GetMediationServerMessage())

				// Put the parsed message on the endpoint's channel
				// - this will block till the upper layer receives this message
				msgChannel := endpoint.MessageReceiver()
				if msgChannel != nil { // checking if the channel was closed before putting the message
					msgChannel <- parsedMsg
				}

				glog.V(3).Infof(logPrefix + "parsed message delivered on the message channel, continue to listen from transport ...")
			} //end select
		} //end for
	}()
	glog.V(4).Infof(logPrefix + "DONE")
}

func (endpoint *ClientProtobufEndpoint) waitForSingleServerMessage() {
	logPrefix := "[" + endpoint.Name + "][[waitForSingleServerMessage] : "
	glog.V(4).Infof(logPrefix + "waiting for server response")

	go func() {

		// listen for server message
		// - this will block till the message appears on the channel
		rawBytes := <-endpoint.transport.RawMessageReceiver()

		messageHandler := endpoint.GetMessageHandler()
		parsedMsg, err := messageHandler.parse(rawBytes)

		if err != nil {
			glog.Errorf("[" + endpoint.Name + "][waitForSingleServerMessage] : Received null message, dropping it")
			parsedMsg = &ParsedMessage{} //create empty message
		}

		glog.V(4).Infof("["+endpoint.Name+"][waitForSingleServerMessage] : Received: %s\n", parsedMsg)

		// - this will block till the upper layer receives this message
		msgChannel := endpoint.MessageReceiver()
		if msgChannel != nil { // checking if the channel was closed before putting the message
			msgChannel <- parsedMsg
		}

		glog.V(4).Infof(logPrefix + "parsed message delivered on the message channel")
		glog.V(4).Infof(logPrefix + "DONE")
	}()
}

// =====================================================================================
// ---------------------------------------- Not used -----------------------------------
type MessageWaiter interface {
	getMessage(endpoint ProtobufEndpoint) goproto.Message
}

type SingleMessageWaiter struct {
}

func (messageWaiter *SingleMessageWaiter) getMessage(endpoint ProtobufEndpoint) {
	go func() {
		getSingleMessage(endpoint)
	}()
}

type ContinuousMessageWaiter struct {
}

func (messageWaiter *ContinuousMessageWaiter) getMessage(endpoint ProtobufEndpoint) {
	go func() {
		for {
			getSingleMessage(endpoint)
		}
	}()
}

func getSingleMessage(endpoint ProtobufEndpoint) {
	glog.V(4).Infof("[" + endpoint.GetName() + "][waitForSingleServerMessage]: ########## Waiting for server request #######")
	// listen for server message
	// - this will block till the message appears on the channel
	transport := endpoint.GetTransport()
	rawBytes := <-transport.RawMessageReceiver()
	//fmt.Printf("[" + endpoint.Name + "][waitForSingleServerMessage] : Received: message from transport channel %s\n", rawBytes)

	// Parse the input stream using the registered message handler
	messageHandler := endpoint.GetMessageHandler()
	parsedMsg, err := messageHandler.parse(rawBytes)
	if err != nil {
		glog.Errorf("[" + endpoint.GetName() + "][SingleMessageWaiter] : Received null message, dropping it")
		parsedMsg = &ParsedMessage{} //create empty message
	}

	glog.V(4).Infof("["+endpoint.GetName()+"][waitForSingleServerMessage] : Received: %s\n", parsedMsg)

	// - this will block till the upper layer receives this message
	msgChannel := endpoint.MessageReceiver()
	if msgChannel != nil { // TODO: checking if the channel was closed before putting the message
		msgChannel <- parsedMsg
	}
}
