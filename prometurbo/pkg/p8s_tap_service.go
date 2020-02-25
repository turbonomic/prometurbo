package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/conf"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
)

const (
	usernameFilePath = "/etc/turbonomic-credentials/username"
	passwordFilePath = "/etc/turbonomic-credentials/password"
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

	// Disconnect from Turbo server when Kubeturbo is shutdown
	handleExit(func() { p.tapService.DisconnectFromTurbo() })

	// Connect to the Turbo server
	p.tapService.ConnectToTurbo()
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

	if err := loadOpsMgrCredentialsFromSecret(conf); err != nil {
		return nil, err
	}

	communicator := conf.Communicator
	metricExporter := exporter.NewMetricExporter(conf.MetricExporterEndpoint)
	var targetAddr, scope string
	if conf.TargetConf != nil {
		targetAddr = conf.TargetConf.Address
		scope = conf.TargetConf.Scope
	}
	keepStandalone := args.KeepStandalone

	registrationClient := &registration.P8sRegistrationClient{conf.TargetTypeSuffix}
	targetType := registrationClient.TargetType()

	var optionalTargetAddr *string
	if len(targetAddr) > 0 {
		optionalTargetAddr = &targetAddr
	}
	discoveryClient := discovery.NewDiscoveryClient(*keepStandalone,
		scope, optionalTargetAddr, targetType, metricExporter)

	builder := probe.NewProbeBuilder(targetType, registration.ProbeCategory).
		WithDiscoveryOptions(probe.FullRediscoveryIntervalSecondsOption(int32(*args.DiscoveryIntervalSec))).
		WithEntityMetadata(registrationClient).
		RegisteredBy(registrationClient)

	if len(targetAddr) > 0 {
		glog.Infof("Should discover target %s", targetAddr)
		builder = builder.DiscoversTarget(targetAddr, discoveryClient)
	} else {
		glog.Infof("Not discovering target")
		builder = builder.WithDiscoveryClient(discoveryClient)
	}

	return service.NewTAPServiceBuilder().
		WithTurboCommunicator(communicator).
		WithTurboProbe(builder).
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

func loadOpsMgrCredentialsFromSecret(conf *conf.PrometurboConf) error {
	// Return unchanged if the mounted file isn't present
	// for backward compatibility.
	if _, err := os.Stat(usernameFilePath); os.IsNotExist(err) {
		glog.V(3).Infof("credentials from secret unavailable. Checked path: %s", usernameFilePath)
		return nil
	}
	if _, err := os.Stat(passwordFilePath); os.IsNotExist(err) {
		glog.V(3).Infof("credentials from secret unavailable. Checked path: %s", passwordFilePath)
		return nil
	}

	username, err := ioutil.ReadFile(usernameFilePath)
	if err != nil {
		return fmt.Errorf("error reading credentials from secret: username: %v", err)
	}
	password, err := ioutil.ReadFile(passwordFilePath)
	if err != nil {
		return fmt.Errorf("error reading credentials from secret: password: %v", err)
	}

	conf.Communicator.OpsManagerUsername = strings.TrimSpace(string(username))
	conf.Communicator.OpsManagerPassword = strings.TrimSpace(string(password))

	return nil
}
