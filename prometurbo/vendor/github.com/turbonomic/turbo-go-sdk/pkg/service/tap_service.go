package service

import (
	"fmt"
	"net/url"

	restclient "github.com/turbonomic/turbo-api/pkg/client"

	"github.com/turbonomic/turbo-go-sdk/pkg/mediationcontainer"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"

	"github.com/golang/glog"
)

const (
	defaultTurboAPIPath = "/vmturbo/rest"
)

// Turbonomic Automation Service that will discover the environment for the Turbo server
// and receive recommendations to control the infrastructure environment
type TAPService struct {
	// TurboProbe provides the communication interface between the Turbo server
	// and the infrastructure environment
	*probe.TurboProbe
	*restclient.Client
	disconnectFromTurbo chan struct{}
}

func (tapService *TAPService) DisconnectFromTurbo() {
	glog.V(4).Infof("[DisconnectFromTurbo] Enter *******")
	close(tapService.disconnectFromTurbo)
	glog.V(4).Infof("[DisconnectFromTurbo] End *********")
}

func (tapService *TAPService) addTarget(isRegistered chan bool) {
	c := tapService.ProbeConfiguration
	pinfo := c.ProbeCategory + "::" + c.ProbeType

	//1. wait until probe is registered.
	select {
	case status := <-isRegistered:
		if !status {
			glog.Errorf("Probe %v registration failed.", pinfo)
			close(tapService.disconnectFromTurbo)
			glog.Errorf("Notified to close TAP service.")
			return
		}
		break
	case <-tapService.disconnectFromTurbo:
		glog.V(2).Infof("Abort adding target: Kubeturbo service is stopped.")
		return
	}

	//2. register the targets
	glog.V(2).Infof("Probe %v is registered. Begin to add targets.", pinfo)

	// Login to obtain the session cookie
	targetInfos := tapService.GetProbeTargets()
	if len(targetInfos) > 0 {
		if _, err := tapService.Client.Login(); err != nil {
			glog.Errorf("Cannot login to the Turbo API Client: %v", err)
			return
		}
		glog.V(2).Infof("Successfully logged in to Turbo server.")

		for _, targetInfo := range targetInfos {
			target := targetInfo.GetTargetInstance()
			glog.V(2).Infof("Now adding target %++v", target)
			resp, err := tapService.AddTarget(target)
			if err != nil {
				glog.Errorf("Error while adding %s %s [%s] target: %s", targetInfo.TargetCategory(),
					targetInfo.TargetType(), targetInfo.TargetIdentifierField(), err)
				// TODO, do we want to return an error?
				continue
			}
			glog.V(3).Infof("Successfully add target: %v", resp)
		}
	}
}

func (tapService *TAPService) ConnectToTurbo() {
	glog.V(4).Infof("[ConnectToTurbo] Enter ******* ")

	if tapService.RESTClient == nil {
		glog.V(1).Infof("TAP service cannot be started - cannot create target")
		return
	}

	tapService.disconnectFromTurbo = make(chan struct{})
	isRegistered := make(chan bool, 1)
	defer close(isRegistered)

	// start a separate go routine to connect to the Turbo server
	go mediationcontainer.InitMediationContainer(isRegistered)
	go tapService.addTarget(isRegistered)

	// Once connected the mediation container will keep running till a disconnect message is sent to the tap service
	select {
	case <-tapService.disconnectFromTurbo:
		glog.V(1).Infof("Begin to stop TAP service.")
		mediationcontainer.CloseMediationContainer()
		glog.V(1).Infof("TAP service is stopped.")
		return
	}
}

// ==============================================================================

// Convenience builder for building a TAPService
type TAPServiceBuilder struct {
	tapService *TAPService

	err error
}

// Get an instance of TAPServiceBuilder
func NewTAPServiceBuilder() *TAPServiceBuilder {
	serviceBuilder := &TAPServiceBuilder{}
	service := &TAPService{}
	serviceBuilder.tapService = service
	return serviceBuilder
}

// Build a new instance of TAPService.
func (builder *TAPServiceBuilder) Create() (*TAPService, error) {
	if builder.err != nil {
		return nil, builder.err
	}

	return builder.tapService, nil
}

// Build the mediation container and Turbo API client.
func (builder *TAPServiceBuilder) WithTurboCommunicator(commConfig *TurboCommunicationConfig) *TAPServiceBuilder {
	if builder.err != nil {
		return builder
	}

	// Create the mediation container. This is a singleton.
	containerConfig := &mediationcontainer.MediationContainerConfig{
		ServerMeta:      commConfig.ServerMeta,
		WebSocketConfig: commConfig.WebSocketConfig,
	}
	mediationcontainer.CreateMediationContainer(containerConfig)

	// The RestAPI Handler
	serverAddress, err := url.Parse(commConfig.TurboServer)
	if err != nil {
		builder.err = fmt.Errorf("Error during create Turbo API client config: Incorrect URL: %s\n", err)
		return builder
	}
	config := restclient.NewConfigBuilder(serverAddress).
		APIPath(defaultTurboAPIPath).
		BasicAuthentication(commConfig.OpsManagerUsername, commConfig.OpsManagerPassword).
		Create()
	glog.V(4).Infof("The Turbo API client config authentication is: %s, %s", commConfig.OpsManagerUsername, commConfig.OpsManagerPassword)
	glog.V(4).Infof("The Turbo API client config is create successfully: %v", config)
	apiClient, err := restclient.NewAPIClientWithBA(config)
	if err != nil {
		builder.err = fmt.Errorf("Error during create Turbo API client: %s\n", err)
		return builder
	}
	glog.V(4).Infof("The Turbo API client is created successfully: %v", apiClient)

	builder.tapService.Client = apiClient

	return builder
}

// The TurboProbe representing the service in the Turbo server
func (builder *TAPServiceBuilder) WithTurboProbe(probeBuilder *probe.ProbeBuilder) *TAPServiceBuilder {
	if builder.err != nil {
		return builder
	}
	// Create the probe
	turboProbe, err := probeBuilder.Create()
	if err != nil {
		builder.err = err
		return builder
	}

	builder.tapService.TurboProbe = turboProbe
	// Register the probe
	err = mediationcontainer.LoadProbe(turboProbe)
	if err != nil {
		builder.err = err
		return builder
	}

	return builder
}
