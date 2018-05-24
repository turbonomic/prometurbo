package inter

const (
	//CommodityType
	// transaction/request per second
	TPS = "tps"

	// unit of latency is millisecond
	Latency = "latency"

	//Labels
	IP       = "ip"
	Port     = "port"
	Name     = "name"
	Category = "category"
)

type EntityType int32
type MetricType string

const (
	ApplicationEntity        EntityType = 1
	VirtualApplicationEntity EntityType = 2
	VirtualMachineEntity     EntityType = 3
	PhysicalMachineEntity    EntityType = 4
	ContainerPodEntity       EntityType = 5
	ContainerEntity          EntityType = 6
	ServiceEntity            EntityType = 7
	LoadBalancerEntity       EntityType = 8
)

const (
	TPSSoldMetric     MetricType = "tps_sold"
	LatencySoldMetric MetricType = "latency_sold"
	CPUSoldMetric     MetricType = "cpu_sold"
	CPUBoughtMetric   MetricType = "cpu_buy"
	CPUCapacityMetric MetricType = "cpu_capacity"
)
