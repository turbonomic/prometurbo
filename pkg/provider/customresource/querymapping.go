package customresource

import (
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.com/turbonomic/prometurbo/pkg/provider"
)

type queryMapping struct {
	labels     map[string]string
	entityDefs []*provider.EntityDef
}

func queryMappingFromCustomResource(prometheusQueryMapping v1alpha1.PrometheusQueryMapping) *queryMapping {
	var entityDefs []*provider.EntityDef
	for _, entityConfig := range prometheusQueryMapping.Spec.EntityConfigs {
		entityDef, err := entityDefFromCustomResource(entityConfig)
		if err != nil {
			glog.Errorf("Failed to parse EntityConfiguration in %v/%v.",
				prometheusQueryMapping.GetNamespace(), prometheusQueryMapping.GetName())
			// TODO: Post status
			continue
		}
		entityDefs = append(entityDefs, entityDef)
	}
	return &queryMapping{
		labels:     prometheusQueryMapping.Labels,
		entityDefs: entityDefs,
	}
}
