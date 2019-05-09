package builder

import (
	"errors"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type Protocol int32

const (
	TCP Protocol = 1
	UDP Protocol = 2
)

type FlowDTOBuilder struct {
	sourceAddress      string
	sourcePort         int32
	destinationAddress string
	destinationPort    int32
	protocol           Protocol
	flowAmount         float64
	latency            int64
	rx                 int64
	tx                 int64
	err                error
}

func NewFlowDTOBuilder() *FlowDTOBuilder {
	return &FlowDTOBuilder{
		sourcePort:      0,
		destinationPort: 0,
		protocol:        TCP,
		flowAmount:      0,
		latency:         0,
		rx:              0,
		tx:              0,
	}
}

// Sets the source
func (builder *FlowDTOBuilder) Source(source string) *FlowDTOBuilder {
	builder.sourceAddress = source
	return builder
}

// Sets the destination
func (builder *FlowDTOBuilder) Destination(destination string, port int32) *FlowDTOBuilder {
	builder.destinationAddress = destination
	builder.destinationPort = port
	return builder
}

// Set the protocol
func (builder *FlowDTOBuilder) Protocol(protocol Protocol) *FlowDTOBuilder {
	if protocol != TCP && protocol != UDP {
		builder.err = errors.New("unsupported protocol")
		return builder
	}
	builder.protocol = protocol
	return builder
}

// Sets the flow amount
func (builder *FlowDTOBuilder) FlowAmount(flowAmount float64) *FlowDTOBuilder {
	builder.flowAmount = flowAmount
	return builder
}

// Sets the latency
func (builder *FlowDTOBuilder) Latency(latency int64) *FlowDTOBuilder {
	builder.latency = latency
	return builder
}

// Sets the received amount
func (builder *FlowDTOBuilder) Received(received int64) *FlowDTOBuilder {
	builder.rx = received
	return builder
}

// Sets the transmitted amount
func (builder *FlowDTOBuilder) Transmitted(transmitted int64) *FlowDTOBuilder {
	builder.tx = transmitted
	return builder
}

// Creates the DTO
func (builder *FlowDTOBuilder) Create() (*proto.FlowDTO, error) {
	if builder.err != nil {
		return nil, builder.err
	}

	src := &proto.EntityIdentityData{
		IpAddress: &builder.sourceAddress,
		Port:      &builder.sourcePort,
	}
	dst := &proto.EntityIdentityData{
		IpAddress: &builder.destinationAddress,
		Port:      &builder.destinationPort,
	}
	protocol := proto.FlowDTO_Protocol(builder.protocol)
	return &proto.FlowDTO{
		SourceEntityIdentityData: src,
		DestEntityIdentityData:   dst,
		Protocol:                 &protocol,
		FlowAmount:               &builder.flowAmount,
		Latency:                  &builder.latency,
		ReceivedAmount:           &builder.rx,
		TransmittedAmount:        &builder.tx,
	}, nil
}
