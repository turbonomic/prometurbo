# Prometurbo
Get metrics from [Prometheus](https://prometheus.io) for applications, and expose these applications and metrics in JSON format via REST API. The [`data ingestion framework probe`](https://github.com/turbonomic/data-ingestion-framework) (i.e., DIF Probe) will access the REST API, convert the JSON output to Turbonomic DTO to be consumed by Turbonomic server.


<img width="800" alt="appmetric" src="https://user-images.githubusercontent.com/10012486/80402653-34bfb780-888c-11ea-82f8-f102452047ff.png">


Applications are distinguished by their IP address. For example, each [Kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod/) Pod corresponds to one Application.

Currently, `Prometurbo` can get application metrics and attributes from Prometheus servers that are configured to scape metrices from the following exporters:
- [Istio exporter](https://istio.io/docs/reference/config/adapters/prometheus.html)
- [Redis exporter](https://github.com/oliver006/redis_exporter)
- [Cassandra exporter](https://github.com/criteo/cassandra_exporter)
- [WebDriver exporter](https://github.com/mattbostock/webdriver_exporter)
- [MySQL exporter](https://github.com/prometheus/mysqld_exporter)
- [JMX exporter](https://github.com/prometheus/jmx_exporter) 

The applications to create, as well as the queries to run to get the metrics of those applications are defined in the `configmap-prometurbo.yaml` file. The configuration can be extended to support more exporters. If you deploy with helm or operator, define additional exporters through `extraPrometheusExporters` in the value yaml.

# Output of Prometurbo: Applications with their metrics
The application metrics are served via REST API at endpoint `/metrics`. The output JSON format is defined at [turbo-go-sdk](https://github.com/turbonomic/turbo-go-sdk/tree/master/pkg/dataingestionframework/data):
```golang
type Topology struct {
	Version    string       `json:"version"`
	Updatetime int64        `json:"updateTime"`
	Scope      string       `json:"scope"`
	Source     string       `json:"source"`
	Entities   []*DIFEntity `json:"topology"`
}

type DIFEntity struct {
	UID                 string                     `json:"uniqueId"`
	Type                string                     `json:"type"`
	Name                string                     `json:"name"`
	HostedOn            *DIFHostedOn               `json:"hostedOn"`
	MatchingIdentifiers *DIFMatchingIdentifiers    `json:"matchIdentifiers"`
	PartOf              []*DIFPartOf               `json:"partOf"`
	Metrics             map[string][]*DIFMetricVal `json:"metrics"`
	partOfSet           set.Set
	hostTypeSet         set.Set
}
```


# Deploy
Follow the deployment instructions at [here](./deploy/) to deploy **Prometurbo** and **DIFProbe** container in the same Pod. 
