package provider

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/config"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"strings"
)

type entityDef struct {
	eType         proto.EntityDTO_EntityType
	hostedOnVM    bool
	attributeDefs map[string]*attributeValueDef
	metricDefs    []*metricDef
}

func newEntityDef(entityConfig config.EntityConfig) (*entityDef, error) {
	if entityConfig.Type == "" {
		return nil, fmt.Errorf("empty entityDef type")
	}
	eType, ok := proto.EntityDTO_EntityType_value[strings.ToUpper(entityConfig.Type)]
	if !ok {
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
		eType:         proto.EntityDTO_EntityType(eType),
		hostedOnVM:    entityConfig.HostedOnVM,
		metricDefs:    metrics,
		attributeDefs: attributes,
	}, nil
}

func (e *entityDef) reconcileAttributes(labels map[string]string) (string, map[string]string, error) {
	var id string
	var reconciledAttributes = map[string]string{}
	for name, def := range e.attributeDefs {
		key := def.labelKey
		value, exist := labels[key]
		if !exist {
			if def.isIdentifier {
				return "", reconciledAttributes, fmt.Errorf("required identifer label key %q does not exist", key)
			}
			continue
		}
		glog.V(4).Infof("Reconcile label with key: %q, value: %q", key, value)
		matchIndex := def.valueMatches.FindStringSubmatchIndex(value)
		if matchIndex == nil {
			return "", reconciledAttributes, fmt.Errorf("label %q's value %q did not match expected pattern %q",
				key, value, def.valueMatches.String())
		}
		expandedValue := def.valueMatches.ExpandString(nil, def.valueAs, value, matchIndex)
		reconciledAttributes[name] = string(expandedValue)
		if def.isIdentifier {
			id = reconciledAttributes[name]
		}
	}
	glog.V(4).Infof("Reconciled attributes: %s", spew.Sdump(reconciledAttributes))
	return id, reconciledAttributes, nil
}
