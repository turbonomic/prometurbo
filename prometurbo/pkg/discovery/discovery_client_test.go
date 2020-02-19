package discovery

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/turbonomic/prometurbo/prometurbo/pkg/conf"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/constant"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

var (
	namespace      = "DEFAULT"
	ipAttr         = "IP"
	appPrefix      = "APPLICATION-"
	appType        = proto.EntityDTO_APPLICATION_COMPONENT
	useTopoExt     = true
	keepStandalone = false
	scope          = "k8s-cluster-foo"
	idKey          = registration.TargetIdField
	scopeKey       = registration.Scope
	targetAddr     = "foo"
	targetType     = "test"
	accountValues  = []*proto.AccountValue{{
		Key:         &idKey,
		StringValue: &targetAddr,
	}, {
		Key:         &scopeKey,
		StringValue: &scope,
	}}

	replacementMetaData = builder.NewReplacementEntityMetaDataBuilder().
		Matching(ipAttr).
		MatchingExternal(&proto.ServerEntityPropDef{
			Entity:     &appType,
			Attribute:  &ipAttr,
			UseTopoExt: &useTopoExt,
		}).
		PatchSellingWithProperty(proto.CommodityDTO_TRANSACTION, []string{constant.Used}).
		PatchSellingWithProperty(proto.CommodityDTO_RESPONSE_TIME, []string{constant.Used}).
		Build()

	metrics = []*exporter.EntityMetric{
		newMetric("1.2.3.4", 13.4, 66.7, appType),
		newMetric("5.6.7.8", 0, 0, appType),
	}

	inCompleteMetrics = []*exporter.EntityMetric{
		newIncompleteMetric("9.10.11.12", appType),
		newMetric("13.14.15.16", 0, 0, appType),
	}
	bizAppConfBySource = conf.BusinessAppConfBySource{}
)

func TestP8sDiscoveryClient_GetAccountValues(t *testing.T) {
	ex := mockExporter{metrics: metrics}
	d := NewDiscoveryClient(keepStandalone, scope, &targetAddr, targetType, bizAppConfBySource, ex)

	for _, f := range d.GetAccountValues().GetTargetInstance().InputFields {
		if f.Name == "targetIdentifier" && f.Value == targetAddr {
			return
		}
	}

	t.Errorf("AccountValues does not contain targetIdentifier with value %s", targetAddr)
}

func TestP8sDiscoveryClient_Discover(t *testing.T) {
	ex := &mockExporter{metrics: metrics}
	d := NewDiscoveryClient(keepStandalone, scope, &targetAddr, targetType, bizAppConfBySource, ex)

	testDiscoverySucceeded(d, metrics)
}

func TestP8sDiscoveryClient_Discover_Query_Failed(t *testing.T) {
	ex := mockExporter{err: fmt.Errorf("query failed with the mocked exporter")}
	d := NewDiscoveryClient(keepStandalone, scope, &targetAddr, targetType, bizAppConfBySource, ex)

	res, err := d.Discover(accountValues)

	if err != nil {
		t.Errorf("P8sDiscoveryClient.Discover() error = %v", err)
		return
	}

	// Expect to see error in the response
	if len(res.ErrorDTO) != 1 || *res.ErrorDTO[0].Severity != proto.ErrorDTO_CRITICAL {
		t.Errorf("Expected one error DTO with serverity CRITICAL but got %v", res.ErrorDTO)
	}
}

func TestP8sDiscoveryClient_Discover_Incomplete_Metrics(t *testing.T) {
	d := NewDiscoveryClient(false, scope, &targetAddr, targetType, bizAppConfBySource,
		&mockExporter{
			metrics: inCompleteMetrics,
			err:     nil,
		})
	res, err := d.Discover(accountValues)
	if err != nil {
		t.Errorf("P8sDiscoveryClient.Discover() error = %v", err)
		return
	}
	// Expect one DTO in response
	if len(res.EntityDTO) != 2 {
		t.Errorf("expected 2 DTOs but got %v", len(res.EntityDTO))
	}
}

type mockExporter struct {
	metrics []*exporter.EntityMetric
	err     error
}

func (m mockExporter) Query(targetAddr string, scope string) ([]*exporter.EntityMetric, error) {
	return m.metrics, m.err
}

func (m mockExporter) Validate(targetAddr string) error {
	return nil
}

func newMetric(ip string, tpsUsed, latUsed float64, entityType proto.EntityDTO_EntityType) *exporter.EntityMetric {
	m := map[proto.CommodityDTO_CommodityType]map[string]float64{
		proto.CommodityDTO_TRANSACTION:   {exporter.Used: tpsUsed},
		proto.CommodityDTO_RESPONSE_TIME: {exporter.Used: latUsed},
	}
	return &exporter.EntityMetric{
		UID:     ip,
		Type:    entityType,
		Metrics: m,
	}
}

func newIncompleteMetric(ip string, entityType proto.EntityDTO_EntityType) *exporter.EntityMetric {
	m := map[proto.CommodityDTO_CommodityType]map[string]float64{
		proto.CommodityDTO_HEAP: {},
	}
	return &exporter.EntityMetric{
		UID:     ip,
		Type:    entityType,
		Metrics: m,
	}
}

func checkAppResult(metric *exporter.EntityMetric, entity *proto.EntityDTO) error {
	ip := metric.UID
	tpsUsed := metric.Metrics[proto.CommodityDTO_TRANSACTION][exporter.Used]
	latUsed := metric.Metrics[proto.CommodityDTO_RESPONSE_TIME][exporter.Used]

	commodities := []*proto.CommodityDTO{
		newTransactionCommodity(tpsUsed, ip),
		newResponseTimeCommodity(latUsed, ip),
	}

	entityProperty := &proto.EntityDTO_EntityProperty{
		Namespace: &namespace,
		Name:      &ipAttr,
		Value:     &ip,
	}

	dto, err := builder.NewEntityDTOBuilder(proto.EntityDTO_APPLICATION_COMPONENT, appPrefix+scope+"/"+ip).
		DisplayName(appPrefix + scope + "/" + ip).
		SellsCommodities(commodities).
		WithProperty(entityProperty).
		ReplacedBy(replacementMetaData).
		Create()

	if err != nil {
		return err
	}

	if !reflect.DeepEqual(dto, entity) {
		return fmt.Errorf("the entity %v doesn't match the metric %v", entity, metric)
	}

	return nil
}

func newTransactionCommodity(used float64, key string) *proto.CommodityDTO {
	comm, _ := builder.NewCommodityDTOBuilder(proto.CommodityDTO_TRANSACTION).
		Used(used).Key(key).Create()
	return comm
}

func newResponseTimeCommodity(used float64, key string) *proto.CommodityDTO {
	comm, _ := builder.NewCommodityDTOBuilder(proto.CommodityDTO_RESPONSE_TIME).
		Used(used).Key(key).Create()
	return comm
}

func testDiscoverySucceeded(d *P8sDiscoveryClient, metrics []*exporter.EntityMetric) error {
	res, err := d.Discover([]*proto.AccountValue{})

	if err != nil {
		return err
	}

	if len(res.ErrorDTO) != 0 {
		return fmt.Errorf("ErrorDTO is not empty: %v", res.ErrorDTO[0])
	}

	if len(res.EntityDTO) != len(metrics) {
		return fmt.Errorf("expected %d entities but got %d entities", len(metrics), len(res.EntityDTO))
	}

	for i := range metrics {
		if err := checkAppResult(metrics[i], res.EntityDTO[i]); err != nil {
			return fmt.Errorf("P8sDiscoveryClient.Discover() error = %v", err)
		}
	}

	return nil
}
