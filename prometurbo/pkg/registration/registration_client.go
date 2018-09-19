package registration

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"hash/fnv"
)

const (
	TargetIdField string = "targetIdentifier"
	ProbeCategory string = "Cloud Native"
	targetType    string = "Prometheus"
	Scope         string = "Scope"
)

// Implements the TurboRegistrationClient interface
type P8sRegistrationClient struct {
}

func (p *P8sRegistrationClient) GetSupplyChainDefinition() []*proto.TemplateDTO {
	glog.Infoln("Building a supply chain ..........")

	supplyChainFactory := &SupplyChainFactory{}
	templateDtos, err := supplyChainFactory.CreateSupplyChain()
	if err != nil {
		glog.Error("Error creating Supply chain for Prometurbo")
		return nil
	}
	glog.Infoln("Supply chain for Prometurbo is created.")
	return templateDtos
}

func (p *P8sRegistrationClient) GetIdentifyingFields() string {
	return TargetIdField
}

func (p *P8sRegistrationClient) GetAccountDefinition() []*proto.AccountDefEntry {

	targetIDAcctDefEntry := builder.NewAccountDefEntryBuilder(TargetIdField, "URL",
		"URL of the Prometheus target", ".*", true, false).Create()

	scopeAcctDefEntry := builder.NewAccountDefEntryBuilder(Scope, Scope,
		"The associated target name (e.g., Kubernetes target)", ".*", false, false).Create()

	return []*proto.AccountDefEntry{
		targetIDAcctDefEntry,
		scopeAcctDefEntry,
	}
}

// Return the target type as the default target type appended with hash number from target Id
func TargetType(targetId string) string {
	return appendRandomName(targetType, targetId)
}

func appendRandomName(name, append string) string {
	return name + "-" + fmt.Sprint(hash(append))
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
