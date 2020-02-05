package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/constant"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/supplychain"
)

type entityBuilder struct {
	// TODO: Add the scope to the property for stitching, which needs corresponding change at kubeturbo side
	keepStandalone bool
	scope          string
	metric         *exporter.EntityMetric
}

func NewEntityBuilder(keepStandalone bool, scope string, metric *exporter.EntityMetric) *entityBuilder {
	return &entityBuilder{
		keepStandalone: keepStandalone,
		scope:          scope,
		metric:         metric,
	}
}

func (b *entityBuilder) Build() ([]*proto.EntityDTO, error) {
	var provider *proto.EntityDTO
	var entities []*proto.EntityDTO
	var err error
	ip := b.metric.UID

	if b.metric.HostedOnVM {
		provider, err = b.createProviderEntity(ip)
		if err != nil {
			return nil, err
		}
		entities = append(entities, provider)
	}

	entity, err := b.createEntity(provider)
	if err != nil {
		return nil, err
	}

	entities = append(entities, entity)

	consumer, err := b.createConsumerEntity(entity, ip)
	if err != nil {
		return nil, err
	}

	entities = append(entities, consumer)

	return entities, nil

}

func (b *entityBuilder) BuildBusinessApp(vapps []*proto.EntityDTO, name string) (*proto.EntityDTO, error) {
	bizAppDtoBuilder := builder.NewEntityDTOBuilder(proto.EntityDTO_BUSINESS_APPLICATION,
		b.getEntityId(proto.EntityDTO_BUSINESS_APPLICATION, name))
	for _, vapp := range vapps {
		provider := builder.CreateProvider(proto.EntityDTO_VIRTUAL_APPLICATION, *vapp.Id)
		bizAppDtoBuilder.Provider(provider).BuysCommodities(vapp.CommoditiesSold)
	}
	bizAppDto, err := bizAppDtoBuilder.DisplayName(name).
		WithProperty(getEntityProperty(constant.BizAppPrefix + name)).
		Monitored(false).
		Create()

	if err != nil {
		return nil, err
	}
	bizAppDto.KeepStandalone = &b.keepStandalone
	glog.V(4).Infof("Entity DTO: %+v", bizAppDto)
	return bizAppDto, nil
}

func (b *entityBuilder) getEntityId(entityType proto.EntityDTO_EntityType, entityName string) string {
	eType := proto.EntityDTO_EntityType_name[int32(entityType)]

	return fmt.Sprintf("%s-%s:%s", eType, b.scope, entityName)
}

func getReplacementMetaData(entityType proto.EntityDTO_EntityType, commTypes []proto.CommodityDTO_CommodityType, bought bool) *proto.EntityDTO_ReplacementEntityMetaData {
	attr := constant.StitchingAttr
	useTopoExt := true

	b := builder.NewReplacementEntityMetaDataBuilder().
		Matching(attr).
		MatchingExternal(&proto.ServerEntityPropDef{
			Entity:     &entityType,
			Attribute:  &attr,
			UseTopoExt: &useTopoExt,
		})

	for _, commType := range commTypes {
		if bought {
			b.PatchBuyingWithProperty(commType, []string{constant.Used})
		} else {
			b.PatchSellingWithProperty(commType, []string{constant.Used})
		}
	}

	return b.Build()
}

func getEntityProperty(value string) *proto.EntityDTO_EntityProperty {
	attr := constant.StitchingAttr
	ns := constant.DefaultPropertyNamespace

	return &proto.EntityDTO_EntityProperty{
		Namespace: &ns,
		Name:      &attr,
		Value:     &value,
	}
}

// Creates provider entity from the application entity. Currently, the use case is to create VM for Application.
func (b *entityBuilder) createProviderEntity(ip string) (*proto.EntityDTO, error) {
	// For application entity, we also want to create proxy VM entity.
	VMType := proto.EntityDTO_VIRTUAL_MACHINE
	id := b.getEntityId(VMType, ip)

	var commodities []*proto.CommodityDTO

	// If metric exporter doesn't provide the necessary commodity usage, create one with value 0.
	// TODO: This is to match the supply chain and should be removed.
	for commType := range constant.SupportedVMCommodities {
		commodity, err := builder.NewCommodityDTOBuilder(commType).Used(0).Create()
		if err != nil {
			glog.Errorf("Error building a commodity: %s", err)
			continue
		}
		commodities = append(commodities, commodity)
	}

	vmDto, err := builder.NewEntityDTOBuilder(VMType, id).
		DisplayName(id).
		SellsCommodities(commodities).
		WithProperty(getEntityProperty(ip)).
		ReplacedBy(builder.NewReplacementEntityMetaDataBuilder().
			Matching(constant.StitchingAttr).
			MatchingExternal(supplychain.VM_IP).Build()).
		Monitored(false).
		Create()

	if err != nil {
		return nil, err
	}

	vmDto.KeepStandalone = &b.keepStandalone
	glog.V(4).Infof("Entity DTO: %+v", vmDto)
	return vmDto, nil
}

// Creates consumer entity from a given provider entity. Currently, the use case is to create vApp from Application
// and Database Server.
func (b *entityBuilder) createConsumerEntity(providerDTO *proto.EntityDTO, ip string) (*proto.EntityDTO, error) {
	entityType := *providerDTO.EntityType
	providerId := b.getEntityId(entityType, ip)
	commodities := providerDTO.CommoditiesSold

	if b.metric.HostedOnVM {
		if entityType != proto.EntityDTO_APPLICATION && entityType != proto.EntityDTO_DATABASE_SERVER {
			return nil, fmt.Errorf("unsupported provider type %v to create consumer, " +
				"only APPLICATION and DATABASE_SERVER is supported when hosted on VM ", entityType)
		}
		// Hosted on VM, create non-proxy Virtual Application entity
		provider := builder.CreateProvider(entityType, providerId)
		vAppType := proto.EntityDTO_VIRTUAL_APPLICATION
		id := b.getEntityId(vAppType, ip)
		vappDto, err := builder.NewEntityDTOBuilder(vAppType, id).
			DisplayName(id).
			Provider(provider).
			BuysCommodities(commodities).
			//Added the sell commodity for VApp due to the businessApp, I don't see any problem doing so even without BizApp
			SellsCommodities(commodities).
			Monitored(false).
			Create()
		if err != nil {
			return nil, err
		}
		glog.V(4).Infof("Entity DTO: %+v", vappDto)
		return vappDto, nil
	}

	// Hosted on Container, create proxy Virtual Application entity
	if entityType != proto.EntityDTO_APPLICATION {
		return nil, fmt.Errorf("unsupported provider type %v to create consumer, " +
			"only APPLICATION is supported when hosted on Container", entityType)
	}
	var commTypes []proto.CommodityDTO_CommodityType
	for _, comm := range commodities {
		commTypes = append(commTypes, *comm.CommodityType)
	}
	provider := builder.CreateProvider(entityType, providerId)
	vAppType := proto.EntityDTO_VIRTUAL_APPLICATION
	id := b.getEntityId(vAppType, ip)
	vappDto, err := builder.NewEntityDTOBuilder(vAppType, id).
		DisplayName(id).
		Provider(provider).
		BuysCommodities(commodities).
		//Added the sell commodity for VApp due to the businessApp, I don't see any problem doing so even without BizApp
		SellsCommodities(commodities).
		WithProperty(getEntityProperty(constant.VAppPrefix + ip)).
		ReplacedBy(getReplacementMetaData(vAppType, commTypes, true)).
		Monitored(false).
		Create()

	if err != nil {
		return nil, err
	}

	vappDto.KeepStandalone = &b.keepStandalone
	glog.V(4).Infof("Entity DTO: %+v", vappDto)
	return vappDto, nil
}

// Creates entity DTO from the EntityMetric
func (b *entityBuilder) createEntity(provider *proto.EntityDTO) (*proto.EntityDTO, error) {
	metric := b.metric

	entityType := metric.Type
	supportedCommodities, ok := constant.EntityTypeMap[entityType]
	if !ok {
		return nil, fmt.Errorf("unsupported entity type %v", metric.Type)
	}

	ip := metric.UID
	labels := metric.Labels

	var commKey, serviceName, serviceNamespace string
	serviceName, serviceNameExists := labels["service_name"]
	serviceNamespace, serviceNamespaceExists := labels["service_ns"]

	if serviceNameExists && serviceNamespaceExists {
		commKey = fmt.Sprintf("%s/%s", serviceNamespace, serviceName)
	} else {
		commKey = ip
	}

	if serviceNamespace != "" && serviceName != "" {
		commKey = fmt.Sprintf("%s/%s", serviceNamespace, serviceName)
	} else {
		commKey = ip
	}

	var commodities []*proto.CommodityDTO
	var commTypes []proto.CommodityDTO_CommodityType

	for commType, value := range metric.Metrics {
		defaultValue, ok := supportedCommodities[commType]
		if !ok {
			glog.Warningf("Unsupported commodity type %v for entity type %v", commType, entityType)
			continue
		}
		if _, found := value[exporter.Used]; !found {
			glog.Errorf("Missing used value for commodity type %v, entity type %v", commType, entityType)
			continue
		}
		commodityBuilder := builder.NewCommodityDTOBuilder(commType).
			Used(value[exporter.Used]).Key(commKey)
		capacity, found := value[exporter.Capacity]
		if found && capacity > 0 {
			commodityBuilder.Capacity(capacity)
		} else if defaultValue.Capacity > 0 {
			commodityBuilder.Capacity(defaultValue.Capacity)
		}

		commodity, err := commodityBuilder.Create()
		if err != nil {
			glog.Errorf("Error building a commodity: %s", err)
			continue
		}

		commodities = append(commodities, commodity)
		commTypes = append(commTypes, commType)
	}

	id := b.getEntityId(entityType, ip)

	if provider != nil {
		providerEntityType := *provider.EntityType
		providerId := b.getEntityId(providerEntityType, ip)
		commoditiesBought := provider.CommoditiesSold
		provider := builder.CreateProvider(providerEntityType, providerId)
		entityDto, err := builder.NewEntityDTOBuilder(entityType, id).
			DisplayName(id).
			SellsCommodities(commodities).
			Provider(provider).
			BuysCommodities(commoditiesBought).
			WithProperty(getEntityProperty(ip)).
			Monitored(false).
			Create()

		if err != nil {
			glog.Errorf("Error building EntityDTO from metric %v: %s", metric, err)
			return nil, err
		}

		glog.V(4).Infof("Entity DTO: %+v", entityDto)
		return entityDto, nil
	}

	// Create a proxy entity
	entityDto, err := builder.NewEntityDTOBuilder(entityType, id).
		DisplayName(id).
		SellsCommodities(commodities).
		WithProperty(getEntityProperty(ip)).
		ReplacedBy(getReplacementMetaData(entityType, commTypes, false)).
		Monitored(false).
		Create()

	if err != nil {
		glog.Errorf("Error building EntityDTO from metric %v: %s", metric, err)
		return nil, err
	}

	entityDto.KeepStandalone = &b.keepStandalone

	glog.V(4).Infof("Entity DTO: %+v", entityDto)
	return entityDto, nil
}
