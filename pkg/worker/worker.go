package worker

import "github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"

type ITask interface {
	Run() []*data.DIFEntity
}

type worker struct {
	id       string
	taskChan chan ITask
}

func newWorker(id string) *worker {
	return &worker{
		id:       id,
		taskChan: make(chan ITask),
	}
}

func (w *worker) execute(task ITask) []*data.DIFEntity {
	return task.Run()
}
