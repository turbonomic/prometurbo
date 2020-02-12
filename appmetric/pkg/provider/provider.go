package provider

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
)

type Provider struct {
	promHost     string
	promClient   *prometheus.RestClient
	exporterDefs []*exporterDef
}

type ProviderFactory struct {
	exporterDefs []*exporterDef
}

func NewProviderFactory(exporterDefs []*exporterDef) *ProviderFactory {
	return &ProviderFactory{exporterDefs: exporterDefs}
}

func (pf *ProviderFactory) NewProvider(promHost string) (*Provider, error) {
	promClient, err := prometheus.NewRestClient(promHost)
	if err != nil {
		glog.Errorf("Unable to create new provider due to error: %v", err)
		return nil, err
	}

	return &Provider{
		promHost:     promHost,
		promClient:   promClient,
		exporterDefs: pf.exporterDefs,
	}, nil
}

func (p *Provider) GetMetrics() ([]*EntityMetric, error) {
	var metrics []*EntityMetric

	// TODO: use goroutine
	for _, exporterDef := range p.exporterDefs {
		metricsForExporters := p.getMetricsForExporter(exporterDef)
		metrics = append(metrics, metricsForExporters...)
	}

	return metrics, nil
}

func (p *Provider) getMetricsForExporter(exporterDef *exporterDef) []*EntityMetric {
	var metricsForExporter []*EntityMetric
	for _, entityDef := range exporterDef.entityDefs {
		metricsForEntity := p.getMetricsForEntity(entityDef)
		metricsForExporter = append(metricsForExporter, metricsForEntity...)
	}
	return metricsForExporter
}

func (p *Provider) getMetricsForEntity(entityDef *entityDef) []*EntityMetric {
	var metricsForEntity []*EntityMetric
	var metricsForEntityMap = map[string]*EntityMetric{}
	for _, metricDef := range entityDef.metricDefs {
		for metricKind, metricQuery := range metricDef.queries {
			metricSeries, err := p.promClient.GetMetrics(metricQuery)
			if err != nil {
				glog.Errorf("Failed to query metric %v [%v] for entity type %v: %v.",
					metricKind, metricQuery, entityDef.eType, err)
				continue
			}
			for _, metricData := range metricSeries {
				basicMetricData, ok := metricData.(*prometheus.BasicMetricData)
				if !ok {
					// TODO: Enhance error messages
					glog.Errorf("Type assertion failed for metricData %+v obtained from %v [%v] for entity type %v.",
						metricData, metricKind, metricQuery, entityDef.eType)
					continue
				}
				id, attr, err := entityDef.reconcileAttributes(basicMetricData.Labels)
				if err != nil {
					glog.Errorf("Failed to reconcile attributes from labels %+v obtained from %v [%v] for entity %v: %v.",
						basicMetricData.Labels, metricKind, metricQuery, entityDef.eType, err)
					continue
				}
				if id == "" {
					glog.Warningf("Failed to get identifier from labels %+v obtained from %v [%v] for entity %v.",
						basicMetricData.Labels, metricKind, metricQuery, entityDef.eType)
					continue
				}

				if _, found := metricsForEntityMap[id]; !found {
					metricsForEntityMap[id] = NewEntityMetric(entityDef.eType, id, p.promHost).OnVM(entityDef.hostedOnVM)
				}
				for name, value := range attr {
					metricsForEntityMap[id].SetLabel(name, value)
				}
				metricsForEntityMap[id].SetMetric(metricDef.mType, metricKind, basicMetricData.GetValue())
			}
		}
	}
	for _, metric := range metricsForEntityMap {
		metricsForEntity = append(metricsForEntity, metric)
	}
	return metricsForEntity
}
