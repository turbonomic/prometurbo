package configmap

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/turbonomic/prometurbo/pkg/config"
	"github.com/turbonomic/prometurbo/pkg/provider"
)

func attributesFromConfigMap(attributeConfigs map[string]config.ValueMapping) (map[string]*provider.AttributeValueDef, error) {
	var attributes = map[string]*provider.AttributeValueDef{}
	var identifier []string
	for name, valueMapping := range attributeConfigs {
		value, err := attributeFromConfigMap(valueMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create attribute %q: %v", name, err)
		}
		if value.IsIdentifier {
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

func attributeFromConfigMap(valueMapping config.ValueMapping) (*provider.AttributeValueDef, error) {
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
	return &provider.AttributeValueDef{
		LabelKey:     valueMapping.Label,
		ValueMatches: valueMatches,
		ValueAs:      valueAs,
		IsIdentifier: valueMapping.IsIdentifier,
	}, nil
}
