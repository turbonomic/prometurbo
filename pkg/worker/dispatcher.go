package worker

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

type Dispatcher struct {
	workerCount int
	workerPool  chan chan ITask
	collector   *Collector
}

func NewDispatcher(workerCount int) *Dispatcher {
	return &Dispatcher{
		workerCount: workerCount,
		workerPool:  make(chan chan ITask, workerCount),
	}
}

func (d *Dispatcher) WithCollector(collector *Collector) *Dispatcher {
	d.collector = collector
	return d
}

func (d *Dispatcher) Start() {
	// Create workers
	for i := 0; i < d.workerCount; i++ {
		// Create and launch a worker in a separate goroutine
		go d.launchWorker(fmt.Sprintf("%d", i))
	}
}

func (d *Dispatcher) launchWorker(id string) {
	worker := newWorker(id)
	// Put the worker into the pool
	d.workerPool <- worker.taskChan
	for {
		// Wait for a task to appear on the channel
		select {
		case t := <-worker.taskChan:
			glog.V(2).Infof("worker %s has received a task.", worker.id)
			result := worker.execute(t)
			d.collector.resultPool <- result
			glog.V(2).Infof("worker %s has finished.", worker.id)
			d.workerPool <- worker.taskChan
		}
	}
}

// Dispatch a task, block when there is no free worker
func (d *Dispatcher) Dispatch(t ITask) {
	glog.V(4).Infof("Waiting for a free worker")
	// Pick a free worker from the worker pool, when its channel frees up
	taskChannel := <-d.workerPool
	// Assign a task to the worker
	taskChannel <- t
}

// Collect results from this round of discovery
func (d *Dispatcher) CollectResult(taskCount int) []*data.DIFEntity {
	return d.collector.collect(taskCount)
}
