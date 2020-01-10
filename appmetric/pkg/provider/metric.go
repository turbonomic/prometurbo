package provider

import (
	"fmt"
	"github.com/turbonomic/prometurbo/appmetric/pkg/config"
	"github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"regexp"
	"strings"
)

type metricDef struct {
	mType proto.CommodityDTO_CommodityType
	query prometheus.Request
}

type attributeValueDef struct {
	labelKey     string
	valueMatches *regexp.Regexp
	valueAs      string
	isIdentifier bool
}

func newMetricDef(metricConfig config.MetricConfig) (*metricDef, error) {
	if metricConfig.Type == "" {
		return nil, fmt.Errorf("empty metricDef type")
	}
	mType, ok := proto.CommodityDTO_CommodityType_value[strings.ToUpper(metricConfig.Type)]
	if !ok {
		return nil, fmt.Errorf("unsupported metricDef type %q", metricConfig.Type)
	}
	if metricConfig.Query == "" {
		return nil, fmt.Errorf("empty query for metricDef type %q", metricConfig.Type)
	}
	return &metricDef{
		mType: proto.CommodityDTO_CommodityType(mType),
		query: prometheus.NewBasicRequest().SetQuery(metricConfig.Query),
	}, nil
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
	var identifier = map[string]bool{}
	for name, valueMapping := range attributeConfigs {
		value, err := newAttributeValue(valueMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create attribute %q: %v", name, err)
		}
		if value.isIdentifier {
			identifier[name] = true
			if len(identifier) > 1 {
				var l []string
				for n := range identifier {
					l = append(l, n)
				}
				return nil, fmt.Errorf("failed to create attribute %q: duplicated identifiers %v",
					name, strings.Join(l, " "))
			}
		}
		attributes[name] = value
	}
	if len(identifier) < 1 {
		return nil, fmt.Errorf("missing identifier")
	}
	return attributes, nil
}
