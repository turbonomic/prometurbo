package supplychain

import "github.com/turbonomic/turbo-go-sdk/pkg/proto"

const (
	getIPAddressMethodName string = "getAddress"

	serverIPProperty   = "UsesEndPoints"
	serverUUIDProperty = "Uuid"
)

func getIPHandler() *proto.PropertyHandler {
	directlyApply := false
	ipEntityType := proto.EntityDTO_IP
	methodName := getIPAddressMethodName

	return &proto.PropertyHandler{
		MethodName:    &methodName,
		DirectlyApply: &directlyApply,
		EntityType:    &ipEntityType,
	}
}

var VM_IP *proto.ServerEntityPropDef = getVirtualMachineIPProperty()

func getVirtualMachineIPProperty() *proto.ServerEntityPropDef {
	attribute := serverIPProperty
	vmEntityType := proto.EntityDTO_VIRTUAL_MACHINE

	return &proto.ServerEntityPropDef{
		Entity:          &vmEntityType,
		Attribute:       &attribute,
		PropertyHandler: getIPHandler(),
	}
}

var VM_UUID *proto.ServerEntityPropDef = getVirtualMachineUUIDProperty()

func getVirtualMachineUUIDProperty() *proto.ServerEntityPropDef {
	attribute := serverUUIDProperty
	vmEntityType := proto.EntityDTO_VIRTUAL_MACHINE

	return &proto.ServerEntityPropDef{
		Entity:    &vmEntityType,
		Attribute: &attribute,
	}
}
