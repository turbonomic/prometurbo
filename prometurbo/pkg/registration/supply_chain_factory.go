package registration

import (
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/constant"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/supplychain"
)

var (
	respTimeType    = proto.CommodityDTO_RESPONSE_TIME
	transactionType = proto.CommodityDTO_TRANSACTION
	key             = "key-placeholder"

	respTimeTemplateComm *proto.TemplateCommodity = &proto.TemplateCommodity{
		CommodityType: &respTimeType,
		Key:           &key,
	}

	transactionTemplateComm *proto.TemplateCommodity = &proto.TemplateCommodity{
		CommodityType: &transactionType,
		Key:           &key,
	}
)

const (
	StitchingAttr string = "IP"
	DEFAULT_DELIMITER string = ","
)

type SupplyChainFactory struct{}

func (f *SupplyChainFactory) CreateSupplyChain() ([]*proto.TemplateDTO, error) {
	appNode, err := f.buildAppSupplyBuilder()

	if err != nil {
		return nil, err
	}

	f.setAppStitchingMetaData(appNode)

	vAppNode, err := f.buildVAppSupplyBuilder()

	if err != nil {
		return nil, err
	}

	f.setVAppStitchingMetaData(vAppNode)

	return supplychain.NewSupplyChainBuilder().Top(vAppNode).Entity(appNode).
		Create()
}

func (f *SupplyChainFactory) buildAppSupplyBuilder() (*proto.TemplateDTO, error) {
	builder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_APPLICATION).
		Sells(transactionTemplateComm).
		Sells(respTimeTemplateComm)
	builder.SetPriority(-1)
	builder.SetTemplateType(proto.TemplateDTO_BASE)
	//builder.SetTemplateType(proto.TemplateDTO_EXTENSION)

	return builder.Create()
}

func (f *SupplyChainFactory) buildVAppSupplyBuilder() (*proto.TemplateDTO, error) {
	builder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_VIRTUAL_APPLICATION).
		Provider(proto.EntityDTO_APPLICATION, proto.Provider_LAYERED_OVER).
		Buys(transactionTemplateComm).
		Buys(respTimeTemplateComm)
	builder.SetPriority(-1)
	builder.SetTemplateType(proto.TemplateDTO_BASE)
	//builder.SetTemplateType(proto.TemplateDTO_EXTENSION)

	return builder.Create()
}

func (f *SupplyChainFactory) setAppStitchingMetaData(appNode *proto.TemplateDTO) {
	commodityList := []proto.CommodityDTO_CommodityType{respTimeType, transactionType}

	var mbuilder *builder.MergedEntityMetadataBuilder

	mbuilder = builder.NewMergedEntityMetadataBuilder().
		InternalMatchingProperty(constant.StitchingAttr).
		InternalMatchingType(builder.MergedEntityMetadata_STRING).
		ExternalMatchingProperty(constant.StitchingAttr).
		ExternalMatchingType(builder.MergedEntityMetadata_STRING).
		PatchSoldList(commodityList)

	metadata, _ := mbuilder.Build()
	appNode.MergedEntityMetaData = metadata
	return
}

func (f *SupplyChainFactory) setVAppStitchingMetaData(vappNode *proto.TemplateDTO) {
	commodityList := []proto.CommodityDTO_CommodityType{respTimeType, transactionType}

	mbuilder := builder.NewMergedEntityMetadataBuilder().
		InternalMatchingProperty(constant.StitchingAttr).
		InternalMatchingType(builder.MergedEntityMetadata_STRING).
		ExternalMatchingProperty(constant.StitchingAttr).
		ExternalMatchingType(builder.MergedEntityMetadata_LIST_STRING).
		PatchBoughtList(proto.EntityDTO_APPLICATION, commodityList)

	metadata, _ := mbuilder.Build()
	vappNode.MergedEntityMetaData = metadata
	return
}