package builder

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	PropertyUsed        = "used"
	PropertyCapacity    = "capacity"
	PropertyResizable   = "resizable"
	PropertyLimit       = "limit"
	PropertyPeak        = "peak"
	PropertyComputeUsed = "computeUsed"
	PropertyReservation = "reservation"
)

var (
	defaultPropertyNames = []string{PropertyUsed, PropertyResizable, PropertyComputeUsed, PropertyCapacity}
)

type ReplacementEntityMetaDataBuilder struct {
	metaData *proto.EntityDTO_ReplacementEntityMetaData
}

func NewReplacementEntityMetaDataBuilder() *ReplacementEntityMetaDataBuilder {
	replacementEntityMetaData := &proto.EntityDTO_ReplacementEntityMetaData{
		IdentifyingProp:  []string{},
		BuyingCommTypes:  []*proto.EntityDTO_ReplacementCommodityPropertyData{},
		SellingCommTypes: []*proto.EntityDTO_ReplacementCommodityPropertyData{},
	}
	return &ReplacementEntityMetaDataBuilder{
		metaData: replacementEntityMetaData,
	}
}

func (builder *ReplacementEntityMetaDataBuilder) Build() *proto.EntityDTO_ReplacementEntityMetaData {
	return builder.metaData
}

// Specifies the name of the property whose value will be used to find the server entity
// for which builder entity is a proxy. The value for the property must be set while building the
// entity.
// Specific properties are pre-defined for some entity types. See the constants defined in
// supply_chain_constants for the names of the specific properties.
func (builder *ReplacementEntityMetaDataBuilder) Matching(property string) *ReplacementEntityMetaDataBuilder {
	builder.metaData.IdentifyingProp = append(builder.metaData.GetIdentifyingProp(), property)
	return builder
}

func (builder *ReplacementEntityMetaDataBuilder) MatchingExternal(propertyDef *proto.ServerEntityPropDef) *ReplacementEntityMetaDataBuilder {
	builder.metaData.ExtEntityPropDef = append(builder.metaData.GetExtEntityPropDef(), propertyDef)
	return builder
}

// Set the commodity type whose metric values will be transferred to the entity
// builder DTO will be replaced by.
func (builder *ReplacementEntityMetaDataBuilder) PatchBuying(commType proto.CommodityDTO_CommodityType) *ReplacementEntityMetaDataBuilder {
	return builder.PatchBuyingWithProperty(commType, defaultPropertyNames)
}

func (builder *ReplacementEntityMetaDataBuilder) PatchBuyingWithProperty(commType proto.CommodityDTO_CommodityType, names []string) *ReplacementEntityMetaDataBuilder {
	builder.metaData.BuyingCommTypes = append(builder.metaData.GetBuyingCommTypes(),
		&proto.EntityDTO_ReplacementCommodityPropertyData{
			CommodityType: &commType,
			PropertyName:  names,
		})
	return builder
}

// Set the commodity type whose metric values will be transferred to the entity
//  builder DTO will be replaced by.
func (builder *ReplacementEntityMetaDataBuilder) PatchSelling(commType proto.CommodityDTO_CommodityType) *ReplacementEntityMetaDataBuilder {
	return builder.PatchSellingWithProperty(commType, defaultPropertyNames)
}

func (builder *ReplacementEntityMetaDataBuilder) PatchSellingWithProperty(commType proto.CommodityDTO_CommodityType, names []string) *ReplacementEntityMetaDataBuilder {
	builder.metaData.SellingCommTypes = append(builder.metaData.GetSellingCommTypes(),
		&proto.EntityDTO_ReplacementCommodityPropertyData{
			CommodityType: &commType,
			PropertyName:  names,
		})

	return builder
}
