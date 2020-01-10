package config

type MetricsDiscoveryConfig struct {
	ExporterConfigs []ExporterConfig `yaml:"exporters"`
}

type ExporterConfig struct {
	Name          string         `yaml:"name"`
	EntityConfigs []EntityConfig `yaml:"entities"`
}

type EntityConfig struct {
	Type             string                  `yaml:"type"`
	MetricConfigs    []MetricConfig          `yaml:"metrics"`
	AttributeConfigs map[string]ValueMapping `yaml:"attributes"`
}

type MetricConfig struct {
	Type  string `yaml:"type"`
	Query string `yaml:"query"`
}

type ValueMapping struct {
	Label        string `yaml:"label"`
	Matches      string `yaml:"matches,omitempty"`
	As           string `yaml:"as,omitempty"`
	IsIdentifier bool   `yaml:"isIdentifier"`
}
