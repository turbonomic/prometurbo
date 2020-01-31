package main

import (
	"flag"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/config"
	"github.com/turbonomic/prometurbo/appmetric/pkg/provider"

	"github.com/turbonomic/prometurbo/appmetric/pkg/server"
)

const (
	defaultPort           = 8081
	defaultSampleDuration = "3m"
)

var (
	port           int
	configFileName string
	sampleDuration string
)

func parseFlags() {
	flag.IntVar(&port, "port", defaultPort, "port to expose metrics (default 8081)")
	flag.StringVar(&configFileName, "config", "", "path to the metrics discovery config file")
	flag.StringVar(&sampleDuration, "sampleDuration", defaultSampleDuration, "the sample duration for prometheus query")
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
	if len(configFileName) < 1 {
		glog.Fatal("Failed to get metric discovery configuration.")
	}
	metricConf, err := config.FromYAML(configFileName)
	if err != nil {
		glog.Fatalf("Failed to load metric discovery configuration %s: %v.", configFileName, err)
	}
	glog.V(2).Infof("%s", spew.Sdump(metricConf))

	// Construct exporter provider from configuration
	promExporters, err := provider.ExportersFromConfig(metricConf)
	if err != nil {
		glog.Fatalf("Failed to construct exporters from configuration %s: %v.", configFileName, err)
	}

	// Start metric server to serve entity metrics queries
	s := server.NewMetricServer(port, provider.NewProviderFactory(promExporters))
	s.Run()
	return
}
