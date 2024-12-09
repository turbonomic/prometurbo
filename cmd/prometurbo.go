package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.ibm.com/turbonomic/turbo-metrics/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spf13/viper"
	"github.ibm.com/turbonomic/prometurbo/pkg/config"
	"github.ibm.com/turbonomic/prometurbo/pkg/provider"
	"github.ibm.com/turbonomic/prometurbo/pkg/provider/configmap"
	"github.ibm.com/turbonomic/prometurbo/pkg/provider/customresource"
	"github.ibm.com/turbonomic/prometurbo/pkg/server"
	"github.ibm.com/turbonomic/prometurbo/pkg/topology"
	"github.ibm.com/turbonomic/prometurbo/pkg/worker"
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

	// Watch the configmap and detect the change on it
	go WatchConfigMap()

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
	// For turbo-on-turbo use case, we still use configMap to define prometheus query mappings.
	// We also support configuring the servers/exporters in helm-chart configmap template.
	// We also need to make sure that if either the servers or exporter fields are not configured (empty)
	// in the configmap, than the turbo metrics CR is read
	configMapMetricProvider, err := configmap.GetMetricProvider(prometheusConfigFileName) //config map is mounted as 'prometheus.config' file
	if err == nil {
		metricProvider = configMapMetricProvider
	} else {
		// if we cannot read the config file, or if either the servers/exporter config is missing
		// look at the turbo-metrics CRs
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

func WatchConfigMap() {
	//Check if the /etc/prometurbo/turbo-autoreload.config exists
	autoReloadConfigFilePath := "/etc/prometurbo"
	autoReloadConfigFileName := "turbo-autoreload.config"

	viper.AddConfigPath(autoReloadConfigFilePath)
	viper.SetConfigType("json")
	viper.SetConfigName(autoReloadConfigFileName)
	for {
		verr := viper.ReadInConfig()
		if verr == nil {
			break
		} else {
			glog.V(4).Infof("Can't read the autoreload config file %s/%s due to the error: %v, will retry in 3 seconds", autoReloadConfigFilePath, autoReloadConfigFileName, verr)
			time.Sleep(30 * time.Second)
		}
	}

	glog.V(1).Infof("Start watching the autoreload config file %s/%s", autoReloadConfigFilePath, autoReloadConfigFileName)
	updateConfig := func() {
		newLoggingLevel := viper.GetString("logging.level")
		currentLoggingLevel := flag.Lookup("v").Value.String()
		if newLoggingLevel != currentLoggingLevel {
			if newLogVInt, err := strconv.Atoi(newLoggingLevel); err != nil || newLogVInt < 0 {
				glog.Errorf("Invalid log verbosity %v in the autoreload config file", newLoggingLevel)
			} else {
				err := flag.Lookup("v").Value.Set(newLoggingLevel)
				if err != nil {
					glog.Errorf("Can't apply the new logging level setting due to the error:%v", err)
				} else {
					glog.V(1).Infof("Logging level is changed from %v to %v", currentLoggingLevel, newLoggingLevel)
				}
			}
		}
	}
	updateConfig() //update the logging level during startup
	viper.OnConfigChange(func(in fsnotify.Event) {
		updateConfig()
	})

	viper.WatchConfig()
}
