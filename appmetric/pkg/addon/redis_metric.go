package addon

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/alligator"
	"github.com/turbonomic/prometurbo/appmetric/pkg/inter"
	xfire "github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
	"strings"
)

const (
	redis_OPS_TOTAL = "redis_commands_processed_total"
	// ops_per_sec is too sensitive, so commands_processed_total will be used.
	//redis_OPS_PER_SEC = "redis_instantaneous_ops_per_sec"

	default_Redis_Port = 6379
)

type RedisEntityGetter struct {
	name  string
	query *redisQuery
}

// ensure RedisEntityGetter implement the requisite interfaces
var _ alligator.EntityMetricGetter = &RedisEntityGetter{}

func NewRedisEntityGetter(name, du string) *RedisEntityGetter {
	return &RedisEntityGetter{
		name:  name,
		query: newRedisQuery(du),
	}
}

func (r *RedisEntityGetter) Name() string {
	return r.name
}

func (r *RedisEntityGetter) Category() string {
	return "Redis"
}

func (r *RedisEntityGetter) GetEntityMetric(client *xfire.RestClient) ([]*inter.EntityMetric, error) {
	result := []*inter.EntityMetric{}
	midResult := make(map[string]*inter.EntityMetric)

	//1. get TPS data
	r.query.SetQueryType(false)
	tpsDat, err := client.GetMetrics(r.query)
	if err != nil {
		glog.Errorf("Failed to get Redis TPS metrics: %v", err)
		return result, err
	} else {
		r.addEntity(tpsDat, midResult, inter.TPS)
	}

	//2. get Latency data
	r.query.SetQueryType(true)
	latencyDat, err := client.GetMetrics(r.query)
	if err != nil {
		glog.Errorf("Failed to get Redis Latency metrics: %v", err)
		//return result, err
	} else {
		r.addEntity(latencyDat, midResult, inter.Latency)
	}

	//3. reform map to list
	for _, v := range midResult {
		result = append(result, v)
	}

	return result, nil
}

// key should be inter.TPS or inter.Latency
func (r *RedisEntityGetter) addEntity(mdat []xfire.MetricData, result map[string]*inter.EntityMetric, key string) error {
	addrName := "addr"

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

		ip, port, err := r.parseIP(addr)
		if err != nil {
			glog.Errorf("Failed to parse IP from addr[%v]: %v", addr, err)
			continue
		}

		//2. add entity metrics
		entity, ok := result[ip]
		if !ok {
			entity = inter.NewEntityMetric(ip, inter.ApplicationType)
			entity.SetLabel(inter.IP, ip)
			entity.SetLabel(inter.Port, port)
			entity.SetLabel(inter.Category, r.Category())
			result[ip] = entity
		}

		entity.SetMetric(key, metric.GetValue())
	}

	return nil
}

func (r *RedisEntityGetter) parseIP(addr string) (string, string, error) {
	addr = strings.TrimSpace(addr)
	if len(addr) < 2 {
		return "", "", fmt.Errorf("Illegal addr[%v]", addr)
	}

	items := strings.Split(addr, ":")
	if len(items) >= 2 {
		return items[0], items[1], nil
	}
	return items[0], fmt.Sprintf("%v", default_Redis_Port), nil
}

//------------------ Get and Parse the metrics ---------------
// QueryTypes
//    0: TPS
//    1: Latency
type redisQuery struct {
	qtype    int
	du       string // summary sample duration
	queryMap map[int]string
}

func newRedisQuery(du string) *redisQuery {
	q := &redisQuery{
		qtype:    0,
		du:       du,
		queryMap: make(map[int]string),
	}

	q.queryMap[0] = q.getRPSExp()
	q.queryMap[1] = q.getLatencyExp()
	return q
}

func (q *redisQuery) SetQueryType(isLatency bool) {
	if isLatency {
		q.qtype = 1
	} else {
		q.qtype = 0
	}
}

func (q *redisQuery) GetQuery() string {
	return q.queryMap[q.qtype]
}

// rate(redis_commands_processed_total[3m])
func (q *redisQuery) getRPSExp() string {
	result := fmt.Sprintf("rate(%v[%v])", redis_OPS_TOTAL, q.du)
	glog.V(3).Infof("Redis TPS: %v", result)
	return result
}

func (q *redisQuery) getLatencyExp() string {
	glog.Errorf("Redis has no Latency metric.")
	return ""
}

func (q *redisQuery) Parse(m *xfire.RawMetric) (xfire.MetricData, error) {
	d := xfire.NewBasicMetricData()
	if err := d.Parse(m); err != nil {
		return nil, err
	}

	return d, nil
}
