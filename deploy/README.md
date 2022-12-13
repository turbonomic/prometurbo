
Prometurbo can be deployed in Kubernetes with the following steps:

## Setup prometheus exporters

The following is an example on how to set up and configure an Istio prometheus exporter in an Istio 1.4 environment:

Creating some Istio resources to collect  http-related metrics of the Pods and Services. 

The definition of these Istio metrics, handlers and rule are defined in [`scripts/istio/ip.turbo.metric.istio-1.4.yaml`](../scripts/istio/ip.turbo.metric.istio-1.4.yaml), and can be deployed with:

```bash
istioctl create -f scripts/istio/ip.turbo.metric.istio-1.4.yaml
```
 
 With these resources, `Response time` and `Transactions` of Applications can be monitored through Istio.
 

## Create a namespace

Use an existing namespace, or create one where to deploy prometurbo. The yaml examples will use `turbo`.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: turbo 
```

## Create a service account, and add the role of cluster-admin
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: turbo-user
  namespace: turbo
```

## Create a configMap for Prometurbo
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometurbo-config
data:
  prometheus.config: |-
    # A map of prometheus servers and metrics to scrape
    servers:
      # The unique name of the prometheus server
      server1:
        # The URL of the prometheus server
        url: http://Prometheus_Server_URL
        # The list of configured exporters to discover entities and metrics
        exporters:
          - cassandra
          - istio
          - jmx-tomcat
          - node
          - redis
          - webdriver
    # A map of exporter configurations to discover entities and related metrics
    exporters:
      istio:
        entities:
...
...
```

## Create a configMap for Turbodif
The <TURBONOMIC_SERVER_VERSION> is the release version of Turbonomic release, e.g. 7.22.0
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: turbodif-config
data:
  turbodif-config.json: |-
    {
        "communicationConfig": {
            "serverMeta": {
                "version": "<TURBONOMIC_SERVER_VERSION>",
                "turboServer": "https://<TURBO-SERVER-ADDRESS>:<PORT>"
            },
            "restAPIConfig": {
                "opsManagerUserName": "administrator",
                "opsManagerPassword": "<TURBO-SERVER-PASSWORD>"
            }
        },
        "targetConfig": {
            "targetName": "Prometheus",
            "targetAddress": "http://127.0.0.1:8081/metrics"
        },
        "targetTypeSuffix": "Prometheus" <-- Adjust this value as necessary to change the DIFProbe type name.
    }
```

## Create a deployment for prometurbo
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometurbo
  labels:
    app: prometurbo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometurbo
  template:
    metadata:
      labels:
        app: prometurbo
    spec:
      containers:
        # Replace the image with desired version:8.6.3 or snapshot version:8.7.3-SNAPSHOT from icr.io
        - image: icr.io/cpopen/turbonomic/prometurbo:8.6.3
          imagePullPolicy: IfNotPresent
          name: prometurbo
          args:
            - --v=2
          ports:
            - containerPort: 8081
          volumeMounts:
            - name: prometurbo-config
              mountPath: /etc/prometurbo
              readOnly: true
        - name: turbodif
          # Replace the image with desired version:8.6.3 or snapshot version:8.7.3-SNAPSHOT from icr.io
          image: icr.io/cpopen/turbonomic/turbodif:8.6.3
          imagePullPolicy: IfNotPresent
          args:
            - --v=2
          volumeMounts:
          - name: turbodif-config
            mountPath: /etc/turbodif
            readOnly: true
          - name: varlog
            mountPath: /var/log
      volumes:
        - name: prometurbo-config
          configMap:
            name: prometurbo-config
        - name: turbodif-config
          configMap:
            name: turbodif-config
        - name: varlog
          emptyDir: {}
      restartPolicy: Always

```
