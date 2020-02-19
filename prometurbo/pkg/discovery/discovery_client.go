package discovery

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/conf"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/dtofactory"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// Implements the TurboDiscoveryClient interface
type P8sDiscoveryClient struct {
	keepStandalone        bool
	scope                 string
	optionalTargetAddress *string
	targetType            string
	bizAppConfBySource    conf.BusinessAppConfBySource
	metricExporter        exporter.MetricExporter
}

func NewDiscoveryClient(keepStandalone bool, scope string, optionalTargetAddress *string,
	targetType string, bizAppConfBySource conf.BusinessAppConfBySource,
	metricExporter exporter.MetricExporter) *P8sDiscoveryClient {
	return &P8sDiscoveryClient{
		keepStandalone:        keepStandalone,
		scope:                 scope,
		optionalTargetAddress: optionalTargetAddress,
		targetType:            targetType,
		bizAppConfBySource:    bizAppConfBySource,
		metricExporter:        metricExporter,
	}
}

// Get the Account Values to create VMTTarget in the turbo server corresponding to this client
func (d *P8sDiscoveryClient) GetAccountValues() *probe.TurboTargetInfo {
	targetAddr := ""
	if d.optionalTargetAddress != nil {
		targetAddr = *d.optionalTargetAddress
	}

	targetId := registration.TargetIdField
	targetIdVal := &proto.AccountValue{
		Key:         &targetId,
		StringValue: &targetAddr,
	}

	scope := registration.Scope
	scopeVal := &proto.AccountValue{
		Key:         &scope,
		StringValue: &d.scope,
	}

	accountValues := []*proto.AccountValue{
		targetIdVal,
		scopeVal,
	}

	targetInfo := probe.NewTurboTargetInfoBuilder(registration.ProbeCategory, d.targetType,
		registration.TargetIdField, accountValues).Create()

	return targetInfo
}

// Validate the Target
func (d *P8sDiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	targetAddr, found := targetAddress(accountValues)
	if !found {
		description := fmt.Sprintf("No target address (%s) in account values %v",
			registration.TargetIdField, accountValueKeyNames(accountValues))
		return d.failValidation(description), nil
	}

	validationResponse := &proto.ValidationResponse{}

	// Attempt validation
	glog.V(2).Infof("Validating to validate target %s", targetAddr)
	err := d.metricExporter.Validate(targetAddr)
	if err != nil {
		return d.failValidationWithError(targetAddr, err), err
	}

	return validationResponse, nil
}

// Discover the Target Topology
func (d *P8sDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	glog.V(2).Infof("Discovering the target %s", accountValues)
	targetAddr, found := targetAddress(accountValues)
	if !found {
		description := fmt.Sprintf("No target address (%s) in account values %v",
			registration.TargetIdField, accountValueKeyNames(accountValues))
		return d.failDiscovery(description), nil
	}
	scope, found := targetScope(accountValues)
	if !found {
		glog.V(3).Infof(fmt.Sprintf("No target scope (%s) in account values %v",
			registration.Scope, accountValueKeyNames(accountValues)))
	}

	metrics, err := d.metricExporter.Query(targetAddr, scope)
	if err != nil {
		return d.failDiscoveryWithError(targetAddr, err), nil
	}
	dtos, err := d.buildEntities(metrics)
	if err != nil {
		return d.failDiscoveryWithError(targetAddr, err), nil
	}

	glog.Infof("Discovered %d entities (%v) from provider %v (targetAddress=%s)", len(dtos),
		entityCountByType(dtos), d.metricExporter, targetAddr)
	glog.V(4).Infof("Entities built from exporter %v: %v", d.metricExporter, dtos)

	return &proto.DiscoveryResponse{EntityDTO: dtos}, nil
}

func accountValueKeyNames(accountValues []*proto.AccountValue) []*string {
	names := make([]*string, len(accountValues))
	for i := range accountValues {
		names[i] = accountValues[i].Key
	}
	return names
}

// targetAddress reads the target address from the array of account values.
// The first value returned is the address, if found.
// The second value returned is a bool indicating whether or not the address was successfully found.
func targetAddress(accountValues []*proto.AccountValue) (string, bool) {
	return matchingAccountValue(accountValues, registration.TargetIdField)
}

// targetScope reads the target scope from the array of account values.
// The first value returned is the scope, if found.
// The second value returned is a bool indicating whether or not the scope was successfully found.
func targetScope(accountValues []*proto.AccountValue) (string, bool) {
	return matchingAccountValue(accountValues, registration.Scope)
}

func matchingAccountValue(accountValues []*proto.AccountValue, matchKey string) (string, bool) {
	for _, value := range accountValues {
		if *value.Key == matchKey {
			return *value.StringValue, true
		}
	}

	return "", false
}

func entityCountByType(entities []*proto.EntityDTO) map[string]int {
	var types = make(map[string]int)
	for _, entity := range entities {
		types[proto.EntityDTO_EntityType_name[int32(*entity.EntityType)]]++
	}
	return types
}

func (d *P8sDiscoveryClient) buildEntities(metrics []*exporter.EntityMetric) ([]*proto.EntityDTO, error) {
	var entities []*proto.EntityDTO
	var bizAppInfoBySource = dtofactory.BusinessAppInfoBySource{}

	for _, metric := range metrics {
		bizAppInfo, ok := bizAppInfoBySource[metric.Source]
		if !ok {
			// Create a new entry
			bizAppInfo = dtofactory.NewBusinessAppInfo()
			bizAppInfoBySource[metric.Source] = bizAppInfo
		}
		var entityDTOs []*proto.EntityDTO
		var err error
		switch eType := metric.Type; eType {
		case proto.EntityDTO_APPLICATION_COMPONENT, proto.EntityDTO_DATABASE_SERVER:
			// For APPLICATION_COMPONENT or DATABASE_SERVER metrics, we create entity DTOs with corresponding
			// VIRTUAL_MACHINE provider DTO (if needed) and SERVICE consumer DTO if needed
			entityDTOs, err = dtofactory.NewApplicationBuilder(d.keepStandalone, d.scope, metric).Build()
			if err != nil {
				glog.Errorf("Error building entity from metric %v: %s", metric, err)
				continue
			}
			entities = append(entities, entityDTOs...)
		case proto.EntityDTO_BUSINESS_TRANSACTION:
			bizAppInfo.Transactions[metric.UID] = metric
		default:
			glog.V(2).Infof("Metric with entity type %v is not supported yet.", eType)
			continue
		}
		for _, entityDTO := range entityDTOs {
			if entityDTO.GetEntityType() == proto.EntityDTO_SERVICE {
				svcName, ok := metric.Labels["service"]
				if !ok || svcName == "" {
					continue
				}
				bizAppInfo.Services[svcName] = entityDTO
			}
		}
	}

	glog.V(4).Infof("%s", spew.Sdump(bizAppInfoBySource))

	// Create BUSINESS_APPLICATION and BUSINESS_TRANSACTION entity DTOs
	entityDTOS, err := dtofactory.NewBusinessAppBuilder(d.scope,
		bizAppInfoBySource, d.bizAppConfBySource).Build()
	if err != nil {
		glog.Errorf("Error building business app entities: %v", err)
	} else {
		entities = append(entities, entityDTOS...)
	}

	return entities, nil
}

func (d *P8sDiscoveryClient) failDiscoveryWithError(targetAddr string, err error) *proto.DiscoveryResponse {
	return d.failDiscovery(fmt.Sprintf("Discovery of %s failed due to error: %v", targetAddr, err))
}

func (d *P8sDiscoveryClient) failDiscovery(description string) *proto.DiscoveryResponse {
	glog.Errorf(description)
	severity := proto.ErrorDTO_CRITICAL
	errorDTO := &proto.ErrorDTO{
		Severity:    &severity,
		Description: &description,
	}
	discoveryResponse := &proto.DiscoveryResponse{
		ErrorDTO: []*proto.ErrorDTO{errorDTO},
	}
	return discoveryResponse
}

func (d *P8sDiscoveryClient) failValidationWithError(targetAddr string, err error) *proto.ValidationResponse {
	return d.failValidation(fmt.Sprintf("Validation of %s failed due to error: %v", err, targetAddr))
}

func (d *P8sDiscoveryClient) failValidation(description string) *proto.ValidationResponse {
	glog.Errorf(description)
	severity := proto.ErrorDTO_CRITICAL
	errorDto := &proto.ErrorDTO{
		Severity:    &severity,
		Description: &description,
	}

	validationResponse := &proto.ValidationResponse{
		ErrorDTO: []*proto.ErrorDTO{errorDto},
	}
	return validationResponse
}
