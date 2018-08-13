# appMetric
Get metrics from [Prometheus](https://prometheus.io) for applications, and expose these applications via REST API. [`probe`](../prometurbo) will access the REST API, and consume the results.

<img width="800" alt="appmetric" src="https://user-images.githubusercontent.com/27221807/41060294-2d58206e-699d-11e8-93f8-dae4cc775e49.png">


Applications are distinguished by mainly their IP address. For example, each [Kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod/) Pod corresponds to one Application.
Currently, it can get applications from [Istio exporter](https://istio.io/docs/reference/config/adapters/prometheus.html), [Redis exporter](https://github.com/oliver006/redis_exporter) and [Cassandra exporter](https://github.com/criteo/cassandra_exporter). More exporters can be supported by implementing
their [`addon`](https://github.com/songbinliu/appMetric/tree/v2.0/pkg/addon).

# Output of appMetric: Applications with their metrics
The application metrics are served via REST API. Access endpoint `/pod/metrics`, and will get json data:
```json
{
	"status": 0,
	"message:omitemtpy": "Success",
	"data:omitempty": [{
		"uid": "10.2.6.38",
		"type": 33,
		"labels": {
			"category": "Istio",
			"ip": "10.2.6.38",
			"name": "default/image-nkqq6"
		},
		"metrics": {
			"49": 0.2857142857142857,
			"52": 3758.488515119534
		}
	}, {
		"uid": "10.2.7.55",
		"type": 33,
		"labels": {
			"category": "Istio",
			"ip": "10.2.7.55",
			"name": "default/music-jfrpw"
		},
		"metrics": {
			"49": 3.1314285714285712,
			"52": 2388.7400252478587
		}
	}, {
		"uid": "10.2.3.31",
		"type": 33,
		"labels": {
			"category": "Redis",
			"ip": "10.2.3.31",
			"port": "6379"
		},
		"metrics": {
			"49": 1.5028571428571427
		}
	}]
}
```

The output json format is defined as:
```golang
type EntityMetric struct {
	UID     string                                       `json:"uid"`
	Type    proto.EntityDTO_EntityType                   `json:"type,omitempty"`
	Labels  map[string]string                            `json:"labels,omitempty"`
	Metrics map[proto.CommodityDTO_CommodityType]float64 `json:"metrics,omitempty"`
}

type MetricResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message:omitemtpy"`
	Data    []*EntityMetric `json:"data:omitempty"`
}

```


# Deploy
**appMetric** can be deployed in the same Pod with *Prometurbo*, as suggested [here](../deploy/). It can also be deployed
a standalone service in Kubernetes as specified in following steps.

## Prerequisites
* [Kubernetes](https://kubernetes.io) 1.7.3 +
* [Istio](https://istio.io) 0.3 + (with Prometheus addon)

## Deploy metrics and rules in Istio
Istio metrics, handlers and rules are defined in [script](https://github.com/turbonomic/prometurbo/blob/master/appmetric/scripts/istio/ip.turbo.metric.yaml), deploy it with:
```console
istioctl create -f scripts/istio/ip.turbo.metric.yaml
```
**Four Metrics**: pod latency, pod request count, service latency and service request count.

**One Handler**: a `Prometheus handler` to consume the four metrics, and generate metrics in [Prometheus](https://prometheus.io) format. This server will provide REST API to get the metrics from Prometheus.

**One Rule**: Only the `http` based metrics will be handled by the defined handler.

## Run REST API Server

#### Run in terminal
build and run this go application:
```console
make build
./_output/appMetric --v=3 --promUrl=http://localhost:9090 --port=8081
```

Then the server will serve on port `8081`; access the REST API by:
```console
curl http://localhost:8081/pod/metrics
```
```json
{"status":0,"message:omitemtpy":"Success","data:omitempty":[{"uid":"10.0.2.3","type":1,"labels":{"ip":"10.0.2.3","name":"default/curl-1xfj"},"metrics":{"latency":133.2,"tps":12}},{"uid":"10.0.3.2","type":1,"labels":{"ip":"10.0.3.2","name":"istio/music-ftaf2"},"metrics":{"latency":13.2,"tps":10}}]}
```

#### Run in docker container
```console
 docker run -d -p 18081:8081 beekman9527/appmetric:v2 --promUrl=http://10.10.200.34:9090 --v=3 --logtostderr
```

#### Deploy it in Kubernetes
This REST API service can also be deployed in Kubernetes:
```console
kubectl create -f scripts/k8s/deploy.yaml

# Access it in Kubernetes by service name:
curl http://appmetric.default:8081/service/metrics
```


