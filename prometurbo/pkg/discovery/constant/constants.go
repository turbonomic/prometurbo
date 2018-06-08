package constant

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	// EntityType
	ApplicationType = int32(1)

	// CommodityType
	TPS     = "tps"
	Latency = "latency"

	// MetricType
	Used     = "used"
	Capacity = "capacity"

	// Capacity
	TPSCap     = 20.0
	LatencyCap = 500.0 //millisec

	// The default namespace of entity property
	DefaultPropertyNamespace = "DEFAULT"

	// The attribute used for stitching with other probes (e.g., prometurbo) with app and vapp
	StitchingAttr string = "IP"

	VAppPrefix = "vApp-"
)

var EntityTypeMap = map[proto.EntityDTO_EntityType]struct{}{
	proto.EntityDTO_APPLICATION: {},
}

var CommodityTypeMap = map[proto.CommodityDTO_CommodityType]struct{}{
	proto.CommodityDTO_TRANSACTION:   {},
	proto.CommodityDTO_RESPONSE_TIME: {},
}

var CommodityCapMap = map[proto.CommodityDTO_CommodityType]float64{
	proto.CommodityDTO_TRANSACTION:   TPSCap,
	proto.CommodityDTO_RESPONSE_TIME: LatencyCap,
}
