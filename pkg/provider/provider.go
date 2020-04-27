package provider

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/prometheus"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"math"
)

var metricKindToKey = map[string]data.DIFMetricValKey{
	Used:     data.AVERAGE,
	Capacity: data.CAPACITY,
}

type MetricProvider struct {
	serverDefs   map[string]*serverDef
	exporterDefs map[string]*exporterDef
}

func NewProvider(serverDefs map[string]*serverDef, exporterDefs map[string]*exporterDef) *MetricProvider {
	return &MetricProvider{
		serverDefs:   serverDefs,
		exporterDefs: exporterDefs,
	}
}

func (p *MetricProvider) GetEntityMetrics() ([]*data.DIFEntity, error) {
	var entityMetrics []*data.DIFEntity

	// TODO: use goroutine
	for _, serverDef := range p.serverDefs {
		var metricsForProms []*data.DIFEntity
		for _, exporter := range serverDef.exporters {
			exporterDef, found := p.exporterDefs[exporter]
			if !found {
				continue
			}
			metricsForExporters := getMetricsForExporter(serverDef.promClient, exporterDef)
			metricsForProms = append(metricsForProms, metricsForExporters...)
		}
		entityMetrics = append(entityMetrics, metricsForProms...)
	}

	return entityMetrics, nil
}

func getMetricsForExporter(promClient *prometheus.RestClient, exporterDef *exporterDef) []*data.DIFEntity {
	var entityMetricsForExporter []*data.DIFEntity
	for _, entityDef := range exporterDef.entityDefs {
		metricsForEntity := getMetricsForEntity(promClient, entityDef)
		entityMetricsForExporter = append(entityMetricsForExporter, metricsForEntity...)
	}
	return entityMetricsForExporter
}

func getMetricsForEntity(promClient *prometheus.RestClient, entityDef *entityDef) []*data.DIFEntity {
	var entityMetrics []*data.DIFEntity
	var entityMetricsMap = map[string]*data.DIFEntity{}
	for _, metricDef := range entityDef.metricDefs {
		entityType := entityDef.eType
		for metricKind, metricQuery := range metricDef.queries {
			metricType := metricDef.mType
			metricSeries, err := promClient.GetMetrics(metricQuery)
			if err != nil {
				glog.Errorf("Failed to query metric %v [%v] for entity type %v: %v.",
					metricKind, metricQuery, entityType, err)
				continue
			}
			for _, metricData := range metricSeries {
				basicMetricData, ok := metricData.(*prometheus.BasicMetricData)
				if !ok {
					// TODO: Enhance error messages
					glog.Errorf("Type assertion failed for metricData %+v obtained from %v [%v] for entity type %v.",
						metricData, metricKind, metricQuery, entityType)
					continue
				}
				metricValue := basicMetricData.GetValue()
				if math.IsNaN(metricValue) || math.IsInf(metricValue, 0) {
					glog.Warningf("Invalid value for metricData %+v obtained from %v [%v] for entity type %v.",
						metricData, metricKind, metricQuery, entityType)
					continue
				}
				id, attr, err := entityDef.reconcileAttributes(basicMetricData.Labels)
				if err != nil {
					glog.Errorf("Failed to reconcile attributes from labels %+v obtained from %v [%v] for entity %v: %v.",
						basicMetricData.Labels, metricKind, metricQuery, entityType, err)
					continue
				}
				if id == "" {
					glog.Warningf("Failed to get identifier from labels %+v obtained from %v [%v] for entity %v.",
						basicMetricData.Labels, metricKind, metricQuery, entityType)
					continue
				}
				difEntity, found := entityMetricsMap[id]
				if !found {
					// Create new entity if it does not exist
					difEntity = data.NewDIFEntity(id, entityType).Matching(id)
					if entityDef.hostedOnVM {
						difEntity.HostedOnType(data.VM).HostedOnIP(id)
					}
					processOwner(difEntity, attr)
					entityMetricsMap[id] = difEntity
				}
				// Process metrics
				if metricKey, ok := metricKindToKey[metricKind]; ok {
					glog.V(4).Infof("Processing %v, %v, %v",
						difEntity.Name, metricType, metricKey)
					difEntity.AddMetric(metricType, metricKey, basicMetricData.GetValue())
				}
			}
		}
	}
	for _, metric := range entityMetricsMap {
		entityMetrics = append(entityMetrics, metric)
	}
	return entityMetrics
}

func processOwner(entity *data.DIFEntity, attr map[string]string) {
	for key, label := range attr {
		if key == "service" {
			ServicePrefix := "Service-"
			svcID := ServicePrefix + entity.UID
			entity.PartOfEntity("service", svcID, label)
		}
	}
}
