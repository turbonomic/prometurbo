package probe

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg"
)

// The configuration for Probe
// probeCategory - information about probe category, used for categorizing the probe in UI
// probeType - information about probe type
// discoveryMetadata - information about the discovery intervals for the probe
type ProbeConfig struct {
	ProbeType         string
	ProbeCategory     string
	discoveryMetadata *DiscoveryMetadata
}

// Create new instance of ProbeConfig.
// Sets default discovery intervals for the full, incremental and performance discoveries.
// Returns error if the probe type and category fields cannot be validated.
//
func NewProbeConfig(probeType string, probeCategory string) (*ProbeConfig, error) {
	if probeType == "" {
		return nil, ErrorInvalidProbeType()
	}

	if probeCategory == "" {
		return nil, ErrorInvalidProbeCategory()
	}

	probeConf := &ProbeConfig{
		ProbeCategory:     probeCategory,
		ProbeType:         probeType,
		discoveryMetadata: NewDiscoveryMetadata(),
	}

	return probeConf, nil
}

// Validate the probe config instance
// Returns error if the probe type and category fields cannot be validated.
// Sets default discovery intervals for the full, incremental and performance discoveries.
func (probeConfig *ProbeConfig) Validate() error {
	if probeConfig.ProbeType == "" {
		return ErrorInvalidProbeType()
	}

	if probeConfig.ProbeCategory == "" {
		return ErrorInvalidProbeCategory()
	}

	if probeConfig.discoveryMetadata == nil {
		probeConfig.discoveryMetadata = NewDiscoveryMetadata()
	}

	return nil
}

// Sets the discovery metadata with intervals for the full, incremental and performance discoveries.
func (probeConfig *ProbeConfig) SetDiscoveryMetadata(discoveryMetadata *DiscoveryMetadata) {
	// validate the discovery intervals
	checkRediscoveryIntervalValidity(discoveryMetadata.fullDiscovery,
		discoveryMetadata.incrementalDiscovery,
		discoveryMetadata.performanceDiscovery)
	probeConfig.discoveryMetadata = discoveryMetadata
}

func checkRediscoveryIntervalValidity(rediscoveryIntervalSec,
	incrementalDiscoverySec,
	performanceDiscoverySec int32) {

	if performanceDiscoverySec >= rediscoveryIntervalSec {
		glog.Warning(discoveryConfigError("performance", "full"))
	}

	if incrementalDiscoverySec >= rediscoveryIntervalSec {
		glog.Warning(discoveryConfigError("incremental", "full"))
	}

	if incrementalDiscoverySec >= performanceDiscoverySec &&
		performanceDiscoverySec != pkg.DISCOVERY_NOT_SUPPORTED {
		glog.Warning(discoveryConfigError("incremental", "performance"))
	}
}

func discoveryConfigError(discoveryType1, discoveryType2 string) string {
	return fmt.Sprintf("%s rediscovery interval is greater than %s rediscovery interval, "+
		"will be skipped!", discoveryType1, discoveryType2)
}
