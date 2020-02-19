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

type appBuilder struct {
	// TODO: Add the scope to the property for stitching, which needs corresponding change at kubeturbo side
	keepStandalone bool
	scope          string
	metric         *exporter.EntityMetric
}

func NewApplicationBuilder(keepStandalone bool, scope string, metric *exporter.EntityMetric) *appBuilder {
	return &appBuilder{
		keepStandalone: keepStandalone,
		scope:          scope,
		metric:         metric,
	}
}

func (b *appBuilder) Build() ([]*proto.EntityDTO, error) {
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

// Creates provider entity from the application entity. Currently, the use case is to create VM for Application.
func (b *appBuilder) createProviderEntity(ip string) (*proto.EntityDTO, error) {
	// For application entity, we also want to create proxy VM entity.
	VMType := proto.EntityDTO_VIRTUAL_MACHINE
	id := getEntityId(VMType, b.scope, ip)

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

// Creates consumer entity from a given provider entity.
// Currently, the use case is to create Service from Application and Database Server.
func (b *appBuilder) createConsumerEntity(providerDTO *proto.EntityDTO, ip string) (*proto.EntityDTO, error) {
	entityType := *providerDTO.EntityType
	providerId := getEntityId(entityType, b.scope, ip)

	commoditiesBought := providerDTO.CommoditiesSold

	id := getEntityId(proto.EntityDTO_SERVICE, b.scope, ip)

	// Create application commodity sold by the Service
	var commoditiesSold = []*proto.CommodityDTO{createAppCommodity()}
	var commTypes []proto.CommodityDTO_CommodityType
	for _, comm := range commoditiesBought {
		// Add SLO commodities from the bought list to the sold commodities
		if isSLOCommodity(comm.GetCommodityType()) {
			// Clear the commodity key, as we don't need to set key for
			// SLO commodities on the sold side of Service
			comm.Key = nil
			commoditiesSold = append(commoditiesSold, comm)
			commTypes = append(commTypes, *comm.CommodityType)
		}
	}
	if b.metric.HostedOnVM {
		if entityType != proto.EntityDTO_APPLICATION_COMPONENT && entityType != proto.EntityDTO_DATABASE_SERVER {
			return nil, fmt.Errorf("unsupported provider type %v to create consumer, "+
				"only APPLICATION and DATABASE_SERVER is supported when hosted on VM ", entityType)
		}
		// Hosted on VM, create non-proxy Service entity
		provider := builder.CreateProvider(entityType, providerId)
		serviceDTO, err := builder.NewEntityDTOBuilder(proto.EntityDTO_SERVICE, id).
			DisplayName(id).
			Provider(provider).
			BuysCommodities(commoditiesBought).
			SellsCommodities(commoditiesSold).
			Monitored(false).
			Create()
		if err != nil {
			return nil, err
		}
		glog.V(4).Infof("Entity DTO: %+v", serviceDTO)
		return serviceDTO, nil
	}

	// Hosted on Container, create proxy Service entity
	if entityType != proto.EntityDTO_APPLICATION_COMPONENT {
		return nil, fmt.Errorf("unsupported provider type %v to create consumer, "+
			"only APPLICATION is supported when hosted on Container", entityType)
	}
	provider := builder.CreateProvider(entityType, providerId)
	serviceDTO, err := builder.NewEntityDTOBuilder(proto.EntityDTO_SERVICE, id).
		DisplayName(id).
		Provider(provider).
		BuysCommodities(commoditiesBought).
		SellsCommodities(commoditiesSold).
		WithProperty(getEntityProperty(constant.ServicePrefix + ip)).
		ReplacedBy(getReplacementMetaData(proto.EntityDTO_SERVICE, commTypes, true)).
		Monitored(false).
		Create()
	if err != nil {
		return nil, err
	}
	serviceDTO.KeepStandalone = &b.keepStandalone
	glog.V(4).Infof("Entity DTO: %+v", serviceDTO)
	return serviceDTO, nil
}

// Creates entity DTO from the EntityMetric
// Create application commodity for non-proxy app entities only
func (b *appBuilder) createEntity(provider *proto.EntityDTO) (*proto.EntityDTO, error) {
	metric := b.metric
	entityType := metric.Type
	ip := metric.UID
	labels := metric.Labels

	// Get the entity ID
	id := getEntityId(entityType, b.scope, ip)

	// Get the commodity key
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

	if provider != nil {
		providerEntityType := *provider.EntityType
		providerId := getEntityId(providerEntityType, b.scope, ip)
		commoditiesBought := provider.CommoditiesSold
		provider := builder.CreateProvider(providerEntityType, providerId)
		// Create an application commodity
		commoditiesSold := []*proto.CommodityDTO{createAppCommodity()}
		commodities, err := createCommodities(metric, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create commodities for "+
				"entity %v: %v", id, err)
		}
		commoditiesSold = append(commoditiesSold, commodities...)
		entityDto, err := builder.NewEntityDTOBuilder(entityType, id).
			DisplayName(id).
			SellsCommodities(commoditiesSold).
			Provider(provider).
			BuysCommodities(commoditiesBought).
			WithProperty(getEntityProperty(ip)).
			Monitored(false).
			Create()
		if err != nil {
			return nil, err
		}
		glog.V(4).Infof("Entity DTO: %+v", entityDto)
		return entityDto, nil
	}

	// Create a proxy entity
	// Do not create application commodity for the proxy app entity
	commoditiesSold, err := createCommodities(metric, commKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create commodities sold for "+
			"entity %v: %v", id, err)
	}
	if len(commoditiesSold) == 0 {
		return nil, fmt.Errorf("missing commodities sold for "+
			"entity %v: %v", id, err)
	}
	var commTypes []proto.CommodityDTO_CommodityType
	for _, commodity := range commoditiesSold {
		commTypes = append(commTypes, *commodity.CommodityType)
	}
	entityDto, err := builder.NewEntityDTOBuilder(entityType, id).
		DisplayName(id).
		SellsCommodities(commoditiesSold).
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
