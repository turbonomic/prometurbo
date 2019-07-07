package main

import (
	"flag"
	"github.com/golang/glog"
	"os"
	"strconv"

	"fmt"
	"github.com/turbonomic/prometurbo/appmetric/pkg/addon"
	ali "github.com/turbonomic/prometurbo/appmetric/pkg/alligator"
	"github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
	"github.com/turbonomic/prometurbo/appmetric/pkg/server"
)

const (
	defaultPort           = 8081
	defaultSampleDuration = "3m"
)

var (
	prometheusHost string
	port           int
	configfname    string
	sampleDuration string
)

func parseFlags() {
	flag.StringVar(&prometheusHost, "promUrl", "", "the address of prometheus server")
	flag.IntVar(&port, "port", 0, "port to expose metrics (default 8081)")
	flag.StringVar(&configfname, "config", "", "path of the config file")
	flag.StringVar(&sampleDuration, "sampleDuration", defaultSampleDuration, "the sample duration for prometheus query")
	flag.Parse()
}

func getJobs(mclient *prometheus.RestClient) {
	msg, err := mclient.GetJobs()
	if err != nil {
		glog.Errorf("Failed to get jobs: %v", err)
		return
	}
	glog.V(1).Infof("jobs: %v", msg)
}

func test_prometheus(mclient *prometheus.RestClient) {
	glog.V(2).Infof("Begin to test prometheus client...")
	getJobs(mclient)
	glog.V(2).Infof("End of testing prometheus client.")
	return
}

func parseConf() error {
	if prometheusHost == "" && configfname == "" {
		err := fmt.Errorf("neither promUrl nor config flags is set")
		glog.Errorf(err.Error())
		return err
	}

	if len(configfname) > 0 {
		mconf, err := readConfig(configfname)
		if err != nil {
			glog.Errorf("Failed to load config file: %v", err)
			return err
		}

		if len(prometheusHost) < 1 {
			prometheusHost = mconf.Address
		}

		if port < 1 && len(mconf.Port) > 1 {
			port, err = strconv.Atoi(mconf.Port)
			if err != nil {
				glog.Errorf("Failed to convert port from string to int: %v", err)
				return err
			}
		}
	}

	if len(prometheusHost) < 1 {
		err := fmt.Errorf("Failed to get prometheus server address")
		glog.Error(err.Error())
		return err
	}

	if port < 1 {
		port = defaultPort
	}

	return nil
}

func main() {
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/var/log")
	defer glog.Flush()

	parseFlags()
	glog.Info("Starting Prometurbo...")
	glog.Infof("GIT_COMMIT: %s", os.Getenv("GIT_COMMIT"))

	if err := parseConf(); err != nil {
		glog.Errorf("Failed to parse configurations : %v", err)
		glog.Errorf("Quit now")
		return
	}

	pclient, err := prometheus.NewRestClient(prometheusHost)
	if err != nil {
		glog.Fatalf("Failed to generate client: %v", err)
	}
	//mclient.SetUser("", "")
	test_prometheus(pclient)

	factory := addon.NewGetterFactory()

	//1. Application Metrics
	appClient := ali.NewAlligator(pclient)
	istioGetter, err := factory.CreateEntityGetter(addon.IstioGetterCategory, "istio.app.metric", sampleDuration)
	if err != nil {
		glog.Errorf("Failed to create Istio App getter: %v", err)
		return
	}
	appClient.AddGetter(istioGetter)

	redisGetter, err := factory.CreateEntityGetter(addon.RedisGetterCategory, "redis.app.metric", sampleDuration)
	if err != nil {
		glog.Errorf("Failed to create Redis App getter: %v", err)
		return
	}
	appClient.AddGetter(redisGetter)

	cassandraGetter, err := factory.CreateEntityGetter(addon.CassandraGetterCategory, "cassandra.app.metric", sampleDuration)
	if err != nil {
		glog.Errorf("Failed to create Cassandra App getter: %v", err)
		return
	}
	glog.V(2).Infof("Added cassandraGetter: %+v", cassandraGetter)
	appClient.AddGetter(cassandraGetter)

	webdriverGetter, err := factory.CreateEntityGetter(addon.WebdriverGetterCategory, "webdriver.app.metric", sampleDuration)
	if err != nil {
		glog.Errorf("Failed to create Webdriver App getter: %v", err)
		return
	}
	glog.V(2).Infof("Added webdriverGetter: %+v", webdriverGetter)
	appClient.AddGetter(webdriverGetter)

	//2. Virtual Application Metrics
	vappClient := ali.NewAlligator(pclient)
	istioVAppGetter, err := factory.CreateEntityGetter(addon.IstioVAppGetterCategory, "istio.vapp.metric", sampleDuration)
	if err != nil {
		glog.Errorf("Failed to create Istio VApp getter: %v", err)
		return
	}
	vappClient.AddGetter(istioVAppGetter)

	webdriverVAppGetter, err := factory.CreateEntityGetter(addon.WebdriverVAppGetterCategory, "webdriver.vapp.metric", sampleDuration)
	if err != nil {
		glog.Errorf("Failed to create Webdriver VApp getter: %v", err)
		return
	}
	vappClient.AddGetter(webdriverVAppGetter)

	s := server.NewMetricServer(port, appClient, vappClient)
	s.Run()
	return
}
