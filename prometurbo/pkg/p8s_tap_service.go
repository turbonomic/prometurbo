package pkg

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/conf"
	"github.com/turbonomic/prometurbo/pkg/discovery"
	"github.com/turbonomic/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/prometurbo/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type disconnectFromTurboFunc func()

type P8sTAPService struct {
	tapService *service.TAPService
}

func NewP8sTAPService(args *conf.PrometurboArgs) (*P8sTAPService, error) {
	tapService, err := createTAPService(args)

	if err != nil {
		glog.Errorf("Error while building turbo TAP service on target %v", err)
		return nil, err
	}

	return &P8sTAPService{tapService}, nil
}

func (p *P8sTAPService) Start() {
	glog.V(0).Infof("Starting prometheus TAP service...")

	// Before running service, wait for the exporter to start up
	// TODO: Check the readiness of the exporter
	time.Sleep(5 * time.Second)

	// Disconnect from Turbo server when Kubeturbo is shutdown
	handleExit(func() { p.tapService.DisconnectFromTurbo() })

	// Connect to the Turbo server
	p.tapService.ConnectToTurbo()

	select {}
}

func createTAPService(args *conf.PrometurboArgs) (*service.TAPService, error) {
	confPath := conf.DefaultConfPath

	if os.Getenv("PROMETURBO_LOCAL_DEBUG") == "1" {
		confPath = conf.LocalDebugConfPath
		glog.V(2).Infof("Using config file %s for local debugging", confPath)
	}

	conf, err := conf.NewPrometurboConf(confPath)
	if err != nil {
		glog.Errorf("Error while parsing the service config file %s: %v", confPath, err)
		os.Exit(1)
	}

	glog.V(3).Infof("Read service configuration from %s: %++v", confPath, conf)

	communicator := conf.Communicator
	targetAddr := conf.TargetConf.Address
	scope := conf.TargetConf.Scope
	metricExporters := []exporter.MetricExporter{exporter.NewMetricExporter(conf.MetricExporterEndpoint)}

	registrationClient := &registration.P8sRegistrationClient{}
	discoveryClient := discovery.NewDiscoveryClient(targetAddr, scope, metricExporters)

	return service.NewTAPServiceBuilder().
		WithTurboCommunicator(communicator).
		WithTurboProbe(probe.NewProbeBuilder(registration.TargetType, registration.ProbeCategory).
			WithDiscoveryOptions(probe.FullRediscoveryIntervalSecondsOption(int32(*args.DiscoveryIntervalSec))).
			RegisteredBy(registrationClient).
			DiscoversTarget(targetAddr, discoveryClient)).
		Create()
}

// TODO: Move the handle to turbo-sdk-probe as it should be common logic for similar probes
// handleExit disconnects the tap service from Turbo service when prometurbo is terminated
func handleExit(disconnectFunc disconnectFromTurboFunc) {
	glog.V(4).Infof("*** Handling Prometurbo Termination ***")
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGHUP)

	go func() {
		select {
		case sig := <-sigChan:
			// Close the mediation container including the endpoints. It avoids the
			// invalid endpoints remaining in the server side. See OM-28801.
			glog.V(2).Infof("Signal %s received. Disconnecting from Turbo server...\n", sig)
			disconnectFunc()
		}
	}()
}
