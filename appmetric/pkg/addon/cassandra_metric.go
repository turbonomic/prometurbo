package addon

import (
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/alligator"
	"github.com/turbonomic/prometurbo/appmetric/pkg/inter"
	xfire "github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
	"github.com/turbonomic/prometurbo/appmetric/pkg/util"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	// query for latency (max of read and write) in milliseconds
	cassandra_latency_query = `0.001*max(cassandra_stats{name=~"org:apache:cassandra:metrics:table:(write|read)latency:99thpercentile"}) by (instance)`

	// query for transaction per second (sum of read and write)
	cassandra_ops_query = `sum(cassandra_stats{name=~"org:apache:cassandra:metrics:table:(write|read)latency:oneminuterate"}) by (instance)`

	default_Cassandra_Port = 8080
)

// Map of Turbo metric type to Cassandra query
var queryMap = map[proto.CommodityDTO_CommodityType]string{
	inter.LatencyType: cassandra_latency_query,
	inter.TpsType:     cassandra_ops_query,
}

type CassandraEntityGetter struct {
	name  string
	du    string
	query *cassandraQuery
}

// ensure CassandraEntityGetter implement the requisite interfaces
var _ alligator.EntityMetricGetter = &CassandraEntityGetter{}

func NewCassandraEntityGetter(name, du string) *CassandraEntityGetter {
	return &CassandraEntityGetter{
		name: name,
		du:   du,
	}
}

func (r *CassandraEntityGetter) Name() string {
	return r.name
}

func (r *CassandraEntityGetter) Category() string {
	return "Cassandra"
}

func (r *CassandraEntityGetter) GetEntityMetric(client *xfire.RestClient) ([]*inter.EntityMetric, error) {
	result := []*inter.EntityMetric{}
	midResult := make(map[string]*inter.EntityMetric)

	// Get metrics from Prometheus server
	for metricType := range queryMap {
		query := &cassandraQuery{queryMap[metricType]}
		metrics, err := client.GetMetrics(query)
		if err != nil {
			glog.Errorf("Failed to get Cassandra Latency metrics: %v", err)
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
func (r *CassandraEntityGetter) addEntity(mdat []xfire.MetricData, result map[string]*inter.EntityMetric, key proto.CommodityDTO_CommodityType) error {
	addrName := "instance"

	for _, dat := range mdat {
		metric, ok := dat.(*xfire.BasicMetricData)
		if !ok {
			glog.Errorf("Type assertion failed for[%v].", key)
			continue
		}

		//1. get IP
		addr, ok := metric.Labels[addrName]
		if !ok {
			glog.Errorf("Label %v is not found", addrName)
			continue
		}

		ip, port, err := util.ParseIP(addr, default_Cassandra_Port)
		if err != nil {
			glog.Errorf("Failed to parse IP from addr[%v]: %v", addr, err)
			continue
		}

		//2. add entity metrics
		entity, ok := result[ip]
		if !ok {
			entity = inter.NewEntityMetric(ip, inter.AppEntity)
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
type cassandraQuery struct {
	query string
}

func (q *cassandraQuery) GetQuery() string {
	return q.query
}

func (q *cassandraQuery) Parse(m *xfire.RawMetric) (xfire.MetricData, error) {
	d := xfire.NewBasicMetricData()
	if err := d.Parse(m); err != nil {
		return nil, err
	}

	return d, nil
}
