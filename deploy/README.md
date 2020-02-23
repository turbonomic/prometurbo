
Prometurbo can be deployed in Kubernetes with the following steps:

0. Setup Istio prometheus exporter

Creating some Istio resources to collect  http-related metrics of the Pods and Services. 

The definition of these Istio metrics, handlers and rule are defined in [`appmetric/scripts/istio/ip.turbo.metric.yaml`](../appmetric/scripts/istio/ip.turbo.metric.yaml), and can be deployed with:

```bash
istioctl create -f appmetric/scripts/istio/ip.turbo.metric.yaml
```
 
 With these resources, `Response time` and `Transactions` of Applications can be monitored through Istio.
 

1. Create a namespace

Use an existing namespace, or create one where to deploy prometurbo. The yaml examples will use `turbo`.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: turbo 
```

2. Create a service account, and add the role of cluster-admin
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: turbo-user
  namespace: turbo
```

3. create a configMap for prometurbo, The <TURBONOMIC_SERVER_VERSION> is Turbonomic release version, e.g. 6.2.0
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometurbo-config
data:
  turbo.config: |-
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
        "prometurboTargetConfig": {
            "targetAddress":"<PROMETHEUS-SERVER-ADDRESS>",
            "scope":"<THE-K8S-TARGET-NAME>"
        },
        "targetTypeSuffix": "" <-- Adjust this value as necessary. No suffix is appended to target name if empty.
    }
```


4. Create a deployment for prometurbo
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: prometurbo
  namespace: turbo
  labels:
    app: prometurbo
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: prometurbo
    spec:
      containers:
        - name: prometurbo
          image: turbonomic/prometurbo:7.21.0
          imagePullPolicy: IfNotPresent
          args:
            - --v=2
          volumeMounts:
          - name: prometurbo-config
            mountPath: /etc/prometurbo
            readOnly: true
          - name: varlog
            mountPath: /var/log
        - image: turbonomic/appmetric:7.21.0
          imagePullPolicy: IfNotPresent
          name: appmetric
          args:
            - --v=2
          ports:
          - containerPort: 8081
      volumes:
      - name: prometurbo-config
        configMap: 
          name: prometurbo-config
      - name: varlog
        emptyDir: {}
      restartPolicy: Always
```
