package customresource

import (
	"github.com/golang/glog"
	"github.ibm.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.ibm.com/turbonomic/prometurbo/pkg/provider"
)

type queryMapping struct {
	qryMapping *v1alpha1.PrometheusQueryMapping
	entityDefs []*provider.EntityDef
}

func queryMappingFromCustomResource(prometheusQueryMapping v1alpha1.PrometheusQueryMapping) *queryMapping {
	var entityDefs []*provider.EntityDef
	for _, entityConfig := range prometheusQueryMapping.Spec.EntityConfigs {
		entityDef, err := entityDefFromCustomResource(entityConfig)
		if err != nil {
			glog.Errorf("Failed to parse EntityConfiguration in %v/%v: %s",
				prometheusQueryMapping.GetNamespace(), prometheusQueryMapping.GetName(), err)
			// TODO: Post status
			continue
		}
		entityDefs = append(entityDefs, entityDef)
	}
	return &queryMapping{
		qryMapping: &prometheusQueryMapping,
		entityDefs: entityDefs,
	}
}
