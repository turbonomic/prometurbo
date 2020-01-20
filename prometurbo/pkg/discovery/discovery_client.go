package discovery

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/dtofactory"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// Implements the TurboDiscoveryClient interface
type P8sDiscoveryClient struct {
	targetAddr      string
	keepStandalone  bool
	createProxyVM   bool
	scope           string
	metricExporters []exporter.MetricExporter
}

func NewDiscoveryClient(targetAddr string, keepStandalone bool, createProxyVM bool, scope string, metricExporters []exporter.MetricExporter) *P8sDiscoveryClient {
	return &P8sDiscoveryClient{
		targetAddr:      targetAddr,
		keepStandalone:  keepStandalone,
		createProxyVM:   createProxyVM,
		scope:           scope,
		metricExporters: metricExporters,
	}
}

// Get the Account Values to create VMTTarget in the turbo server corresponding to this client
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

	// Validation fails if no exporter responses
	for _, metricExporter := range d.metricExporters {
		if metricExporter.Validate() {
			return validationResponse, nil
		}

		glog.Errorf("Unable to connect to metric exporter %v", metricExporter)
	}
	return d.failValidation(), nil
}

// Discover the Target Topology
func (d *P8sDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	glog.V(2).Infof("Discovering the target %s", accountValues)
	var entities []*proto.EntityDTO
	allExportersFailed := true

	for _, metricExporter := range d.metricExporters {
		dtos, err := d.buildEntities(metricExporter)
		if err != nil {
			glog.Errorf("Error while querying metrics exporter %v: %v", metricExporter, err)
			continue
		}
		allExportersFailed = false
		entities = append(entities, dtos...)

		glog.V(4).Infof("Entities built from exporter %v: %v", metricExporter, dtos)
	}

	// The discovery fails if all queries to exporters fail
	if allExportersFailed {
		return d.failDiscovery(), nil
	}

	discoveryResponse := &proto.DiscoveryResponse{
		EntityDTO: entities,
	}

	return discoveryResponse, nil
}

func (d *P8sDiscoveryClient) buildEntities(metricExporter exporter.MetricExporter) ([]*proto.EntityDTO, error) {
	var entities []*proto.EntityDTO
	businessAppMap := make(map[string][]*proto.EntityDTO)
	metrics, err := metricExporter.Query()
	if err != nil {
		glog.Errorf("Error while querying metrics exporter: %v", err)
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
	if (len(businessAppMap) > 0) {
		for k, v := range businessAppMap {
			dto, err := dtofactory.NewEntityBuilder(d.keepStandalone, d.createProxyVM, d.scope,nil).BuildBusinessApp(v,k)
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
	description := fmt.Sprintf("All exporter queries failed: %v", d.metricExporters)
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
	description := fmt.Sprintf("All exporter queries failed: %v", d.metricExporters)
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
