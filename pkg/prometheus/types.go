package prometheus

import (
	"bytes"
	"fmt"
	"github.com/prometheus/common/model"
	"math"
)

// RawMetric the raw metric from Prometheus: its labels and a time/value pair
type RawMetric struct {
	Labels map[string]string `json:"metric"`
	Value  model.SamplePair  `json:"value"`
}

func (m RawMetric) Parse() (MetricData, error) {
	metricData := NewBasicMetricData()
	for k, v := range m.Labels {
		metricData.Labels[k] = v
	}
	metricData.Value = float64(m.Value.Value)
	if math.IsNaN(metricData.Value) {
		return nil, fmt.Errorf("failed to convert value: NaN")
	}
	return metricData, nil
}

// MetricData is the interface to transform the RawMetric to customer defined data structure
type MetricData interface {
	GetValue() float64
}

// BasicMetricData implements Request and MetricData
type BasicMetricData struct {
	Labels map[string]string
	Value  float64
	ID     string
}

func NewBasicMetricData() *BasicMetricData {
	return &BasicMetricData{
		Labels: make(map[string]string),
	}
}

func (d *BasicMetricData) GetValue() float64 {
	return d.Value
}

func (d *BasicMetricData) Parse(m *RawMetric) error {
	for k, v := range m.Labels {
		d.Labels[k] = v
	}

	d.Value = float64(m.Value.Value)
	if math.IsNaN(d.Value) {
		return fmt.Errorf("failed to convert value: NaN")
	}

	return nil
}

func (d *BasicMetricData) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("value=%.6f\n", d.Value))
	for k, v := range d.Labels {
		buffer.WriteString(fmt.Sprintf("\t%v=%v\n", k, v))
	}
	return buffer.String()
}
