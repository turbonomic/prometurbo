package provider

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/inter"
	"github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
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

func (p *Provider) GetMetrics() ([]*inter.EntityMetric, error) {
	var metrics []*inter.EntityMetric
	// TODO: use goroutine
	for _, promClient := range p.promClients {
		var metricsForProms []*inter.EntityMetric
		for _, exporterDef := range p.exporterDefs {
			metricsForExporters := getMetricsForExporter(promClient, exporterDef)
			metricsForProms = append(metricsForProms, metricsForExporters...)
		}
		metrics = append(metrics, metricsForProms...)
	}
	return metrics, nil
}

func getMetricsForExporter(
	promClient *prometheus.RestClient, exporterDef *exporterDef) []*inter.EntityMetric {
	var metricsForExporter []*inter.EntityMetric
	for _, entityDef := range exporterDef.entityDefs {
		metricsForEntity := getMetricsForEntity(promClient, entityDef)
		metricsForExporter = append(metricsForExporter, metricsForEntity...)
	}
	return metricsForExporter
}

func getMetricsForEntity(
	promClient *prometheus.RestClient, entityDef *entityDef) []*inter.EntityMetric {
	var metricsForEntity []*inter.EntityMetric
	var metricsForEntityMap = map[string]*inter.EntityMetric{}
	for _, metricDef := range entityDef.metricDefs {
		metricSeries, err := promClient.GetMetrics(metricDef.query)
		if err != nil {
			glog.Errorf("Failed to get metric %v for entity type %v: %v.",
				metricDef, entityDef.eType, err)
			continue
		}
		for _, metricData := range metricSeries {
			basicMetricData, ok := metricData.(*prometheus.BasicMetricData)
			if !ok {
				// TODO: Enhance error messages
				glog.Errorf("Type assertion failed for metricData %+v obtained from %v for entity type %v.",
					metricData, metricDef, entityDef.eType)
				continue
			}
			id, attr, err := entityDef.reconcileAttributes(basicMetricData.Labels)
			if err != nil {
				glog.Errorf("Failed to reconcile attributes from labels %+v obtained from %v for entity %v: %v.",
					basicMetricData.Labels, metricDef, entityDef.eType, err)
				continue
			}
			if id == "" {
				glog.Warningf("Failed to get identifier from labels %+v obtained from %v for entity %v.",
					basicMetricData.Labels, metricDef, entityDef.eType)
				continue
			}
			metric, found := metricsForEntityMap[id]
			if !found {
				metric = &inter.EntityMetric{
					UID:     id,
					Type:    entityDef.eType,
					Labels:  attr,
					Metrics: map[proto.CommodityDTO_CommodityType]float64{},
				}
				metricsForEntityMap[id] = metric
			}
			metricsForEntityMap[id].SetMetric(metricDef.mType, basicMetricData.GetValue())
		}
	}
	for _, metric := range metricsForEntityMap {
		metricsForEntity = append(metricsForEntity, metric)
	}
	return metricsForEntity
}
