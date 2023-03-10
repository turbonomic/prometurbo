package customresource

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.com/turbonomic/prometurbo/pkg/prometheus"
)

type serverConfig struct {
	promClient     *prometheus.RestClient
	clusterConfigs []*clusterConfig
}

func serverConfigFromCustomResource(prometheusServerConfig v1alpha1.PrometheusServerConfig,
	queryMappingMap map[string][]*queryMapping) (*serverConfig, error) {
	glog.V(2).Infof("Loading PrometheusServerConfig %v/%v.",
		prometheusServerConfig.GetNamespace(), prometheusServerConfig.GetName())
	address := prometheusServerConfig.Spec.Address
	if len(address) == 0 {
		return nil, fmt.Errorf("no prometheus server address defined")
	}
	promClient, err := prometheus.NewRestClient(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client from %v: %v",
			address, err)
	}
	// Find all converted queryMappings in the same namespace
	queryMappings, found := queryMappingMap[prometheusServerConfig.GetNamespace()]
	if !found {
		return nil, fmt.Errorf("there is no PrometheusQueryMapping resource in namespace %v",
			prometheusServerConfig.GetNamespace())
	}
	glog.V(2).Infof("There are %v PrometheusQueryMapping resources in namespace %v",
		len(queryMappings), prometheusServerConfig.GetNamespace())
	var clusterConfigs []*clusterConfig
	if len(prometheusServerConfig.Spec.ClusterConfigs) == 0 {
		clusterConfigs = []*clusterConfig{
			{
				queryMappings: queryMappings,
			},
		}
	} else {
		for _, specClusterConfig := range prometheusServerConfig.Spec.ClusterConfigs {
			clusterCfg, err := clusterConfigFromCustomResource(specClusterConfig, queryMappings)
			if err != nil {
				glog.Errorf("Failed to load cluster configuration for %v/%v: %v",
					prometheusServerConfig.GetNamespace(), prometheusServerConfig.GetName(), err)
				continue
			}
			clusterConfigs = append(clusterConfigs, clusterCfg)
		}
	}
	return &serverConfig{
		promClient:     promClient,
		clusterConfigs: clusterConfigs,
	}, nil
}
