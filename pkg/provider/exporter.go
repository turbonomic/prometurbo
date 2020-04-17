package provider

import (
	"fmt"

	"github.com/turbonomic/prometurbo/pkg/config"
)

type exporterDef struct {
	name       string
	entityDefs []*entityDef
}

func newExporterDef(exporterConfig config.ExporterConfig) (*exporterDef, error) {
	if exporterConfig.Name == "" {
		return nil, fmt.Errorf("empty exporterDef name")
	}
	if len(exporterConfig.EntityConfigs) == 0 {
		return nil, fmt.Errorf("no entityDefs defined for exporterDef %v", exporterConfig.Name)
	}
	var entities []*entityDef
	for _, entityConfig := range exporterConfig.EntityConfigs {
		entity, err := newEntityDef(entityConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create entityDefs: %v", err)
		}
		entities = append(entities, entity)
	}
	return &exporterDef{
		name:       exporterConfig.Name,
		entityDefs: entities,
	}, nil
}

func ExportersFromConfig(cfg *config.MetricsDiscoveryConfig) ([]*exporterDef, error) {
	var exporters []*exporterDef
	for _, exporterConfig := range cfg.ExporterConfigs {
		exporter, err := newExporterDef(exporterConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporterDefs for %v: %v",
				exporterConfig.Name, err)
		}
		exporters = append(exporters, exporter)
	}
	return exporters, nil
}
