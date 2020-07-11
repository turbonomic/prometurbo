package provider

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/prometheus"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"math"
)

type Task struct {
	source    *prometheus.RestClient
	entityDef *entityDef
}

func NewTask(source *prometheus.RestClient, entityDef *entityDef) *Task {
	return &Task{
		source:    source,
		entityDef: entityDef,
	}
}

// Implement the ITask Run() interface
func (t *Task) Run() []*data.DIFEntity {
	return t.getMetricsForEntity()
}

func (t *Task) getMetricsForEntity() []*data.DIFEntity {
	promClient := t.source
	entityDef := t.entityDef
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
				entityAttr, err := entityDef.reconcileAttributes(basicMetricData.Labels)
				if err != nil {
					glog.Errorf("Failed to reconcile attributes from labels %+v obtained from %v [%v] for entity %v: %v.",
						basicMetricData.Labels, metricKind, metricQuery, entityType, err)
					continue
				}
				difEntity, found := entityMetricsMap[entityAttr.id]
				if !found {
					difEntity = data.NewDIFEntity(entityAttr.id, entityType).
						WithNamespace(entityAttr.namespace)
					if entityAttr.ip != "" {
						difEntity.Matching(entityAttr.ip)
					}
					if entityDef.hostedOnVM {
						difEntity.HostedOnType(data.VM).HostedOnIP(entityAttr.ip)
					}
					processOwner(difEntity, entityAttr)
					entityMetricsMap[entityAttr.id] = difEntity
				}
				// Process metrics
				if difMetricValKind, ok := metricKindToDIFMetricValKind[metricKind]; ok {
					glog.V(4).Infof("Processing %v, %v, %v",
						difEntity.Name, metricType, difMetricValKind)
					difEntity.AddMetric(metricType, difMetricValKind, basicMetricData.GetValue(), "")
				}
			}
		}
	}
	for _, metric := range entityMetricsMap {
		entityMetrics = append(entityMetrics, metric)
	}
	return entityMetrics
}

func processOwner(entity *data.DIFEntity, entityAttr *entityAttribute) {
	if entityAttr.service != "" {
		ServicePrefix := "Service-"
		svcID := ServicePrefix + entity.UID
		entity.PartOfEntity("service", svcID, entityAttr.service)
	}
}
