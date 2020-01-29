package constant

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	// MetricType
	Used = "used"

	// Internal matching property
	// The default namespace of entity property
	DefaultPropertyNamespace = "DEFAULT"

	// The attribute used for stitching with other probes (e.g., prometurbo) with app and vapp
	StitchingAttr string = "IP"

	// External matching property
	// The attribute used for stitching with other probes (e.g., prometurbo) with vm
	SUPPLY_CHAIN_CONSTANT_IP_ADDRESS string = "ipAddress"
	SUPPLY_CHAIN_CONSTANT_VIRTUAL_MACHINE_DATA = "virtual_machine_data"

	VAppPrefix = "vApp-"
	BizAppPrefix = "businessApp-"
)

type DefaultValue struct {
	Capacity float64
}

var EntityTypeMap = map[proto.EntityDTO_EntityType]struct{}{
	proto.EntityDTO_APPLICATION: {},
}

var VMCommodityTypeMap = map[proto.CommodityDTO_CommodityType]struct{}{
	proto.CommodityDTO_VCPU:   {},
	proto.CommodityDTO_VMEM: {},
}

var AppCommodityTypeMap = map[proto.CommodityDTO_CommodityType]DefaultValue{
	proto.CommodityDTO_TRANSACTION:   {},
	proto.CommodityDTO_RESPONSE_TIME: {},
	proto.CommodityDTO_VCPU: {},
	proto.CommodityDTO_VMEM: {},
	proto.CommodityDTO_COLLECTION_TIME: {100.0},
	proto.CommodityDTO_HEAP: {},
	proto.CommodityDTO_THREADS: {},
}

var DBCommodityTypeMap = map[proto.CommodityDTO_CommodityType]DefaultValue{
	proto.CommodityDTO_DB_CACHE_HIT_RATE: {100.0},
	proto.CommodityDTO_DB_MEM: {},
	proto.CommodityDTO_CONNECTION: {},
}