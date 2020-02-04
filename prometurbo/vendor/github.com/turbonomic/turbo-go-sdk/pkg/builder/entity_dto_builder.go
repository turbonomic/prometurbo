package builder

import (
	"fmt"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// The actions in the Action Eligibility settings section of the EntityDTO
type ActionEligibilityField string

const (
	SUSPEND   ActionEligibilityField = "suspend"
	PROVISION ActionEligibilityField = "provision"
	MOVE      ActionEligibilityField = "move"
	SCALE     ActionEligibilityField = "scale"
	START     ActionEligibilityField = "start"
)

type ProviderDTO struct {
	providerType proto.EntityDTO_EntityType
	id           string
}

func CreateProvider(pType proto.EntityDTO_EntityType, id string) *ProviderDTO {
	return &ProviderDTO{
		providerType: pType,
		id:           id,
	}
}

func (pDto *ProviderDTO) GetProviderType() proto.EntityDTO_EntityType {
	return pDto.providerType
}

func (pDto *ProviderDTO) GetId() string {
	return pDto.id
}

// Action Eligibility settings section in the EntityDTO
type ActionEligibility struct {
	// These two settings will have an effect on the entity
	// setting for provision action, if entity is eligible for provision actions
	isProvisionable *bool
	// setting for suspend action, if the entity is eligible for suspend actions
	isSuspendable *bool

	// Settings for action that will be have an effect on the entity and a type of provider
	actionEligibilityByProviderMap map[proto.EntityDTO_EntityType]*ActionEligibilityByProvider
}

// ActionEligibilityByProvider settings section in the EntityDTO
type ActionEligibilityByProvider struct {
	// provider type
	pType proto.EntityDTO_EntityType
	// setting for move action, if the entity can move across on the providers of the given type
	isMovable *bool
	// setting for scale action, if the entity can be scaled on the providers of the given type
	isScalable *bool
	// setting for start action, if the entity can start on the providers of the given type
	isStartable *bool
}

type EntityDTOBuilder struct {
	entityType      *proto.EntityDTO_EntityType
	id              *string
	displayName     *string
	commoditiesSold []*proto.CommodityDTO

	underlying            []string
	entityProperties      []*proto.EntityDTO_EntityProperty
	origin                *proto.EntityDTO_EntityOrigin
	replacementEntityData *proto.EntityDTO_ReplacementEntityMetaData
	monitored             *bool
	powerState            *proto.EntityDTO_PowerState
	consumerPolicy        *proto.EntityDTO_ConsumerPolicy
	providerPolicy        *proto.EntityDTO_ProviderPolicy
	ownedBy               *string
	notification          []*proto.NotificationDTO
	keepStandalone        *bool
	profileID             *string

	// Action Eligibility related
	actionEligibility *ActionEligibility

	// Bought Commodity related
	// map of provider id and its entity type
	providerMap map[string]proto.EntityDTO_EntityType
	// map of provider and the list of commodities bought from it
	commoditiesBoughtProviderMap map[string][]*proto.CommodityDTO

	storageData            *proto.EntityDTO_StorageData
	diskArrayData          *proto.EntityDTO_DiskArrayData
	applicationData        *proto.EntityDTO_ApplicationData
	virtualMachineData     *proto.EntityDTO_VirtualMachineData
	physicalMachineData    *proto.EntityDTO_PhysicalMachineData
	virtualDataCenterData  *proto.EntityDTO_VirtualDatacenterData
	storageControllerData  *proto.EntityDTO_StorageControllerData
	logicalPoolData        *proto.EntityDTO_LogicalPoolData
	virtualApplicationData *proto.EntityDTO_VirtualApplicationData
	containerPodData       *proto.EntityDTO_ContainerPodData
	containerData          *proto.EntityDTO_ContainerData

	virtualMachineRelatedData    *proto.EntityDTO_VirtualMachineRelatedData
	physicalMachineRelatedData   *proto.EntityDTO_PhysicalMachineRelatedData
	storageControllerRelatedData *proto.EntityDTO_StorageControllerRelatedData

	currentProvider  *ProviderDTO
	entityDataHasSet bool

	err error
}

func NewEntityDTOBuilder(eType proto.EntityDTO_EntityType, id string) *EntityDTOBuilder {
	return &EntityDTOBuilder{
		entityType: &eType,
		id:         &id,
		actionEligibility: &ActionEligibility{
			actionEligibilityByProviderMap: make(map[proto.EntityDTO_EntityType]*ActionEligibilityByProvider),
		},
		providerMap: make(map[string]proto.EntityDTO_EntityType),
	}
}

func (eb *EntityDTOBuilder) Create() (*proto.EntityDTO, error) {
	if eb.err != nil {
		return nil, eb.err
	}

	entityDTO := &proto.EntityDTO{
		EntityType:            eb.entityType,
		Id:                    eb.id,
		DisplayName:           eb.displayName,
		CommoditiesSold:       eb.commoditiesSold,
		Underlying:            eb.underlying,
		EntityProperties:      eb.entityProperties,
		Origin:                eb.origin,
		ReplacementEntityData: eb.replacementEntityData,
		Monitored:             eb.monitored,
		PowerState:            eb.powerState,
		ConsumerPolicy:        eb.consumerPolicy,
		ProviderPolicy:        eb.providerPolicy,
		OwnedBy:               eb.ownedBy,
		Notification:          eb.notification,
	}
	if eb.storageData != nil {
		entityDTO.EntityData = &proto.EntityDTO_StorageData_{eb.storageData}
	} else if eb.diskArrayData != nil {
		entityDTO.EntityData = &proto.EntityDTO_DiskArrayData_{eb.diskArrayData}
	} else if eb.applicationData != nil {
		entityDTO.EntityData = &proto.EntityDTO_ApplicationData_{eb.applicationData}
	} else if eb.virtualMachineData != nil {
		entityDTO.EntityData = &proto.EntityDTO_VirtualMachineData_{eb.virtualMachineData}
	} else if eb.physicalMachineData != nil {
		entityDTO.EntityData = &proto.EntityDTO_PhysicalMachineData_{eb.physicalMachineData}
	} else if eb.virtualDataCenterData != nil {
		entityDTO.EntityData = &proto.EntityDTO_VirtualDatacenterData_{eb.virtualDataCenterData}
	} else if eb.storageControllerData != nil {
		entityDTO.EntityData = &proto.EntityDTO_StorageControllerData_{eb.storageControllerData}
	} else if eb.logicalPoolData != nil {
		entityDTO.EntityData = &proto.EntityDTO_LogicalPoolData_{eb.logicalPoolData}
	} else if eb.virtualApplicationData != nil {
		entityDTO.EntityData = &proto.EntityDTO_VirtualApplicationData_{eb.virtualApplicationData}
	} else if eb.containerPodData != nil {
		entityDTO.EntityData = &proto.EntityDTO_ContainerPodData_{eb.containerPodData}
	} else if eb.containerData != nil {
		entityDTO.EntityData = &proto.EntityDTO_ContainerData_{eb.containerData}
	}

	if eb.virtualMachineRelatedData != nil {
		entityDTO.RelatedEntityData = &proto.EntityDTO_VirtualMachineRelatedData_{eb.virtualMachineRelatedData}
	} else if eb.physicalMachineRelatedData != nil {
		entityDTO.RelatedEntityData = &proto.EntityDTO_PhysicalMachineRelatedData_{eb.physicalMachineRelatedData}
	} else if eb.storageControllerRelatedData != nil {
		entityDTO.RelatedEntityData = &proto.EntityDTO_StorageControllerRelatedData_{eb.storageControllerRelatedData}
	}

	// Create the action eligibility spec for the entity
	if eb.actionEligibility != nil {
		ae := eb.actionEligibility
		entityActionEligibility := &proto.EntityDTO_ActionEligibility{}
		// Is the provisionable setting available
		if ae.isProvisionable != nil {
			entityActionEligibility.Cloneable = ae.isProvisionable
		}
		// Is the suspendable setting available
		if ae.isSuspendable != nil {
			entityActionEligibility.Suspendable = ae.isSuspendable
		}

		entityDTO.ActionEligibility = entityActionEligibility
	}

	// Create the bought commodities
	entityDTO.CommoditiesBought = eb.buildCommodityBoughtFromMap()

	return entityDTO, nil
}

func (eb *EntityDTOBuilder) DisplayName(displayName string) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	eb.displayName = &displayName
	return eb
}

// Add a list of commodities to entity commodities sold list.
func (eb *EntityDTOBuilder) SellsCommodities(commDTOs []*proto.CommodityDTO) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	eb.commoditiesSold = append(eb.commoditiesSold, commDTOs...)
	return eb
}

// Add a single commodity to entity commodities sold list.
func (eb *EntityDTOBuilder) SellsCommodity(commDTO *proto.CommodityDTO) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	if eb.commoditiesSold == nil {
		eb.commoditiesSold = []*proto.CommodityDTO{}
	}
	eb.commoditiesSold = append(eb.commoditiesSold, commDTO)
	return eb
}

// Set the current provider with provided entity type and ID.
func (eb *EntityDTOBuilder) Provider(provider *ProviderDTO) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	eb.currentProvider = provider
	// Save the current provider type
	eb.providerMap[provider.id] = provider.providerType
	return eb
}

// entity buys a list of commodities.
func (eb *EntityDTOBuilder) BuysCommodities(commDTOs []*proto.CommodityDTO) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	if eb.currentProvider == nil {
		eb.err = fmt.Errorf("Provider has not been set for current list of commodities: %++v", commDTOs)
		return eb
	}
	for _, commDTO := range commDTOs {
		eb.BuysCommodity(commDTO)
	}
	return eb
}

// entity buys a single commodity
func (eb *EntityDTOBuilder) BuysCommodity(commDTO *proto.CommodityDTO) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	if eb.currentProvider == nil {
		eb.err = fmt.Errorf("Porvider has not been set for %++v", commDTO)
		return eb
	}

	if eb.commoditiesBoughtProviderMap == nil {
		eb.commoditiesBoughtProviderMap = make(map[string][]*proto.CommodityDTO)
	}

	// add commodity bought to map
	commoditiesSoldByCurrentProvider, exist := eb.commoditiesBoughtProviderMap[eb.currentProvider.id]
	if !exist {
		commoditiesSoldByCurrentProvider = []*proto.CommodityDTO{}
	}
	commoditiesSoldByCurrentProvider = append(commoditiesSoldByCurrentProvider, commDTO)
	eb.commoditiesBoughtProviderMap[eb.currentProvider.id] = commoditiesSoldByCurrentProvider

	return eb
}

// Add a single property to entity
func (eb *EntityDTOBuilder) WithProperty(property *proto.EntityDTO_EntityProperty) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}

	if eb.entityProperties == nil {
		eb.entityProperties = []*proto.EntityDTO_EntityProperty{}
	}
	// add the property to list.
	eb.entityProperties = append(eb.entityProperties, property)

	return eb
}

// Add multiple properties to entity
func (eb *EntityDTOBuilder) WithProperties(properties []*proto.EntityDTO_EntityProperty) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}

	if eb.entityProperties == nil {
		eb.entityProperties = []*proto.EntityDTO_EntityProperty{}
	}
	// add the property to list.
	eb.entityProperties = append(eb.entityProperties, properties...)

	return eb
}

// Set the ReplacementEntityMetadata that will contain the information about the external entity
// that this entity will patch with the metrics data it collected.
func (eb *EntityDTOBuilder) ReplacedBy(replacementEntityMetaData *proto.EntityDTO_ReplacementEntityMetaData) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	origin := proto.EntityDTO_PROXY
	eb.origin = &origin
	eb.replacementEntityData = replacementEntityMetaData
	return eb
}

func (eb *EntityDTOBuilder) WithPowerState(state proto.EntityDTO_PowerState) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	eb.powerState = &state
	return eb
}

func (eb *EntityDTOBuilder) Monitored(monitored bool) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	eb.monitored = &monitored
	return eb
}

func (eb *EntityDTOBuilder) ConsumerPolicy(cp *proto.EntityDTO_ConsumerPolicy) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	eb.consumerPolicy = cp
	return eb
}

func (eb *EntityDTOBuilder) ApplicationData(appData *proto.EntityDTO_ApplicationData) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	if eb.entityDataHasSet {
		eb.err = fmt.Errorf("EntityData has already been set. Cannot use %v as entity data.", appData)

		return eb
	}
	eb.applicationData = appData
	eb.entityDataHasSet = true
	return eb
}

func (eb *EntityDTOBuilder) VirtualMachineData(vmData *proto.EntityDTO_VirtualMachineData) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	if eb.entityDataHasSet {
		eb.err = fmt.Errorf("EntityData has already been set. Cannot use %v as entity data.", vmData)

		return eb
	}
	eb.virtualMachineData = vmData
	eb.entityDataHasSet = true
	return eb
}

func (eb *EntityDTOBuilder) ContainerPodData(podData *proto.EntityDTO_ContainerPodData) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	if eb.entityDataHasSet {
		eb.err = fmt.Errorf("EntityData has already been set. Cannot use %v as entity data.", podData)

		return eb
	}
	eb.containerPodData = podData
	eb.entityDataHasSet = true
	return eb
}

func (eb *EntityDTOBuilder) ContainerData(containerData *proto.EntityDTO_ContainerData) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	if eb.entityDataHasSet {
		eb.err = fmt.Errorf("EntityData has already been set. Cannot use %v as entity data.", containerData)

		return eb
	}
	eb.containerData = containerData
	eb.entityDataHasSet = true
	return eb
}

func (eb *EntityDTOBuilder) VirtualApplicationData(vAppData *proto.EntityDTO_VirtualApplicationData) *EntityDTOBuilder {
	if eb.err != nil {
		return eb
	}
	if eb.entityDataHasSet {
		eb.err = fmt.Errorf("EntityData has already been set. Cannot use %v as entity data.", vAppData)

		return eb
	}
	eb.virtualApplicationData = vAppData
	eb.entityDataHasSet = true
	return eb
}

// Build the commodity DTOs for the commodity bought from different providers
func (eb *EntityDTOBuilder) buildCommodityBoughtFromMap() []*proto.EntityDTO_CommodityBought {
	var commoditiesBought []*proto.EntityDTO_CommodityBought

	if len(eb.commoditiesBoughtProviderMap) == 0 {
		return commoditiesBought
	}

	var actionEligibilityByProviderMap map[proto.EntityDTO_EntityType]*ActionEligibilityByProvider
	if eb.actionEligibility != nil {
		actionEligibilityByProviderMap = eb.actionEligibility.actionEligibilityByProviderMap
	}

	// Create CommodityBought for each provider
	for providerId, commodities := range eb.commoditiesBoughtProviderMap {
		// Commodity bought per provider
		p := providerId

		commodityBought := &proto.EntityDTO_CommodityBought{
			ProviderId: &p,
			Bought:     commodities,
		}

		// update provider type and the action eligibility of the entity across that provider type
		providerType, providerTypeExists := eb.providerMap[p]
		if providerTypeExists {

			commodityBought.ProviderType = &providerType
			aeByProvider, exists := actionEligibilityByProviderMap[providerType]
			if exists {
				commodityBought.ActionEligibility = &proto.EntityDTO_ActionOnProviderEligibility{}
				if aeByProvider.isMovable != nil {
					commodityBought.ActionEligibility.Movable = aeByProvider.isMovable
				}
				if aeByProvider.isStartable != nil {
					commodityBought.ActionEligibility.Startable = aeByProvider.isStartable
				}
				if aeByProvider.isScalable != nil {
					commodityBought.ActionEligibility.Scalable = aeByProvider.isScalable
				}
			}
		}

		commoditiesBought = append(commoditiesBought, commodityBought)
	}

	return commoditiesBought
}

// Specifies if the entity is eligible to be cloned by Turbonomic analysis
func (eb *EntityDTOBuilder) IsProvisionable(provisionable bool) *EntityDTOBuilder {
	eb.actionEligibility.isProvisionable = &provisionable
	return eb
}

// Specifies if the entity is eligible to be suspended by Turbonomic analysis
func (eb *EntityDTOBuilder) IsSuspendable(suspendable bool) *EntityDTOBuilder {
	eb.actionEligibility.isSuspendable = &suspendable
	return eb
}

// Specifies if the entity is eligible to move across the given provider
func (eb *EntityDTOBuilder) IsMovable(providerType proto.EntityDTO_EntityType, movable bool) *EntityDTOBuilder {
	eb.setActionEligibilityByProviderField(providerType, "move", movable)
	return eb
}

// Specifies if the entity is eligible to scale on the given provider
func (eb *EntityDTOBuilder) IsScalable(providerType proto.EntityDTO_EntityType, scalable bool) *EntityDTOBuilder {
	eb.setActionEligibilityByProviderField(providerType, "scale", scalable)
	return eb
}

// Specifies if the entity is eligible to scale on the given provider
func (eb *EntityDTOBuilder) IsStartable(providerType proto.EntityDTO_EntityType, startable bool) *EntityDTOBuilder {
	eb.setActionEligibilityByProviderField(providerType, "start", startable)
	return eb
}

func (eb *EntityDTOBuilder) setActionEligibilityFlag(flagName ActionEligibilityField, flagVal bool) {
	// Create action eligibility if null
	ae := eb.actionEligibility
	if ae == nil {
		ae = &ActionEligibility{
			actionEligibilityByProviderMap: make(map[proto.EntityDTO_EntityType]*ActionEligibilityByProvider),
		}
		eb.actionEligibility = ae
	}

	switch flagName {
	case SUSPEND:
		ae.isSuspendable = &flagVal
	case PROVISION:
		ae.isProvisionable = &flagVal
	default:
	}
}

func (eb *EntityDTOBuilder) setActionEligibilityByProviderField(providerType proto.EntityDTO_EntityType,
	flagName ActionEligibilityField, flagVal bool) {
	// Create action eligibility if null
	ae := eb.actionEligibility
	if ae == nil {
		ae = &ActionEligibility{
			actionEligibilityByProviderMap: make(map[proto.EntityDTO_EntityType]*ActionEligibilityByProvider),
		}
		eb.actionEligibility = ae
	}

	actionEligibilityByProvider, exists := ae.actionEligibilityByProviderMap[providerType]
	if !exists {
		ae.actionEligibilityByProviderMap[providerType] = &ActionEligibilityByProvider{}
		actionEligibilityByProvider = ae.actionEligibilityByProviderMap[providerType]
	}

	switch flagName {
	case MOVE:
		actionEligibilityByProvider.isMovable = &flagVal
	case SCALE:
		actionEligibilityByProvider.isScalable = &flagVal
	case START:
		actionEligibilityByProvider.isStartable = &flagVal
	default:
	}
}
