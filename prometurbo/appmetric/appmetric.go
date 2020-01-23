package appmetric

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/appmetric/internal/config"
	"github.com/turbonomic/prometurbo/prometurbo/appmetric/provider"

	"github.com/turbonomic/prometurbo/prometurbo/appmetric/internal/prometheus"
)

type Args struct {
	PrometheusHosts arrayFlags
	ConfigFileName  string
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Provider will construct a new MetricFetcher for retrieving Prometheus metrics from remote
// Prometheus servers.
func Provider(args Args) *provider.Provider {

	// Config pretty print for debugging
	spew.Config = spew.ConfigState{
		Indent:                  "  ",
		MaxDepth:                0,
		DisableMethods:          true,
		DisablePointerMethods:   true,
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		ContinueOnMethod:        false,
		SortKeys:                true,
		SpewKeys:                false,
	}

	glog.Info("Starting Prometurbo...")
	glog.Infof("GIT_COMMIT: %s", os.Getenv("GIT_COMMIT"))

	// Parse prometheus servers
	if len(args.PrometheusHosts) < 1 {
		glog.Fatal("Failed to get prometheus server address.")
	}
	var promClients []*prometheus.RestClient
	for _, promHost := range args.PrometheusHosts {
		promClient, err := prometheus.NewRestClient(promHost)
		if err != nil {
			glog.Fatalf("Failed to create prometheus client: %v.", err)
		}
		targets, err := promClient.Validate()
		if err != nil {
			glog.Errorf("Failed to validate prometheus server %q: %v.", promHost, err)
		}
		glog.V(4).Infof("Targets from prometheus server %q: %v.", promHost, targets)
		promClients = append(promClients, promClient)
	}

	// Load metric discovery configuration
	if len(args.ConfigFileName) < 1 {
		glog.Fatal("Failed to get metric discovery configuration.")
	}
	metricConf, err := config.FromYAML(args.ConfigFileName)
	if err != nil {
		glog.Fatalf("Failed to load metric discovery configuration %s: %v.", args.ConfigFileName, err)
	}
	glog.V(2).Infof("%s", spew.Sdump(metricConf))

	// Construct exporter provider from configuration
	promExporters, err := provider.ExportersFromConfig(metricConf)
	if err != nil {
		glog.Fatalf("Failed to construct exporters from configuration %s: %v.", args.ConfigFileName, err)
	}

	// Start metric server to serve entity metrics queries
	return provider.NewProvider(promClients, promExporters)
}
