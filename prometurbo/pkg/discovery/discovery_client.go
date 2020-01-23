package discovery

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/appmetric/metrics"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/dtofactory"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// MetricsProvider is a wrapper interface for getting metrics from and validating
// a metrics provider
type MetricsProvider interface {
	// Get the latest metrics from the provider
	GetMetrics() ([]*metrics.EntityMetric, error)

	// Validate access to the provider
	Validate() error
}

// P8sDiscoveryClient implements the TurboDiscoveryClient interface
type P8sDiscoveryClient struct {
	targetAddr     string
	keepStandalone bool
	createProxyVM  bool
	scope          string
	providers      []MetricsProvider
}

func NewDiscoveryClient(targetAddr string, keepStandalone bool, createProxyVM bool,
	scope string, providers []MetricsProvider) *P8sDiscoveryClient {
	return &P8sDiscoveryClient{
		targetAddr:     targetAddr,
		keepStandalone: keepStandalone,
		createProxyVM:  createProxyVM,
		scope:          scope,
		providers:      providers,
	}
}

// GetAccountValues gets the Account Values to create VMTTarget in the turbo
// server corresponding to this client
func (d *P8sDiscoveryClient) GetAccountValues() *probe.TurboTargetInfo {
	targetId := registration.TargetIdField
	targetIdVal := &proto.AccountValue{
		Key:         &targetId,
		StringValue: &d.targetAddr,
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

	targetInfo := probe.NewTurboTargetInfoBuilder(registration.ProbeCategory, registration.TargetType(d.targetAddr),
		registration.TargetIdField, accountValues).Create()

	return targetInfo
}

// Validate the Target
func (d *P8sDiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	validationResponse := &proto.ValidationResponse{}

	// Validation fails if any provider fails validation
	for _, provider := range d.providers {
		if err := provider.Validate(); err != nil {
			return d.failValidation(), err
		}
		glog.Infof("Successfully validated [%v]", provider)
	}
	return validationResponse, nil
}

// Discover the Target Topology
func (d *P8sDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	glog.V(2).Infof("Discovering the target %s", accountValues)
	var entities []*proto.EntityDTO
	allProvidersFailed := true

	for _, provider := range d.providers {
		dtos, err := d.buildEntities(provider)
		if err != nil {
			glog.Errorf("Error while querying metrics provider %v: %v", provider, err)
			continue
		}
		allProvidersFailed = false
		entities = append(entities, dtos...)

		glog.Infof("Discovered %d entities (%v) from provider %v", len(dtos),
			entityTypes(dtos), provider)
		glog.V(4).Infof("Entities built from provider %v: %v", provider, dtos)
	}

	// The discovery fails if all queries to providers fail
	if allProvidersFailed {
		return d.failDiscovery(), nil
	}

	discoveryResponse := &proto.DiscoveryResponse{
		EntityDTO: entities,
	}

	return discoveryResponse, nil
}

func entityTypes(entities []*proto.EntityDTO) map[string]int {
	var types = make(map[string]int)
	for _, entity := range entities {
		types[proto.EntityDTO_EntityType_name[int32(*entity.EntityType)]]++
	}
	return types
}

func (d *P8sDiscoveryClient) buildEntities(provider MetricsProvider) ([]*proto.EntityDTO, error) {
	var entities []*proto.EntityDTO
	businessAppMap := make(map[string][]*proto.EntityDTO)
	metrics, err := provider.GetMetrics()
	if err != nil {
		glog.Errorf("Error while querying metrics provider: %v", err)
		return nil, err
	}

	for _, metric := range metrics {
		dtos, err := dtofactory.NewEntityBuilder(d.keepStandalone, d.createProxyVM, d.scope, metric).Build()
		if err != nil {
			glog.Errorf("Error building entity from metric %v: %s", metric, err)
			continue
		}
		//Create a map with key: businessAppName (based on relabeling) and value: vapp dtos
		if v, ok := metric.Labels["business_app"]; ok {
			for _, dto := range dtos {
				if *dto.EntityType == proto.EntityDTO_VIRTUAL_APPLICATION {
					businessAppMap[v] = append(businessAppMap[v], dto)
				}
			}
		}
		entities = append(entities, dtos...)
	}
	if len(businessAppMap) > 0 {
		for k, v := range businessAppMap {
			dto, err := dtofactory.NewEntityBuilder(d.keepStandalone, d.createProxyVM, d.scope, nil).BuildBusinessApp(v, k)
			if err != nil {
				glog.Errorf("Error building business app entity for %s", k)
				continue
			}
			entities = append(entities, dto)
		}
	}
	return entities, nil
}

func (d *P8sDiscoveryClient) failDiscovery() *proto.DiscoveryResponse {
	description := fmt.Sprintf("All provider queries failed: %v", d.providers)
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

func (d *P8sDiscoveryClient) failValidation() *proto.ValidationResponse {
	description := fmt.Sprintf("All provider queries failed: %v", d.providers)
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
