package provider

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
)

type Provider struct {
	promClients  []*prometheus.RestClient
	exporterDefs []*exporterDef
}

func NewProvider(promClients []*prometheus.RestClient, exporterDefs []*exporterDef) *Provider {
	return &Provider{
		promClients:  promClients,
		exporterDefs: exporterDefs,
	}
}

func (p *Provider) GetMetrics() ([]*EntityMetric, error) {
	var metrics []*EntityMetric
	// TODO: use goroutine
	for _, promClient := range p.promClients {
		var metricsForProms []*EntityMetric
		for _, exporterDef := range p.exporterDefs {
			metricsForExporters := getMetricsForExporter(promClient, exporterDef)
			metricsForProms = append(metricsForProms, metricsForExporters...)
		}
		metrics = append(metrics, metricsForProms...)
	}
	return metrics, nil
}

func getMetricsForExporter(
	promClient *prometheus.RestClient, exporterDef *exporterDef) []*EntityMetric {
	var metricsForExporter []*EntityMetric
	for _, entityDef := range exporterDef.entityDefs {
		metricsForEntity := getMetricsForEntity(promClient, entityDef)
		metricsForExporter = append(metricsForExporter, metricsForEntity...)
	}
	return metricsForExporter
}

func getMetricsForEntity(
	promClient *prometheus.RestClient, entityDef *entityDef) []*EntityMetric {
	var metricsForEntity []*EntityMetric
	var metricsForEntityMap = map[string]*EntityMetric{}
	for metricType, metricDef := range entityDef.metricDefs {
		for metricKind, metricQuery := range metricDef.queries {
			metricSeries, err := promClient.GetMetrics(metricQuery)
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
					metricsForEntityMap[id] = NewEntityMetric(id, entityDef.eType)
				}
				for name, value := range attr {
					metricsForEntityMap[id].SetLabel(name, value)
				}
				metricsForEntityMap[id].SetMetric(metricType, metricKind, basicMetricData.GetValue())
			}
		}
	}
	for _, metric := range metricsForEntityMap {
		metricsForEntity = append(metricsForEntity, metric)
	}
	return metricsForEntity
}
