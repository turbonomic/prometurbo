package exporter

type EntityMetric struct {
	UID     string             `json:"uid,omitempty"`
	Type    int32              `json:"type,omitempty"`
	Labels  map[string]string  `json:"labels,omitempty"`
	Metrics map[string]float64 `json:"metrics,omitempty"`
}

type MetricResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message:omitemtpy"`
	Data    []*EntityMetric `json:"data:omitempty"`
}
