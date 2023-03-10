package configmap

import (
	"fmt"

	"github.com/turbonomic/prometurbo/pkg/config"
	"github.com/turbonomic/prometurbo/pkg/provider"
)

type exporterDef struct {
	entityDefs []*provider.EntityDef
}

func exporterDefFromConfigMap(exporterConfig config.ExporterConfig) (*exporterDef, error) {
	if len(exporterConfig.EntityConfigs) == 0 {
		return nil, fmt.Errorf("no entityDefs defined")
	}
	var entities []*provider.EntityDef
	for _, entityConfig := range exporterConfig.EntityConfigs {
		entity, err := entityDefFromConfigMap(entityConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create entityDefs: %v", err)
		}
		entities = append(entities, entity)
	}
	return &exporterDef{
		entityDefs: entities,
	}, nil
}

func exportersFromConfigMap(cfg *config.MetricsDiscoveryConfig) (map[string]*exporterDef, error) {
	exporters := make(map[string]*exporterDef)
	for name, exporterConfig := range cfg.ExporterConfigs {
		exporter, err := exporterDefFromConfigMap(exporterConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporterDef for %v: %v",
				name, err)
		}
		exporters[name] = exporter
	}
	return exporters, nil
}
