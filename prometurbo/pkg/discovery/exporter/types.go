package exporter

import "github.com/turbonomic/turbo-go-sdk/pkg/proto"

const (
	Used     = "used"
	Capacity = "capacity"

	TargetAddress = "target_address"
	Scope         = "scope"
)

type EntityMetric struct {
	UID        string                                                  `json:"uid,omitempty"`
	Type       proto.EntityDTO_EntityType                              `json:"type,omitempty"`
	Labels     map[string]string                                       `json:"labels,omitempty"`
	Metrics    map[proto.CommodityDTO_CommodityType]map[string]float64 `json:"metrics,omitempty"`
	HostedOnVM bool                                                    `json:"hostedOnVM"`
	Source     string                                                  `json:"source"`
}

type MetricResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message:omitemtpy"`
	Data    []*EntityMetric `json:"data:omitempty"`
}
