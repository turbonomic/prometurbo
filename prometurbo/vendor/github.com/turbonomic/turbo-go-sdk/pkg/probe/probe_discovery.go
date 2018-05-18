package probe

import "github.com/turbonomic/turbo-go-sdk/pkg/proto"

// Interface for incremental discovery.
// External probes which want to support incremental discovery should implement
// this interface in order for the Mediation Container to call it.
type IIncrementalDiscovery interface {
	//Discovers the target, creating EntityDTO representation of changed objects.
	// @param accountValues object, holding all the account values
	// @return discovery response, only entities changed after the previous full/incremental discovery
	DiscoverIncremental(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error)
}

// Interface for performance discovery.
// External probes which want to support performance discovery should implement
// this interface in order for the Mediation Container to call it.
type IPerformanceDiscovery interface {
	//Discovers the target, creating EntityDTOs with associated commodities used/capacity values.
	//@param accountValues object, holding all the account values
	// @return discovery response, EntityDTOs with associated commodities used/capacity values
	DiscoverPerformance(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error)
}
