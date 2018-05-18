package pkg

type DiscoveryType string

const (
	DEFAULT_FULL_DISCOVERY_IN_SECS int32 = 600
	DEFAULT_MIN_DISCOVERY_IN_SECS  int32 = 60
	DISCOVERY_NOT_SUPPORTED        int32 = -1

	FULL_DISCOVERY        DiscoveryType = "Full"
	INCREMENTAL_DISCOVERY DiscoveryType = "Incremental"
	PERFORMANCE_DISCOVERY DiscoveryType = "Performance"
)
