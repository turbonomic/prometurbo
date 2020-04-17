# appMetric
Get metrics from [Prometheus](https://prometheus.io) for applications, and expose these applications via REST API. [`probe`](../prometurbo) will access the REST API, and consume the results.

<img width="800" alt="appmetric" src="https://user-images.githubusercontent.com/27221807/41060294-2d58206e-699d-11e8-93f8-dae4cc775e49.png">


Applications are distinguished by their IP address. For example, each [Kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod/) Pod corresponds to one Application.

Currently, `appMetric` can get application metrics and attributes from [Istio exporter](https://istio.io/docs/reference/config/adapters/prometheus.html), [Redis exporter](https://github.com/oliver006/redis_exporter), [Cassandra exporter](https://github.com/criteo/cassandra_exporter), [WebDriver exporter](https://github.com/mattbostock/webdriver_exporter), [MySQL exporter](https://github.com/prometheus/mysqld_exporter), and [JMX exporter](https://github.com/prometheus/jmx_exporter) based on the definition in the `appmetric-config.yaml` configuration file. More exporters can be supported by specifying their definition in the `appmetric-config.yml` configuration file.

# Output of appMetric: Applications with their metrics
The application metrics are served via REST API. Access endpoint `/pod/metrics`, and will get json data:
```json
{
  "status": 0,
  "message:omitemtpy": "Success",
  "data:omitempty": [
    {
      "uid": "10.10.169.38",
      "type": 4,
      "hostedOnVM": true,
      "labels": {
        "business_app": "xl",
        "ip": "10.10.169.38"
      },
      "metrics": {
        "49": {
          "used": 0.008333333333333333
        },
        "67": {
          "capacity": 131072,
          "used": 109616
        },
        "69": {
          "used": 95.80342211850882
        }
      }
    },
    {
      "uid": "10.10.169.38",
      "type": 10,
      "hostedOnVM": false,
      "labels": {
        "ip": "10.10.169.38",
        "target": "kubernetes-service-endpoints"
      },
      "metrics": {
        "26": {
          "used": 3.813936434391914
        },
        "53": {
          "used": 19748298752
        }
      }
    },
    {
      "uid": "10.233.90.42",
      "type": 33,
      "hostedOnVM": false,
      "labels": {
        "business_app": "xl",
        "ip": "10.233.90.42"
      },
      "metrics": {
        "1": {
          "capacity": 66,
          "used": 46
        },
        "71": {
          "used": 0.023234433925960134
        },
        "77": {
          "capacity": 262144,
          "used": 68836.046875
        }
      }
    },
  ]
}  
```

The output json format is defined as:
```golang
type EntityMetric struct {
	UID        string                                                  `json:"uid"`
	Type       proto.EntityDTO_EntityType                              `json:"type,omitempty"`
	HostedOnVM bool                                                    `json:"hostedOnVM"`
	Labels     map[string]string                                       `json:"labels,omitempty"`
	Metrics    map[proto.CommodityDTO_CommodityType]map[string]float64 `json:"metrics,omitempty"`
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
./_output/appMetric --v=3 --port=8081
```

Then the server will serve on port `8081`; access the REST API by:
```console
curl http://localhost:8081/metrics
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


