package data

import (
	"fmt"
)

type DIFMetric struct {
	MetricMap map[string]*DIFMetricVal
}

type DIFMetricVal struct {
	Average     *float64       `json:"average"`
	Min         *float64       `json:"min,omitempty"`
	Max         *float64       `json:"max,omitempty"`
	Capacity    *float64       `json:"capacity,omitempty"`
	Unit        *DIFMetricUnit `json:"unit,omitempty"`
	Key         *string        `json:"key,omitempty"`
	Resizable   *bool          `json:"resizable,omitempty"`
	Description *string        `json:"description,omitempty"`
	RawMetrics  interface{}    `json:"rawData,omitempty"`
}

type DIFMetricUnit string

const (
	COUNT DIFMetricUnit = "count"
	TPS   DIFMetricUnit = "tps"
	MS    DIFMetricUnit = "ms"
	MB    DIFMetricUnit = "mb"
	MHZ   DIFMetricUnit = "mhz"
	PCT   DIFMetricUnit = "pct"
)

type DIFMetricValKind string

const (
	KEY         DIFMetricValKind = "key"
	DESCRIPTION DIFMetricValKind = "description"
	RAWDATA     DIFMetricValKind = "rawData"
	AVERAGE     DIFMetricValKind = "average"
	MAX         DIFMetricValKind = "max"
	MIN         DIFMetricValKind = "min"
	CAPACITY    DIFMetricValKind = "capacity"
	UNIT        DIFMetricValKind = "unit"
)

const (
	UNSET_FLOAT  = 0.0
	UNSET_STRING = ""
)

func (m *DIFMetricVal) String() string {
	s := ""
	if m.Average != nil {
		s += fmt.Sprintf("Average:%v ", *m.Average)
	}
	if m.Capacity != nil {
		s += fmt.Sprintf("Capacity:%v ", *m.Capacity)
	}
	if m.Unit != nil {
		s += fmt.Sprintf("Unit:%v ", *m.Unit)
	}
	if m.Key != nil {
		s += fmt.Sprintf("Key:%v ", *m.Key)
	}
	if m.Resizable != nil {
		s += fmt.Sprintf("Resizable:%v", *m.Resizable)
	}
	return s
}

// Clone creates a copy of this DIFMetricVal.
func (m *DIFMetricVal) Clone() *DIFMetricVal {
	clone := DIFMetricVal{
		RawMetrics: m.RawMetrics,
	}
	if m.Average != nil {
		average := *m.Average
		clone.Average = &average
	}
	if m.Min != nil {
		minimum := *m.Min
		clone.Min = &minimum
	}
	if m.Max != nil {
		maximum := *m.Max
		clone.Max = &maximum
	}
	if m.Capacity != nil {
		capacity := *m.Capacity
		clone.Capacity = &capacity
	}
	if m.Unit != nil {
		unit := *m.Unit
		clone.Unit = &unit
	}
	if m.Key != nil {
		key := *m.Key
		clone.Key = &key
	}
	if m.Resizable != nil {
		resizable := *m.Resizable
		clone.Resizable = &resizable
	}
	if m.Description != nil {
		description := *m.Description
		clone.Description = &description
	}
	return &clone
}

// Sum adds the values of the other DIFMetricVal to the current one.
// Currently, we only sum up the average and the capacity.
// We also assume Unit/Key/Resizable are the same.
// Note: This method is used to aggregate GPU metrics for a container,
// where a container may use multiple GPUs, but the metrics obtained
// are at per GPU level.
func (m *DIFMetricVal) Sum(other *DIFMetricVal) {
	// Average
	m.sumAverage(other)
	// Capacity
	m.sumCapacity(other)
}

func (m *DIFMetricVal) sumAverage(other *DIFMetricVal) {
	if m.Average == nil {
		m.Average = other.Average
	} else if other.Average != nil {
		newAvg := *m.Average + *other.Average
		m.Average = &newAvg
	}
}

func (m *DIFMetricVal) sumCapacity(other *DIFMetricVal) {
	if m.Capacity == nil {
		m.Capacity = other.Capacity
	} else if other.Capacity != nil {
		newCapacity := *m.Capacity + *other.Capacity
		m.Capacity = &newCapacity
	}
}
