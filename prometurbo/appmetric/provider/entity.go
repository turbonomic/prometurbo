package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/appmetric/internal/config"
	"github.com/turbonomic/prometurbo/prometurbo/appmetric/metrics"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type metricDef struct {
	queries map[string]string
}

type attributeValueDef struct {
	labelKey     string
	valueMatches *regexp.Regexp
	valueAs      string
	isIdentifier bool
}

type entityDef struct {
	eType         proto.EntityDTO_EntityType
	attributeDefs map[string]*attributeValueDef
	metricDefs    map[proto.CommodityDTO_CommodityType]*metricDef
}

func newmetricDef(metricConfig config.MetricConfig) (*metricDef, error) {
	if len(metricConfig.Queries) == 0 {
		return nil, fmt.Errorf("empty queries")
	}
	if _, exist := metricConfig.Queries[metrics.Used]; !exist {
		return nil, fmt.Errorf("missing query for used value")
	}
	metricDef := metricDef{
		queries: make(map[string]string),
	}
	for k, v := range metricConfig.Queries {
		metricDef.queries[k] = v
	}
	return &metricDef, nil
}

func newAttributeValue(valueMapping config.ValueMapping) (*attributeValueDef, error) {
	if valueMapping.Label == "" {
		return nil, fmt.Errorf("empty label")
	}
	var err error
	var valueMatches *regexp.Regexp
	if valueMapping.Matches != "" {
		if valueMatches, err = regexp.Compile(valueMapping.Matches); err != nil {
			return nil, fmt.Errorf("failed to compile match expression %q for label %q: %v",
				valueMapping.Matches, valueMapping.Label, err)
		}
	} else {
		// this will always succeed
		valueMatches = regexp.MustCompile(".*")
	}
	valueAs := valueMapping.As
	if valueAs == "" {
		subexpNames := valueMatches.SubexpNames()
		if len(subexpNames) == 1 {
			// use the whole string as the capture group
			valueAs = "$0"
		} else if len(subexpNames) == 2 {
			// one capture group
			valueAs = "$1"
		} else {
			return nil, fmt.Errorf("missing 'as' for the value matcher %q for label %q",
				valueMapping.Matches, valueMapping.Label)
		}
	}
	return &attributeValueDef{
		labelKey:     valueMapping.Label,
		valueMatches: valueMatches,
		valueAs:      valueAs,
		isIdentifier: valueMapping.IsIdentifier,
	}, nil
}

func newAttributes(attributeConfigs map[string]config.ValueMapping) (map[string]*attributeValueDef, error) {
	var attributes = map[string]*attributeValueDef{}
	var identifier []string
	for name, valueMapping := range attributeConfigs {
		value, err := newAttributeValue(valueMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create attribute %q: %v", name, err)
		}
		if value.isIdentifier {
			identifier = append(identifier, name)
			if len(identifier) > 1 {
				return nil, fmt.Errorf("failed to create attribute %q: duplicated identifiers: [%v]",
					name, strings.Join(identifier, ","))
			}
		}
		attributes[name] = value
	}
	if len(identifier) < 1 {
		return nil, fmt.Errorf("missing identifier")
	}
	return attributes, nil
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
	var metrics = map[proto.CommodityDTO_CommodityType]*metricDef{}
	for metricType, metricConfig := range entityConfig.MetricConfigs {
		mType, ok := proto.CommodityDTO_CommodityType_value[strings.ToUpper(metricType)]
		if !ok {
			return nil, fmt.Errorf("unsupported metric type %q", metricType)
		}
		metric, err := newmetricDef(metricConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create metricDefs for %v [%v]: %v",
				entityConfig.Type, metricType, err)
		}
		metrics[proto.CommodityDTO_CommodityType(mType)] = metric
	}
	attributes, err := newAttributes(entityConfig.AttributeConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to create attributeDefs for entityDef type %v: %v", entityConfig.Type, err)
	}
	return &entityDef{
		eType:         proto.EntityDTO_EntityType(eType),
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
