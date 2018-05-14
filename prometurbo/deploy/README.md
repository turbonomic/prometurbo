> NOTE: The user performing the steps to create a namespace, service account, and `cluster-admin` clusterrolebinding, will need to have cluster-admin role.

Deploy the prometurbo pod with the following resources:

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

Assign `cluster-admin` role by cluster role binding:
```yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1    
metadata:
  name: turbo-all-binding
  namespace: turbo
subjects:
- kind: ServiceAccount
  name: turbo-user
  namespace: turbo
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io  
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
        }
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
      serviceAccount: turbo-user
      containers:
        - name: prometurbo
          image: vmturbo/prometurbo:6.2dev
          imagePullPolicy: IfNotPresent
          args:
            - --v=2
          volumeMounts:
          - name: prometurbo-config
            mountPath: /etc/prometurbo
            readOnly: true
          - name: varlog
            mountPath: /var/log
        - image: docker.io/beekman9527/appmetric:v2
          imagePullPolicy: IfNotPresent
          name: appmetric
          args:
          - --promUrl=http://prometheus.istio-system:9090
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
