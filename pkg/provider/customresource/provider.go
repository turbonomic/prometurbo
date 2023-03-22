package customresource

import (
	"context"
	"fmt"
	"regexp"

	"github.com/golang/glog"
	"github.com/turbonomic/turbo-metrics/api/v1alpha1"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/turbonomic/prometurbo/pkg/provider"
)

const (
	defaultNamespace   = "default"
	defaultServiceName = "kubernetes"
)

var (
	listOptions = client.ListOptions{
		Namespace: v1.NamespaceAll,
	}
)

type MetricProviderImpl struct {
	kubeClient client.Client
	k8sSvcId   string
}

func (p *MetricProviderImpl) GetTasks() (tasks []*provider.Task) {
	// Discover custom resources and assemble tasks
	for _, serverCfg := range p.discoverServerConfigs() {
		for _, clusterCfg := range serverCfg.clusterConfigs {
			for _, qryMapping := range clusterCfg.queryMappings {
				for _, entityDef := range qryMapping.entityDefs {
					tasks = append(append(tasks, provider.
						NewTask(serverCfg.promClient, entityDef).
						WithClusterId(clusterCfg.clusterId).
						WithK8sSvcId(p.k8sSvcId)))
				}
			}
		}
	}
	return
}

func (p *MetricProviderImpl) discoverServerConfigs() (serverConfigs []*serverConfig) {
	prometheusQueryMappingList := &v1alpha1.PrometheusQueryMappingList{}
	if err := p.kubeClient.List(context.TODO(), prometheusQueryMappingList, &listOptions); err != nil {
		glog.V(2).Infof("Unable to list PrometheusQueryMapping resource: %v.", err)
		return
	}
	prometheusQueryMappings := prometheusQueryMappingList.Items
	if len(prometheusQueryMappings) == 0 {
		glog.V(2).Info("There is no PrometheusQueryMapping resource found in the cluster.")
		return
	}
	glog.V(2).Infof("Discovered %v PrometheusQueryMapping resources.", len(prometheusQueryMappings))
	prometheusServerConfigList := &v1alpha1.PrometheusServerConfigList{}
	if err := p.kubeClient.List(context.TODO(), prometheusServerConfigList, &listOptions); err != nil {
		glog.V(2).Infof("Unable to list PrometheusServerConfig resource: %v.", err)
		return
	}
	prometheusServerConfigs := prometheusServerConfigList.Items
	if len(prometheusServerConfigs) == 0 {
		glog.V(2).Info("There is no PrometheusServerConfig resource found in the cluster.")
		return
	}
	glog.V(2).Infof("Discovered %v PrometheusServerConfig resources.", len(prometheusServerConfigs))
	return convertToServerConfigs(prometheusQueryMappings, prometheusServerConfigs)
}

func convertToServerConfigs(
	prometheusQueryMappings []v1alpha1.PrometheusQueryMapping,
	prometheusServerConfigs []v1alpha1.PrometheusServerConfig) (serverConfigs []*serverConfig) {
	queryMappingMap := make(map[string][]*queryMapping)
	for _, prometheusQueryMapping := range prometheusQueryMappings {
		qryMapping := queryMappingFromCustomResource(prometheusQueryMapping)
		if queryMappings, found := queryMappingMap[prometheusQueryMapping.GetNamespace()]; found {
			queryMappingMap[prometheusQueryMapping.GetNamespace()] = append(queryMappings, qryMapping)
		} else {
			queryMappingMap[prometheusQueryMapping.GetNamespace()] = []*queryMapping{qryMapping}
		}
	}
	for _, prometheusServerConfig := range prometheusServerConfigs {
		serverCfg, err := serverConfigFromCustomResource(prometheusServerConfig, queryMappingMap)
		if err != nil {
			glog.Errorf("Failed to load %v %v/%v: %v.",
				prometheusServerConfig.GetObjectKind().GroupVersionKind(),
				prometheusServerConfig.GetNamespace(), prometheusServerConfig.GetName(), err)
			continue
		}
		serverConfigs = append(serverConfigs, serverCfg)
	}
	return
}

func GetMetricProvider(kubeClient client.Client) (provider.MetricProvider, error) {
	k8sSvcId, err := getKubernetesServiceID(kubeClient)
	if err != nil {
		return nil, err
	}
	return &MetricProviderImpl{
		kubeClient: kubeClient,
		k8sSvcId:   k8sSvcId,
	}, nil
}

func getKubernetesServiceID(kubeClient client.Client) (string, error) {
	svc := &v1.Service{}
	err := kubeClient.Get(context.Background(), client.ObjectKey{
		Namespace: defaultNamespace,
		Name:      defaultServiceName,
	}, svc)
	if err != nil {
		return "", fmt.Errorf("failed to get default kubernetes service %s/%s: %v",
			defaultNamespace, defaultServiceName, err)
	}
	svcUID := string(svc.GetUID())
	regex := regexp.MustCompile("^[0-9a-fA-F]{8}")
	match := regex.FindStringSubmatch(svcUID)
	if len(match) != 1 {
		return "", fmt.Errorf("failed to parse UUID %v of the default kubernetes service %s/%s",
			svcUID, defaultNamespace, defaultServiceName)
	}
	return match[0], nil
}
