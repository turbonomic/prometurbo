package dtofactory

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/conf"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// BusinessAppInfo defines the following information discovered from a specific source
//   - A map of business transaction metrics by transaction path
//   - A map of service DTOs by service name
type BusinessAppInfo struct {
	Services     map[string]*proto.EntityDTO
	Transactions map[string]*exporter.EntityMetric
}

func NewBusinessAppInfo() *BusinessAppInfo {
	return &BusinessAppInfo{
		Services:     make(map[string]*proto.EntityDTO),
		Transactions: make(map[string]*exporter.EntityMetric),
	}
}

type BusinessAppInfoBySource map[string]*BusinessAppInfo

type businessAppBuilder struct {
	scope              string
	bizAppInfoBySource BusinessAppInfoBySource      // Business application metrics by source
	bizAppConfBySource conf.BusinessAppConfBySource // Business application configuration by source
}

func NewBusinessAppBuilder(scope string, bizAppInfoBySource BusinessAppInfoBySource,
	bizAppConfBySource conf.BusinessAppConfBySource) *businessAppBuilder {
	return &businessAppBuilder{
		scope:              scope,
		bizAppInfoBySource: bizAppInfoBySource,
		bizAppConfBySource: bizAppConfBySource,
	}
}

func (b *businessAppBuilder) Build() ([]*proto.EntityDTO, error) {
	var bizAppDTOs []*proto.EntityDTO
	for source, bizAppConfByName := range b.bizAppConfBySource {
		bizAppInfo, ok := b.bizAppInfoBySource[source]
		if !ok {
			glog.Warningf("No metrics are discovered from source %v", source)
			continue
		}
		if len(bizAppInfo.Services) == 0 {
			glog.Warningf("No services are discovered from source %v", source)
		}
		// Create business application and business transactions from a specific source
		for bizAppName, bizAppConf := range bizAppConfByName {
			// 1. Create business transaction DTOs
			bizTranDTOs, err := NewBusinessTransactionBuilder(
				b.scope, bizAppName, source, bizAppInfo, bizAppConf.Transactions).Build()
			if err != nil {
				glog.Errorf("Failed to create business transactions for "+
					"business application %v from source %v: %v", bizAppName, source, err)
			}
			if len(bizTranDTOs) > 0 {
				bizAppDTOs = append(bizAppDTOs, bizTranDTOs...)
			}
			// 2. Create business application DTO
			// 2.1 Create business application DTO builder
			bizAppID := bizAppName + "-" + source
			bizAppDtoBuilder := builder.
				NewEntityDTOBuilder(proto.EntityDTO_BUSINESS_APPLICATION,
					getEntityId(proto.EntityDTO_BUSINESS_APPLICATION, b.scope, bizAppID)).
				DisplayName(bizAppName).
				Monitored(false)
			// 2.2 Create service providers
			for _, serviceName := range bizAppConf.Services {
				// From the list of defined services of this business application
				if serviceDTO, found := bizAppInfo.Services[serviceName]; found {
					// Found the service DTO created from appmetric
					provider := builder.CreateProvider(proto.EntityDTO_SERVICE, *serviceDTO.Id)
					bizAppDtoBuilder.Provider(provider).BuysCommodities(serviceDTO.CommoditiesSold)
				}
			}
			// 2.3 Create business transaction providers
			for _, bizTran := range bizTranDTOs {
				provider := builder.CreateProvider(proto.EntityDTO_BUSINESS_TRANSACTION, *bizTran.Id)
				bizAppDtoBuilder.Provider(provider).BuysCommodities(bizTran.CommoditiesSold)
			}
			// 2.4 Create the business transaction DTO
			bizAppDto, err := bizAppDtoBuilder.Create()
			if err != nil {
				return nil, err
			}
			glog.V(4).Infof("Business Application DTO: %+v", bizAppDto)
			bizAppDTOs = append(bizAppDTOs, bizAppDto)
		}
	}
	return bizAppDTOs, nil
}
