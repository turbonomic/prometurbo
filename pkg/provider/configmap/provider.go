package configmap

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"

	"github.ibm.com/turbonomic/prometurbo/pkg/config"
	"github.ibm.com/turbonomic/prometurbo/pkg/provider"

	"github.ibm.com/turbonomic/turbo-metrics/api/v1alpha1"
)

type MetricProviderImpl struct {
	serverDefs   map[string]*serverDef
	exporterDefs map[string]*exporterDef
}

func (p *MetricProviderImpl) GetTasks() (tasks []*provider.Task) {
	for _, svrDef := range p.serverDefs {
		for _, exporter := range svrDef.exporters {
			expDef, found := p.exporterDefs[exporter]
			if !found {
				continue
			}
			for _, entityDef := range expDef.entityDefs {
				clusterId := v1alpha1.ClusterIdentifier{ID: svrDef.clusterId}
				tasks = append(tasks, provider.NewTask(svrDef.promClient, entityDef).WithClusterId(&clusterId))
			}
		}
	}
	return
}

func GetMetricProvider(prometheusConfigFileName string) (provider.MetricProvider, error) {
	_, err := os.Stat(prometheusConfigFileName)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("metrics config file %v does not exist", prometheusConfigFileName)
	}
	// Load metric discovery configuration
	metricConf, err := config.NewMetricsDiscoveryConfig(prometheusConfigFileName)
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("%s", spew.Sdump(metricConf))

	// Construct prometheus servers from configuration
	promServers, err := serversFromConfigMap(metricConf)
	if err != nil || len(promServers) == 0 {
		return nil, fmt.Errorf("failed to construct servers from configuration %s: %v",
			prometheusConfigFileName, err)
	}

	// Construct exporter provider from configuration
	promExporters, err := exportersFromConfigMap(metricConf)
	if err != nil || len(promExporters) == 0 {
		return nil, fmt.Errorf("failed to construct exporters from configuration %s: %v",
			prometheusConfigFileName, err)
	}

	return &MetricProviderImpl{
		serverDefs:   promServers,
		exporterDefs: promExporters,
	}, nil
}
