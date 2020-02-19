package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/conf"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"strings"
)

type bizTranBuilder struct {
	scope        string
	bizAppName   string
	source       string
	bizAppInfo   *BusinessAppInfo   // Business application metrics
	transactions []conf.Transaction // Business transaction configuration
}

func NewBusinessTransactionBuilder(scope, bizAppName, source string,
	bizAppInfo *BusinessAppInfo, transactions []conf.Transaction) *bizTranBuilder {
	return &bizTranBuilder{
		scope:        scope,
		bizAppName:   bizAppName,
		source:       source,
		bizAppInfo:   bizAppInfo,
		transactions: transactions,
	}
}

func (b *bizTranBuilder) Build() ([]*proto.EntityDTO, error) {
	var bizTranDTOs []*proto.EntityDTO
	for _, transaction := range b.transactions {
		services := transaction.DependOn
		// 1. Create business transaction DTO builder
		displayName := transaction.Name
		if displayName == "" {
			displayName = transaction.Path
		}
		bizTranName := strings.Join([]string{b.bizAppName,
			strings.TrimLeft(displayName, "/")}, "/")
		bizTranID := bizTranName + "-" + b.source
		bizTranDTOBuilder := builder.
			NewEntityDTOBuilder(proto.EntityDTO_BUSINESS_TRANSACTION,
				getEntityId(proto.EntityDTO_BUSINESS_TRANSACTION, b.scope, bizTranID)).
			DisplayName(bizTranName).
			Monitored(false)
		// 2. Create bought commodities from service providers
		serviceDTOs := b.bizAppInfo.Services
		var providerDTOs []*proto.EntityDTO
		for _, service := range services {
			if serviceDTO, ok := serviceDTOs[service]; ok {
				providerDTOs = append(providerDTOs, serviceDTO)
			}
		}
		if len(providerDTOs) == 0 {
			return nil, fmt.Errorf("no metrics are discovered for provider "+
				"services %v configured for transaction %v",
				strings.Join(services, ","), bizTranName)
		}
		for _, providerDTO := range providerDTOs {
			bizTranDTOBuilder.
				Provider(builder.CreateProvider(proto.EntityDTO_SERVICE, *providerDTO.Id)).
				BuysCommodities(providerDTO.CommoditiesSold)
		}
		// 3. Create sold commodities
		commoditiesSold := []*proto.CommodityDTO{createAppCommodity()}
		if transactionMetric, ok := b.bizAppInfo.Transactions[transaction.Path]; ok {
			commodities, err := createCommodities(transactionMetric, "")
			if err != nil {
				return nil, fmt.Errorf("failed to create commodities sold for business "+
					"transaction %v discovered from %v: %v", bizTranName, b.source, err)
			}
			commoditiesSold = append(commoditiesSold, commodities...)
		}
		bizTranDTOBuilder.SellsCommodities(commoditiesSold)
		// 4. Create the business transaction DTO
		bizTranDTO, err := bizTranDTOBuilder.Create()
		if err != nil {
			return nil, err
		}
		glog.V(4).Infof("Business Transaction DTO: %+v", bizTranDTO)
		bizTranDTOs = append(bizTranDTOs, bizTranDTO)
	}
	return bizTranDTOs, nil
}
