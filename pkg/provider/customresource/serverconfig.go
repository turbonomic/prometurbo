package customresource

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.ibm.com/turbonomic/prometurbo/pkg/prometheus"
	"github.ibm.com/turbonomic/turbo-metrics/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type serverConfig struct {
	promClient     *prometheus.RestClient
	clusterConfigs []*clusterConfig
}

func serverConfigFromCustomResource(
	prometheusServerConfig v1alpha1.PrometheusServerConfig,
	queryMappingMap map[string][]*queryMapping,
	kubeClient client.Client) (*serverConfig, error) {
	glog.V(2).Infof("Loading PrometheusServerConfig %v/%v.",
		prometheusServerConfig.GetNamespace(), prometheusServerConfig.GetName())
	address := prometheusServerConfig.Spec.Address
	bearerToken := getServerBearerToken(
		prometheusServerConfig.ObjectMeta.Namespace,
		prometheusServerConfig.Spec.BearerToken,
		kubeClient)
	if len(address) == 0 {
		return nil, fmt.Errorf("no prometheus server address defined")
	}
	promClient, err := prometheus.NewRestClient(address, bearerToken)
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

func getServerBearerToken(namespace string, source v1alpha1.BearerTokenSource, kubeClient client.Client) string {
	secretName := source.SecretKeyRef.Name
	secretKey := source.SecretKeyRef.Key

	if len(secretName) == 0 || len(secretKey) == 0 {
		return ""
	}

	glog.V(2).Infof("Reading Prometheus Auth Token %v/%v:%v", namespace, secretName, secretKey)

	secret := &v1.Secret{}
	err := kubeClient.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      secretName,
	}, secret)
	if err != nil {
		glog.Errorf("Failed to read Secret %s %v", secretName, err)
		return ""
	}
	if secret.Type != v1.SecretTypeOpaque {
		glog.Errorf("Incorrect secret type of Prometheus Auth Token %s", secret.Type)
		return ""
	}
	tokenBytes := secret.Data[secretKey]
	token := string(tokenBytes)
	glog.V(2).Infof("Prometheus Auth Token len: %v", len(token))
	return token
}
