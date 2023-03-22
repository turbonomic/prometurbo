package customresource

import (
	"fmt"

	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"github.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.com/turbonomic/prometurbo/pkg/provider"
)

func metricDefFromCustomResource(metricConfig v1alpha1.MetricConfiguration) (*provider.MetricDef, error) {
	if !data.IsValidDIFMetric(metricConfig.Type) {
		return nil, fmt.Errorf("unsupported metric type %q", metricConfig.Type)
	}
	if len(metricConfig.QueryConfigs) == 0 {
		return nil, fmt.Errorf("empty querie configurations")
	}
	metricDef := provider.MetricDef{
		MType:   metricConfig.Type,
		Queries: make(map[string]string),
	}
	for _, queryConfig := range metricConfig.QueryConfigs {
		metricDef.Queries[queryConfig.Type] = queryConfig.PromQL
	}
	return &metricDef, nil
}
