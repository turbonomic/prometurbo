package topology

import (
	"fmt"
	set "github.com/deckarep/golang-set"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg/config"
	"github.com/turbonomic/prometurbo/pkg/util"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

type businessAppConfBySource map[string]businessAppConfByName
type businessAppConfByName map[string]*config.BusinessApplication
type serviceMap map[string][]*data.DIFEntity
type transactionMap map[string]*data.DIFEntity
type serviceByNamespace map[string]serviceMap
type transactionByNamespace map[string]transactionMap

type BusinessTopology struct {
	bizAppConfs []config.BusinessApplication
}

func NewBusinessTopology(bizAppConfs []config.BusinessApplication) *BusinessTopology {
	return &BusinessTopology{
		bizAppConfs: bizAppConfs,
	}
}

func (t *BusinessTopology) BuildTopologyEntities(entities []*data.DIFEntity) []*data.DIFEntity {
	// Build a transaction map by namespace and transaction ID needed to create BT entities
	// The topologyEntities so far contain all discovered entities that are not business transactions
	topologyEntities, discoveredTrans := buildTransByNamespace(entities)
	// Create Service entities, and build a service map by namespace and service ID needed to create BT and BA entities
	discoveredSvcs := buildSvcByNamespace(topologyEntities)
	// Add all created Service entities to the topologyEntities
	for namespace, svcMap := range discoveredSvcs {
		for svcName, svcEntities := range svcMap {
			for _, svcEntity := range svcEntities {
				topologyEntities = append(topologyEntities, svcEntity)
			}
			glog.V(2).Infof("Created %v services for %v",
				len(svcEntities), util.GetDisplay(svcName, namespace))
		}
	}
	// Create BT and BA entities from the transaction map and service map
	bizEntities := t.buildBizDIFEntities(discoveredSvcs, discoveredTrans)
	if bizEntities != nil {
		glog.Infof("Number of business entities: %d.", len(bizEntities))
		// All to the final topologyEntities
		topologyEntities = append(topologyEntities, bizEntities...)
	}
	return topologyEntities
}

func buildTransByNamespace(entities []*data.DIFEntity) ([]*data.DIFEntity, transactionByNamespace) {
	var filteredEntites []*data.DIFEntity
	discoveredTrans := make(transactionByNamespace)
	for _, entity := range entities {
		if entity.Type != "businessTransaction" {
			filteredEntites = append(filteredEntites, entity)
			continue
		}
		namespace := entity.GetNamespace()
		transMap, ok := discoveredTrans[namespace]
		if !ok {
			transMap = make(transactionMap)
			discoveredTrans[namespace] = transMap
		}
		transMap[entity.UID] = entity
	}
	return filteredEntites, discoveredTrans
}

func buildSvcByNamespace(entities []*data.DIFEntity) serviceByNamespace {
	discoveredSvcs := make(serviceByNamespace)
	for _, entity := range entities {
		if entity.Type != "databaseServer" &&
			entity.Type != "application" {
			continue
		}
		// Only create service entities from application and databaseServer
		ServicePrefix := "Service-"
		svcID := ServicePrefix + entity.UID
		namespace := entity.GetNamespace()
		svc := data.NewDIFEntity(svcID, "service").WithNamespace(namespace)
		for meType, meList := range entity.Metrics {
			svc.AddMetrics(meType, meList)
		}
		hostedOnTypes := entity.GetHostedOnType()
		if len(hostedOnTypes) != 1 ||
			hostedOnTypes[0] != data.VM {
			// The application is not hosted on VM, assume it is hosted on
			// container, so create a proxy service (i.e., with stitching attribute)
			if entity.MatchingIdentifiers != nil {
				matchingID := ServicePrefix + entity.MatchingIdentifiers.IPAddress
				svc.Matching(matchingID)
			}
		}
		// Add the service to the service by namespace map
		svcMap, ok := discoveredSvcs[namespace]
		if !ok {
			svcMap = make(serviceMap)
			discoveredSvcs[namespace] = svcMap
		}
		for _, partOf := range entity.PartOf {
			svcName := partOf.Label
			svcMap[svcName] = append(svcMap[svcName], svc)
		}
		glog.V(3).Infof("Created service entity: %v", svc)
	}
	glog.V(4).Infof("Discovered services by namespace %v", spew.Sdump(discoveredSvcs))
	return discoveredSvcs
}

func (t *BusinessTopology) buildBizDIFEntities(discoveredSvcs serviceByNamespace,
	discoveredTrans transactionByNamespace) (bizEntities []*data.DIFEntity) {
	bizAppConfBySource, err := t.buildBizAppConfBySource(discoveredSvcs)
	if err != nil {
		glog.Warningf("Failed to build business entities: %v", err)
		return
	}
	for source, bizAppConfByName := range bizAppConfBySource {
		for name, bizAppConf := range bizAppConfByName {
			svcMap := discoveredSvcs[bizAppConf.Namespace]
			transMap := discoveredTrans[bizAppConf.Namespace]
			bizAppID := fmt.Sprintf("%s-%s", name, source)
			glog.V(4).Infof("BizApp ID: %v", bizAppID)
			var allDefinedSvcs []string
			allDefinedSvcs = append(allDefinedSvcs, bizAppConf.Services...)
			allDefinedSvcs = append(allDefinedSvcs, bizAppConf.OptionalServices...)
			for _, definedSvc := range allDefinedSvcs {
				svcEntities, exists := svcMap[definedSvc]
				if !exists {
					// Skip services that are configured but don't have metrics
					continue
				}
				for _, svcEntity := range svcEntities {
					svcEntity.PartOfEntity("businessApplication", bizAppID, "")
				}
			}
			for _, definedTrans := range bizAppConf.Transactions {
				bizTransID := util.GetName(definedTrans.Path, bizAppConf.Namespace)
				for _, service := range definedTrans.DependOn {
					svcEntities, exists := svcMap[service]
					if !exists {
						// Skip services that are configured but don't have metrics
						continue
					}
					for _, svcEntity := range svcEntities {
						svcEntity.PartOfEntity("businessTransaction", bizTransID, "")
					}
				}
				bizTransEntity := bizTransToDIFEntity(definedTrans, bizAppConf.Namespace, bizAppID)
				if bizTransEntityDiscovered, found := transMap[bizTransID]; found {
					// Specify the part of relationship for discovered business transaction
					bizTransEntityDiscovered.PartOf = bizTransEntity.PartOf
					// Update the display name of the discovered business transaction
					bizEntities = append(bizEntities, bizTransEntityDiscovered.WithName(bizTransEntity.Name))
				} else {
					// Add configured business transaction that is not discovered
					bizEntities = append(bizEntities, bizTransEntity)
				}
			}
			bizAppEntity := bizAppToDIFEntity(bizAppConf, bizAppID)
			bizEntities = append(bizEntities, bizAppEntity)
		}
	}
	return
}

func bizAppToDIFEntity(bizApp *config.BusinessApplication, bizAppID string) *data.DIFEntity {
	bizAppName := util.GetDisplay(bizApp.Name, bizApp.Namespace)
	bizAppDIFEntity := data.NewDIFEntity(bizAppID, "businessApplication").
		WithName(bizAppName)
	glog.V(2).Infof("Created business app entity %v", bizAppDIFEntity)
	return bizAppDIFEntity
}

func bizTransToDIFEntity(bizTrans config.Transaction, namespace, bizAppID string) *data.DIFEntity {
	name := bizTrans.Name
	if name == "" {
		name = bizTrans.Path
	}
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	bizTransID := util.GetName(bizTrans.Path, namespace)
	bizTransName := util.GetDisplay(name, namespace)
	bizTransDIFEntity := data.NewDIFEntity(bizTransID, "businessTransaction").
		WithName(bizTransName).
		PartOfEntity("businessApplication", bizAppID, "")
	glog.V(2).Infof("Created business transaction entity %v", bizTransDIFEntity)
	return bizTransDIFEntity
}

func (t *BusinessTopology) buildBizAppConfBySource(discoveredSvcs serviceByNamespace) (businessAppConfBySource, error) {
	var bizAppConfBySource = businessAppConfBySource{}
	for _, bizAppConf := range t.bizAppConfs {
		// Determine if at least one defined mandatory services for a business application are discovered under any
		// namespace. Create one business application for each of such namespaces.
		namespaces := reconcileNamespaces(discoveredSvcs, bizAppConf.Services)
		if len(namespaces) < 1 {
			glog.V(2).Infof("No services have been discovered for defined business application %v from"+
				" source %v", bizAppConf.Name, bizAppConf.From)
			continue
		}
		glog.V(2).Infof("Services for business application %v from source %v have been discovered in"+
			" namespaces %v", bizAppConf.Name, bizAppConf.From, strings.Join(namespaces, ","))
		// Namespace, Name and Source combination uniquely identifies a business application.
		// There cannot be two configured business applications with the same name, namespace and source.
		bizAppConfByName, ok := bizAppConfBySource[bizAppConf.From]
		if !ok {
			bizAppConfByName = make(map[string]*config.BusinessApplication)
			bizAppConfBySource[bizAppConf.From] = bizAppConfByName
		}
		for _, namespace := range namespaces {
			bizAppName := util.GetName(bizAppConf.Name, namespace)
			if _, found := bizAppConfByName[bizAppName]; found {
				return nil, fmt.Errorf("business app %v in namespace %v from source %v has been defined"+
					" more than once", bizAppConf.Name, namespace, bizAppConf.From)
			}
			bizAppCopy := bizAppConf
			bizAppCopy.Namespace = namespace
			bizAppConfByName[bizAppName] = &bizAppCopy
		}
	}
	return bizAppConfBySource, nil
}

func reconcileNamespaces(discoveredSvcs serviceByNamespace, definedSvcs []string) (namespaces []string) {
	definedSet := set.NewSet()
	for _, definedSvc := range definedSvcs {
		definedSet.Add(definedSvc)
	}
	for namespace, svcMap := range discoveredSvcs {
		discoveredSet := set.NewSet()
		for svc := range svcMap {
			discoveredSet.Add(svc)
		}
		glog.V(3).Infof("Services discovered in namespace %v: %v", namespace, discoveredSet)
		discovered := definedSet.Intersect(discoveredSet).Cardinality()
		if discovered >= 1 {
			// At least one defined service(s) have been discovered in this namespace!
			namespaces = append(namespaces, namespace)
			continue
		}
		glog.V(4).Infof("Namespace %v does not contain at least 1 defined services. Missing services: %v",
			namespace, definedSet.Difference(discoveredSet))
	}
	return
}
