package configmap

import (
	"fmt"

	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"

	"github.ibm.com/turbonomic/prometurbo/pkg/config"
	"github.ibm.com/turbonomic/prometurbo/pkg/provider"
)

func metricDefFromConfigMap(metricConfig config.MetricConfig) (*provider.MetricDef, error) {
	if !data.IsValidDIFMetric(metricConfig.Type) {
		return nil, fmt.Errorf("unsupported metric type %q", metricConfig.Type)
	}
	if len(metricConfig.Queries) == 0 {
		return nil, fmt.Errorf("empty queries")
	}
	if _, exist := metricConfig.Queries[provider.Used]; !exist {
		return nil, fmt.Errorf("missing query for used value")
	}
	metricDef := provider.MetricDef{
		MType:   metricConfig.Type,
		Queries: make(map[string]string),
	}
	for k, v := range metricConfig.Queries {
		metricDef.Queries[k] = v
	}
	return &metricDef, nil
}
