package provider

import (
	"fmt"
	"math"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"github.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.com/turbonomic/prometurbo/pkg/prometheus"
	"github.com/turbonomic/prometurbo/pkg/util"
)

type Task struct {
	source    *prometheus.RestClient
	entityDef *EntityDef
	clusterId *v1alpha1.ClusterIdentifier
	k8sSvcId  string
}

func NewTask(source *prometheus.RestClient, entityDef *EntityDef) *Task {
	return &Task{
		source:    source,
		entityDef: entityDef,
	}
}

func (t *Task) WithClusterId(clusterId *v1alpha1.ClusterIdentifier) *Task {
	t.clusterId = clusterId
	return t
}

func (t *Task) WithK8sSvcId(k8sSvcId string) *Task {
	t.k8sSvcId = k8sSvcId
	return t
}

// Run implements the ITask Run() interface
func (t *Task) Run() []*data.DIFEntity {
	return t.getMetricsForEntity()
}

func (t *Task) getMetricsForEntity() []*data.DIFEntity {
	promClient := t.source
	entityDef := t.entityDef
	var entityMetrics []*data.DIFEntity
	var entityMetricsMap = map[string]*data.DIFEntity{}
	for _, metricDef := range entityDef.MetricDefs {
		entityType := entityDef.EType
		for metricKind, metricQuery := range metricDef.Queries {
			metricType := metricDef.MType
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
				entityAttr, err := reconcileAttributes(basicMetricData.Labels, entityDef.AttributeDefs)
				if err != nil {
					glog.Errorf("Failed to reconcile attributes from labels %+v obtained from %v [%v] for entity %v: %v.",
						basicMetricData.Labels, metricKind, metricQuery, entityType, err)
					continue
				}
				difEntity, found := entityMetricsMap[entityAttr.ID]
				if !found {
					difEntity = data.NewDIFEntity(entityAttr.ID, entityType).
						WithNamespace(entityAttr.Namespace)
					if entityAttr.IP != "" {
						if t.k8sSvcId != "" {
							difEntity.Matching(fmt.Sprintf("%s-%s", entityAttr.IP, t.k8sSvcId))
						} else {
							difEntity.Matching(entityAttr.IP)
						}
					}
					if entityDef.HostedOnVM {
						difEntity.HostedOnType(data.VM).HostedOnIP(entityAttr.IP)
					}
					processOwner(difEntity, entityAttr)
					entityMetricsMap[entityAttr.ID] = difEntity
				}
				// Process metrics
				if difMetricValKind, ok := MetricKindToDIFMetricValKind[metricKind]; ok {
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

func processOwner(entity *data.DIFEntity, entityAttr *EntityAttribute) {
	if entityAttr.Service != "" {
		ServicePrefix := "Service-"
		svcID := ServicePrefix + entity.UID
		entity.PartOfEntity("service", svcID, entityAttr.Service)
	}
}

func reconcileAttributes(labels map[string]string, attributeDefs map[string]*AttributeValueDef) (*EntityAttribute, error) {
	var id string
	var reconciledAttributes = map[string]string{}
	for name, def := range attributeDefs {
		key := def.LabelKey
		value, exist := labels[key]
		if !exist {
			if def.IsIdentifier {
				return nil, fmt.Errorf("required identifer label key %q does not exist", key)
			}
			continue
		}
		glog.V(4).Infof("Reconcile attribute %v with label: %q, value: %q", name, key, value)
		matchIndex := def.ValueMatches.FindStringSubmatchIndex(value)
		if matchIndex == nil {
			return nil, fmt.Errorf("label %q's value %q did not match expected pattern %q",
				key, value, def.ValueMatches.String())
		}
		expandedValue := def.ValueMatches.ExpandString(nil, def.ValueAs, value, matchIndex)
		reconciledAttributes[name] = string(expandedValue)
		if def.IsIdentifier {
			id = reconciledAttributes[name]
			if id == "" {
				return nil, fmt.Errorf("empty identifier from label key %q and value %q", key, value)
			}
		}
	}
	glog.V(4).Infof("Reconciled attributes: %s", spew.Sdump(reconciledAttributes))
	namespace := reconciledAttributes["namespace"]
	entityAttr := &EntityAttribute{
		ID:        util.GetName(id, namespace),
		IP:        reconciledAttributes["ip"],
		Service:   reconciledAttributes["service"],
		Namespace: namespace,
	}
	return entityAttr, nil
}
