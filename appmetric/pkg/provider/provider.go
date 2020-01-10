package provider

import (
	"fmt"
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
			metricsForExporters, err := getMetricsForExporter(promClient, exporterDef)
			if err != nil {
				glog.Errorf("Failed to get entity metrics for exporter %v from host %v: %v",
					exporterDef.name, promClient.GetHost(), err)
				continue
			}
			metricsForProms = append(metricsForProms, metricsForExporters...)
		}
		metrics = append(metrics, metricsForProms...)
	}
	return metrics, nil
}

func getMetricsForExporter(
	promClient *prometheus.RestClient, exporterDef *exporterDef) ([]*inter.EntityMetric, error) {
	var metricsForExporter []*inter.EntityMetric
	for _, entityDef := range exporterDef.entityDefs {
		metricsForEntity, err := getMetricsForEntity(promClient, entityDef)
		if err != nil {
			return nil, err
		}
		metricsForExporter = append(metricsForExporter, metricsForEntity...)
	}
	return metricsForExporter, nil
}

func getMetricsForEntity(
	promClient *prometheus.RestClient, entityDef *entityDef) ([]*inter.EntityMetric, error) {
	var metricsForEntity []*inter.EntityMetric
	var metricsForEntityMap = map[string]*inter.EntityMetric{}
	for _, metricDef := range entityDef.metricDefs {
		metricSeries, err := promClient.GetMetrics(metricDef.query)
		if err != nil {
			return nil, fmt.Errorf("failed to query metric for entity type %v: %v", entityDef.eType, err)
		}
		for _, metricData := range metricSeries {
			basicMetricData, ok := metricData.(*prometheus.BasicMetricData)
			if !ok {
				// TODO: Enhance error messages
				return nil, fmt.Errorf("type assertion failed for metricData %+v", metricData)
			}
			id, attr, err := entityDef.reconcileAttributes(basicMetricData.Labels)
			if err != nil {
				return nil, fmt.Errorf("failed to reconcile attributes from labels: %v", err)
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
	return metricsForEntity, nil
}