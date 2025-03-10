package provider

type MetricProvider interface {
	GetTasks() (tasks []*Task)
}
