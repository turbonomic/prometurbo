package worker

import (
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"sync"
)

type Collector struct {
	resultPool chan []*data.DIFEntity
}

func NewCollector(maxWorkerNumber int) *Collector {
	return &Collector{
		resultPool: make(chan []*data.DIFEntity, maxWorkerNumber),
	}
}

func (m *Collector) collect(count int) (mergedResult []*data.DIFEntity) {
	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(count)
	go func() {
		for {
			select {
			case <-stopChan:
				return
			case result := <-m.resultPool:
				mergedResult = append(mergedResult, result...)
				wg.Done()
			}
		}
	}()
	wg.Wait()
	// stop the result waiting goroutine.
	close(stopChan)
	glog.V(2).Infof("Collected results from all %d tasks.", count)
	return
}
