package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
)

const (
	apiPath      = "/api/v1/"
	apiQueryPath = "/api/v1/query"

	defaultTimeOut = 60 * time.Second
)

// for internal use only
type Response struct {
	Status    string   `json:"status"`
	Data      *RawData `json:"data,omitempty"`
	ErrorType string   `json:"errorType,omitempty"`
	Error     string   `json:"error,omitempty"`
}

type RawData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

type RestClient struct {
	client      *http.Client
	host        string
	username    string
	password    string
	bearerToken string
}

// NewRestClient create a new prometheus HTTP API client
func NewRestClient(host string, bearerToken string) (*RestClient, error) {
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

	glog.V(2).Infof("Creating client for Prometheus server: %v", host)

	if addr.Path == "" {
		// Append the default query path if not already there
		host = fmt.Sprintf("%v%v", host, apiQueryPath)
	}
	return &RestClient{
		client:      client,
		host:        host,
		bearerToken: bearerToken,
	}, nil
}

// GetHost get the host associated with the prometheus client
func (c *RestClient) GetHost() string {
	return c.host
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
		err := fmt.Errorf("prometheus query is empty")
		glog.Errorf(err.Error())
		return nil, err
	}

	req, err := http.NewRequest("GET", c.host, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return nil, err
	}

	// 1. set query
	// Always appending the current unix timestamp as some server implementation such as the one in IBM Cloud
	// doesn't conform to the Prometheus specs and treat the time parameter as optional.
	q := req.URL.Query()
	q.Set("query", query)
	q.Set("time", strconv.FormatInt(time.Now().Unix(), 10))
	req.URL.RawQuery = q.Encode()

	//2. set headers
	addHttpHeaders(*req, *c)

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
	glog.V(4).Infof("resp: %+v", string(result))

	// Invalid requests that reach the Prometheus server API handlers return a JSON error object with
	// a >400 status code instead of an error in the response.
	if resp.StatusCode >= 400 {
		err = fmt.Errorf(string(result))
		return nil, err
	}
	var ss Response
	if err := json.Unmarshal(result, &ss); err != nil {
		glog.Errorf("Failed to unmarshall response: %v", err)
		return nil, err
	}

	if ss.Status == "error" {
		return nil, fmt.Errorf(ss.Error)
	}
	return ss.Data, nil
}

// GetMetrics send a query to prometheus server, and return a list of MetricData
//
//	Note: it only support 'vector query: the data in the response is a 'vector'
//	       not a 'matrix' (range query), 'string', or 'scalar'
//	(1) the Request will generate a query;
//	(2) the Request will parse the response into a list of MetricData
func (c *RestClient) GetMetrics(request string) ([]MetricData, error) {
	var result []MetricData

	//1. query
	response, err := c.Query(request)
	if err != nil {
		glog.Errorf("Failed to get metrics from prometheus; url: %v, query: %v, error: %v", c.host, request, err)
		return result, err
	}

	if response == nil || response.ResultType != "vector" {
		err := fmt.Errorf("unsupported result type: %v", response.ResultType)
		glog.Errorf(err.Error())
		return result, err
	}

	//2. parse/decode the value
	var rawMetrics []RawMetric
	if err := json.Unmarshal(response.Result, &rawMetrics); err != nil {
		glog.Errorf("Failed to unmarshal: %v", err)
		return result, err
	}

	//3. assign the values
	for i := range rawMetrics {
		d, err := rawMetrics[i].Parse()
		if err != nil {
			glog.Errorf("Failed to parse value: %v", err)
			continue
		}
		glog.V(4).Infof("Successfully parsed metric data: %v", spew.Sdump(d))
		result = append(result, d)
	}

	return result, nil
}

func (c *RestClient) Validate() (string, error) {
	jobs, err := c.getJobs()
	if err != nil {
		return "", err
	}
	return jobs, nil
}

func (c *RestClient) getJobs() (string, error) {
	p := fmt.Sprintf("%v%v%v", c.host, apiPath, "label/job/values")
	glog.V(4).Infof("path=%v", p)

	//1. prepare result
	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return "", err
	}
	addHttpHeaders(*req, *c)

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
	return string(result), nil
}

func addHttpHeaders(req http.Request, client RestClient) {
	req.Header.Set("Accept", "application/json")
	if len(client.username) > 0 {
		req.SetBasicAuth(client.username, client.password)
	} else if len(client.bearerToken) > 0 {
		req.Header.Add("Authorization", "Bearer "+client.bearerToken)
	}
}
