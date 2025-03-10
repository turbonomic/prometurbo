package configmap

import (
	"fmt"

	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"

	"github.ibm.com/turbonomic/prometurbo/pkg/config"
	"github.ibm.com/turbonomic/prometurbo/pkg/provider"
)

func entityDefFromConfigMap(entityConfig config.EntityConfig) (*provider.EntityDef, error) {
	if entityConfig.Type == "" {
		return nil, fmt.Errorf("empty EntityDef type")
	}
	if !data.IsValidDIFEntity(entityConfig.Type) {
		return nil, fmt.Errorf("unsupported EntityDef type %v", entityConfig.Type)
	}
	if len(entityConfig.MetricConfigs) == 0 {
		return nil, fmt.Errorf("empty MetricDef configuration for EntityDef type %v", entityConfig.Type)
	}
	var metrics []*provider.MetricDef
	for _, metricConfig := range entityConfig.MetricConfigs {
		metric, err := metricDefFromConfigMap(metricConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create metricDefs for %v [%v]: %v",
				entityConfig.Type, metricConfig.Type, err)
		}
		metrics = append(metrics, metric)
	}
	attributes, err := attributesFromConfigMap(entityConfig.AttributeConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to create attributeDefs for EntityDef type %v: %v", entityConfig.Type, err)
	}
	return &provider.EntityDef{
		EType:         entityConfig.Type,
		HostedOnVM:    entityConfig.HostedOnVM,
		MetricDefs:    metrics,
		AttributeDefs: attributes,
	}, nil
}
