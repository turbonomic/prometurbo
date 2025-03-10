package customresource

import (
	"fmt"
	"regexp"
	"strings"

	"github.ibm.com/turbonomic/turbo-metrics/api/v1alpha1"

	"github.ibm.com/turbonomic/prometurbo/pkg/provider"
)

func attributesFromCustomResource(
	attributeConfigs []v1alpha1.AttributeConfiguration,
) (map[string]*provider.AttributeValueDef, error) {
	attributes := map[string]*provider.AttributeValueDef{}
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
	labels := []string{}
	if len(attributeConfig.Labels) > 0 {
		labels = attributeConfig.Labels
	} else if attributeConfig.Label != "" {
		labels = append(labels, attributeConfig.Label)
	} else {
		labels = append(labels, attributeConfig.Name)
	}
	var err error
	var valueMatches *regexp.Regexp
	if attributeConfig.Matches != "" {
		if valueMatches, err = regexp.Compile(attributeConfig.Matches); err != nil {
			return nil, fmt.Errorf("failed to compile match expression %q for label %v: %v",
				attributeConfig.Matches, labels, err)
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
			return nil, fmt.Errorf("missing 'as' for the value matcher %q for label %v",
				attributeConfig.Matches, labels)
		}
	}
	delim := "-"
	if attributeConfig.Delimeter != "" {
		delim = attributeConfig.Delimeter
	}
	return &provider.AttributeValueDef{
		LabelKeys:    labels,
		LabelDelim:   delim,
		ValueMatches: valueMatches,
		ValueAs:      valueAs,
		IsIdentifier: attributeConfig.IsIdentifier,
	}, nil
}
