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
	defaultPort = 8081
)

var (
	prometheusHost string
	port           int
	configfname    string
)

func parseFlags() {
	flag.StringVar(&prometheusHost, "promUrl", "", "the address of prometheus server")
	flag.IntVar(&port, "port", 0, "port to expose metrics (default 8081)")
	flag.StringVar(&configfname, "config", "", "path of the config file")
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
	istioGetter, err := factory.CreateEntityGetter(addon.IstioGetterCategory, "istio.app.metric")
	if err != nil {
		glog.Errorf("Failed to create Istio App getter: %v", err)
		return
	}
	appClient.AddGetter(istioGetter)

	redisGetter, err := factory.CreateEntityGetter(addon.RedisGetterCategory, "redis.app.metric")
	if err != nil {
		glog.Errorf("Failed to create Redis App getter: %v", err)
		return
	}
	appClient.AddGetter(redisGetter)

	//2. Virtual Application Metrics
	vappClient := ali.NewAlligator(pclient)
	vappGetter, err := factory.CreateEntityGetter(addon.IstioVAppGetterCategory, "istio.vapp.metric")
	if err != nil {
		glog.Errorf("Failed to create Istio VApp getter: %v", err)
		return
	}
	vappClient.AddGetter(vappGetter)

	s := server.NewMetricServer(port, appClient, vappClient)
	s.Run()
	return
}
