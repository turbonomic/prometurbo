package data

import "github.com/turbonomic/turbo-go-sdk/pkg/proto"

// USING the common DIF Data
var DIFEntityType = map[proto.EntityDTO_EntityType]string{
	proto.EntityDTO_VIRTUAL_MACHINE:       "virtualMachine",
	proto.EntityDTO_APPLICATION_COMPONENT: "application",
	proto.EntityDTO_BUSINESS_APPLICATION:  "businessApplication",
	proto.EntityDTO_BUSINESS_TRANSACTION:  "businessTransaction",
	proto.EntityDTO_DATABASE_SERVER:       "databaseServer",
	proto.EntityDTO_SERVICE:               "service",
}

type DIFHostType string

const (
	VM        DIFHostType = "virtualMachine"
	CONTAINER DIFHostType = "container"
)

var DIFMetricType = map[proto.CommodityDTO_CommodityType]string{
	proto.CommodityDTO_RESPONSE_TIME:     "responseTime",
	proto.CommodityDTO_TRANSACTION:       "transaction",
	proto.CommodityDTO_VCPU:              "cpu",
	proto.CommodityDTO_VMEM:              "memory",
	proto.CommodityDTO_THREADS:           "threads",
	proto.CommodityDTO_HEAP:              "heap",
	proto.CommodityDTO_COLLECTION_TIME:   "collectionTime",
	proto.CommodityDTO_DB_MEM:            "dbMem",
	proto.CommodityDTO_DB_CACHE_HIT_RATE: "dbCacheHitRate",
	proto.CommodityDTO_CONNECTION:        "connection",
	proto.CommodityDTO_KPI:               "kpi",
}
