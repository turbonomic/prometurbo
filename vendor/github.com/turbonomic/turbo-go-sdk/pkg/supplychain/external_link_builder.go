package supplychain

import (
	"fmt"

	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type ExternalEntityLinkBuilder struct {
	buyerRef                   *proto.EntityDTO_EntityType
	sellerRef                  *proto.EntityDTO_EntityType
	relationship               *proto.Provider_ProviderType
	commodityDefs              []*proto.ExternalEntityLink_CommodityDef
	key                        *string
	hasExternalEntity          *bool
	probeEntityPropertyDef     []*proto.ExternalEntityLink_EntityPropertyDef
	externalEntityPropertyDefs []*proto.ServerEntityPropDef

	err error
}

func NewExternalEntityLinkBuilder() *ExternalEntityLinkBuilder {
	return &ExternalEntityLinkBuilder{}
}

// Get the ExternalEntityLink that you have built.
func (builder *ExternalEntityLinkBuilder) Build() (*proto.ExternalEntityLink, error) {
	if builder.err != nil {
		return nil, builder.err
	}
	return &proto.ExternalEntityLink{
		BuyerRef:                   builder.buyerRef,
		SellerRef:                  builder.sellerRef,
		Relationship:               builder.relationship,
		CommodityDefs:              builder.commodityDefs,
		Key:                        builder.key,
		HasExternalEntity:          builder.hasExternalEntity,
		ProbeEntityPropertyDef:     builder.probeEntityPropertyDef,
		ExternalEntityPropertyDefs: builder.externalEntityPropertyDefs,
	}, nil
}

// Initialize the buyer/seller external link that you're building.
// This method sets the entity types for the buyer and seller, as well as the type of provider
// relationship HOSTING or code LAYERED_OVER.
func (builder *ExternalEntityLinkBuilder) Link(buyer, seller proto.EntityDTO_EntityType,
	relationship proto.Provider_ProviderType) *ExternalEntityLinkBuilder {
	if builder.err != nil {
		return builder
	}

	builder.buyerRef = &buyer
	builder.sellerRef = &seller
	builder.relationship = &relationship

	return builder
}

// Add a single bought commodity to the link.
func (builder *ExternalEntityLinkBuilder) Commodity(comm proto.CommodityDTO_CommodityType, hasKey bool) *ExternalEntityLinkBuilder {
	if builder.err != nil {
		return builder
	}
	commodityDef := &proto.ExternalEntityLink_CommodityDef{
		Type:   &comm,
		HasKey: &hasKey,
	}
	builder.commodityDefs = append(builder.commodityDefs, commodityDef)
	return builder
}

// Set a property of the discovered entity to the link. Operations Manager will use builder property to
// stitch the discovered entity into the Operations Manager topology. This setting includes the property name
// and an arbitrary description.
func (builder *ExternalEntityLinkBuilder) ProbeEntityPropertyDef(name, description string) *ExternalEntityLinkBuilder {
	if builder.err != nil {
		return builder
	}
	entityProperty := &proto.ExternalEntityLink_EntityPropertyDef{
		Name:        &name,
		Description: &description,
	}
	builder.probeEntityPropertyDef = append(builder.probeEntityPropertyDef, entityProperty)

	return builder
}

// Set an ServerEntityPropertyDef to the link you're building.
// The ServerEntityPropertyDef includes metadata for the properties of the  external entity. Operations Manager can
// use the metadata to stitch entities discovered by the probe together with external entities.
// An external entity is one that exists in the Operations Manager topology, but has not been discovered by the probe.
func (builder *ExternalEntityLinkBuilder) ExternalEntityPropertyDef(propertyDef *proto.ServerEntityPropDef) *ExternalEntityLinkBuilder {
	if propertyDef == nil {
		builder.err = fmt.Errorf("Nil service entity property definition.")
		return builder
	}
	builder.externalEntityPropertyDefs = append(builder.externalEntityPropertyDefs, propertyDef)

	return builder
}
