package probe

import (
	"errors"

	"github.com/golang/glog"
)

type ProbeBuilder struct {
	probeConf              *ProbeConfig
	registrationClient     TurboRegistrationClient
	targetsToAdd           map[string]bool
	discoveryClient        TurboDiscoveryClient
	actionClient           TurboActionExecutorClient
	builderError           error
	supplyChainProvider    ISupplyChainProvider
	accountDefProvider     IAccountDefinitionProvider
	actionPolicyProvider   IActionPolicyProvider
	entityMetadataProvider IEntityMetadataProvider
}

func ErrorInvalidTargetIdentifier() error {
	return errors.New("Null Target Identifier")
}

func ErrorInvalidProbeType() error {
	return errors.New("Null Probe type")
}

func ErrorInvalidProbeCategory() error {
	return errors.New("Null Probe category")
}

func ErrorInvalidRegistrationClient() error {
	return errors.New("Null registration client")
}

func ErrorInvalidActionClient() error {
	return errors.New("Null action client")
}

func ErrorInvalidDiscoveryClient(targetId string) error {
	return errors.New("Invalid discovery client for target [" + targetId + "]")
}

func ErrorUndefinedDiscoveryClient() error {
	return errors.New("No discovery clients defined")
}

func ErrorCreatingProbe(probeType string, probeCategory string) error {
	return errors.New("Error creating probe for " + probeCategory + "::" + probeType)
}

// Get an instance of ProbeBuilder
func NewProbeBuilder(probeType string, probeCategory string) *ProbeBuilder {
	probeBuilder := &ProbeBuilder{}

	// Validate probe type and category
	probeConf, err := NewProbeConfig(probeType, probeCategory)
	if err != nil {
		glog.Errorf("Encountered error constructing NewProbeBuilder: %v", err)
		probeBuilder.builderError = err
		return probeBuilder
	}

	return &ProbeBuilder{
		probeConf:    probeConf,
		targetsToAdd: make(map[string]bool),
	}
}

// Build an instance of TurboProbe.
func (pb *ProbeBuilder) Create() (*TurboProbe, error) {
	if pb.builderError != nil {
		glog.Errorf(pb.builderError.Error())
		return nil, pb.builderError
	}
	if pb.registrationClient == nil {
		return nil, ErrorInvalidRegistrationClient()
	}

	turboProbe, err := newTurboProbe(pb.probeConf)
	if err != nil {
		pb.builderError = ErrorCreatingProbe(pb.probeConf.ProbeType,
			pb.probeConf.ProbeCategory)
		glog.Errorf(pb.builderError.Error())
		return nil, pb.builderError
	}

	turboProbe.RegistrationClient.ISupplyChainProvider = pb.registrationClient
	turboProbe.RegistrationClient.IAccountDefinitionProvider = pb.registrationClient

	if pb.supplyChainProvider != nil {
		turboProbe.RegistrationClient.ISupplyChainProvider = pb.supplyChainProvider
	}

	if pb.accountDefProvider != nil {
		turboProbe.RegistrationClient.IAccountDefinitionProvider = pb.accountDefProvider
	}

	if pb.actionPolicyProvider != nil {
		turboProbe.RegistrationClient.IActionPolicyProvider = pb.actionPolicyProvider
	}

	if pb.entityMetadataProvider != nil {
		turboProbe.RegistrationClient.IEntityMetadataProvider = pb.entityMetadataProvider
	}

	turboProbe.ActionClient = pb.actionClient
	turboProbe.DiscoveryClient.TurboDiscoveryClient = pb.discoveryClient
	turboProbe.TargetsToAdd = pb.targetsToAdd

	return turboProbe, nil
}

func (pb *ProbeBuilder) WithDiscoveryOptions(options ...DiscoveryMetadataOption) *ProbeBuilder {
	discoveryMetadata := NewDiscoveryMetadata()
	for _, option := range options {
		option(discoveryMetadata)
	}

	pb.probeConf.SetDiscoveryMetadata(discoveryMetadata)
	return pb
}

// Set the supply chain provider for the probe
func (pb *ProbeBuilder) WithSupplyChain(supplyChainProvider ISupplyChainProvider) *ProbeBuilder {
	if supplyChainProvider == nil {
		pb.builderError = ErrorInvalidRegistrationClient()
		return pb
	}
	pb.supplyChainProvider = supplyChainProvider

	return pb
}

// Set the provider that for the account definition for discovering the probe targets
func (pb *ProbeBuilder) WithAccountDef(accountDefProvider IAccountDefinitionProvider) *ProbeBuilder {
	if accountDefProvider == nil {
		pb.builderError = ErrorInvalidRegistrationClient()
		return pb
	}
	pb.accountDefProvider = accountDefProvider

	return pb
}

// Set the provider for the policies regarding the supported action types
func (pb *ProbeBuilder) WithActionPolicies(actionPolicyProvider IActionPolicyProvider) *ProbeBuilder {
	if actionPolicyProvider == nil {
		pb.builderError = ErrorInvalidRegistrationClient()
		return pb
	}
	pb.actionPolicyProvider = actionPolicyProvider

	return pb
}

// Set the provider for the metadata for generating unique identifiers for the probe entities
func (pb *ProbeBuilder) WithEntityMetadata(entityMetadataProvider IEntityMetadataProvider) *ProbeBuilder {
	if entityMetadataProvider == nil {
		pb.builderError = ErrorInvalidRegistrationClient()
		return pb
	}
	pb.entityMetadataProvider = entityMetadataProvider

	return pb
}

// Set the registration client for the probe
func (pb *ProbeBuilder) RegisteredBy(registrationClient TurboRegistrationClient) *ProbeBuilder {
	if registrationClient == nil {
		pb.builderError = ErrorInvalidRegistrationClient()
		return pb
	}
	pb.registrationClient = registrationClient

	return pb
}

// Set a target and discovery client for the probe
func (pb *ProbeBuilder) DiscoversTarget(targetId string, discoveryClient TurboDiscoveryClient) *ProbeBuilder {
	if targetId == "" {
		pb.builderError = ErrorInvalidTargetIdentifier()
		return pb
	}
	if discoveryClient == nil {
		pb.builderError = ErrorInvalidDiscoveryClient(targetId)
		return pb
	}

	pb.targetsToAdd[targetId] = true
	if pb.discoveryClient != nil {
		glog.Warningf("Overwriting discovery client %v with %v", pb.discoveryClient, discoveryClient)
	}
	pb.discoveryClient = discoveryClient

	return pb
}

// Set a target and discovery client for the probe
func (pb *ProbeBuilder) WithDiscoveryClient(discoveryClient TurboDiscoveryClient) *ProbeBuilder {
	if discoveryClient == nil {
		pb.builderError = ErrorInvalidDiscoveryClient("")
		return pb
	}

	if pb.discoveryClient != nil {
		glog.Warningf("Overwriting discovery client %v with %v", pb.discoveryClient, discoveryClient)
	}
	pb.discoveryClient = discoveryClient
	return pb
}

// Set the action client for the probe
func (pb *ProbeBuilder) ExecutesActionsBy(actionClient TurboActionExecutorClient) *ProbeBuilder {
	if actionClient == nil {
		pb.builderError = ErrorInvalidActionClient()
		return pb
	}
	pb.actionClient = actionClient

	return pb
}
