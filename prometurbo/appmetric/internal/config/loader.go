package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// FromYAML loads the configuration from a yaml file.
func FromYAML(filename string) (*MetricsDiscoveryConfig, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to load metrics discovery config file: %v", err)
	}
	var cfg MetricsDiscoveryConfig
	if err := yaml.UnmarshalStrict(contents, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse metrics discovery config: %v", err)
	}
	return &cfg, nil
}
