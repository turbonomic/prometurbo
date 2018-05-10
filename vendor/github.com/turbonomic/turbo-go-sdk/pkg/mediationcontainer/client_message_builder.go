package mediationcontainer

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// ========== Builder for proto messages created from the probe =============
// A ClientMessageBuilder builds a ClientMessage instance.
type ClientMessageBuilder struct {
	clientMessage *proto.MediationClientMessage
}

// Get an instance of ClientMessageBuilder
func NewClientMessageBuilder(messageID int32) *ClientMessageBuilder {
	clientMessage := &proto.MediationClientMessage{
		MessageID: &messageID,
	}
	return &ClientMessageBuilder{
		clientMessage: clientMessage,
	}
}

// Build an instance of ClientMessage.
func (cmb *ClientMessageBuilder) Create() *proto.MediationClientMessage {
	return cmb.clientMessage
}

// set the validation response
func (cmb *ClientMessageBuilder) SetValidationResponse(validationResponse *proto.ValidationResponse) *ClientMessageBuilder {

	response := &proto.MediationClientMessage_ValidationResponse{
		ValidationResponse: validationResponse,
	}

	cmb.clientMessage.MediationClientMessage = response

	return cmb
}

// set discovery response
func (cmb *ClientMessageBuilder) SetDiscoveryResponse(discoveryResponse *proto.DiscoveryResponse) *ClientMessageBuilder {
	response := &proto.MediationClientMessage_DiscoveryResponse{
		DiscoveryResponse: discoveryResponse,
	}
	cmb.clientMessage.MediationClientMessage = response
	return cmb
}

// set discovery keep alive
func (cmb *ClientMessageBuilder) SetKeepAlive(keepAlive *proto.KeepAlive) *ClientMessageBuilder {
	response := &proto.MediationClientMessage_KeepAlive{
		KeepAlive: keepAlive,
	}
	cmb.clientMessage.MediationClientMessage = response

	return cmb
}

// set action progress
func (cmb *ClientMessageBuilder) SetActionProgress(actionProgress *proto.ActionProgress) *ClientMessageBuilder {
	response := &proto.MediationClientMessage_ActionProgress{
		ActionProgress: actionProgress,
	}
	cmb.clientMessage.MediationClientMessage = response

	return cmb
}

// set action response
func (cmb *ClientMessageBuilder) SetActionResponse(actionResponse *proto.ActionResult) *ClientMessageBuilder {
	response := &proto.MediationClientMessage_ActionResponse{
		ActionResponse: actionResponse,
	}
	cmb.clientMessage.MediationClientMessage = response

	return cmb
}
