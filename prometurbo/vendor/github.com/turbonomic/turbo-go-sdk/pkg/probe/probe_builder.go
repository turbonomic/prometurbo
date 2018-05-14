package probe

import (
	"errors"
	"github.com/golang/glog"
)

type ProbeBuilder struct {
	probeConf              *ProbeConfig
	registrationClient     TurboRegistrationClient
	discoveryClientMap     map[string]TurboDiscoveryClient
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
		probeBuilder.builderError = err
		return probeBuilder
	}

	return &ProbeBuilder{
		probeConf:          probeConf,
		discoveryClientMap: make(map[string]TurboDiscoveryClient),
	}
}

// Build an instance of TurboProbe.
func (pb *ProbeBuilder) Create() (*TurboProbe, error) {
	if pb.builderError != nil {
		glog.Errorf(pb.builderError.Error())
		return nil, pb.builderError
	}

	if len(pb.discoveryClientMap) == 0 {
		pb.builderError = ErrorUndefinedDiscoveryClient()
		glog.Errorf(pb.builderError.Error())
		return nil, pb.builderError
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

	turboProbe.ActionClient = pb.actionClient
	for targetId, discoveryClient := range pb.discoveryClientMap {
		targetDiscoveryAgent := NewTargetDiscoveryAgent(targetId)
		targetDiscoveryAgent.TurboDiscoveryClient = discoveryClient
		turboProbe.DiscoveryClientMap[targetId] = targetDiscoveryAgent //discoveryClient
	}

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

	pb.discoveryClientMap[targetId] = discoveryClient

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
