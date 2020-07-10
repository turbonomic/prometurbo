package main

import (
	"flag"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/config"
	"github.com/turbonomic/prometurbo/pkg/provider"
	"github.com/turbonomic/prometurbo/pkg/server"
	"github.com/turbonomic/prometurbo/pkg/topology"
)

const (
	defaultPort                 = 8081
	defaultPrometheusConfigPath = "/etc/prometurbo/prometheus.config"
	defaultTopologyConfigPath   = "/etc/prometurbo/businessapp.config"
)

var (
	port                     int
	prometheusConfigFileName string
	topologyConfigFileName   string
)

func parseFlags() {
	flag.IntVar(&port, "port", defaultPort, "port to expose metrics (default 8081)")
	flag.StringVar(&prometheusConfigFileName, "prometheusConfig",
		defaultPrometheusConfigPath, "path to the metrics discovery config file")
	flag.StringVar(&topologyConfigFileName, "topologyConfig",
		defaultTopologyConfigPath, "path to the topology config file")
	flag.Parse()
}

func main() {
	// Ignore errors
	_ = flag.Set("logtostderr", "false")
	_ = flag.Set("alsologtostderr", "true")
	_ = flag.Set("log_dir", "/var/log")
	defer glog.Flush()

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

	// Parse command line flags
	parseFlags()

	glog.Info("Starting Prometurbo...")
	glog.Infof("GIT_COMMIT: %s", os.Getenv("GIT_COMMIT"))

	// Load metric discovery configuration
	if len(prometheusConfigFileName) < 1 {
		glog.Fatal("Failed to get metric discovery configuration.")
	}
	metricConf, err := config.NewMetricsDiscoveryConfig(prometheusConfigFileName)
	if err != nil {
		glog.Fatalf("Failed to create metric discovery configuration from %s: %v.",
			prometheusConfigFileName, err)
	}
	glog.V(2).Infof("%s", spew.Sdump(metricConf))
	// Construct prometheus servers from configuration
	promServers, err := provider.ServersFromConfig(metricConf)
	if err != nil {
		glog.Fatalf("Failed to construct servers from configuration %s: %v.",
			prometheusConfigFileName, err)
	}
	// Construct exporter provider from configuration
	promExporters, err := provider.ExportersFromConfig(metricConf)
	if err != nil {
		glog.Fatalf("Failed to construct exporters from configuration %s: %v.",
			prometheusConfigFileName, err)
	}
	bizApps, err := config.NewBusinessApplicationConfig(topologyConfigFileName)
	if err != nil {
		glog.Fatalf("Failed to parse topology configuration from %v: %v.",
			topologyConfigFileName, err)
	}
	glog.V(2).Infof("Business application topology configuration: %s",
		spew.Sdump(bizApps))
	server.NewServer(port).
		MetricProvider(provider.NewProvider(promServers, promExporters)).
		Topology(topology.NewBusinessTopology(bizApps)).
		Run()

	return
}
