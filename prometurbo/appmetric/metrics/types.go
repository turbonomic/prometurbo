package metrics

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	Used     = "used"
	Capacity = "capacity"
)

type EntityMetric struct {
	UID     string                                                  `json:"uid"`
	Type    proto.EntityDTO_EntityType                              `json:"type,omitempty"`
	Labels  map[string]string                                       `json:"labels,omitempty"`
	Metrics map[proto.CommodityDTO_CommodityType]map[string]float64 `json:"metrics,omitempty"`
}

type MetricResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message:omitemtpy"`
	Data    []*EntityMetric `json:"data:omitempty"`
}

func NewEntityMetric(id string, t proto.EntityDTO_EntityType) *EntityMetric {
	m := &EntityMetric{
		UID:     id,
		Type:    t,
		Labels:  make(map[string]string),
		Metrics: make(map[proto.CommodityDTO_CommodityType]map[string]float64),
	}

	return m
}

func (e *EntityMetric) SetLabel(name, value string) {
	e.Labels[name] = value
}

func (e *EntityMetric) SetMetric(cname proto.CommodityDTO_CommodityType, kind string, value float64) {
	if _, ok := e.Metrics[cname]; !ok {
		e.Metrics[cname] = map[string]float64{}
	}
	e.Metrics[cname][kind] = value
}

func NewMetricResponse() *MetricResponse {
	return &MetricResponse{
		Status:  0,
		Message: "",
		Data:    []*EntityMetric{},
	}
}

func (r *MetricResponse) SetStatus(v int, msg string) {
	r.Status = v
	r.Message = msg
}

func (r *MetricResponse) SetMetrics(dat []*EntityMetric) {
	r.Data = dat
}

func (r *MetricResponse) AddMetric(m *EntityMetric) {
	r.Data = append(r.Data, m)
}
