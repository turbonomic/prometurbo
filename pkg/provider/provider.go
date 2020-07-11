package provider

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/worker"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

var metricKindToDIFMetricValKind = map[string]data.DIFMetricValKind{
	Used:     data.AVERAGE,
	Capacity: data.CAPACITY,
}

type MetricProvider struct {
	serverDefs   map[string]*serverDef
	exporterDefs map[string]*exporterDef
	dispatcher   *worker.Dispatcher
}

func NewProvider(serverDefs map[string]*serverDef, exporterDefs map[string]*exporterDef) *MetricProvider {
	return &MetricProvider{
		serverDefs:   serverDefs,
		exporterDefs: exporterDefs,
	}
}

func (p *MetricProvider) WithDispatcher(dispatcher *worker.Dispatcher) *MetricProvider {
	p.dispatcher = dispatcher
	return p
}

func (p *MetricProvider) Start() {
	p.dispatcher.Start()
}

func (p *MetricProvider) GetEntityMetrics() ([]*data.DIFEntity, error) {
	var tasks []*Task
	for _, serverDef := range p.serverDefs {
		for _, exporter := range serverDef.exporters {
			exporterDef, found := p.exporterDefs[exporter]
			if !found {
				continue
			}
			for _, entityDef := range exporterDef.entityDefs {
				// Dispatch the discovery task
				tasks = append(tasks, NewTask(serverDef.promClient, entityDef))
			}
		}
	}
	total := len(tasks)
	glog.V(2).Infof("Total entity types to discover %v", total)
	// Dispatch tasks in a separate goroutine to avoid deadlock
	go func() {
		for _, task := range tasks {
			p.dispatcher.Dispatch(task)
		}
	}()
	// Collect the result
	return p.dispatcher.CollectResult(total), nil
}
