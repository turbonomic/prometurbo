package inter

const (
	//EntityType
	ApplicationType        = int32(1)
	VirtualApplicationType = int32(2)
	VirtualMachineType     = int32(3)

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
