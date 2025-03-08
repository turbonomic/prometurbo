package provider

import (
	"fmt"
	"math"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"github.ibm.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.ibm.com/turbonomic/prometurbo/pkg/prometheus"
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
	entityMetricsMap := map[string]*data.DIFEntity{}
	for _, metricDef := range entityDef.MetricDefs {
		entityType := entityDef.EType
		for metricKind, metricQuery := range metricDef.Queries {
			metricType := metricDef.MType
			metricSeries, err := promClient.GetMetrics(metricQuery)
			if err != nil {
				glog.Errorf("Failed to query metric %v[%v] [%v] for entity type %v: %v.",
					metricType, metricKind, metricQuery, entityType, err)
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
					hostedOnVM := entityDef.HostedOnVM
					entityId := t.getEntityId(hostedOnVM, entityAttr)
					difEntity = data.NewDIFEntity(entityId, entityType).
						WithClusterId(t.getClusterId()).
						WithName(entityAttr.ID).
						WithNamespace(entityAttr.Namespace).
						WithAttributes(entityAttr.AsMap)
					// Setting matching attributes
					matchingAttr := t.getMatchingAttribute(hostedOnVM, entityAttr)
					if matchingAttr != "" {
						difEntity.Matching(matchingAttr)
					}
					if hostedOnVM {
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

func (t *Task) getEntityId(hostedOnVM bool, entityAttr *EntityAttribute) (entityId string) {
	entityId = entityAttr.ID
	// Use ID directly when app is hosted on VM
	if hostedOnVM {
		return
	}
	// For containerized applications, append clusterId to the entityId if possible to make it unique
	clusterId := t.getClusterId()
	if clusterId != "" {
		entityId = entityId + "-" + clusterId
	}
	return
}

func (t *Task) getClusterId() string {
	if t.clusterId != nil && t.clusterId.ID != "" {
		return t.clusterId.ID
	}
	return t.k8sSvcId
}

func (t *Task) getMatchingAttribute(hostedOnVM bool, entityAttr *EntityAttribute) (matchingAttr string) {
	if entityAttr.IP == "" {
		// No IP is found
		return
	}
	matchingAttr = entityAttr.IP
	if hostedOnVM {
		// Use IP directly when app is hosted on VM
		return
	}
	// For containerized applications, we must append kubernetes clusterId to the IP address
	// to avoid potential collision of IPs from different clusters during server side stitching
	if t.clusterId != nil && t.clusterId.ID != "" {
		// This is a multi-cluster configuration
		matchingAttr = matchingAttr + "-" + t.clusterId.ID
	} else if t.k8sSvcId != "" {
		// This is a single-cluster configuration
		matchingAttr = matchingAttr + "-" + t.k8sSvcId
	}
	return
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
	reconciledAttributes := map[string]string{}
	for name, def := range attributeDefs {
		values := []string{}
		for _, key := range def.LabelKeys {
			if v, exist := labels[key]; exist {
				values = append(values, v)
			} else {
				glog.V(4).Infof("Label key %s in attribute %s does not exist in %+v", key, name, labels)
			}
		}
		if len(values) < 1 {
			if def.IsIdentifier {
				return nil, fmt.Errorf("required identifer label %v does not exist", def.LabelKeys)
			}
			continue
		}
		value := strings.Join(values, def.LabelDelim)
		glog.V(4).Infof("Reconcile attribute %v with label: %q, value: %q", name, def.LabelKeys, value)
		matchIndex := def.ValueMatches.FindStringSubmatchIndex(value)
		if matchIndex == nil {
			return nil, fmt.Errorf("%v label's value %q did not match expected pattern %q. "+
				"Please consider changing the pattern in promql."+
				"This may also happen if label key or it's expected value is longer than 63 chars. "+
				"Kubernetes character limit for label key and value is 63",
				def.LabelKeys, value, def.ValueMatches.String())
		}
		expandedValue := def.ValueMatches.ExpandString(nil, def.ValueAs, value, matchIndex)
		reconciledAttributes[name] = string(expandedValue)
		if def.IsIdentifier {
			id = reconciledAttributes[name]
			if id == "" {
				return nil, fmt.Errorf("empty identifier from label key %v and value %q", def.LabelKeys, value)
			}
		}
	}
	glog.V(4).Infof("Reconciled attributes: %s", spew.Sdump(reconciledAttributes))
	namespace := reconciledAttributes["namespace"]
	entityAttr := &EntityAttribute{
		ID:        id,
		IP:        reconciledAttributes["ip"],
		Service:   reconciledAttributes["service"],
		Namespace: namespace,
		AsMap:     reconciledAttributes,
	}
	return entityAttr, nil
}
