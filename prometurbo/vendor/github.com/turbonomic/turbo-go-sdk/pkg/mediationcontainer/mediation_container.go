package mediationcontainer

import (
	"errors"
	"sync"

	"github.com/turbonomic/turbo-go-sdk/pkg/probe"

	"github.com/golang/glog"
)

type mediationContainer struct {
	// Configuration for making the transport connection
	containerConfig *MediationContainerConfig
	// Map of probes registered with the container
	allProbes map[string]*ProbeProperties
	// The Mediation client that will handle the messages from the server
	theRemoteMediationClient *remoteMediationClient
}

type ProbeSignature struct {
	ProbeType     string
	ProbeCategory string
}

type ProbeProperties struct {
	ProbeSignature *ProbeSignature
	Probe          *probe.TurboProbe
}

var (
	theInstance *mediationContainer
	once        sync.Once
)

func singletonMediationContainer() *mediationContainer {
	once.Do(func() {
		if theInstance == nil {
			theInstance = &mediationContainer{
				allProbes: make(map[string]*ProbeProperties),
				// TODO: create the probe store and mediation client here
			}
		}
	})
	return theInstance
}

// Static method to get the singleton instance of the Mediation Container
func CreateMediationContainer(containerConfig *MediationContainerConfig) *mediationContainer {
	// Validate the container config
	containerConfig.ValidateMediationContainerConfig()
	glog.Infof("---------- Created MediationContainer ----------")
	theContainer := singletonMediationContainer() //&mediationContainer {} // TODO: make a singleton instance

	//  Load the main container configuration file and validate it
	theContainer.containerConfig = containerConfig

	// Create the RemoteMediationClient to start the session with the server
	theContainer.theRemoteMediationClient = CreateRemoteMediationClient(theContainer.allProbes, theContainer.containerConfig)

	return theContainer
}

// Start the RemoteMediationClient
func InitMediationContainer(probeRegisteredMsg chan bool) {
	theContainer := singletonMediationContainer()
	glog.Infof("Initializing mediation container .....")
	// Assert that the probes are registered before starting the handshake
	if len(theContainer.allProbes) == 0 {
		glog.Errorf("No probes are registered with the container")
		return
	}
	// Open connection to the server and start server handshake to register probes
	glog.V(2).Infof("Registering %d probes", len(theContainer.allProbes))

	remoteMediationClient := theContainer.theRemoteMediationClient
	remoteMediationClient.Init(probeRegisteredMsg)
}

func CloseMediationContainer() {
	glog.Infof("[CloseMediationContainer] Closing mediation container .....")
	theContainer := singletonMediationContainer()
	theContainer.theRemoteMediationClient.Stop()
	// TODO: clear probe map ?
}

// ============================= Probe Management ==================
func LoadProbe(probe *probe.TurboProbe) error {
	// load the probe config
	config := &ProbeSignature{
		ProbeCategory: probe.ProbeConfiguration.ProbeCategory,
		ProbeType:     probe.ProbeConfiguration.ProbeType,
	}

	probeProp := &ProbeProperties{
		ProbeSignature: config,
		Probe:          probe,
	}
	theContainer := singletonMediationContainer()
	if theContainer == nil {
		return errors.New("[LoadProbe] Null mediation container")
	}
	// TODO: check if the probe type already exists and warn before overwriting
	theContainer.allProbes[config.ProbeType] = probeProp
	glog.Infof("Registered " + config.ProbeCategory + "::" + config.ProbeType)
	return nil
}

func GetProbe(probeType string) (*probe.TurboProbe, error) {
	theContainer := singletonMediationContainer()
	if theContainer == nil {
		return nil, errors.New("[GetProbe] Null mediation container")
	}
	probeProps := theContainer.allProbes[probeType]

	if probeProps != nil {
		probe := probeProps.Probe
		registrationClient := probe.RegistrationClient
		acctDefProps := registrationClient.GetAccountDefinition()
		glog.V(2).Infof("Found "+probeProps.ProbeSignature.ProbeCategory+"::"+probeProps.ProbeSignature.ProbeType+" ==> ", acctDefProps)
		return probe, nil
	}
	return nil, errors.New("[GetProbe] Cannot find Probe of type " + probeType)

}
