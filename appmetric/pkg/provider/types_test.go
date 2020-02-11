package provider

import (
	"encoding/json"
	"fmt"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"testing"
)

const (
	AppEntity   = proto.EntityDTO_APPLICATION_COMPONENT
	LatencyType = proto.CommodityDTO_RESPONSE_TIME
	TpsType     = proto.CommodityDTO_TRANSACTION
	URL         = "http://prometheus-server.turbonomic:9090"
)

func TestEntityMetric_Marshall(t *testing.T) {
	em := NewEntityMetric(AppEntity, "aid1", URL)
	em.SetLabel("name", "default/curl-1xfj")
	em.SetLabel("ip", "10.0.2.3")
	em.SetLabel("scope", "k8s1")

	em.SetMetric(LatencyType, Used, 133.2)
	em.SetMetric(TpsType, Used, 12)

	//1. marshal
	ebytes, err := json.Marshal(em)
	if err != nil {
		t.Errorf("Failed to marshall EntityMetric %+v", em)
		return
	}

	fmt.Println(string(ebytes))

	//2. unmarshal it
	var em2 EntityMetric
	if err = json.Unmarshal(ebytes, &em2); err != nil {
		t.Errorf("Failed to un-marshal bytes: %v", string(ebytes))
		return
	}
	fmt.Printf("em2=%+v\n", em2)
}

func TestNewMetricResponse(t *testing.T) {
	em := NewEntityMetric(AppEntity, "aid1", URL)
	em.SetLabel("name", "default/curl-1xfj")
	em.SetLabel("ip", "10.0.2.3")
	em.SetLabel("scope", "k8s1")

	em.SetMetric(LatencyType, Used, 133.2)
	em.SetMetric(TpsType, Used, 12)

	em2 := NewEntityMetric(AppEntity, "aid2", URL)
	em2.SetLabel("name", "istio/music-ftaf2")
	em2.SetLabel("ip", "10.0.3.2")
	em2.SetLabel("scope", "k8s1")

	em2.SetMetric(LatencyType, Used, 13.2)
	em2.SetMetric(TpsType, Used, 10)

	res := NewMetricResponse()
	res.SetStatus(0, "good")
	res.AddMetric(em)
	res.AddMetric(em2)

	//1. marshal it
	ebytes, err := json.Marshal(res)
	if err != nil {
		t.Errorf("Failed to marshall EntityMetric %+v", res)
		return
	}

	fmt.Println(string(ebytes))

	//2. unmarshal it
	var mr MetricResponse
	if err = json.Unmarshal(ebytes, &mr); err != nil {
		t.Errorf("Failed to un-marshal bytes: %v", string(ebytes))
		return
	}
	if mr.Status != 0 || len(mr.Data) < 1 {
		t.Errorf("Failed to un-marshal MetricResponse: %+v", res)
		return
	}

	fmt.Printf("mr=%+v, len=%d\n", mr, len(mr.Data))
	for i, e := range mr.Data {
		fmt.Printf("[%d] %+v\n", i, e)
	}
}

func TestNewMetricResponse2(t *testing.T) {
	res := NewMetricResponse()
	res.SetStatus(-1, "error")

	//1. marshal it
	ebytes, err := json.Marshal(res)
	if err != nil {
		t.Errorf("Failed to marshall EntityMetric %+v", res)
		return
	}

	fmt.Println(string(ebytes))

	//2. unmarshal it
	var mr MetricResponse
	if err = json.Unmarshal(ebytes, &mr); err != nil {
		t.Errorf("Failed to un-marshal bytes: %v", string(ebytes))
		return
	}
	if mr.Status == 0 || len(mr.Data) > 0 {
		t.Errorf("Failed to un-marshal MetricResponse: %+v", res)
		return
	}

	fmt.Printf("%+v\n", mr)
}
