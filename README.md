# Prometurbo

Get metrics from [Prometheus](https://prometheus.io) for applications, and expose these applications and metrics in JSON
format via REST API. The [`data ingestion framework probe`](https://github.com/turbonomic/data-ingestion-framework) (
i.e. DIF Probe) will access the REST API, convert the JSON output to Turbonomic DTO to be consumed by Turbonomic
server. This enables Turbonomic to collect and analyze Prometheus metrics and make intelligent decisions about
application scaling, placement and optimization.

<img width="800" alt="appmetric" src="https://user-images.githubusercontent.com/10012486/80402653-34bfb780-888c-11ea-82f8-f102452047ff.png">

To configure the Prometheus server and map query results into applications and metrics, you need to create the following
custom resources in the Kubernetes cluster (from Prometurbo **8.8.4**):

* [PrometheusQueryMapping](https://pkg.go.dev/github.com/turbonomic/turbo-metrics@v0.0.0-20230222215340-3cdff28ffdaf/api/v1alpha1#PrometheusQueryMapping):
  allows users to define mappings between Turbonomic entities (such as **ApplicationComponents**, **Services**, or **VirtualMachines**) and Prometheus metrics exposed by different prometheus exporters.
* [PrometheusServerConfig](https://pkg.go.dev/github.com/turbonomic/turbo-metrics@v0.0.0-20230222215340-3cdff28ffdaf/api/v1alpha1#PrometheusServerConfig):
  specifies the address of the Prometheus server, as well as optional label selectors to filter
  out [PrometheusQueryMapping](https://pkg.go.dev/github.com/turbonomic/turbo-metrics@v0.0.0-20230222215340-3cdff28ffdaf/api/v1alpha1#PrometheusQueryMapping)
  resources applicable to that server. This allows users to configure multiple Prometheus servers and use different
  mappings for each server.

Custom resource definitions for the above two resources must be installed first in the Kubernetes cluster. Get
them [here](https://github.com/turbonomic/turbo-metrics/tree/main/config/crd/bases).

Sample custom resource instances can be
found [here](https://github.com/turbonomic/turbo-metrics/tree/main/config/samples).

# Output of Prometurbo: Applications with their metrics

The application metrics are served via REST API at endpoint `/metrics`. The output JSON format is defined
at [turbo-go-sdk](https://github.com/turbonomic/turbo-go-sdk/tree/master/pkg/dataingestionframework/data):

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

Follow the deployment instructions at [here](./deploy/) to deploy **Prometurbo** and **DIFProbe** container in the same
Pod.
