package addon

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/alligator"
	"github.com/turbonomic/prometurbo/appmetric/pkg/inter"
	xfire "github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
	"github.com/turbonomic/prometurbo/appmetric/pkg/util"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"net/url"
)

const (
	// query for latency (max of read and write) in milliseconds
	webdriver_latency_query = `navigation_timing_load_event_end_seconds{job="webdriver"}-navigation_timing_start_seconds{job="webdriver"}`

	default_Webdriver_Port = 80

	appType  = 1
	vAppType = 2
)

// Map of Turbo metric type to Webdriver query
var webDriverQueryMap = map[proto.CommodityDTO_CommodityType]string{
	inter.LatencyType: webdriver_latency_query,
}

type WebdriverEntityGetter struct {
	name  string
	du    string
	query *webdriverQuery
	etype int //Pod(Application), or Service
}

// ensure WebdriverEntityGetter implement the requisite interfaces
var _ alligator.EntityMetricGetter = &WebdriverEntityGetter{}

func NewWebdriverEntityGetter(name, du string) *WebdriverEntityGetter {
	return &WebdriverEntityGetter{
		name:  name,
		du:    du,
		etype: appType,
	}
}

func (webdriver *WebdriverEntityGetter) Name() string {
	return webdriver.name
}

func (webdriver *WebdriverEntityGetter) SetType(isVirtualApp bool) {
	if isVirtualApp {
		webdriver.etype = vAppType
	} else {
		webdriver.etype = appType
	}
}

func (webdriver *WebdriverEntityGetter) Category() string {
	if webdriver.etype == appType {
		return "Webdriver"
	}

	return "Webdriver.VApp"
}

func (r *WebdriverEntityGetter) GetEntityMetric(client *xfire.RestClient) ([]*inter.EntityMetric, error) {
	result := []*inter.EntityMetric{}
	midResult := make(map[string]*inter.EntityMetric)

	// Get metrics from Prometheus server
	for metricType := range webDriverQueryMap {
		query := &webdriverQuery{webDriverQueryMap[metricType]}
		metrics, err := client.GetMetrics(query)
		if err != nil {
			glog.Errorf("Failed to get webdriver Latency metrics: %v", err)
			return result, err
		} else {
			r.addEntity(metrics, midResult, metricType)
		}
	}

	// Reform map to list
	for _, v := range midResult {
		result = append(result, v)
	}

	return result, nil
}

// addEntity creates entities from the metric data
func (r *WebdriverEntityGetter) addEntity(mdat []xfire.MetricData, result map[string]*inter.EntityMetric, key proto.CommodityDTO_CommodityType) error {
	addrName := "instance"

	for _, dat := range mdat {
		metric, ok := dat.(*xfire.BasicMetricData)
		if !ok {
			glog.Errorf("Type assertion failed for[%v].", key)
			continue
		}

		//1. get IP
		var addr, err = url.Parse(metric.Labels[addrName])
		if err != nil {
			glog.Errorf("Label %v is not found", addrName)
			continue
		}

		ip, port, err := util.ParseIP(addr.Host, default_Webdriver_Port)
		if err != nil {
			glog.Errorf("Failed to parse IP from url[%v]: %v", addr.Host, err)
			continue
		}

		//2. add entity metrics
		entity, ok := result[ip]
		if !ok {
			if r.etype == vAppType {
				entity = inter.NewEntityMetric(ip, inter.VAppEntity)
			} else {
				entity = inter.NewEntityMetric(ip, inter.AppEntity)
			}
			entity.SetLabel(inter.IP, ip)
			entity.SetLabel(inter.Port, port)
			entity.SetLabel(inter.Category, r.Category())
			result[ip] = entity
		}

		entity.SetMetric(key, metric.GetValue())
	}

	return nil
}

//------------------ Get and Parse the metrics ---------------
type webdriverQuery struct {
	query string
}

func (q *webdriverQuery) GetQuery() string {
	return q.query
}

func (q *webdriverQuery) Parse(m *xfire.RawMetric) (xfire.MetricData, error) {
	d := xfire.NewBasicMetricData()
	if err := d.Parse(m); err != nil {
		return nil, err
	}

	return d, nil
}
