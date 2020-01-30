package exporter

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
)

type MetricExporter interface {
	Query(targetAddr string, scope string) ([]*EntityMetric, error)
	Validate(targetAddr string) error
}

type metricExporter struct {
	endpoint string
}

func NewMetricExporter(endpoint string) metricExporter {
	return metricExporter{
		endpoint: endpoint,
	}
}

func (m metricExporter) Validate(targetAddr string) error {
	params := map[string]string{TargetAddress: targetAddr}
	if _, err := sendRequest(m.endpoint, params); err != nil {
		glog.Error("Failed connecting to the exporter. Retrying...")
		// Retry once with 3-second wait
		time.Sleep(3 * time.Second)
		if _, err = sendRequest(m.endpoint, params); err != nil {
			return err
		}
	}

	return nil
}

func (m metricExporter) Query(targetAddr string, scope string) ([]*EntityMetric, error) {
	params := map[string]string{TargetAddress: targetAddr, Scope: scope}
	resp, err := sendRequest(m.endpoint, params)
	if err != nil {
		return nil, err
	}

	var mr MetricResponse
	if err := json.Unmarshal(resp, &mr); err != nil {
		glog.Errorf("Failed to un-marshal bytes from target [%s]: %v", targetAddr, string(resp))
		return nil, err
	}
	if mr.Status != 0 || len(mr.Data) < 1 {
		glog.Errorf("Failed to un-marshal MetricResponse from target [%s]: %+v", targetAddr, string(resp))
		return nil, nil
	}

	glog.V(4).Infof("mr=%+v, len=%d\n", mr, len(mr.Data))
	for i, e := range mr.Data {
		glog.V(4).Infof("[%d] %+v\n", i, e)
	}

	return mr.Data, nil
}

// Send a request to the given endpoint. Params are encoded as query parameters
func sendRequest(endpoint string, params map[string]string) ([]byte, error) {
	url, err := encodeRequest(endpoint, params)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("Sending request to %s", url)

	resp, err := http.Get(url)
	if err != nil {
		glog.Errorf("Failed getting response from %s: %v", url, err)
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

func encodeRequest(endpoint string, params map[string]string) (string, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	return req.URL.String(), nil
}
