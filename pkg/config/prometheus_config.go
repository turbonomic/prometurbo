package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type MetricsDiscoveryConfig struct {
	ServerConfigs   []ServerConfig   `yaml:"servers"`
	ExporterConfigs []ExporterConfig `yaml:"exporters"`
}

type ServerConfig struct {
	Name     string `yaml:"name,omitempty"`
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ExporterConfig struct {
	Name          string         `yaml:"name"`
	EntityConfigs []EntityConfig `yaml:"entities"`
}

type EntityConfig struct {
	Type             string                  `yaml:"type"`
	HostedOnVM       bool                    `yaml:"hostedOnVM,omitempty"`
	MetricConfigs    []MetricConfig          `yaml:"metrics"`
	AttributeConfigs map[string]ValueMapping `yaml:"attributes"`
}

type MetricConfig struct {
	Type    string            `yaml:"type"`
	Queries map[string]string `yaml:"queries"`
}

type ValueMapping struct {
	Label        string `yaml:"label"`
	Matches      string `yaml:"matches,omitempty"`
	As           string `yaml:"as,omitempty"`
	IsIdentifier bool   `yaml:"isIdentifier"`
}

// NewMetricsDiscoveryConfig loads the configuration from a yaml file.
func NewMetricsDiscoveryConfig(filename string) (*MetricsDiscoveryConfig, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to load metrics discovery config file: %v", err)
	}
	var cfg MetricsDiscoveryConfig
	if err := yaml.UnmarshalStrict(contents, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse metrics discovery config: %v", err)
	}
	// Parse prometheus servers
	if len(cfg.ServerConfigs) < 1 {
		return nil, fmt.Errorf("missing prometheus servers")
	}
	return &cfg, nil
}
