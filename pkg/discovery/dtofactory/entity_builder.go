package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/discovery/constant"
	"github.com/turbonomic/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type entityBuilder struct {
	// TODO: Add the scope to the property for stitching, which needs corresponding change at kubeturbo side
	scope string

	metric *exporter.EntityMetric
}

func NewEntityBuilder(scope string, metric *exporter.EntityMetric) *entityBuilder {
	return &entityBuilder{
		scope:  scope,
		metric: metric,
	}
}

func (b *entityBuilder) Build() ([]*proto.EntityDTO, error) {
	metric := b.metric

	entityType, ok := constant.EntityTypeMap[metric.Type]
	if !ok {
		err := fmt.Errorf("Unsupported entity type %v", metric.Type)
		glog.Errorf(err.Error())
		return nil, err
	}

	ip := metric.UID

	commodities := []*proto.CommodityDTO{}
	commTypes := []proto.CommodityDTO_CommodityType{}
	commMetrics := metric.Metrics
	for key, value := range commMetrics {
		var commType proto.CommodityDTO_CommodityType
		commType, ok := constant.CommodityTypeMap[key]

		if !ok {
			err := fmt.Errorf("Unsupported commodity type %s", key)
			glog.Errorf(err.Error())
			continue
		}

		capacity, ok := constant.CommodityCapMap[commType]
		if !ok {
			err := fmt.Errorf("Missing commodity capacity for type %s", commType)
			glog.Errorf(err.Error())
			continue
		}

		// TODO: Remove this if using 'millisec' unit at exporter side
		if commType == proto.CommodityDTO_RESPONSE_TIME {
			//value *= 1000 // Convert second to millisecond
		}

		// Adjust the capacity in case utilization > 1
		if value >= capacity {
			capacity = value // + 1
		}

		commodity, err := builder.NewCommodityDTOBuilder(commType).
			Used(value).Capacity(capacity).Key(ip).Create()

		if err != nil {
			glog.Errorf("Error building a commodity: %s", err)
			continue
		}

		commodities = append(commodities, commodity)
		commTypes = append(commTypes, commType)
	}

	id := b.getEntityId(entityType, ip)

	dto, err := builder.NewEntityDTOBuilder(entityType, id).
		DisplayName(id).
		SellsCommodities(commodities).
		WithProperty(getEntityProperty(ip)).
		ReplacedBy(getReplacementMetaData(entityType, commTypes)).
		Create()

	if err != nil {
		glog.Errorf("Error building EntityDTO from metric %v: %s", metric, err)
		return nil, err
	}

	dtos := []*proto.EntityDTO{dto}

	return dtos, nil
}

func (b *entityBuilder) getEntityId(entityType proto.EntityDTO_EntityType, entityName string) string {
	eType := proto.EntityDTO_EntityType_name[int32(entityType)]

	return fmt.Sprintf("%s-%s/%s", eType, b.scope, entityName)
}

func getReplacementMetaData(entityType proto.EntityDTO_EntityType, commTypes []proto.CommodityDTO_CommodityType) *proto.EntityDTO_ReplacementEntityMetaData {
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
		b.PatchSellingWithProperty(commType, []string{constant.Used, constant.Capacity})
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
