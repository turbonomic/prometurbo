package configmap

import (
	"fmt"

	"github.ibm.com/turbonomic/prometurbo/pkg/config"
	"github.ibm.com/turbonomic/prometurbo/pkg/prometheus"
)

type serverDef struct {
	promClient *prometheus.RestClient
	username   string
	password   string
	clusterId  string
	exporters  []string
}

func serverDefFromConfigMap(serverConfig config.ServerConfig) (*serverDef, error) {
	if len(serverConfig.URL) == 0 {
		return nil, fmt.Errorf("no url defined")
	}
	if len(serverConfig.Exporters) == 0 {
		return nil, fmt.Errorf("missing exporters")
	}
	promClient, err := prometheus.NewRestClient(serverConfig.URL, serverConfig.BearerToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client from %v: %v",
			serverConfig.URL, err)
	}
	promClient.SetUser(serverConfig.Username, serverConfig.Password)
	return &serverDef{
		promClient: promClient,
		clusterId:  serverConfig.ClusterId,
		exporters:  serverConfig.Exporters,
	}, nil
}

func serversFromConfigMap(cfg *config.MetricsDiscoveryConfig) (map[string]*serverDef, error) {
	servers := make(map[string]*serverDef)
	for name, serverConfig := range cfg.ServerConfigs {
		server, err := serverDefFromConfigMap(serverConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create serverDef for %v: %v",
				name, err)
		}
		servers[name] = server
	}
	return servers, nil
}
