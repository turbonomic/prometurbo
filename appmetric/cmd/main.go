package main

import (
	"flag"
	"github.com/golang/glog"

	"github.com/songbinliu/xfire/pkg/prometheus"
	"github.com/turbonomic/prometurbo/appmetric/pkg/addon"
	ali "github.com/turbonomic/prometurbo/appmetric/pkg/alligator"
	"github.com/turbonomic/prometurbo/appmetric/pkg/server"
)

var (
	prometheusHost string
	port           int
)

func parseFlags() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&prometheusHost, "promUrl", "http://localhost:9090", "the address of prometheus server")
	flag.IntVar(&port, "port", 8081, "port to expose metrics")
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

func main() {
	parseFlags()
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
