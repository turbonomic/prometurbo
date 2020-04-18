package provider

import (
	"fmt"
	"github.com/turbonomic/prometurbo/pkg/config"
	"github.com/turbonomic/prometurbo/pkg/prometheus"
)

type serverDef struct {
	promClient *prometheus.RestClient
	username   string
	password   string
	exporters  []string
}

func newServerDef(serverConfig config.ServerConfig) (*serverDef, error) {
	if len(serverConfig.URL) == 0 {
		return nil, fmt.Errorf("no url defined")
	}
	if len(serverConfig.Exporters) == 0 {
		return nil, fmt.Errorf("missing exporters")
	}
	promClient, err := prometheus.NewRestClient(serverConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client from %v: %v",
			serverConfig.URL, err)
	}
	promClient.SetUser(serverConfig.Username, serverConfig.Password)
	return &serverDef{
		promClient: promClient,
		exporters:  serverConfig.Exporters,
	}, nil
}

func ServersFromConfig(cfg *config.MetricsDiscoveryConfig) (map[string]*serverDef, error) {
	servers := make(map[string]*serverDef)
	for name, serverConfig := range cfg.ServerConfigs {
		server, err := newServerDef(serverConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create serverDef for %v: %v",
				name, err)
		}
		servers[name] = server
	}
	return servers, nil
}
