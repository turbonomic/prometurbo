package mediationcontainer

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// DiscoveryResponseSender handles sending messages for DiscoveryResponse to
// a specific channel.
type DiscoveryResponseSender struct {
}

// Send an input DiscoveryResponse to a specific channel. It will block till the channel is ready to receive.
// Before being sent, the DiscoveryResponse will be processed by rearranging it into multiple
// chunks with each chunk contains DTOs of one type.
// For example, if the input DiscoveryDesponse contains DTOs for Entity and Notification,
// two DiscoveryResponse instances (chunks) for Entity and Notification will be sent to the channel.
func (d *DiscoveryResponseSender) Send(discoveryResponse *proto.DiscoveryResponse,
	msgID int32, probeMsgChan chan *proto.MediationClientMessage) {

	// Send chunked discovery response for each DTO type
	for _, chunk := range d.chunkDiscoveryResponse(discoveryResponse) {
		clientMsg := NewClientMessageBuilder(msgID).SetDiscoveryResponse(chunk).Create()

		// Send the response on the callback channel to send to the server
		probeMsgChan <- clientMsg // This will block till the channel is ready to receive
	}
}

// Send discovery response. The message will be chunked so that for each DTO type in DiscoveryResponse, one chunk
// will be generated and sent. This is required to get the DTOs passed at server side.
func (d *DiscoveryResponseSender) chunkDiscoveryResponse(dr *proto.DiscoveryResponse) []*proto.DiscoveryResponse {
	chunks := []*proto.DiscoveryResponse{}

	if len(dr.EntityDTO) > 0 {
		chunk := &proto.DiscoveryResponse{
			EntityDTO: dr.EntityDTO,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.ErrorDTO) > 0 {
		chunk := &proto.DiscoveryResponse{
			ErrorDTO: dr.ErrorDTO,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.DiscoveredGroup) > 0 {
		chunk := &proto.DiscoveryResponse{
			DiscoveredGroup: dr.DiscoveredGroup,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.EntityProfile) > 0 {
		chunk := &proto.DiscoveryResponse{
			EntityProfile: dr.EntityProfile,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.DeploymentProfile) > 0 {
		chunk := &proto.DiscoveryResponse{
			DeploymentProfile: dr.DeploymentProfile,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.Notification) > 0 {
		chunk := &proto.DiscoveryResponse{
			Notification: dr.Notification,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.MetadataDTO) > 0 {
		chunk := &proto.DiscoveryResponse{
			MetadataDTO: dr.MetadataDTO,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.DerivedTarget) > 0 {
		chunk := &proto.DiscoveryResponse{
			DerivedTarget: dr.DerivedTarget,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.NonMarketEntityDTO) > 0 {
		chunk := &proto.DiscoveryResponse{
			NonMarketEntityDTO: dr.NonMarketEntityDTO,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.FlowDTO) > 0 {
		chunk := &proto.DiscoveryResponse{
			FlowDTO: dr.FlowDTO,
		}
		chunks = append(chunks, chunk)
	}

	// Send an empty response to signal completion of discovery
	chunks = append(chunks, &proto.DiscoveryResponse{})

	return chunks
}
