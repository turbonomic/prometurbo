package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/constant"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

func getEntityId(entityType proto.EntityDTO_EntityType, scope, entityName string) string {
	eType := proto.EntityDTO_EntityType_name[int32(entityType)]

	return fmt.Sprintf("%s::%s::%s", eType, scope, entityName)
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

func createAppCommodity() *proto.CommodityDTO {
	// There is no reason why the call will fail. Ignore error here to avoid unnecessary error checking.
	appCommodity, _ := builder.
		NewCommodityDTOBuilder(proto.CommodityDTO_APPLICATION).Create()
	return appCommodity
}

func isSLOCommodity(commType proto.CommodityDTO_CommodityType) bool {
	return commType == proto.CommodityDTO_TRANSACTION ||
		commType == proto.CommodityDTO_RESPONSE_TIME
}

func createCommodities(metric *exporter.EntityMetric, SLOCommKey string) ([]*proto.CommodityDTO, error) {
	supportedCommodities, supported := constant.EntityTypeMap[metric.Type]
	if !supported {
		return nil, fmt.Errorf("unsupported entity type %v", metric.Type)
	}
	var commodities []*proto.CommodityDTO
	// Create commodities based on available metrics
	for commType, value := range metric.Metrics {
		attribute, ok := supportedCommodities[commType]
		if !ok {
			glog.Warningf("Unsupported commodity type %v for entity type %v", commType, metric.Type)
			continue
		}
		if _, found := value[exporter.Used]; !found {
			glog.Errorf("Missing used value for commodity type %v, entity type %v", commType, metric.Type)
			continue
		}
		// Set used value of the commodity
		commodityBuilder := builder.NewCommodityDTOBuilder(commType).
			Used(value[exporter.Used])
		// Set commodity key for SLO commodities only
		if SLOCommKey != "" && isSLOCommodity(commType) {
			commodityBuilder.Key(SLOCommKey)
		}
		// Set capacity if it exists in the query result, or there is a default value defined
		capacity, found := value[exporter.Capacity]
		if found && capacity > 0 {
			commodityBuilder.Capacity(capacity)
		} else if attribute.DefaultCapacity > 0 {
			commodityBuilder.Capacity(attribute.DefaultCapacity)
		}
		commodity, err := commodityBuilder.Create()
		if err != nil {
			glog.Errorf("Error building a commodity: %s", err)
			continue
		}
		commodities = append(commodities, commodity)
	}
	return commodities, nil
}
