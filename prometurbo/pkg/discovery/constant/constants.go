package constant

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	// MetricType
	Used = "used"

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
