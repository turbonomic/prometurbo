package main

import (
	"flag"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-metrics/api/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/turbonomic/prometurbo/pkg/config"
	"github.com/turbonomic/prometurbo/pkg/provider"
	"github.com/turbonomic/prometurbo/pkg/provider/configmap"
	"github.com/turbonomic/prometurbo/pkg/provider/customresource"
	"github.com/turbonomic/prometurbo/pkg/server"
	"github.com/turbonomic/prometurbo/pkg/topology"
	"github.com/turbonomic/prometurbo/pkg/worker"
)

const (
	defaultPort                 = 8081
	defaultPrometheusConfigPath = "/etc/prometurbo/prometheus.config"
	defaultTopologyConfigPath   = "/etc/prometurbo/businessapp.config"
	defaultWorkerCount          = 4
)

var (
	port                     int
	workerCount              int
	prometheusConfigFileName string
	topologyConfigFileName   string
	// custom resource scheme for controller runtime client
	customScheme = runtime.NewScheme()
)

func parseFlags() {
	flag.IntVar(&port, "port", defaultPort, "port to expose metrics (default 8081)")
	flag.StringVar(&prometheusConfigFileName, "prometheusConfig",
		defaultPrometheusConfigPath, "path to the metrics discovery config file")
	flag.StringVar(&topologyConfigFileName, "topologyConfig",
		defaultTopologyConfigPath, "path to the topology config file")
	flag.IntVar(&workerCount, "workerCount", defaultWorkerCount, "the number of concurrent workers to"+
		"discover metrics")
	flag.Parse()
}

func init() {
	utilruntime.Must(v1.AddToScheme(customScheme))
	// Add registered custom types to the custom scheme
	utilruntime.Must(v1alpha1.AddToScheme(customScheme))
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
}

func main() {
	// Ignore errors
	_ = flag.Set("logtostderr", "false")
	_ = flag.Set("alsologtostderr", "true")
	_ = flag.Set("log_dir", "/var/log")
	defer glog.Flush()

	// Parse command line flags
	parseFlags()

	glog.Infof("Running prometurbo GIT_COMMIT: %s", os.Getenv("GIT_COMMIT"))

	if workerCount < 1 {
		glog.Warningf("The specified number of concurrent workers %v is invalid. Set it 1.", workerCount)
		workerCount = 1
	} else {
		glog.V(2).Infof("Number of concurrent workers to discover metrics: %v", workerCount)
	}

	server.NewServer(port).
		MetricProvider(getMetricProvider()).
		Topology(topology.NewBusinessTopology(getBizAppsConfig())).
		Dispatcher(worker.NewDispatcher(workerCount).
			WithCollector(worker.NewCollector(workerCount * 2))).
		Run()

	return
}

func getMetricProvider() (metricProvider provider.MetricProvider) {
	if configmap.HasMetricProvider(prometheusConfigFileName) {
		// This is ONLY used for turbo-on-turbo use case, where we still use configMap to define
		// prometheus query mappings.
		configMapMetricProvider, err := configmap.GetMetricProvider(prometheusConfigFileName)
		if err != nil {
			glog.Fatalf("Failed to get metric provider from configMap: %v.", err)
		}
		metricProvider = configMapMetricProvider
	} else {
		kubeClient := createKubeClientOrDie()
		customResourceMetricProvider, err := customresource.GetMetricProvider(kubeClient)
		if err != nil {
			glog.Fatalf("Failed to get metric provider from custom resource: %v.", err)
		}
		metricProvider = customResourceMetricProvider
	}
	return
}

func createKubeClientOrDie() client.Client {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Fatal error: failed to get in-cluster config: %v.", err)
	}
	// This specifies the number and the max number of query per second to the api server.
	kubeConfig.QPS = 20.0
	kubeConfig.Burst = 30
	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: customScheme})
	if err != nil {
		glog.Fatalf("Failed to create controller runtime client: %v.", err)
	}
	return kubeClient
}

func getBizAppsConfig() []config.BusinessApplication {
	bizApps, err := config.NewBusinessApplicationConfigMap(topologyConfigFileName)
	if err != nil {
		glog.Errorf("Failed to parse topology configuration from %v: %v.",
			topologyConfigFileName, err)
	}
	glog.V(2).Infof("Business application topology configuration: %s",
		spew.Sdump(bizApps))
	return bizApps
}
