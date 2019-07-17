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
	keepStandalone  bool
	createProxyVM   bool
	scope string

	metric *exporter.EntityMetric
}

func NewEntityBuilder(keepStandalone bool, createProxyVM bool, scope string, metric *exporter.EntityMetric) *entityBuilder {
	return &entityBuilder{
		keepStandalone: keepStandalone,
		createProxyVM:  createProxyVM,
		scope:  		scope,
		metric: 		metric,
	}
}

func (b *entityBuilder) Build() ([]*proto.EntityDTO, error) {
	metric := b.metric
	ip := metric.UID


	if b.createProxyVM {
		providerDto, err := b.createProviderEntity(ip)

		if err != nil {
			glog.Errorf("Error building provider EntityDTO from metric %v: %s", metric, err)
			return nil, err
		}
		dtos := []*proto.EntityDTO{providerDto}

		entityDto, err := b.createEntityDto(providerDto)

		if err != nil {
			glog.Errorf("Error building EntityDTO from metric %v: %s", metric, err)
			return nil, err
		}

		dtos = append(dtos, entityDto)

		consumerDto, err := b.createConsumerEntity(entityDto, ip)

		if err != nil {
			glog.Errorf("Error building consumer EntityDTO from metric %v: %s", metric, err)
		} else {
			dtos = append(dtos, consumerDto)
		}

		return dtos, nil
	} else {
		entityDto, err := b.createEntityDto(nil)

		if err != nil {
			glog.Errorf("Error building EntityDTO from metric %v: %s", metric, err)
			return nil, err
		}

		dtos := []*proto.EntityDTO{entityDto}

		consumerDto, err := b.createConsumerEntity(entityDto, ip)

		if err != nil {
			glog.Errorf("Error building consumer EntityDTO from metric %v: %s", metric, err)
		} else {
			dtos = append(dtos, consumerDto)
		}

		return dtos, nil
	}
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

	commodities := []*proto.CommodityDTO{}

	// If metric exporter doesn't provide the necessary commodity usage, create one with value 0.
	// TODO: This is to match the supply chain and should be removed.
	for commType := range constant.VMCommodityTypeMap {
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


	return vmDto, nil
}

// Creates consumer entity from a given provider entity. Currently, the use case is to create vApp from Application.
func (b *entityBuilder) createConsumerEntity(provider *proto.EntityDTO, ip string) (*proto.EntityDTO, error) {
	entityType := *provider.EntityType
	providerId := b.getEntityId(entityType, ip)
	commodities := provider.CommoditiesSold
	commTypes := []proto.CommodityDTO_CommodityType{}
	for _, comm := range commodities {
		commTypes = append(commTypes, *comm.CommodityType)
	}

	// For application entity, we also want to create proxy entities for vApp.
	// The logic may or may not apply to other entity types depending on future use cases, if any.
	if b.createProxyVM {
		if entityType == proto.EntityDTO_APPLICATION {
			provider := builder.CreateProvider(entityType, providerId)
			vAppType := proto.EntityDTO_VIRTUAL_APPLICATION
			id := b.getEntityId(vAppType, ip)
			vappDto, err := builder.NewEntityDTOBuilder(vAppType, id).
				DisplayName(id).
				Provider(provider).
				BuysCommodities(commodities).
				WithProperty(getEntityProperty(constant.VAppPrefix + ip)).
				Monitored(false).
				Create()

			if err != nil {
				return nil, err
			}

			return vappDto, nil
		}
	} else {
		if entityType == proto.EntityDTO_APPLICATION {
			provider := builder.CreateProvider(entityType, providerId)
			vAppType := proto.EntityDTO_VIRTUAL_APPLICATION
			id := b.getEntityId(vAppType, ip)
			vappDto, err := builder.NewEntityDTOBuilder(vAppType, id).
				DisplayName(id).
				Provider(provider).
				BuysCommodities(commodities).
				WithProperty(getEntityProperty(constant.VAppPrefix + ip)).
				ReplacedBy(getReplacementMetaData(vAppType, commTypes, true)).
				Monitored(false).
				Create()

			if err != nil {
				return nil, err
			}

			vappDto.KeepStandalone = &b.keepStandalone

			return vappDto, nil
		}
	}

	return nil, fmt.Errorf("Unsupported provider type %v to create consumer", entityType)
}

// Creates entity DTO from the EntityMetric
func (b *entityBuilder) createEntityDto(provider *proto.EntityDTO) (*proto.EntityDTO, error) {
	metric := b.metric

	entityType := metric.Type
	if _, ok := constant.EntityTypeMap[entityType]; !ok {
		err := fmt.Errorf("Unsupported entity type %v", metric.Type)
		glog.Errorf(err.Error())
		return nil, err
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

	commodities := []*proto.CommodityDTO{}
	commTypes := []proto.CommodityDTO_CommodityType{}
	commMetrics := metric.Metrics

	// If metric exporter doesn't provide the necessary commodity usage, create one with value 0.
	// TODO: This is to match the supply chain and should be removed.
	for commType := range constant.AppCommodityTypeMap {
		if _, ok := commMetrics[commType]; !ok {
			commMetrics[commType] = 0
		}
	}

	for key, value := range commMetrics {
		commType := key

		if _, ok := constant.AppCommodityTypeMap[commType]; !ok {
			err := fmt.Errorf("Unsupported commodity type %s", key)
			glog.Errorf(err.Error())
			continue
		}

		commodity, err := builder.NewCommodityDTOBuilder(commType).
			Used(value).Key(commKey).Create()

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

		glog.V(4).Infof("Entity DTO: %++v", entityDto)
		return entityDto, nil

	} else {
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

		glog.V(4).Infof("Entity DTO: %++v", entityDto)
		return entityDto, nil
	}
}
