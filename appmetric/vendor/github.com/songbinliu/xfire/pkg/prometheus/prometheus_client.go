package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiPath      = "/api/v1/"
	apiQueryPath = "/api/v1/query"
	apiRangePath = "/api/v1/query_range"

	defaultTimeOut = time.Duration(60 * time.Second)
)

type RestClient struct {
	client   *http.Client
	host     string
	username string
	password string
}

// NewRestClient create a new prometheus HTTP API client
func NewRestClient(host string) (*RestClient, error) {
	//1. get http client
	client := &http.Client{
		Timeout: defaultTimeOut,
	}

	//2. check whether it is using ssl
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}

	addr, err := url.Parse(host)
	if err != nil {
		glog.Errorf("Invalid url:%v, %v", host, err)
		return nil, err
	}
	if addr.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	glog.V(2).Infof("Prometheus server address is: %v", host)

	return &RestClient{
		client: client,
		host:   host,
	}, nil
}

// SetUser set the login user/password for the prometheus client
func (c *RestClient) SetUser(username, password string) {
	c.username = username
	c.password = password
}

// Query query the prometheus server, and return the rawData
func (c *RestClient) Query(query string) (*RawData, error) {
	query = strings.TrimSpace(query)
	if len(query) < 1 {
		err := fmt.Errorf("Prometheus query is empty")
		glog.Errorf(err.Error())
		return nil, err
	}

	p := fmt.Sprintf("%v%v", c.host, apiQueryPath)
	glog.V(4).Infof("path=%v, query=%v", p, query)

	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return nil, err
	}

	//1. set query
	q := req.URL.Query()
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	//2. set headers
	req.Header.Set("Accept", "application/json")
	if len(c.username) > 0 {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		glog.Errorf("Failed to send http request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Failed to read response: %v", err)
		return nil, err
	}

	var ss promeResponse
	if err := json.Unmarshal(result, &ss); err != nil {
		glog.Errorf("Failed to unmarshall respone: %v", err)
		return nil, err
	}

	if ss.Status == "error" {
		return nil, fmt.Errorf(ss.Error)
	}

	glog.V(4).Infof("resp: %++v", string(result))
	glog.V(4).Infof("metric: %+++v", ss)
	return ss.Data, nil
}

// GetMetrics send a query to prometheus server, and return a list of MetricData
//   Note: it only support 'vector query: the data in the response is a 'vector'
//          not a 'matrix' (range query), 'string', or 'scalar'
//   (1) the RequestInput will generate a query;
//   (2) the RequestInput will parse the response into a list of MetricData
func (c *RestClient) GetMetrics(input RequestInput) ([]MetricData, error) {
	result := []MetricData{}

	//1. query
	qresult, err := c.Query(input.GetQuery())
	if err != nil {
		glog.Errorf("Failed to get metrics from prometheus: %v", err)
		return result, err
	}

	glog.V(4).Infof("result.type=%v, \n result: %+v",
		qresult.ResultType, string(qresult.Result))

	if qresult.ResultType != "vector" {
		err := fmt.Errorf("Unsupported result type: %v", qresult.ResultType)
		glog.Errorf(err.Error())
		return result, err
	}

	//2. parse/decode the value
	var resp []RawMetric
	if err := json.Unmarshal(qresult.Result, &resp); err != nil {
		glog.Errorf("Failed to unmarshal: %v", err)
		return result, err
	}

	//3. assign the values
	for i := range resp {
		d, err := input.Parse(&(resp[i]))
		if err != nil {
			glog.Errorf("Pase value failed: %v", err)
			continue
		}

		result = append(result, d)
	}

	return result, nil
}

// GetJobs  get the all the jobs in the current prometheus server
//     it is only used for testing.
func (c *RestClient) GetJobs() (string, error) {
	p := fmt.Sprintf("%v%v%v", c.host, apiPath, "label/job/values")
	glog.V(2).Infof("path=%v", p)

	//1. prepare result
	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		glog.Errorf("Failed to send http request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	//2. read response
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Failed to read response: %v", err)
		return "", err
	}

	glog.V(3).Infof("resp: %++v", resp)
	glog.V(3).Infof("result: %++v", string(result))

	return string(result), nil
}
