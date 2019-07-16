package registration

import (
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/constant"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/supplychain"
)

var (
	VMIPFieldPaths = []string{constant.SUPPLY_CHAIN_CONSTANT_VIRTUAL_MACHINE_DATA}

	vCpuType    = proto.CommodityDTO_VCPU
	vMemType = proto.CommodityDTO_VMEM

	vCpuTemplateComm *proto.TemplateCommodity = &proto.TemplateCommodity{
		CommodityType: &vCpuType,
	}

	vMemTemplateComm *proto.TemplateCommodity = &proto.TemplateCommodity{
		CommodityType: &vMemType,
	}

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

type SupplyChainFactory struct{}

func (f *SupplyChainFactory) CreateSupplyChain() ([]*proto.TemplateDTO, error) {
	// VM node
	vmNode, err := f.buildVMSupplyBuilder()

	if err != nil {
		return nil, err
	}

	// Stitching metadata for the vm node
	vmMetadata, err := f.getVMStitchingMetaData()
	if err != nil {
		return nil, err
	}

	vmNode.MergedEntityMetaData = vmMetadata

	// Application node
	appNode, err := f.buildAppSupplyBuilder()

	if err != nil {
		return nil, err
	}

	// Stitching metadata for the application node
	appMetadata, err := f.getAppStitchingMetaData()
	if err != nil {
		return nil, err
	}

	appNode.MergedEntityMetaData = appMetadata

	// vApplication node
	vAppNode, err := f.buildVAppSupplyBuilder()

	if err != nil {
		return nil, err
	}

	// Stitching metadata for the vApp node
	vAppMetadata, err := f.getVAppStitchingMetaData()
	if err != nil {
		return nil, err
	}

	vAppNode.MergedEntityMetaData = vAppMetadata

	return supplychain.NewSupplyChainBuilder().
		Top(vAppNode).
		Entity(appNode).
		Entity(vmNode).
		Create()
}

func (f *SupplyChainFactory) buildVMSupplyBuilder() (*proto.TemplateDTO, error) {
	builder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_VIRTUAL_MACHINE).
		Sells(vCpuTemplateComm).
		Sells(vMemTemplateComm)
	builder.SetPriority(-1)
	builder.SetTemplateType(proto.TemplateDTO_BASE)

	return builder.Create()
}

func (f *SupplyChainFactory) buildAppSupplyBuilder() (*proto.TemplateDTO, error) {
	builder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_APPLICATION).
		Sells(transactionTemplateComm).
		Sells(respTimeTemplateComm).
		Buys(vCpuTemplateComm).
		Buys(vMemTemplateComm)
	builder.SetPriority(-1)
	builder.SetTemplateType(proto.TemplateDTO_BASE)

	return builder.Create()
}

func (f *SupplyChainFactory) buildVAppSupplyBuilder() (*proto.TemplateDTO, error) {
	builder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_VIRTUAL_APPLICATION).
		Provider(proto.EntityDTO_APPLICATION, proto.Provider_LAYERED_OVER).
		Buys(transactionTemplateComm).
		Buys(respTimeTemplateComm)
	builder.SetPriority(-1)
	builder.SetTemplateType(proto.TemplateDTO_BASE)

	return builder.Create()
}

func (f *SupplyChainFactory) getVMStitchingMetaData() (*proto.MergedEntityMetadata, error) {

	var vmbuilder *builder.MergedEntityMetadataBuilder

	vmbuilder = builder.NewMergedEntityMetadataBuilder().
		InternalMatchingType(builder.MergedEntityMetadata_LIST_STRING).
		InternalMatchingPropertyWithDelimiter(constant.StitchingAttr, ",").
		ExternalMatchingType(builder.MergedEntityMetadata_LIST_STRING).
		ExternalMatchingFieldWithDelimiter(constant.SUPPLY_CHAIN_CONSTANT_IP_ADDRESS, VMIPFieldPaths, ",")

	metadata, err := vmbuilder.Build()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (f *SupplyChainFactory) getAppStitchingMetaData() (*proto.MergedEntityMetadata, error) {
	commodityList := []proto.CommodityDTO_CommodityType{respTimeType, transactionType}

	var mbuilder *builder.MergedEntityMetadataBuilder

	mbuilder = builder.NewMergedEntityMetadataBuilder().
		InternalMatchingProperty(constant.StitchingAttr).
		InternalMatchingType(builder.MergedEntityMetadata_STRING).
		ExternalMatchingProperty(constant.StitchingAttr).
		ExternalMatchingType(builder.MergedEntityMetadata_STRING).
		PatchSoldList(commodityList)

	metadata, err := mbuilder.Build()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (f *SupplyChainFactory) getVAppStitchingMetaData() (*proto.MergedEntityMetadata, error) {
	commodityList := []proto.CommodityDTO_CommodityType{respTimeType, transactionType}

	var mbuilder *builder.MergedEntityMetadataBuilder

	mbuilder = builder.NewMergedEntityMetadataBuilder().
		InternalMatchingProperty(constant.StitchingAttr).
		InternalMatchingType(builder.MergedEntityMetadata_STRING).
		ExternalMatchingPropertyWithDelimiter(constant.StitchingAttr, ",").
		ExternalMatchingType(builder.MergedEntityMetadata_LIST_STRING).
		PatchBoughtList(proto.EntityDTO_APPLICATION, commodityList)

	metadata, err := mbuilder.Build()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
