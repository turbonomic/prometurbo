package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/turbonomic/prometurbo/appmetric/pkg/config"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	Used     = "used"
	Capacity = "capacity"

	TargetAddress = "target_address"
	Scope         = "scope"
)

type metricDef struct {
	mType   proto.CommodityDTO_CommodityType
	queries map[string]string
}

type attributeValueDef struct {
	labelKey     string
	valueMatches *regexp.Regexp
	valueAs      string
	isIdentifier bool
}

func newMetricDef(metricConfig config.MetricConfig) (*metricDef, error) {
	mType, ok := proto.CommodityDTO_CommodityType_value[strings.ToUpper(metricConfig.Type)]
	if !ok {
		return nil, fmt.Errorf("unsupported metric type %q", metricConfig.Type)
	}
	if len(metricConfig.Queries) == 0 {
		return nil, fmt.Errorf("empty queries")
	}
	if _, exist := metricConfig.Queries[Used]; !exist {
		return nil, fmt.Errorf("missing query for used value")
	}
	metricDef := metricDef{
		mType:   proto.CommodityDTO_CommodityType(mType),
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
