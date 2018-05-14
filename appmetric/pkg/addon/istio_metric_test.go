package addon

import (
	"fmt"
	pclient "github.com/songbinliu/xfire/pkg/prometheus"
	"strings"
	"testing"
)

func TestTypeAssertion(t *testing.T) {
	metrics := []pclient.MetricData{}

	m1 := newIstioMetricData()
	m1.Labels["destination_uid"] = "uid1"
	metrics = append(metrics, m1)

	m2 := newIstioMetricData()
	m2.Labels["destination_uid"] = "uid2"
	metrics = append(metrics, m2)

	for i, m := range metrics {
		mdat, ok := m.(*istioMetricData)
		if !ok {
			t.Errorf("Type Assertion failed.")
		}

		fmt.Printf("[%d] %+v\n", i, mdat)
	}
}

func Test_ConvertPodUID(t *testing.T) {
	inputs := []string{
		"kubernetes://video-671194421-vpxkh.default",
		"kubernetes://inception-be-41ldc.default",
		"kubernetes://inception-be-41ldc.istio-system",
		"kubernetes://inception-be-41ldc.istio-system.",
	}

	expects := []string{
		"default/video-671194421-vpxkh",
		"default/inception-be-41ldc",
		"istio-system/inception-be-41ldc",
		"istio-system/inception-be-41ldc",
	}

	for i := range inputs {
		ain := inputs[i]
		aout, err := convertPodUID(ain)
		if err != nil {
			t.Errorf("convert UID: %v failed: %v", ain, err)
		}

		if aout != expects[i] {
			t.Errorf("Not equal: %v Vs. %v", aout, expects[i])
		}
	}
}

func Test_ConvertPodUID_Fail(t *testing.T) {
	inputs := []string{
		"netes://video-671194421-vpxkh.default",
		"//inception-be-41ldc.default",
		"kubernetes://inception-be-41ldc",
		"kubernetes://inception-be-41ldc. ",
		"kubernetes://inception-be-41ldc-istio-system",
		"kubernetes:// .inception-be-41ldc-istio-system",
	}

	for i := range inputs {
		_, err := convertPodUID(inputs[i])
		if err == nil {
			t.Errorf("convert UID should have failed with input: %v", inputs[i])
			return
		}

		fmt.Printf("input: %v, err: %v\n", inputs[i], err)
	}
}

func Test_ConvertSVCUID(t *testing.T) {
	inputs := []string{
		"productpage.default.svc.cluster.local",
		"productpage.istio.svc.cluster.local",
		"a.b.svc.cluster.local",
		"a.bb.svc.x.y",
		"aa.bb.svc",
	}

	expects := []string{
		"default/productpage",
		"istio/productpage",
		"b/a",
		"bb/a",
		"bb/aa",
	}

	for i := range inputs {
		ain := inputs[i]
		aout, err := convertSVCUID(ain)
		if err != nil {
			t.Errorf("convert UID: %v failed: %v", ain, err)
		}

		if aout != expects[i] {
			t.Errorf("Not equal: %v Vs. %v", aout, expects[i])
		}
	}
}

func Test_ConvertSVCUID_Fail(t *testing.T) {
	inputs := []string{
		"productpage.default.cluster.local",
		"productpage.istio.cluster.local",
		"a.svc.cluster.local",
		"a.bb..svc.x.y",
	}

	for i := range inputs {
		_, err := convertSVCUID(inputs[i])
		if err == nil {
			t.Errorf("convert UID should have failed with input: %v", inputs[i])
			return
		}

		fmt.Printf("input: %v, err: %v\n", inputs[i], err)
	}
}

func TestParseIP(t *testing.T) {
	d := newIstioMetricData()

	raw := "[0 0 0 0 0 0 0 0 0 0 255 255 10 2 1 84]"
	expected := "10.2.1.84"

	result, err := d.parseIP(raw)
	if err != nil {
		t.Errorf("Failed to parse IP: %v", err)
		return
	}

	if !strings.EqualFold(result, expected) {
		t.Errorf("Wrong result: %v Vs. %v", result, expected)
	}
}
