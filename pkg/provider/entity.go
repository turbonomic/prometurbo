package provider

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/config"
	"github.com/turbonomic/prometurbo/pkg/util"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

type entityDef struct {
	eType         string
	hostedOnVM    bool
	attributeDefs map[string]*attributeValueDef
	metricDefs    []*metricDef
}

type entityAttribute struct {
	id        string
	ip        string
	service   string
	namespace string
}

func newEntityDef(entityConfig config.EntityConfig) (*entityDef, error) {
	if entityConfig.Type == "" {
		return nil, fmt.Errorf("empty entityDef type")
	}
	if !data.IsValidDIFEntity(entityConfig.Type) {
		return nil, fmt.Errorf("unsupported entityDef type %v", entityConfig.Type)
	}
	if len(entityConfig.MetricConfigs) == 0 {
		return nil, fmt.Errorf("empty metricDef configuration for entityDef type %v", entityConfig.Type)
	}
	var metrics []*metricDef
	for _, metricConfig := range entityConfig.MetricConfigs {
		metric, err := newMetricDef(metricConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create metricDefs for %v [%v]: %v",
				entityConfig.Type, metricConfig.Type, err)
		}
		metrics = append(metrics, metric)
	}
	attributes, err := newAttributes(entityConfig.AttributeConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to create attributeDefs for entityDef type %v: %v", entityConfig.Type, err)
	}
	return &entityDef{
		eType:         entityConfig.Type,
		hostedOnVM:    entityConfig.HostedOnVM,
		metricDefs:    metrics,
		attributeDefs: attributes,
	}, nil
}

func (e *entityDef) reconcileAttributes(labels map[string]string) (*entityAttribute, error) {
	var id string
	var reconciledAttributes = map[string]string{}
	for name, def := range e.attributeDefs {
		key := def.labelKey
		value, exist := labels[key]
		if !exist {
			if def.isIdentifier {
				return nil, fmt.Errorf("required identifer label key %q does not exist", key)
			}
			continue
		}
		glog.V(4).Infof("Reconcile attribute %v with label: %q, value: %q", name, key, value)
		matchIndex := def.valueMatches.FindStringSubmatchIndex(value)
		if matchIndex == nil {
			return nil, fmt.Errorf("label %q's value %q did not match expected pattern %q",
				key, value, def.valueMatches.String())
		}
		expandedValue := def.valueMatches.ExpandString(nil, def.valueAs, value, matchIndex)
		reconciledAttributes[name] = string(expandedValue)
		if def.isIdentifier {
			id = reconciledAttributes[name]
			if id == "" {
				return nil, fmt.Errorf("empty identifier from label key %q and value %q", key, value)
			}
		}
	}
	glog.V(4).Infof("Reconciled attributes: %s", spew.Sdump(reconciledAttributes))
	namespace := reconciledAttributes["namespace"]
	entityAttr := &entityAttribute{
		id:        util.GetName(id, namespace),
		ip:        reconciledAttributes["ip"],
		service:   reconciledAttributes["service"],
		namespace: namespace,
	}
	return entityAttr, nil
}
