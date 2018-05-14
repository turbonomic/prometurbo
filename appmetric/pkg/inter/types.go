package inter

type EntityMetric struct {
	UID     string             `json:"uid"`
	Type    int32              `json:"type,omitempty"`
	Labels  map[string]string  `json:"labels,omitempty"`
	Metrics map[string]float64 `json:"metrics,omitempty"`
}

func NewEntityMetric(id string, t int32) *EntityMetric {
	m := &EntityMetric{
		UID:     id,
		Type:    t,
		Labels:  make(map[string]string),
		Metrics: make(map[string]float64),
	}

	return m
}

func (e *EntityMetric) SetLabel(name, value string) {
	e.Labels[name] = value
}

func (e *EntityMetric) SetMetric(name string, value float64) {
	e.Metrics[name] = value
}

type MetricResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message:omitemtpy"`
	Data    []*EntityMetric `json:"data:omitempty"`
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
