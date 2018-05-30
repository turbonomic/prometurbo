package exporter

import (
	"encoding/json"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"time"
)

type MetricExporter interface {
	Query() ([]*EntityMetric, error)
	Validate() bool
}

type metricExporter struct {
	endpoint string
}

func NewMetricExporter(endpoint string) *metricExporter {
	return &metricExporter{
		endpoint: endpoint,
	}
}

func (m *metricExporter) Validate() bool {
	if _, err := sendRequest(m.endpoint); err != nil {
		glog.Error("Failed connecting to the exporter. Retrying...")
		// Retry once with 5-second wait
		time.Sleep(5 * time.Second)
		if _, err = sendRequest(m.endpoint); err != nil {
			glog.Error("Failed connecting to the exporter.")
			return false
		}
	}

	return true
}

func (m *metricExporter) Query() ([]*EntityMetric, error) {
	resp, err := sendRequest(m.endpoint)
	if err != nil {
		return nil, err
	}

	var mr MetricResponse
	if err := json.Unmarshal(resp, &mr); err != nil {
		glog.Errorf("Failed to un-marshal bytes: %v", string(resp))
		return nil, err
	}
	if mr.Status != 0 || len(mr.Data) < 1 {
		glog.Errorf("Failed to un-marshal MetricResponse: %+v", string(resp))
		return nil, nil
	}

	glog.V(4).Infof("mr=%+v, len=%d\n", mr, len(mr.Data))
	for i, e := range mr.Data {
		glog.V(4).Infof("[%d] %+v\n", i, e)
	}

	return mr.Data, nil
}

func sendRequest(endpoint string) ([]byte, error) {
	glog.V(2).Infof("Sending request to %s", endpoint)
	resp, err := http.Get(endpoint)
	if err != nil {
		glog.Errorf("Failed getting response from %s: %v", endpoint, err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Error reading the response %v: %v", resp, err)
		return nil, err
	}
	glog.V(4).Infof("Received resposne: %s", string(body))
	return body, nil
}
