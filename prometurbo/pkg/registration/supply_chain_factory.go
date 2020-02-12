package registration

import (
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/constant"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/supplychain"
)

var (
	VMIPFieldPaths = []string{constant.SupplyChainConstantVirtualMachineData}

	vCpuType         = proto.CommodityDTO_VCPU
	vMemType         = proto.CommodityDTO_VMEM
	respTimeType     = proto.CommodityDTO_RESPONSE_TIME
	transactionType  = proto.CommodityDTO_TRANSACTION
	heapType         = proto.CommodityDTO_HEAP
	collectionType   = proto.CommodityDTO_COLLECTION_TIME
	threadsType      = proto.CommodityDTO_THREADS
	cacheHitRateType = proto.CommodityDTO_DB_CACHE_HIT_RATE
	dbMemType        = proto.CommodityDTO_DB_MEM
	connectionType   = proto.CommodityDTO_CONNECTION
	applicationType  = proto.CommodityDTO_APPLICATION
	key              = "key-placeholder"

	vCpuTemplateComm = &proto.TemplateCommodity{
		CommodityType: &vCpuType,
	}

	vMemTemplateComm = &proto.TemplateCommodity{
		CommodityType: &vMemType,
	}

	respTimeTemplateComm = &proto.TemplateCommodity{
		CommodityType: &respTimeType,
		Key:           &key,
	}

	transactionTemplateComm = &proto.TemplateCommodity{
		CommodityType: &transactionType,
		Key:           &key,
	}

	heapTemplateComm = &proto.TemplateCommodity{
		CommodityType: &heapType,
	}

	collectionTemplateComm = &proto.TemplateCommodity{
		CommodityType: &collectionType,
	}

	threadsTemplateComm = &proto.TemplateCommodity{
		CommodityType: &threadsType,
	}

	cachHitRateTemplateComm = &proto.TemplateCommodity{
		CommodityType: &cacheHitRateType,
	}

	dbMemTemplateComm = &proto.TemplateCommodity{
		CommodityType: &dbMemType,
	}

	connectionTemplateComm = &proto.TemplateCommodity{
		CommodityType: &connectionType,
	}

	applicationTemplateComm = &proto.TemplateCommodity{
		CommodityType: &applicationType,
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

	// DBServer node
	dbServerNode, err := f.buildDBServerSupplyBuilder()

	if err != nil {
		return nil, err
	}

	// Service node
	serviceNode, err := f.buildServiceSupplyBuilder()

	if err != nil {
		return nil, err
	}

	// Stitching metadata for the service node
	serviceMetadata, err := f.getServiceStitchingMetaData()
	if err != nil {
		return nil, err
	}

	serviceNode.MergedEntityMetaData = serviceMetadata

	// BizTransaction node
	bizTranNode, err := f.buildBusinessTranSupplyBuilder()
	if err != nil {
		return nil, err
	}

	// BizApplication node
	bizAppNode, err := f.buildBusinessAppSupplyBuilder()

	if err != nil {
		return nil, err
	}

	return supplychain.NewSupplyChainBuilder().
		Top(bizAppNode).
		Entity(bizTranNode).
		Entity(serviceNode).
		Entity(appNode).
		Entity(dbServerNode).
		Entity(vmNode).
		Create()
}

func (f *SupplyChainFactory) buildVMSupplyBuilder() (*proto.TemplateDTO, error) {
	vmBuilder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_VIRTUAL_MACHINE).
		Sells(vCpuTemplateComm).
		Sells(vMemTemplateComm)
	vmBuilder.SetPriority(-1)
	vmBuilder.SetTemplateType(proto.TemplateDTO_BASE)

	return vmBuilder.Create()
}

// TODO: Currently we only support DATABASE_SERVER links to VIRTUAL_MACHINE
func (f *SupplyChainFactory) buildDBServerSupplyBuilder() (*proto.TemplateDTO, error) {
	dbServerToVMExternalLink, err := supplychain.NewExternalEntityLinkBuilder().
		Link(proto.EntityDTO_DATABASE_SERVER, proto.EntityDTO_VIRTUAL_MACHINE, proto.Provider_HOSTING).
		Commodity(vCpuType, false).Commodity(vMemType, false).
		ProbeEntityPropertyDef(constant.StitchingAttr, "IP Address of the VM hosting the discovered db server").
		ExternalEntityPropertyDef(supplychain.VM_IP).
		Build()
	if err != nil {
		return nil, err
	}
	dbServerBuilder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_DATABASE_SERVER).
		Sells(cachHitRateTemplateComm).
		Sells(dbMemTemplateComm).
		Sells(connectionTemplateComm).
		Sells(respTimeTemplateComm).
		Sells(transactionTemplateComm).
		Sells(applicationTemplateComm).
		ConnectsTo(dbServerToVMExternalLink).
		Provider(proto.EntityDTO_VIRTUAL_MACHINE, proto.Provider_HOSTING).
		Buys(vCpuTemplateComm).
		Buys(vMemTemplateComm)
	dbServerBuilder.SetPriority(-1)
	dbServerBuilder.SetTemplateType(proto.TemplateDTO_BASE)

	return dbServerBuilder.Create()
}

func (f *SupplyChainFactory) buildAppSupplyBuilder() (*proto.TemplateDTO, error) {
	appToVMExternalLink, err := supplychain.NewExternalEntityLinkBuilder().
		Link(proto.EntityDTO_APPLICATION_COMPONENT, proto.EntityDTO_VIRTUAL_MACHINE, proto.Provider_HOSTING).
		Commodity(vCpuType, false).Commodity(vMemType, false).
		Commodity(transactionType, true).Commodity(respTimeType, true).
		ProbeEntityPropertyDef(constant.StitchingAttr, "IP Address of the VM hosting the discovered node").
		ExternalEntityPropertyDef(supplychain.VM_IP).
		Build()

	if err != nil {
		return nil, err
	}

	appBuilder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_APPLICATION_COMPONENT).
		Sells(transactionTemplateComm).
		Sells(respTimeTemplateComm).
		Sells(heapTemplateComm).
		Sells(collectionTemplateComm).
		Sells(threadsTemplateComm).
		Sells(applicationTemplateComm).
		ConnectsTo(appToVMExternalLink).
		Provider(proto.EntityDTO_VIRTUAL_MACHINE, proto.Provider_HOSTING).
		Buys(vCpuTemplateComm).
		Buys(vMemTemplateComm)
	appBuilder.SetPriority(-1)
	appBuilder.SetTemplateType(proto.TemplateDTO_BASE)

	return appBuilder.Create()
}

func (f *SupplyChainFactory) buildServiceSupplyBuilder() (*proto.TemplateDTO, error) {

	serviceBuilder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_SERVICE).
		Provider(proto.EntityDTO_APPLICATION_COMPONENT, proto.Provider_LAYERED_OVER).
		Buys(transactionTemplateComm).
		Buys(respTimeTemplateComm).
		Buys(applicationTemplateComm).
		Provider(proto.EntityDTO_DATABASE_SERVER, proto.Provider_LAYERED_OVER).
		Buys(transactionTemplateComm).
		Buys(respTimeTemplateComm).
		Buys(applicationTemplateComm).
		Sells(transactionTemplateComm).
		Sells(respTimeTemplateComm).
		Sells(applicationTemplateComm)
	serviceBuilder.SetPriority(-1)
	serviceBuilder.SetTemplateType(proto.TemplateDTO_BASE)

	return serviceBuilder.Create()
}

func (f *SupplyChainFactory) buildBusinessTranSupplyBuilder() (*proto.TemplateDTO, error) {

	businessTranBuilder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_BUSINESS_TRANSACTION).
		Provider(proto.EntityDTO_SERVICE, proto.Provider_LAYERED_OVER).
		Buys(transactionTemplateComm).
		Buys(respTimeTemplateComm).
		Buys(applicationTemplateComm).
		Sells(transactionTemplateComm).
		Sells(respTimeTemplateComm).
		Sells(applicationTemplateComm)
	businessTranBuilder.SetPriority(-1)
	businessTranBuilder.SetTemplateType(proto.TemplateDTO_BASE)

	return businessTranBuilder.Create()
}

func (f *SupplyChainFactory) buildBusinessAppSupplyBuilder() (*proto.TemplateDTO, error) {

	businessAppBuilder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_BUSINESS_APPLICATION).
		Provider(proto.EntityDTO_SERVICE, proto.Provider_LAYERED_OVER).
		Buys(transactionTemplateComm).
		Buys(respTimeTemplateComm).
		Buys(applicationTemplateComm).
		Provider(proto.EntityDTO_BUSINESS_TRANSACTION, proto.Provider_LAYERED_OVER).
		Buys(transactionTemplateComm).
		Buys(respTimeTemplateComm).
		Buys(applicationTemplateComm)
	businessAppBuilder.SetPriority(-1)
	businessAppBuilder.SetTemplateType(proto.TemplateDTO_BASE)

	return businessAppBuilder.Create()
}

func (f *SupplyChainFactory) getVMStitchingMetaData() (*proto.MergedEntityMetadata, error) {

	var metadataBuilder *builder.MergedEntityMetadataBuilder

	metadataBuilder = builder.NewMergedEntityMetadataBuilder().
		InternalMatchingType(builder.MergedEntityMetadata_LIST_STRING).
		InternalMatchingPropertyWithDelimiter(constant.StitchingAttr, ",").
		ExternalMatchingType(builder.MergedEntityMetadata_LIST_STRING).
		ExternalMatchingFieldWithDelimiter(constant.SupplyChainConstantIpAddress, VMIPFieldPaths, ",")

	metadata, err := metadataBuilder.Build()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (f *SupplyChainFactory) getAppStitchingMetaData() (*proto.MergedEntityMetadata, error) {
	commodityList := []proto.CommodityDTO_CommodityType{
		respTimeType, transactionType, heapType, collectionType, threadsType}

	var metadataBuilder *builder.MergedEntityMetadataBuilder

	metadataBuilder = builder.NewMergedEntityMetadataBuilder().
		KeepInTopology(false).
		InternalMatchingProperty(constant.StitchingAttr).
		InternalMatchingType(builder.MergedEntityMetadata_STRING).
		ExternalMatchingProperty(constant.StitchingAttr).
		ExternalMatchingType(builder.MergedEntityMetadata_STRING).
		PatchSoldList(commodityList)

	metadata, err := metadataBuilder.Build()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (f *SupplyChainFactory) getServiceStitchingMetaData() (*proto.MergedEntityMetadata, error) {
	commodityList := []proto.CommodityDTO_CommodityType{respTimeType, transactionType}

	var metadataBuilder *builder.MergedEntityMetadataBuilder

	metadataBuilder = builder.NewMergedEntityMetadataBuilder().
		KeepInTopology(false).
		InternalMatchingProperty(constant.StitchingAttr).
		InternalMatchingType(builder.MergedEntityMetadata_STRING).
		ExternalMatchingPropertyWithDelimiter(constant.StitchingAttr, ",").
		ExternalMatchingType(builder.MergedEntityMetadata_LIST_STRING).
		PatchSoldList(commodityList).
		PatchBoughtList(proto.EntityDTO_APPLICATION_COMPONENT, commodityList)

	metadata, err := metadataBuilder.Build()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
