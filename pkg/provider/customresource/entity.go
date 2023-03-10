package customresource

import (
	"fmt"

	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"github.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.com/turbonomic/prometurbo/pkg/provider"
)

func entityDefFromCustomResource(entityConfig v1alpha1.EntityConfiguration) (*provider.EntityDef, error) {
	if entityConfig.Type == "" {
		return nil, fmt.Errorf("empty EntityDef type")
	}
	if !data.IsValidDIFEntity(entityConfig.Type) {
		return nil, fmt.Errorf("unsupported EntityDef type %v", entityConfig.Type)
	}
	var metrics []*provider.MetricDef
	for _, metricConfig := range entityConfig.MetricConfigs {
		metric, err := metricDefFromCustomResource(metricConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create metricDefs for %v [%v]: %v",
				entityConfig.Type, metricConfig.Type, err)
		}
		metrics = append(metrics, metric)
	}
	attributes, err := attributesFromCustomResource(entityConfig.AttributeConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to create AttributeDefs for EntityDef type %v: %v", entityConfig.Type, err)
	}
	return &provider.EntityDef{
		EType:         entityConfig.Type,
		HostedOnVM:    entityConfig.HostedOnVM,
		MetricDefs:    metrics,
		AttributeDefs: attributes,
	}, nil
}
