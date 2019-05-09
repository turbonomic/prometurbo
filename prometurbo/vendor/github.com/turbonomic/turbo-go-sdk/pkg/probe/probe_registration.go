package probe

import (
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// ISupplyChainProvider provides the entities defined in the supply chain for a probe
type ISupplyChainProvider interface {
	GetSupplyChainDefinition() []*proto.TemplateDTO
}

// IAccountDefinitionProvider provides the definitions used to uniquely create a target
// for the probe
type IAccountDefinitionProvider interface {
	GetAccountDefinition() []*proto.AccountDefEntry
	GetIdentifyingFields() string
}

// IFullDiscoveryMetadata specifies the interval at which discoveries will be executed
// for the probe that supplies this interface during registration.
// The value is specified in seconds. If the interface implementation is not provided, a default
// of 600 seconds (10 minutes) will be used.
// The minimum value allowed for this field is 60 seconds (1 minute).
type IFullDiscoveryMetadata interface {
	GetFullRediscoveryIntervalSeconds() int32
}

// IIncrementalDiscoveryMetadata specifies the interval at which incremental discoveries
// will be executed for the probe that supplies this interface during registration.
// The value is specified in seconds. If the interface implementation is not provided,
// the probe does not support incremental discovery.
type IIncrementalDiscoveryMetadata interface {
	GetIncrementalRediscoveryIntervalSeconds() int32
}

// IPerformanceDiscoveryMetadata specifies the interval at which performance discoveries
// will be executed for the probe that supplues this interface during registration.
// The value is specified in seconds. If the interface implementation is not provided,
// the probe does not support performance discovery
type IPerformanceDiscoveryMetadata interface {
	GetPerformanceRediscoveryIntervalSeconds() int32
}

// Implementation for the providing the full, incremental, performance
// discovery intervals for the probe.
type DiscoveryMetadata struct {
	fullDiscovery        int32
	incrementalDiscovery int32
	performanceDiscovery int32
}

// Create a DiscoveryMetadata structure with default values for the different discovery intervals.
// Full discovery default is set to 600 seconds which means full discovery will occur every 10 minutes.
// Incremental discovery is set to -1 implies that incremental discovery is not supported by the probe.
// Performance discovery is set to -1 implies that the performance discovery is not supported by the probe.
func NewDiscoveryMetadata() *DiscoveryMetadata {
	return &DiscoveryMetadata{
		fullDiscovery:        pkg.DEFAULT_FULL_DISCOVERY_IN_SECS,
		incrementalDiscovery: pkg.DISCOVERY_NOT_SUPPORTED,
		performanceDiscovery: pkg.DISCOVERY_NOT_SUPPORTED,
	}
}

type DiscoveryMetadataOption func(*DiscoveryMetadata)

func IncrementalRediscoveryIntervalSecondsOption(incrementalDiscovery int32) func(dm *DiscoveryMetadata) {
	return func(dm *DiscoveryMetadata) {
		dm.SetIncrementalRediscoveryIntervalSeconds(incrementalDiscovery)
	}
}

func PerformanceRediscoveryIntervalSecondsOption(performanceDiscovery int32) func(dm *DiscoveryMetadata) {
	return func(dm *DiscoveryMetadata) {
		dm.SetPerformanceRediscoveryIntervalSeconds(performanceDiscovery)
	}
}

func FullRediscoveryIntervalSecondsOption(fullDiscovery int32) func(dm *DiscoveryMetadata) {
	return func(dm *DiscoveryMetadata) {
		dm.SetFullRediscoveryIntervalSeconds(fullDiscovery)
	}
}

// Return the time interval in seconds for running the full discovery
func (dMetadata *DiscoveryMetadata) GetFullRediscoveryIntervalSeconds() int32 {
	return dMetadata.fullDiscovery
}

// Return the time interval in seconds for running the incremental discovery
func (dMetadata *DiscoveryMetadata) GetIncrementalRediscoveryIntervalSeconds() int32 {
	return dMetadata.incrementalDiscovery
}

// Return the time interval in seconds for running the performance discovery
func (dMetadata *DiscoveryMetadata) GetPerformanceRediscoveryIntervalSeconds() int32 {
	return dMetadata.performanceDiscovery
}

// Set the time interval in seconds for running the incremental discovery
func (dMetadata *DiscoveryMetadata) SetIncrementalRediscoveryIntervalSeconds(incrementalDiscovery int32) {
	interval := checkSecondaryDiscoveryInterval(incrementalDiscovery, pkg.INCREMENTAL_DISCOVERY)
	dMetadata.incrementalDiscovery = interval
}

// Set the time interval in seconds for running the performance discovery
func (dMetadata *DiscoveryMetadata) SetPerformanceRediscoveryIntervalSeconds(performanceDiscovery int32) {
	interval := checkSecondaryDiscoveryInterval(performanceDiscovery, pkg.PERFORMANCE_DISCOVERY)
	dMetadata.performanceDiscovery = interval
}

// Set the time interval in seconds for running the full discovery
func (dMetadata *DiscoveryMetadata) SetFullRediscoveryIntervalSeconds(fullDiscovery int32) {
	interval := checkFullRediscoveryInterval(fullDiscovery)
	dMetadata.fullDiscovery = interval
}

func checkSecondaryDiscoveryInterval(secondaryDiscoverySec int32, discoveryType pkg.DiscoveryType) int32 {
	if secondaryDiscoverySec <= 0 {
		glog.V(3).Infof("%s discovery is not supported.", discoveryType)
		return pkg.DISCOVERY_NOT_SUPPORTED
	}

	if secondaryDiscoverySec < pkg.DEFAULT_MIN_DISCOVERY_IN_SECS {
		glog.Warningf("%s discovery interval value of %d is below minimum value allowed."+
			" Setting discovery interval to minimum allowed value of %d seconds.",
			discoveryType, secondaryDiscoverySec, pkg.DEFAULT_MIN_DISCOVERY_IN_SECS)
		return pkg.DEFAULT_MIN_DISCOVERY_IN_SECS
	}

	return secondaryDiscoverySec
}

func checkFullRediscoveryInterval(rediscoveryIntervalSec int32) int32 {
	if rediscoveryIntervalSec <= 0 {
		glog.V(3).Infof("No rediscovery interval specified. Using a default value of %d seconds",
			pkg.DEFAULT_FULL_DISCOVERY_IN_SECS)
		return pkg.DEFAULT_FULL_DISCOVERY_IN_SECS
	}

	if rediscoveryIntervalSec < pkg.DEFAULT_MIN_DISCOVERY_IN_SECS {
		glog.Warningf("Rediscovery interval value of %d is below minimum value allowed."+
			" Setting full rediscovery interval to minimum allowed value of %d seconds.",
			rediscoveryIntervalSec, pkg.DEFAULT_MIN_DISCOVERY_IN_SECS)
		return pkg.DEFAULT_MIN_DISCOVERY_IN_SECS
	}

	return rediscoveryIntervalSec
}

// IActionPolicyProvider provides the policies for action types supported for
// different entity types by the probe
type IActionPolicyProvider interface {
	GetActionPolicy() []*proto.ActionPolicyDTO
}

// IEntityMetadataProvider provides the metadata used to generate the unique identifier for
// entities discovered by a probe
type IEntityMetadataProvider interface {
	GetEntityMetadata() []*proto.EntityIdentityMetadata
}
