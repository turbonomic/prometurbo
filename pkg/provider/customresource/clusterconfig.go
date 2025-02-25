package customresource

import (
	"fmt"

	"github.com/golang/glog"
	"github.ibm.com/turbonomic/turbo-metrics/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type clusterConfig struct {
	clusterId     *v1alpha1.ClusterIdentifier
	queryMappings []*queryMapping
}

func clusterConfigFromCustomResource(specClusterConfig v1alpha1.ClusterConfiguration,
	queryMappings []*queryMapping) (*clusterConfig, error) {
	var clusterId *v1alpha1.ClusterIdentifier
	id := specClusterConfig.Identifier
	if id.ID != "" {
		clusterId = &id
	}
	var filteredQueryMappings []*queryMapping
	if specClusterConfig.QueryMappingSelector != nil {
		// Use labelSelectors to filter queryMappings
		selector, err := metav1.LabelSelectorAsSelector(specClusterConfig.QueryMappingSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to parse labelSelector %v: %v", specClusterConfig.QueryMappingSelector, err)
		}
		for _, qryMapping := range queryMappings {
			if selector.Matches(labels.Set(qryMapping.qryMapping.Labels)) {
				filteredQueryMappings = append(filteredQueryMappings, qryMapping)
				continue
			}
			if id.ID != "" {
				glog.V(2).Infof("Excluding %v/%v for cluster %v.",
					qryMapping.qryMapping.GetNamespace(), qryMapping.qryMapping.GetName(), id.ID)
			} else {
				glog.V(2).Infof("Excluding %v/%v.",
					qryMapping.qryMapping.GetNamespace(), qryMapping.qryMapping.GetName())
			}
		}
	} else {
		// If the QueryMappingSelector field is not defined, defaults to all PrometheusQueryMapping resources in the
		// current namespace.
		filteredQueryMappings = queryMappings
	}

	return &clusterConfig{
		clusterId:     clusterId,
		queryMappings: filteredQueryMappings,
	}, nil
}
