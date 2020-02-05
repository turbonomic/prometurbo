package probe

import (
	"fmt"

	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"

	"github.com/golang/glog"
)

// Turbo Probe Abstraction
// Consists of clients that handle probe registration metadata,
// and the discovery and action execution for different probe targets
// A single handler implementation is associated with a TurboProbe.
type TurboProbe struct {
	ProbeConfiguration *ProbeConfig
	RegistrationClient *ProbeRegistrationAgent
	// TargetsToAdd are targets we create for the probe at startup. Targets may also
	// be created in the platform through the UI. Permitted to be empty.
	TargetsToAdd    map[string]bool
	DiscoveryClient *DiscoveryAgent

	ActionClient TurboActionExecutorClient
}

type ProbeRegistrationAgent struct {
	ISupplyChainProvider
	IAccountDefinitionProvider
	IActionPolicyProvider
	IEntityMetadataProvider
}

type TurboRegistrationClient interface {
	ISupplyChainProvider
	IAccountDefinitionProvider
}

func NewProbeRegistrator() *ProbeRegistrationAgent {
	registrator := &ProbeRegistrationAgent{}
	return registrator
}

type DiscoveryAgent struct {
	TurboDiscoveryClient
	IIncrementalDiscovery
	IPerformanceDiscovery
}

func NewDiscoveryAgent() *DiscoveryAgent {
	agent := &DiscoveryAgent{}
	return agent
}

type TurboDiscoveryClient interface {
	// Discovers the target, creating an EntityDTO representation of every object
	// discovered by the  probe.
	// @param accountValues object, holding all the account values
	// @return discovery response, consisting of both entities and errors, if any
	Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error)
	Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error)
	GetAccountValues() *TurboTargetInfo
}

// ===========================================    New Probe ==========================================================
// Create new instance of a probe using the given ProbeConfig.
// Returns nil if the probe config cannot be validated.
func newTurboProbe(probeConf *ProbeConfig) (*TurboProbe, error) {
	err := probeConf.Validate()
	if err != nil {
		return nil, err
	}

	myProbe := &TurboProbe{
		ProbeConfiguration: probeConf,
		DiscoveryClient:    NewDiscoveryAgent(), // Populate fields later.
		RegistrationClient: NewProbeRegistrator(),
		TargetsToAdd:       map[string]bool{},
	}

	glog.V(2).Infof("[NewTurboProbe] Created TurboProbe: %v", myProbe)
	return myProbe, nil
}

// TODO: this method should be synchronized
func (theProbe *TurboProbe) GetTurboDiscoveryClient() TurboDiscoveryClient {
	return theProbe.DiscoveryClient.TurboDiscoveryClient
}

func findTargetId(accountValues []*proto.AccountValue, identifyingField string) string {
	var address string

	for _, accVal := range accountValues {
		if accVal.GetKey() == identifyingField {
			address = accVal.GetStringValue()
			return address
		}
	}
	return ""
}

func (theProbe *TurboProbe) DiscoverTarget(accountValues []*proto.AccountValue) *proto.DiscoveryResponse {
	glog.V(2).Infof("Discover Target: %s", accountValues)
	targetId := findTargetId(accountValues, theProbe.RegistrationClient.GetIdentifyingFields())
	handler := theProbe.DiscoveryClient.TurboDiscoveryClient
	if handler == nil {
		glog.Errorf("Failed to discover target (id=%v): cannot find discovery handler.", targetId)
		return theProbe.createNonSupportedDiscoveryErrorDTO("Full", targetId)
	}

	var discoveryResponse *proto.DiscoveryResponse
	glog.V(4).Infof("Send discovery request to handler %v", handler)
	discoveryResponse, err := handler.Discover(accountValues)

	if err != nil {
		discoveryResponse = theProbe.createDiscoveryTargetErrorDTO("Full", targetId, err)
		glog.Errorf("Error discovering target %s", discoveryResponse)
	}
	glog.V(4).Infof("Discovery response: %s", discoveryResponse)
	return discoveryResponse
}

func (theProbe *TurboProbe) ValidateTarget(accountValues []*proto.AccountValue) *proto.ValidationResponse {
	glog.V(2).Infof("Validate Target: %++v", accountValues)
	var targetId = findTargetId(accountValues, theProbe.RegistrationClient.GetIdentifyingFields())
	handler := theProbe.DiscoveryClient.TurboDiscoveryClient
	if handler == nil {
		glog.Errorf("Failed to validate target (id=%v): cannot find discovery handler.", targetId)
		return theProbe.createNonSupportedValidationErrorDTO(targetId)
	}

	var validationResponse *proto.ValidationResponse
	glog.V(3).Infof("Send validation request to handler %s", handler)
	validationResponse, err := handler.Validate(accountValues)

	// TODO: if the handler is nil, implies the target is added from the UI
	// Create a new discovery client for this target and add it to the map of discovery clients
	// allow to pass a func to instantiate a default discovery client
	if err != nil {
		description := fmt.Sprintf("Error validating target %s", err)
		severity := proto.ErrorDTO_CRITICAL

		validationResponse = theProbe.createValidationErrorDTO(description, severity)
		glog.Errorf("Error validating target %s", validationResponse)
	}
	glog.V(3).Infof("Validation response: %s", validationResponse)
	return validationResponse
}

func (theProbe *TurboProbe) DiscoverTargetIncremental(accountValues []*proto.AccountValue) *proto.DiscoveryResponse {
	glog.V(2).Infof("Incremental discovery for Target: %s", accountValues)
	targetId := findTargetId(accountValues, theProbe.RegistrationClient.GetIdentifyingFields())
	handler := theProbe.DiscoveryClient.IIncrementalDiscovery
	if handler == nil {
		glog.Errorf("Failed to incrementally discover target (id=%v): cannot find discovery handler.", targetId)
		return theProbe.createNonSupportedDiscoveryErrorDTO("Incremental", targetId)
	}

	glog.V(4).Infof("Send incremental discovery request to handler %v", handler)
	discoveryResponse, err := handler.DiscoverIncremental(accountValues)
	if err != nil {
		discoveryResponse = theProbe.createDiscoveryTargetErrorDTO("Incremental", targetId, err)
		glog.Errorf("Error during incremental discovery of target %s", discoveryResponse)
	}
	glog.V(3).Infof("Incremental Discovery response: %s", discoveryResponse)
	return discoveryResponse
}

func (theProbe *TurboProbe) DiscoverTargetPerformance(accountValues []*proto.AccountValue) *proto.DiscoveryResponse {
	glog.V(2).Infof("Performance discovery for Target: %s", accountValues)
	targetId := findTargetId(accountValues, theProbe.RegistrationClient.GetIdentifyingFields())
	handler := theProbe.DiscoveryClient.IPerformanceDiscovery
	if handler == nil {
		glog.Errorf("Failed to performance discover target (id=%v): cannot find discovery handler.", targetId)
		return theProbe.createNonSupportedDiscoveryErrorDTO("Performance", targetId)
	}

	glog.V(4).Infof("Send performance discovery request to handler %v", handler)
	discoveryResponse, err := handler.DiscoverPerformance(accountValues)
	if err != nil {
		discoveryResponse = theProbe.createDiscoveryTargetErrorDTO("Performance", targetId, err)
		glog.Errorf("Error during performance discovery of target %s", discoveryResponse)
	}
	glog.V(3).Infof("Performance Discovery response: %s", discoveryResponse)
	return discoveryResponse
}

func (theProbe *TurboProbe) ExecuteAction(actionExecutionDTO *proto.ActionExecutionDTO, accountValues []*proto.AccountValue,
	progressTracker ActionProgressTracker) *proto.ActionResult {
	if theProbe.ActionClient == nil {
		glog.Errorf("ActionClient not defined for Probe %s", theProbe.ProbeConfiguration.ProbeType)
		return theProbe.createActionErrorDTO("ActionClient not defined for Probe " + theProbe.ProbeConfiguration.ProbeType)
	}
	glog.V(3).Infof("Execute Action for Target: %s", accountValues)
	response, err := theProbe.ActionClient.ExecuteAction(actionExecutionDTO, accountValues, progressTracker)

	if err != nil {
		description := fmt.Sprintf("Error executing action %s", err)
		glog.Errorf("Error executing action")
		return theProbe.createActionErrorDTO(description)
	}
	return response
}

// ==============================================================================================================
// The Targets associated with this probe type
func (theProbe *TurboProbe) GetProbeTargets() []*TurboTargetInfo {
	if theProbe.DiscoveryClient.TurboDiscoveryClient == nil {
		return []*TurboTargetInfo{}
	}

	// Iterate over the discovery client map and send requests to the server
	var targets []*TurboTargetInfo
	for targetId := range theProbe.TargetsToAdd {

		targetInfo := theProbe.DiscoveryClient.GetAccountValues()
		targetInfo.targetType = theProbe.ProbeConfiguration.ProbeType
		targetInfo.targetIdentifierField = targetId

		targets = append(targets, targetInfo)
	}
	return targets
}

// The ProbeInfo for the probe
func (theProbe *TurboProbe) GetProbeInfo() (*proto.ProbeInfo, error) {
	// 1. construct the basic probe info.
	probeConf := theProbe.ProbeConfiguration
	probeCat := probeConf.ProbeCategory
	probeType := probeConf.ProbeType

	probeInfoBuilder := builder.NewBasicProbeInfoBuilder(probeType, probeCat)

	// 2. discovery intervals metadata
	probeInfoBuilder.WithFullDiscoveryInterval(probeConf.discoveryMetadata.GetFullRediscoveryIntervalSeconds())
	probeInfoBuilder.WithIncrementalDiscoveryInterval(probeConf.discoveryMetadata.GetIncrementalRediscoveryIntervalSeconds())
	probeInfoBuilder.WithPerformanceDiscoveryInterval(probeConf.discoveryMetadata.GetPerformanceRediscoveryIntervalSeconds())
	//probeInfo := probeInfoBuilder.Create()
	//glog.V(2).Infof("************** ProbeInfo %++v\n", probeInfo)

	registrationClient := theProbe.RegistrationClient
	// 3. Get the supply chain.
	if registrationClient.ISupplyChainProvider != nil {
		probeInfoBuilder.WithSupplyChain(registrationClient.GetSupplyChainDefinition())
	}

	if registrationClient.IAccountDefinitionProvider != nil {
		// 4. Get the account definition
		probeInfoBuilder.WithAccountDefinition(registrationClient.GetAccountDefinition())

		// 5. Fields that serve to uniquely identify a target
		probeInfoBuilder = probeInfoBuilder.WithIdentifyingField(registrationClient.GetIdentifyingFields())
	}
	// 6. action policy metadata
	if registrationClient.IActionPolicyProvider != nil {
		probeInfoBuilder.WithActionPolicySet(registrationClient.GetActionPolicy())
	}

	// 7. entity metadata
	if registrationClient.IEntityMetadataProvider != nil {
		probeInfoBuilder.WithEntityMetadata(registrationClient.GetEntityMetadata())
	}

	probeInfo := probeInfoBuilder.Create()
	glog.V(3).Infof("ProbeInfo %+v", probeInfo)

	return probeInfo, nil
}

func (theProbe *TurboProbe) createNonSupportedValidationErrorDTO(targetId string) *proto.ValidationResponse {
	errorStr := fmt.Sprintf("Validation not supported for :%s", targetId)
	return theProbe.createValidationErrorDTO(errorStr, proto.ErrorDTO_CRITICAL)
}

func (theProbe *TurboProbe) createMisconfiguredValidationErrorDTO(targetId string) *proto.ValidationResponse {
	errorStr := fmt.Sprintf("Misconfigured target during validation of target %s", targetId)
	fmt.Errorf(errorStr)
	return theProbe.createValidationErrorDTO(errorStr, proto.ErrorDTO_CRITICAL)
}

func (theProbe *TurboProbe) createNonSupportedDiscoveryErrorDTO(discoveryType string, targetId string) *proto.DiscoveryResponse {
	errorStr := fmt.Sprintf("%s discovery not supported for :%s", discoveryType, targetId)
	return theProbe.createDiscoveryErrorDTO(errorStr, proto.ErrorDTO_CRITICAL)
}

func (theProbe *TurboProbe) createDiscoveryTargetErrorDTO(discoveryType string, targetId string, err error) *proto.DiscoveryResponse {
	errorStr := fmt.Sprintf("Error during %s discovery of target %s: %s", discoveryType, targetId, err)
	return theProbe.createDiscoveryErrorDTO(errorStr, proto.ErrorDTO_CRITICAL)
}

func (theProbe *TurboProbe) createMisconfiguredProbeErrorDTO(discoveryType string, targetId string) *proto.DiscoveryResponse {
	errorStr := fmt.Sprintf("Misconfigured target during %s discovery of target %s", discoveryType, targetId)
	fmt.Errorf(errorStr)
	return theProbe.createDiscoveryErrorDTO(errorStr, proto.ErrorDTO_CRITICAL)
}

func (theProbe *TurboProbe) createValidationErrorDTO(errMsg string, severity proto.ErrorDTO_ErrorSeverity) *proto.ValidationResponse {
	errorDTO := &proto.ErrorDTO{
		Severity:    &severity,
		Description: &errMsg,
	}
	var errorDtoList []*proto.ErrorDTO
	errorDtoList = append(errorDtoList, errorDTO)

	validationResponse := &proto.ValidationResponse{
		ErrorDTO: errorDtoList,
	}
	return validationResponse
}

func (theProbe *TurboProbe) createDiscoveryErrorDTO(errMsg string, severity proto.ErrorDTO_ErrorSeverity) *proto.DiscoveryResponse {
	errorDTO := &proto.ErrorDTO{
		Severity:    &severity,
		Description: &errMsg,
	}
	var errorDtoList []*proto.ErrorDTO
	errorDtoList = append(errorDtoList, errorDTO)

	discoveryResponse := &proto.DiscoveryResponse{
		ErrorDTO: errorDtoList,
	}
	return discoveryResponse
}

func (theProbe *TurboProbe) createActionErrorDTO(errMsg string) *proto.ActionResult {
	var progress int32
	progress = 100
	state := proto.ActionResponseState_FAILED
	// build ActionResponse
	actionResponse := &proto.ActionResponse{
		ActionResponseState: &state,
		ResponseDescription: &errMsg,
		Progress:            &progress,
	}

	response := &proto.ActionResult{
		Response: actionResponse,
	}

	return response
}
