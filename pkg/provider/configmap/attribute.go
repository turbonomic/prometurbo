package configmap

import (
	"fmt"
	"regexp"
	"strings"

	"github.ibm.com/turbonomic/prometurbo/pkg/config"
	"github.ibm.com/turbonomic/prometurbo/pkg/provider"
)

func attributesFromConfigMap(attributeConfigs map[string]config.ValueMapping) (map[string]*provider.AttributeValueDef, error) {
	attributes := map[string]*provider.AttributeValueDef{}
	var identifier []string
	for name, valueMapping := range attributeConfigs {
		value, err := attributeFromConfigMap(name, valueMapping)
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

// This function creates the attributes from the configMap
// Since we allow that only the "name" is specified for the attribute in the case where the attribute name is
// the same as the label, therefore we pass the name also to this function as a fallback for this case
// example: we can ommit "label" and use only "name" for this attribute entry
//   - name: pod
//     label: pod
func attributeFromConfigMap(name string, valueMapping config.ValueMapping) (*provider.AttributeValueDef, error) {
	labels := []string{}
	if len(valueMapping.Labels) > 0 {
		labels = valueMapping.Labels
	} else if valueMapping.Label != "" {
		labels = append(labels, valueMapping.Label)
	} else {
		labels = append(labels, name)
	}
	var err error
	var valueMatches *regexp.Regexp
	if valueMapping.Matches != "" {
		if valueMatches, err = regexp.Compile(valueMapping.Matches); err != nil {
			return nil, fmt.Errorf("failed to compile match expression %q for label %v: %v",
				valueMapping.Matches, labels, err)
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
			return nil, fmt.Errorf("missing 'as' for the value matcher %q for label %v",
				valueMapping.Matches, labels)
		}
	}
	delim := "-"
	if valueMapping.Delimeter != "" {
		delim = valueMapping.Delimeter
	}
	return &provider.AttributeValueDef{
		LabelKeys:    labels,
		LabelDelim:   delim,
		ValueMatches: valueMatches,
		ValueAs:      valueAs,
		IsIdentifier: valueMapping.IsIdentifier,
	}, nil
}
