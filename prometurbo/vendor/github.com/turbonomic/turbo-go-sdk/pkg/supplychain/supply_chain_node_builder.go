package supplychain

import (
	"fmt"
	"math"

	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type SupplyChainNodeBuilder struct {
	templateClass              *proto.EntityDTO_EntityType
	templateType               *proto.TemplateDTO_TemplateType
	priority                   *int32
	commoditiesSold            []*proto.TemplateCommodity
	providerCommodityBoughtMap map[*proto.Provider][]*proto.TemplateCommodity
	externalLinks              []*proto.TemplateDTO_ExternalEntityLinkProp

	currentProvider *proto.Provider

	err error
}

// Create a new SupplyChainNode Builder.
// All the new supply chain node are default to use the base template type and priority 0.
func NewSupplyChainNodeBuilder(entityType proto.EntityDTO_EntityType) *SupplyChainNodeBuilder {
	templateType := proto.TemplateDTO_BASE
	priority := int32(0)
	return &SupplyChainNodeBuilder{
		templateClass: &entityType,
		templateType:  &templateType,
		priority:      &priority,
	}
}

// Create a SupplyChainNode
func (scnb *SupplyChainNodeBuilder) Create() (*proto.TemplateDTO, error) {
	if scnb.err != nil {
		return nil, fmt.Errorf("Cannot create supply chain node because of error: %v", scnb.err)
	}
	return &proto.TemplateDTO{
		TemplateClass:    scnb.templateClass,
		TemplateType:     scnb.templateType,
		TemplatePriority: scnb.priority,
		CommoditySold:    scnb.commoditiesSold,
		CommodityBought:  buildCommodityBought(scnb.providerCommodityBoughtMap),
		ExternalLink:     scnb.externalLinks,
	}, nil
}

// The very basic selling method. If want others, use other names
func (scnb *SupplyChainNodeBuilder) Sells(templateComm *proto.TemplateCommodity) *SupplyChainNodeBuilder {
	if scnb.err != nil {
		return scnb
	}

	if scnb.commoditiesSold == nil {
		scnb.commoditiesSold = []*proto.TemplateCommodity{}
	}
	scnb.commoditiesSold = append(scnb.commoditiesSold, templateComm)
	return scnb
}

// set the provider of the SupplyChainNode
func (scnb *SupplyChainNodeBuilder) Provider(provider proto.EntityDTO_EntityType,
	pType proto.Provider_ProviderType) *SupplyChainNodeBuilder {
	if scnb.err != nil {
		return scnb
	}

	if pType == proto.Provider_LAYERED_OVER {
		// TODO, need a separate class to build provider.
		maxCardinality := int32(math.MaxInt32)
		minCardinality := int32(0)
		scnb.currentProvider = &proto.Provider{
			TemplateClass:  &provider,
			ProviderType:   &pType,
			CardinalityMax: &maxCardinality,
			CardinalityMin: &minCardinality,
		}
	} else {
		hostCardinality := int32(1)
		scnb.currentProvider = &proto.Provider{
			TemplateClass:  &provider,
			ProviderType:   &pType,
			CardinalityMax: &hostCardinality,
			CardinalityMin: &hostCardinality,
		}
	}

	return scnb
}

// Add a commodity this node buys from the current provider. The provider must already been specified.
// If there is no provider for this node, does not add the commodity.
func (scnb *SupplyChainNodeBuilder) Buys(templateComm *proto.TemplateCommodity) *SupplyChainNodeBuilder {
	if scnb.err != nil {
		return scnb
	}

	if hasProvider := scnb.requireProvider(); !hasProvider {
		scnb.err = fmt.Errorf("Provider must be set before calling Buys().")
		return scnb
	}

	if scnb.providerCommodityBoughtMap == nil {
		scnb.providerCommodityBoughtMap = make(map[*proto.Provider][]*proto.TemplateCommodity)
	}

	templateCommoditiesSoldByCurrentProvider, exist := scnb.providerCommodityBoughtMap[scnb.currentProvider]
	if !exist {
		templateCommoditiesSoldByCurrentProvider = []*proto.TemplateCommodity{}
	}
	templateCommoditiesSoldByCurrentProvider = append(templateCommoditiesSoldByCurrentProvider, templateComm)
	scnb.providerCommodityBoughtMap[scnb.currentProvider] = templateCommoditiesSoldByCurrentProvider

	return scnb
}

// Adds an external entity link to the current node.
// This means the current node will connect an entity discovered by other probe in the full supply chain. The external
// entity can be a provider or a consumer.The connection configurations, for example type, id and connection type are
// specified in externalEntityLink.
func (scnb *SupplyChainNodeBuilder) ConnectsTo(extEntityLink *proto.ExternalEntityLink) *SupplyChainNodeBuilder {
	if scnb.err != nil {
		return scnb
	}

	linkProp, err := scnb.buildExternalEntityLinkProperty(extEntityLink)
	if err != nil {
		scnb.err = err
		return scnb
	}
	currentLinks := scnb.externalLinks
	if currentLinks == nil {
		currentLinks = []*proto.TemplateDTO_ExternalEntityLinkProp{}
	}
	scnb.externalLinks = append(currentLinks, linkProp)
	return scnb
}

func (scnb SupplyChainNodeBuilder) buildExternalEntityLinkProperty(
	extEntityLink *proto.ExternalEntityLink) (*proto.TemplateDTO_ExternalEntityLinkProp, error) {
	entityType := scnb.templateClass
	var key *proto.EntityDTO_EntityType
	if extEntityLink.GetBuyerRef() == *entityType {
		sellerType := extEntityLink.GetSellerRef()
		key = &sellerType
	} else if extEntityLink.GetSellerRef() == *entityType {
		buyerType := extEntityLink.GetBuyerRef()
		key = &buyerType
	} else {
		return nil, fmt.Errorf("Template entity type %v does match types in this external link", entityType)
	}
	return &proto.TemplateDTO_ExternalEntityLinkProp{
		Key:   key,
		Value: extEntityLink,
	}, nil
}

// Check if the provider has been set.
func (scnb *SupplyChainNodeBuilder) requireProvider() bool {
	if scnb.currentProvider == nil {
		return false
	}
	return true
}

func (scnb *SupplyChainNodeBuilder) SetPriority(p int32) {
	scnb.priority = &p
}

func (scnb *SupplyChainNodeBuilder) SetTemplateType(t proto.TemplateDTO_TemplateType) {
	scnb.templateType = &t
}

func buildCommodityBought(
	providerCommodityBoughtMap map[*proto.Provider][]*proto.TemplateCommodity) []*proto.TemplateDTO_CommBoughtProviderProp {
	if len(providerCommodityBoughtMap) == 0 {
		return nil
	}
	commBought := []*proto.TemplateDTO_CommBoughtProviderProp{}
	for provider, templateCommodities := range providerCommodityBoughtMap {
		commBought = append(commBought, &proto.TemplateDTO_CommBoughtProviderProp{
			Key:   provider,
			Value: templateCommodities,
		})
	}
	return commBought
}
