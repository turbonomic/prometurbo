package addon

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/alligator"
	"github.com/turbonomic/prometurbo/appmetric/pkg/inter"
	xfire "github.com/turbonomic/prometurbo/appmetric/pkg/prometheus"
	"math"
	"strings"
)

const (
	// NOTE: for istio 2.x, the prefix "istio_" should be removed
	turbo_SVC_LATENCY_SUM   = "istio_turbo_service_latency_time_ms_sum"
	turbo_SVC_LATENCY_COUNT = "istio_turbo_service_latency_time_ms_count"
	turbo_SVC_REQUEST_COUNT = "istio_turbo_service_request_count"

	turbo_POD_LATENCY_SUM   = "istio_turbo_pod_latency_time_ms_sum"
	turbo_POD_LATENCY_COUNT = "istio_turbo_pod_latency_time_ms_count"
	turbo_POD_REQUEST_COUNT = "istio_turbo_pod_request_count"

	//turboMetricDuration = "3m"

	k8sPrefix    = "kubernetes://"
	k8sPrefixLen = len(k8sPrefix)

	podTPS     = 0
	podLatency = 1
	svcTPS     = 2
	svcLatency = 3

	podType = 1
	svcType = 2
)

type IstioEntityGetter struct {
	name  string
	query *istioQuery
	etype int //Pod(Application), or Service
}

// ensure IstioEntityGetter implement the requisite interfaces
var _ alligator.EntityMetricGetter = &IstioEntityGetter{}

func newIstioEntityGetter(name, du string) *IstioEntityGetter {
	return &IstioEntityGetter{
		name:  name,
		etype: podType,
		query: newIstioQuery(du),
	}
}

func (istio *IstioEntityGetter) Name() string {
	return istio.name
}

func (istio *IstioEntityGetter) SetType(isVirtualApp bool) {
	if isVirtualApp {
		istio.etype = svcType
	} else {
		istio.etype = podType
	}
}

func (istio *IstioEntityGetter) Category() string {
	if istio.etype == podType {
		return "Istio"
	}

	return "Istio.VApp"
}

func (istio *IstioEntityGetter) GetEntityMetric(client *xfire.RestClient) ([]*inter.EntityMetric, error) {
	result := []*inter.EntityMetric{}

	if istio.etype == podType {
		istio.query.SetQueryType(podTPS)
	} else {
		istio.query.SetQueryType(svcTPS)
	}
	tpsDat, err := client.GetMetrics(istio.query)
	if err != nil {
		glog.Errorf("Failed to get Pod Transaction metrics: %v", err)
		return result, err
	}

	if istio.etype == podType {
		istio.query.SetQueryType(podLatency)
	} else {
		istio.query.SetQueryType(svcLatency)
	}
	latencyDat, err := client.GetMetrics(istio.query)
	if err != nil {
		glog.Errorf("Failed to get pod Latency metrics: %v", err)
		return result, err
	}

	glog.V(4).Infof("len(TPS)=%d, len(Latency)=%d", len(tpsDat), len(latencyDat))

	result = istio.mergeTPSandLatency(tpsDat, latencyDat)

	return result, nil
}

func (istio *IstioEntityGetter) assignMetric(entity *inter.EntityMetric, metric *istioMetricData) {
	for k, v := range metric.Labels {
		entity.SetLabel(k, v)
	}

	//2. other information
	entity.SetLabel(inter.Category, istio.Category())
}

func (istio *IstioEntityGetter) mergeTPSandLatency(tpsDat, latencyDat []xfire.MetricData) []*inter.EntityMetric {
	result := []*inter.EntityMetric{}
	midresult := make(map[string]*inter.EntityMetric)
	etype := inter.AppEntity
	if istio.etype == svcType {
		etype = inter.VAppEntity
	}

	for _, dat := range tpsDat {
		tps, ok := dat.(*istioMetricData)
		if !ok {
			glog.Errorf("Type assertion failed for TPS: not an IstioMetricData")
			continue
		}

		entity := inter.NewEntityMetric(tps.uuid, etype)

		istio.assignMetric(entity, tps)
		entity.SetMetric(inter.TpsType, tps.GetValue())
		midresult[entity.UID] = entity
		glog.V(5).Infof("uid=%v,uid2=%v, %+v", entity.UID, tps.uuid, entity)
	}

	for _, dat := range latencyDat {
		latency, ok := dat.(*istioMetricData)
		if !ok {
			glog.Errorf("Type assertion failed for Latency: not an IstioMetricData")
			continue
		}

		entity, exist := midresult[latency.uuid]
		if !exist {
			glog.V(3).Infof("Some entity does not have TPS metric: %+v", latency)
			entity = inter.NewEntityMetric(latency.uuid, etype)
			midresult[entity.UID] = entity
			istio.assignMetric(entity, latency)
		}
		entity.SetMetric(inter.LatencyType, latency.GetValue())
		glog.V(5).Infof("uid=%v, %+v", entity.UID, entity)
	}

	glog.V(4).Infof("len(midResult) = %d", len(midresult))

	for _, entity := range midresult {
		result = append(result, entity)
	}

	return result
}

// IstioQuery : generate queries for Istio-Prometheus metrics
// qtype 0: pod.request-per-second
//       1: pod.latency
//       2: service.request-per-second
//       3: service.latency
type istioQuery struct {
	qtype    int
	du       string
	queryMap map[int]string
}

// IstioMetricData : hold the result of Istio-Prometheus data
type istioMetricData struct {
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
	uuid   string
	dtype  int //0,1,2,3 same as qtype
}

// NewIstioQuery : create a new IstioQuery
func newIstioQuery(du string) *istioQuery {
	q := &istioQuery{
		qtype:    0,
		du:       du,
		queryMap: make(map[int]string),
	}

	isPod := true
	q.queryMap[podTPS] = q.getRPSExp(isPod)
	q.queryMap[1] = q.getLatencyExp(isPod)
	isPod = false
	q.queryMap[2] = q.getRPSExp(isPod)
	q.queryMap[3] = q.getLatencyExp(isPod)

	return q
}

func (q *istioQuery) SetQueryType(t int) error {
	if t < 0 {
		err := fmt.Errorf("Invalid query type: %d, vs 0|1|2|3", t)
		glog.Error(err)
		return err
	}

	if t > len(q.queryMap) {
		err := fmt.Errorf("Invalid query type: %d, vs 0|1|2|3", t)
		glog.Error(err)
		return err
	}

	q.qtype = t

	return nil
}

func (q *istioQuery) GetQueryType() int {
	return q.qtype
}

func (q *istioQuery) GetQuery() string {
	return q.queryMap[q.qtype]
}

func (q *istioQuery) Parse(m *xfire.RawMetric) (xfire.MetricData, error) {
	d := newIstioMetricData()
	d.SetType(q.qtype)
	if err := d.Parse(m); err != nil {
		glog.Errorf("Failed to parse metrics: %s", err)
		return nil, err
	}

	return d, nil
}

func (q *istioQuery) String() string {
	var buffer bytes.Buffer

	for k, v := range q.queryMap {
		tmp := fmt.Sprintf("qtype:%d, query=%s", k, v)
		buffer.WriteString(tmp)
	}

	return buffer.String()
}

func (q *istioQuery) getLatencyExp(pod bool) string {
	name_sum := ""
	name_count := ""
	if pod {
		name_sum = turbo_POD_LATENCY_SUM
		name_count = turbo_POD_LATENCY_COUNT
	} else {
		name_sum = turbo_SVC_LATENCY_SUM
		name_count = turbo_SVC_LATENCY_COUNT
	}

	du := q.du
	result := fmt.Sprintf("1000.0*rate(%v{response_code=\"200\"}[%v])/rate(%v{response_code=\"200\"}[%v])",
		name_sum, du, name_count, du)
	return result
}

// exp = rate(turbo_request_count{response_code="200",  source_service="unknown"}[3m])
func (q *istioQuery) getRPSExp(pod bool) string {
	name_count := ""
	if pod {
		name_count = turbo_POD_REQUEST_COUNT
	} else {
		name_count = turbo_SVC_REQUEST_COUNT
	}

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])", name_count, q.du)
	return result
}

func newIstioMetricData() *istioMetricData {
	return &istioMetricData{
		Labels: make(map[string]string),
	}
}

func (d *istioMetricData) Parse(m *xfire.RawMetric) error {
	d.Value = float64(m.Value.Value)
	if math.IsNaN(d.Value) {
		return fmt.Errorf("Failed to convert value: NaN")
	}

	labels := m.Labels

	//1. pod/svc Name
	v, ok := labels["destination_uid"]
	if !ok {
		err := fmt.Errorf("No content for destination uid: %v+", m.Labels)
		return err
	}
	uid, err := d.parseUID(v)
	if err != nil {
		glog.Errorf("Failed to parse UID(%v): %v", v, err)
		return err
	}
	d.Labels[inter.Name] = uid
	d.uuid = uid

	//2. ip
	v, ok = labels["destination_ip"]
	if !ok {
		glog.Errorf("No destination_ip label: %v", labels)
		return nil
	}

	ip, err := d.parseIP(v)
	if err != nil {
		glog.Errorf("Failed to parse IP(%v): %v", v, err)
		return nil
	}
	d.Labels[inter.IP] = ip

	//NOTO: set uuid to its IP if available
	d.uuid = ip

	//3. pod service Name and Namespace
	v, ok = labels["destination_svc_ns"]
	if !ok {
		err := fmt.Errorf("No content for destination service namespace: %v+", m.Labels)
		return err
	}

	svc_ns := strings.TrimSpace(v)
	d.Labels[inter.ServiceNamespace] = svc_ns

	v, ok = labels["destination_svc_name"]
	if !ok {
		err := fmt.Errorf("No content for destination service name: %v+", m.Labels)
		return err
	}

	svc_name := strings.TrimSpace(v)
	d.Labels[inter.ServiceName] = svc_name

	return nil
}

func (d *istioMetricData) parseUID(muid string) (string, error) {
	if d.dtype < 2 {
		return convertPodUID(muid)
	}

	return convertSVCUID(muid)
}

func (d *istioMetricData) parseIP(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "[") {
		return d.parseV04IP(raw)
	}

	if len(raw) < 1 {
		return "", fmt.Errorf("IP is empty: %v", raw)
	}

	return raw, nil
}

func (d *istioMetricData) parseService(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	return raw, nil
}

// input: [0 0 0 0 0 0 0 0 0 0 255 255 10 2 1 84]
// output: 10.2.1.84
func (d *istioMetricData) parseV04IP(raw string) (string, error) {
	if len(raw) < 7 {
		return "", fmt.Errorf("Illegal string")
	}

	content := raw[1 : len(raw)-1]
	items := strings.Split(content, " ")
	if len(items) < 4 {
		return "", fmt.Errorf("Illegal IP string: %v", raw)
	}

	i := len(items) - 4

	result := fmt.Sprintf("%v.%v.%v.%v", items[i], items[i+1], items[i+2], items[i+3])
	return result, nil
}

func (d *istioMetricData) SetType(t int) {
	d.dtype = t
}

func (d *istioMetricData) GetEntityID() string {
	return d.uuid
}

func (d *istioMetricData) GetValue() float64 {
	return d.Value
}

func (d *istioMetricData) String() string {
	var buffer bytes.Buffer

	uid := d.GetEntityID()
	content := fmt.Sprintf("uid=%v, value=%.5f", uid, d.GetValue())
	buffer.WriteString(content)

	return buffer.String()
}

// convert the UID from "kubernetes://<podName>.<namespace>" to "<namespace>/<podName>"
// for example, "kubernetes://video-671194421-vpxkh.default" to "default/video-671194421-vpxkh"
func convertPodUID(uid string) (string, error) {
	if !strings.HasPrefix(uid, k8sPrefix) {
		return "", fmt.Errorf("Not start with %v", k8sPrefix)
	}

	items := strings.Split(uid[k8sPrefixLen:], ".")
	if len(items) < 2 {
		return "", fmt.Errorf("Not enough fields: %v", uid[k8sPrefixLen:])
	}

	if len(items) > 2 {
		glog.Warningf("expected 2, got %d for: %v", len(items), uid[k8sPrefixLen:])
	}

	items[0] = strings.TrimSpace(items[0])
	items[1] = strings.TrimSpace(items[1])
	if len(items[0]) < 1 || len(items[1]) < 1 {
		return "", fmt.Errorf("Invalid fields: %v/%v", items[0], items[1])
	}

	nid := fmt.Sprintf("%s/%s", items[1], items[0])
	return nid, nil
}

// 10.10.172.236:9100
// convert UID from "svcName.namespace.svc.cluster.local" to "svcName.namespace"
// for example, "productpage.default.svc.cluster.local" to "default/productpage"
func convertSVCUID(uid string) (string, error) {
	if uid == "unknown" {
		return "", fmt.Errorf("unknown")
	}

	//1. split it
	items := strings.Split(uid, ".")
	if len(items) < 3 {
		err := fmt.Errorf("Not enough fields %d Vs. 3", len(items))
		glog.V(3).Infof(err.Error())
		return "", err
	}

	//2. check the 3rd field
	items[0] = strings.TrimSpace(items[0])
	items[1] = strings.TrimSpace(items[1])
	items[2] = strings.TrimSpace(items[2])
	if items[2] != "svc" {
		err := fmt.Errorf("%v fields[2] should be [svc]: [%v]", uid, items[2])
		glog.V(3).Infof(err.Error())
		return "", err
	}

	//3. construct the new uid
	if len(items[0]) < 1 || len(items[1]) < 1 {
		err := fmt.Errorf("Invalid fields: %v/%v", items[0], items[1])
		glog.V(3).Infof(err.Error())
		return "", err
	}

	nid := fmt.Sprintf("%s/%s", items[1], items[0])
	return nid, nil
}
