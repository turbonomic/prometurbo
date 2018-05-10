package registration

import (
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	TargetIdField string = "targetIdentifier"
	ProbeCategory string = "Cloud Native"
	TargetType    string = "Prometheus"
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
