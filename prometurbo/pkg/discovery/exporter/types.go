package exporter

import "github.com/turbonomic/turbo-go-sdk/pkg/proto"

type EntityMetric struct {
	UID     string                                       `json:"uid,omitempty"`
	Type    proto.EntityDTO_EntityType                   `json:"type,omitempty"`
	Labels  map[string]string                            `json:"labels,omitempty"`
	Metrics map[proto.CommodityDTO_CommodityType]float64 `json:"metrics,omitempty"`
}

type MetricResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message:omitemtpy"`
	Data    []*EntityMetric `json:"data:omitempty"`
}
