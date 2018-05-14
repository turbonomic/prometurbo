package mediationcontainer

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/version"

	"fmt"
	"github.com/golang/glog"
	"time"
)

const (
	waitResponseTimeOut = time.Second * 30
)

type SdkClientProtocol struct {
	allProbes map[string]*ProbeProperties
	version   string
	//TransportReady chan bool
}

func CreateSdkClientProtocolHandler(allProbes map[string]*ProbeProperties, version string) *SdkClientProtocol {
	return &SdkClientProtocol{
		allProbes: allProbes,
		version:   version,
		//TransportReady: done,
	}
}

func (clientProtocol *SdkClientProtocol) handleClientProtocol(transport ITransport, transportReady chan bool) {
	glog.V(2).Infof("Starting Protocol Negotiation ....")
	status := clientProtocol.NegotiateVersion(transport)

	if !status {
		glog.Errorf("Failure during Protocol Negotiation, Registration message will not be sent")
		transportReady <- false
		// clientProtocol.TransportReady <- false
		return
	}
	glog.V(2).Infof("[SdkClientProtocol] Starting Probe Registration ....")
	status = clientProtocol.HandleRegistration(transport)
	if !status {
		glog.Errorf("Failure during Registration, cannot receive server messages")
		transportReady <- false
		// clientProtocol.TransportReady <- false
		return
	}

	transportReady <- true
	// clientProtocol.TransportReady <- true
}

// ============================== Protocol Version Negotiation =========================
func timeOutRead(name string, du time.Duration, ch chan *ParsedMessage) (*ParsedMessage, error) {
	timer := time.NewTimer(du)
	select {
	case msg, ok := <-ch:
		if !ok {
			err := fmt.Errorf("[%s]: Endpoint Receiver channel is closed.", name)
			glog.Error(err.Error())
			return nil, err
		}
		if msg == nil {
			err := fmt.Errorf("[%s]: Endpoint receive null message.", name)
			glog.Error(err.Error())
			return nil, err
		}
		return msg, nil
	case <-timer.C:
		err := fmt.Errorf("[%s]: wait for message from channel timeout(%v seconds).", name, du.Seconds())
		glog.Error(err.Error())
		return nil, err
	}
}

func (clientProtocol *SdkClientProtocol) NegotiateVersion(transport ITransport) bool {
	versionStr := clientProtocol.version
	request := &version.NegotiationRequest{
		ProtocolVersion: &versionStr,
	}
	glog.V(3).Infof("Send negotiation message: %+v", request)

	// Create Protobuf Endpoint to send and handle negotiation messages
	protoMsg := &NegotiationResponse{} // handler for the response
	endpoint := CreateClientProtoBufEndpoint("NegotiationEndpoint", transport, protoMsg, true)
	defer endpoint.CloseEndpoint()

	endMsg := &EndpointMessage{
		ProtobufMessage: request,
	}
	endpoint.Send(endMsg)

	// Wait for the response to be received by the transport and then parsed and put on the endpoint's message channel
	serverMsg, err := timeOutRead(endpoint.GetName(), waitResponseTimeOut, endpoint.MessageReceiver())
	if err != nil {
		glog.Errorf("[%s] : read VersionNegotiation response from channel failed: %v", endpoint.GetName(), err)
		return false
	}
	glog.V(3).Infof("[%s] : Received: %++v\n", endpoint.GetName(), serverMsg)

	// Handler response
	negotiationResponse := protoMsg.NegotiationMsg
	if negotiationResponse == nil {
		glog.Error("Probe Protocol failed, null negotiation response")
		return false
	}
	negotiationResponse.GetNegotiationResult()

	if negotiationResponse.GetNegotiationResult().String() != version.NegotiationAnswer_ACCEPTED.String() {
		glog.Errorf("Protocol version negotiation failed %s",
			negotiationResponse.GetNegotiationResult().String()+") :"+negotiationResponse.GetDescription())
		return false
	}
	glog.V(4).Infof("[SdkClientProtocol] Protocol version is accepted by server: %s", negotiationResponse.GetDescription())
	return true
}

// ======================= Registration ============================
// Send registration message
func (clientProtocol *SdkClientProtocol) HandleRegistration(transport ITransport) bool {
	containerInfo, err := clientProtocol.MakeContainerInfo()
	if err != nil {
		glog.Error("Error creating ContainerInfo")
		return false
	}

	glog.V(3).Infof("Send registration message: %+v", containerInfo)

	// Create Protobuf Endpoint to send and handle registration messages
	protoMsg := &RegistrationResponse{}
	endpoint := CreateClientProtoBufEndpoint("RegistrationEndpoint", transport, protoMsg, true)
	defer endpoint.CloseEndpoint()

	endMsg := &EndpointMessage{
		ProtobufMessage: containerInfo,
	}
	endpoint.Send(endMsg)

	// Wait for the response to be received by the transport and then parsed and put on the endpoint's message channel
	serverMsg, err := timeOutRead(endpoint.GetName(), waitResponseTimeOut, endpoint.MessageReceiver())
	if err != nil {
		glog.Errorf("[%s] : read Registration response from channel failed: %v", endpoint.GetName(), err)
		return false
	}
	glog.V(3).Infof("[%s] : Received: %++v\n", endpoint.GetName(), serverMsg)

	// Handler response
	registrationResponse := protoMsg.RegistrationMsg
	if registrationResponse == nil {
		glog.Errorf("Probe registration failed, null ack")
		return false
	}

	return true
}

func (clientProtocol *SdkClientProtocol) MakeContainerInfo() (*proto.ContainerInfo, error) {
	var probes []*proto.ProbeInfo

	for k, v := range clientProtocol.allProbes {
		glog.V(2).Infof("SdkClientProtocol] Creating Probe Info for", k)
		turboProbe := v.Probe
		var probeInfo *proto.ProbeInfo
		var err error
		probeInfo, err = turboProbe.GetProbeInfo()

		if err != nil {
			return nil, err
		}
		probes = append(probes, probeInfo)
	}

	return &proto.ContainerInfo{
		Probes: probes,
	}, nil
}
