package alligator

import (
	"github.com/golang/glog"

	"github.com/turbonomic/prometurbo/appmetric/pkg/inter"
	"github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
)

type EntityMetricGetter interface {
	GetEntityMetric(client *prometheus.RestClient) ([]*inter.EntityMetric, error)
	Name() string
}

// Alligator: aggregates several kinds of Entity metric getters
type Alligator struct {
	pclient *prometheus.RestClient
	Getters map[string]EntityMetricGetter
}

func NewAlligator(pclient *prometheus.RestClient) *Alligator {
	result := &Alligator{
		pclient: pclient,
		Getters: make(map[string]EntityMetricGetter),
	}

	return result
}

func (c *Alligator) AddGetter(getter EntityMetricGetter) bool {
	name := getter.Name()
	if _, exist := c.Getters[name]; exist {
		glog.Errorf("Entity Metric Getter: %v already exists", name)
		return false
	}

	c.Getters[name] = getter
	return true
}

func (c *Alligator) GetEntityMetrics() ([]*inter.EntityMetric, error) {
	result := []*inter.EntityMetric{}
	for _, getter := range c.Getters {
		dat, err := getter.GetEntityMetric(c.pclient)
		if err != nil {
			glog.Errorf("Failed to get entity metrics: %v", err)
			continue
		}

		result = append(result, dat...)
	}

	return result, nil
}
