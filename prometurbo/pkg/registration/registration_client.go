package registration

import (
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	TargetIdField string = "targetIdentifier"
	ProbeCategory string = "Guest OS Processes"
	targetType    string = "Prometheus"
	Scope         string = "Scope"
	propertyId    string = "id"
)

// Implements the TurboRegistrationClient interface
type P8sRegistrationClient struct {
	TargetTypeSuffix string
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

// TargetType returns the target type as the default target type appended
// an optional (from configuration) suffix
func (p *P8sRegistrationClient) TargetType() string {
	if len(p.TargetTypeSuffix) == 0 {
		return targetType
	}
	return targetType + "-" + p.TargetTypeSuffix
}

func (p *P8sRegistrationClient) GetEntityMetadata() []*proto.EntityIdentityMetadata {
	glog.V(3).Infof("Begin to build EntityIdentityMetadata")

	var result []*proto.EntityIdentityMetadata

	entities := []proto.EntityDTO_EntityType{
		proto.EntityDTO_VIRTUAL_MACHINE,
		proto.EntityDTO_APPLICATION_COMPONENT,
		proto.EntityDTO_SERVICE,
		proto.EntityDTO_BUSINESS_APPLICATION,
		proto.EntityDTO_BUSINESS_TRANSACTION,
		proto.EntityDTO_DATABASE_SERVER,
	}

	for _, etype := range entities {
		meta := p.newIdMetaData(etype, []string{propertyId})
		result = append(result, meta)
	}

	glog.V(4).Infof("EntityIdentityMetaData: %++v", result)

	return result
}

func (p *P8sRegistrationClient) newIdMetaData(etype proto.EntityDTO_EntityType, names []string) *proto.EntityIdentityMetadata {
	var data []*proto.EntityIdentityMetadata_PropertyMetadata
	for _, name := range names {
		dat := &proto.EntityIdentityMetadata_PropertyMetadata{
			Name: &name,
		}
		data = append(data, dat)
	}

	result := &proto.EntityIdentityMetadata{
		EntityType:            &etype,
		NonVolatileProperties: data,
	}

	return result
}
