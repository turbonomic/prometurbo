package inter

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

//Labels
const (
	IP       = "ip"
	Port     = "port"
	Name     = "name"
	Category = "category"

	AppEntity  = proto.EntityDTO_APPLICATION
	VAppEntity = proto.EntityDTO_VIRTUAL_APPLICATION

	LatencyType = proto.CommodityDTO_RESPONSE_TIME
	TpsType     = proto.CommodityDTO_TRANSACTION
)
