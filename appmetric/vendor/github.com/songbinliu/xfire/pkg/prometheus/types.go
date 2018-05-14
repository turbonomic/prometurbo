package prometheus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/prometheus/common/model"
	"math"
)

// for internal use only
type promeResponse struct {
	Status    string   `json:"status"`
	Data      *RawData `json:"data,omitempty"`
	ErrorType string   `json:"errorType,omitempty"`
	Error     string   `json:"error,omitempty"`
}

type RawData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

// RawMetric the raw metric from Prometheus: its labels and a time/value pair
type RawMetric struct {
	Labels map[string]string `json:"metric"`
	Value  model.SamplePair  `json:"value"`
}

// MetricData : interface to transform the RawMetric to customer defined data structure
type MetricData interface {
	GetValue() float64
}

// RequestInput : interface for customer defined query generator, and RawMetric parser.
type RequestInput interface {
	GetQuery() string
	Parse(metric *RawMetric) (MetricData, error)
}

// -----------------------------------------------------------
// an example implementation of RequestInput and MetricData
type BasicMetricData struct {
	Labels map[string]string
	Value  float64
}

// this BasicInput will copy all the labels from the RawData
type BasicInput struct {
	query string
}

func NewBasicInput() *BasicInput {
	return &BasicInput{}
}

func (input *BasicInput) GetQuery() string {
	return input.query
}

func (input *BasicInput) SetQuery(q string) {
	input.query = q
}

func (input *BasicInput) Parse(m *RawMetric) (MetricData, error) {
	d := NewBasicMetricData()
	if err := d.Parse(m); err != nil {
		glog.Errorf("Failed to parse raw metric: %v", err)
		return nil, err
	}
	return d, nil
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
		return fmt.Errorf("Failed to convert value: NaN")
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
