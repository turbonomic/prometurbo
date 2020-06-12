package topology

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/config"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

type BusinessTopology struct {
	BizAppConfBySource config.BusinessAppConfBySource
}

func (t *BusinessTopology) BuildTopologyEntities(entities []*data.DIFEntity) []*data.DIFEntity {
	topologyEntities := entities
	filteredEntities, transMap := buildTransMap(topologyEntities)
	svcMap := buildSvcMap(filteredEntities)
	for _, svcEntities := range svcMap {
		for _, svcEntity := range svcEntities {
			filteredEntities = append(filteredEntities, svcEntity)
		}
	}
	bizEntities := t.buildBizDIFEntities(svcMap, transMap)
	if bizEntities != nil {
		glog.Infof("Number of business entities: %d.", len(bizEntities))
		filteredEntities = append(filteredEntities, bizEntities...)
	}
	return filteredEntities
}

func buildTransMap(entities []*data.DIFEntity) ([]*data.DIFEntity, map[string]*data.DIFEntity) {
	var filteredEntites []*data.DIFEntity
	transMap := make(map[string]*data.DIFEntity)
	for _, entity := range entities {
		if entity.Type != "businessTransaction" {
			filteredEntites = append(filteredEntites, entity)
			continue
		}
		transMap[entity.UID] = entity
	}
	return filteredEntites, transMap
}

func buildSvcMap(entities []*data.DIFEntity) map[string][]*data.DIFEntity {
	svcMap := make(map[string][]*data.DIFEntity)
	for _, entity := range entities {
		if entity.Type != "databaseServer" &&
			entity.Type != "application" {
			continue
		}
		// Only create service entities from application and databaseServer
		ServicePrefix := "Service-"
		svcID := ServicePrefix + entity.UID
		svc := data.NewDIFEntity(svcID, "service")
		for meType, meList := range entity.Metrics {
			svc.AddMetrics(meType, meList)
		}
		hostedOnTypes := entity.GetHostedOnType()
		if len(hostedOnTypes) != 1 ||
			hostedOnTypes[0] != data.VM {
			// The application is not hosted on VM, assume it is hosted on
			// container, so create a proxy service (i.e., with stitching attribute)
			svc.Matching(svcID)
		}
		// Add the service to the service map
		for _, partOf := range entity.PartOf {
			svcName := partOf.Label
			svcMap[svcName] = append(svcMap[svcName], svc)
		}
		glog.V(2).Infof("Service entity: %v", svc)
	}
	return svcMap
}

func (t *BusinessTopology) buildBizDIFEntities(svcMap map[string][]*data.DIFEntity,
	transMap map[string]*data.DIFEntity) []*data.DIFEntity {
	var bizEntities []*data.DIFEntity
	for source, bizAppConfByName := range t.BizAppConfBySource {
		for name, bizAppConf := range bizAppConfByName {
			glog.V(4).Infof("Source %s Name %s BizApp %v", source, name, bizAppConf)
			bizAppId := fmt.Sprintf("%s-%s", bizAppConf.Name, bizAppConf.From)
			for _, service := range bizAppConf.Services {
				svcEntities, exists := svcMap[service]
				if !exists {
					// Skip services that are configured but don't have metrics
					continue
				}
				for _, svcEntity := range svcEntities {
					svcEntity.PartOfEntity("businessApplication", bizAppId, "")
				}
			}
			for _, trans := range bizAppConf.Transactions {
				for _, service := range trans.DependOn {
					svcEntities, exists := svcMap[service]
					if !exists {
						// Skip services that are configured but don't have metrics
						continue
					}
					for _, svcEntity := range svcEntities {
						svcEntity.PartOfEntity("businessTransaction", trans.Path, "")
					}
				}
				bizTransEntity := bizTransToDIFEntity(trans, bizAppId)
				if bizTransEntityDiscovered, found := transMap[trans.Path]; found {
					bizTransEntityDiscovered.PartOf = bizTransEntity.PartOf
					bizEntities = append(bizEntities, bizTransEntityDiscovered)
				} else {
					bizEntities = append(bizEntities, bizTransEntity)
				}
			}
			bizAppEntity := bizAppToDIFEntity(bizAppConf)
			bizEntities = append(bizEntities, bizAppEntity)
		}
	}
	return bizEntities
}

func bizAppToDIFEntity(bizApp *config.BusinessApplication) *data.DIFEntity {
	bizAppDIFEntity := data.NewDIFEntity(fmt.Sprintf("%s-%s", bizApp.Name, bizApp.From),
		"businessApplication").
		WithName(bizApp.Name)
	glog.V(4).Infof("Creating business app entity %v", bizAppDIFEntity)
	return bizAppDIFEntity
}

func bizTransToDIFEntity(bizTrans config.Transaction, bizApp string) *data.DIFEntity {
	name := bizTrans.Name
	if name == "" {
		name = bizTrans.Path
	}
	bizTransDIFEntity := data.NewDIFEntity(bizTrans.Path, "businessTransaction").
		WithName(name).
		PartOfEntity("businessApplication", bizApp, "")
	glog.V(4).Infof("Creating business transaction entity %v", bizTransDIFEntity)
	return bizTransDIFEntity
}
