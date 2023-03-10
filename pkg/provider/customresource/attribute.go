package customresource

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.com/turbonomic/prometurbo/pkg/provider"
)

func attributesFromCustomResource(
	attributeConfigs []v1alpha1.AttributeConfiguration) (map[string]*provider.AttributeValueDef, error) {
	var attributes = map[string]*provider.AttributeValueDef{}
	var identifier []string
	for _, attributeConfig := range attributeConfigs {
		value, err := attributeFromCustomResource(attributeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create attribute %q: %v", attributeConfig.Name, err)
		}
		if value.IsIdentifier {
			identifier = append(identifier, attributeConfig.Name)
			if len(identifier) > 1 {
				return nil, fmt.Errorf("failed to create attribute %q: duplicated identifiers: [%v]",
					attributeConfig.Name, strings.Join(identifier, ","))
			}
		}
		attributes[attributeConfig.Name] = value
	}
	if len(identifier) < 1 {
		return nil, fmt.Errorf("missing identifier")
	}
	return attributes, nil
}

func attributeFromCustomResource(attributeConfig v1alpha1.AttributeConfiguration) (*provider.AttributeValueDef, error) {
	if attributeConfig.Label == "" {
		return nil, fmt.Errorf("empty label")
	}
	var err error
	var valueMatches *regexp.Regexp
	if attributeConfig.Matches != "" {
		if valueMatches, err = regexp.Compile(attributeConfig.Matches); err != nil {
			return nil, fmt.Errorf("failed to compile match expression %q for label %q: %v",
				attributeConfig.Matches, attributeConfig.Label, err)
		}
	} else {
		// this will always succeed
		valueMatches = regexp.MustCompile(".*")
	}
	valueAs := attributeConfig.As
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
				attributeConfig.Matches, attributeConfig.Label)
		}
	}
	return &provider.AttributeValueDef{
		LabelKey:     attributeConfig.Label,
		ValueMatches: valueMatches,
		ValueAs:      valueAs,
		IsIdentifier: attributeConfig.IsIdentifier,
	}, nil
}
