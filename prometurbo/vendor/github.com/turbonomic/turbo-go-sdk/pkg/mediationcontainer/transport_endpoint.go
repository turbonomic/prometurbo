package mediationcontainer

// Transport endpoint that sends and receives raw message bytes
type ITransport interface {
	// Open
	Connect() error
	GetConnectionId() string
	// Send
	Send(messageToSend *TransportMessage) error
	// Receive
	ListenForMessages()
	RawMessageReceiver() chan []byte // Queue or channel for putting byte[] received on the transport
	// Close
	CloseTransportPoint()
	NotifyClosed() chan bool // Channel where connection closed notification is sent
}

type TransportMessage struct {
	RawMsg []byte
}
