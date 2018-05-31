package inter

//Labels
const (
	IP       = "ip"
	Port     = "port"
	Name     = "name"
	Category = "category"

	// These two should be Commodity, but their values are strings
	ClusterMetric = "cluster"

	// its value should be an array of strings
	VMPMAccessMetric = "vmpm_access"
)

type EntityType int32
type MetricType string

// Entity Types
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

// Commodity Types
// Seller has Capacity and Usage, Buyer has Usage and Reservation;
// So only the Usage needs to distinguish buy or sell.
const (
	// transaction/request per second
	TPSSoldMetric     MetricType = "tps"
	TPSBoughtMetric   MetricType = "tps_buy"
	TPSCapacityMetric MetricType = "tps_capacity"

	// unit of latency is millisecond
	LatencySoldMetric     MetricType = "latency"
	LatencyBoughtMetric   MetricType = "latency_buy"
	LatencyCapacityMetric MetricType = "latency_capacity"

	CPUSoldMetric     MetricType = "cpu"
	CPUBoughtMetric   MetricType = "cpu_buy"
	CPUCapacityMetric MetricType = "cpu_capacity"
	CPUReserveMetric  MetricType = "cpu_rev"

	MemSoldMetric     MetricType = "mem"
	MemBoughtMetric   MetricType = "mem_buy"
	MemCapacityMetric MetricType = "mem_capacity"
	MemReserveMetric  MetricType = "mem_rev"
)
