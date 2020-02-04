package discovery

import (
	"reflect"
	"testing"

	"fmt"

	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/constant"
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

var (
	namespace      = "DEFAULT"
	ipAttr         = "IP"
	appPrefix      = "APPLICATION-"
	appType        = proto.EntityDTO_APPLICATION
	useTopoExt     = true
	keepStandalone = false
	scope          = "k8s-cluster-foo"
	targetAddr     = "foo"
	targetType     = "test"

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
)

func TestP8sDiscoveryClient_GetAccountValues(t *testing.T) {
	ex := mockExporter{metrics: metrics}
	d := NewDiscoveryClient(keepStandalone, scope, &targetAddr, targetType, ex)

	for _, f := range d.GetAccountValues().GetTargetInstance().InputFields {
		if f.Name == "targetIdentifier" && f.Value == targetAddr {
			return
		}
	}

	t.Errorf("AccountValues does not contain targetIdentifier with value %s", targetAddr)
}

func TestP8sDiscoveryClient_Discover(t *testing.T) {
	ex := &mockExporter{metrics: metrics}
	d := NewDiscoveryClient(keepStandalone, scope, &targetAddr, targetType, ex)

	testDiscoverySuccedded(d, metrics)
}

func TestP8sDiscoveryClient_Discover_Query_Failed(t *testing.T) {
	ex := mockExporter{err: fmt.Errorf("Query failed with the mocked exporter")}
	d := NewDiscoveryClient(keepStandalone, scope, &targetAddr, targetType, ex)

	res, err := d.Discover([]*proto.AccountValue{})

	if err != nil {
		t.Errorf("P8sDiscoveryClient.Discover() error = %v", err)
		return
	}

	// Expect to see error in the response
	if len(res.ErrorDTO) != 1 || *res.ErrorDTO[0].Severity != proto.ErrorDTO_CRITICAL {
		t.Errorf("Expected one error DTO with serverity CRITICAL but got %v", res.ErrorDTO)
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

func checkAppResult(metric *exporter.EntityMetric, entity *proto.EntityDTO) error {
	ip := metric.UID
	tpsUsed := metric.Metrics[proto.CommodityDTO_TRANSACTION][exporter.Used]
	latUsed := metric.Metrics[proto.CommodityDTO_RESPONSE_TIME][exporter.Used]

	commodities := []*proto.CommodityDTO{
		newTrasactionCommodity(tpsUsed, ip),
		newResponseTimeCommodity(latUsed, ip),
	}

	entityProperty := &proto.EntityDTO_EntityProperty{
		Namespace: &namespace,
		Name:      &ipAttr,
		Value:     &ip,
	}

	dto, err := builder.NewEntityDTOBuilder(proto.EntityDTO_APPLICATION, appPrefix+scope+"/"+ip).
		DisplayName(appPrefix + scope + "/" + ip).
		SellsCommodities(commodities).
		WithProperty(entityProperty).
		ReplacedBy(replacementMetaData).
		Create()

	if err != nil {
		return err
	}

	if !reflect.DeepEqual(dto, entity) {
		return fmt.Errorf("The entity %v doesn't match the metric %v", entity, metric)
	}

	return nil
}

func newTrasactionCommodity(used float64, key string) *proto.CommodityDTO {
	comm, _ := builder.NewCommodityDTOBuilder(proto.CommodityDTO_TRANSACTION).
		Used(used).Key(key).Create()
	return comm
}

func newResponseTimeCommodity(used float64, key string) *proto.CommodityDTO {
	comm, _ := builder.NewCommodityDTOBuilder(proto.CommodityDTO_RESPONSE_TIME).
		Used(used).Key(key).Create()
	return comm
}

func testDiscoverySuccedded(d *P8sDiscoveryClient, metrics []*exporter.EntityMetric) error {
	res, err := d.Discover([]*proto.AccountValue{})

	if err != nil {
		return err
	}

	if len(res.ErrorDTO) != 0 {
		return fmt.Errorf("ErrorDTO is not empty: %v", res.ErrorDTO[0])
	}

	if len(res.EntityDTO) != len(metrics) {
		return fmt.Errorf("Expected %d entities but got %d entities", len(metrics), len(res.EntityDTO))
	}

	for i := range metrics {
		if err := checkAppResult(metrics[i], res.EntityDTO[i]); err != nil {
			return fmt.Errorf("P8sDiscoveryClient.Discover() error = %v", err)
		}
	}

	return nil
}
