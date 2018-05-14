package supplychain

import (
	"fmt"

	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type SupplyChainBuilder struct {
	supplyChainNodes []*proto.TemplateDTO

	err error
}

func NewSupplyChainBuilder() *SupplyChainBuilder {
	return &SupplyChainBuilder{}
}

func (scb *SupplyChainBuilder) Create() ([]*proto.TemplateDTO, error) {
	if scb.err != nil {
		return nil, scb.err
	}

	return scb.supplyChainNodes, nil
}

// To build a supply chain, must specify the top supply chain node first.
// Here it initializes supplyChainNodes.
func (scb *SupplyChainBuilder) Top(topNode *proto.TemplateDTO) *SupplyChainBuilder {
	if topNode == nil {
		scb.err = fmt.Errorf("topNode cannot be nil.")
		return scb
	}
	scb.supplyChainNodes = []*proto.TemplateDTO{}
	scb.supplyChainNodes = append(scb.supplyChainNodes, topNode)

	return scb
}

// Add an entity node to supply chain
func (scb *SupplyChainBuilder) Entity(node *proto.TemplateDTO) *SupplyChainBuilder {
	if scb.err != nil {
		return scb
	}
	if scb.supplyChainNodes == nil || len(scb.supplyChainNodes) == 0 {
		scb.err = fmt.Errorf("Must set top supply chain node first.")
		return scb
	}

	scb.supplyChainNodes = append(scb.supplyChainNodes, node)
	return scb
}
