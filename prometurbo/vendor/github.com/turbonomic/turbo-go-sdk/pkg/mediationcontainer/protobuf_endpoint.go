package mediationcontainer

import (
	"fmt"

	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/version"

	"github.com/golang/glog"
	goproto "github.com/golang/protobuf/proto"
)

// Endpoint to handle communication of a particular protobuf message type with the server
type ProtobufEndpoint interface {
	GetName() string
	GetTransport() ITransport
	CloseEndpoint()
	Send(messageToSend *EndpointMessage)
	GetMessageHandler() ProtobufMessage
	MessageReceiver() chan *ParsedMessage
}

type EndpointMessage struct {
	ProtobufMessage goproto.Message
}

// =====================================================================================
// Parser interface for different server messages
type ProtobufMessage interface {
	parse(rawMsg []byte) (*ParsedMessage, error)
	GetMessage() goproto.Message
}

type ParsedMessage struct {
	ServerMsg       proto.MediationServerMessage
	NegotiationMsg  version.NegotiationAnswer
	RegistrationMsg proto.Ack
}

// Parser for all the Mediation Requests such as Discovery, Validation, Action etc
type MediationRequest struct {
	ServerMsg *proto.MediationServerMessage
}

// Parser for the Negotiation Response
type NegotiationResponse struct {
	NegotiationMsg *version.NegotiationAnswer
}

// Parser for the Registration Response
type RegistrationResponse struct {
	RegistrationMsg *proto.Ack
}

func (sr *MediationRequest) GetMessage() goproto.Message {
	return sr.ServerMsg
}

func (sr *MediationRequest) parse(rawMsg []byte) (*ParsedMessage, error) {
	// Parse the input stream
	serverMsg := &proto.MediationServerMessage{}
	err := goproto.Unmarshal(rawMsg, serverMsg)
	if err != nil {
		glog.Error("[MediationRequest] unmarshaling error: ", err)
		return nil, fmt.Errorf("[MediationRequest] Error unmarshalling transport input stream to protobuf message : %s", err)
	}
	sr.ServerMsg = serverMsg
	parsedMsg := &ParsedMessage{
		ServerMsg: *serverMsg,
	}
	return parsedMsg, nil
}

func (nr *NegotiationResponse) GetMessage() goproto.Message {
	return nr.NegotiationMsg
}

func (nr *NegotiationResponse) parse(rawMsg []byte) (*ParsedMessage, error) {
	// Parse the input stream
	serverMsg := &version.NegotiationAnswer{}
	err := goproto.Unmarshal(rawMsg, serverMsg)
	if err != nil {
		glog.Errorf("[NegotiationResponse] unmarshaling error: %s", err)
		return nil, fmt.Errorf("[NegotiationResponse] Error unmarshalling transport input stream to protobuf message : %s", err)
	}
	nr.NegotiationMsg = serverMsg
	parsedMsg := &ParsedMessage{
		NegotiationMsg: *serverMsg,
	}
	return parsedMsg, nil
}

func (rr *RegistrationResponse) GetMessage() goproto.Message {
	return rr.RegistrationMsg
}

func (rr *RegistrationResponse) parse(rawMsg []byte) (*ParsedMessage, error) {
	// Parse the input stream
	serverMsg := &proto.Ack{}
	err := goproto.Unmarshal(rawMsg, serverMsg)
	if err != nil {
		glog.Error("[RegistrationResponse] unmarshaling error: ", err)
		return nil, fmt.Errorf("[RegistrationResponse] Error unmarshalling transport input stream to protobuf message : %s", err)
	}
	rr.RegistrationMsg = serverMsg
	parsedMsg := &ParsedMessage{
		RegistrationMsg: *serverMsg,
	}
	return parsedMsg, nil
}
