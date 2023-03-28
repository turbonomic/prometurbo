
## Deploy Prometurbo

It is recommended to deploy Prometurbo via operator. The following is an example of deploying Prometurbo through Openshift Operator Hub, create a [PrometheusQueryMapping](https://pkg.go.dev/github.com/turbonomic/turbo-metrics@v0.0.0-20230222215340-3cdff28ffdaf/api/v1alpha1#PrometheusQueryMapping), and a [PrometheusServerConfig](https://pkg.go.dev/github.com/turbonomic/turbo-metrics@v0.0.0-20230222215340-3cdff28ffdaf/api/v1alpha1#PrometheusServerConfig) to be consumed by Prometurbo.

### Install Prometurbo Operator through Openshift OperatorHub

* Create a project (namespace) for your `prometurbo` deployment. For example, the following YAML file creates a `turbo` namespace: 

  ```yaml
  apiVersion: v1
  kind: Namespace
  metadata:
    name: turbo 
  ```

* On the Openshift admin console, navigate to **Operators**, **OperatorHub**, select the project created above, search for `Prometurbo Operator`, and select the Certified Prometurbo Operator:

  ![image](https://user-images.githubusercontent.com/10012486/228285170-fe0c14da-b47f-4007-89e6-849078102563.png)

* Make sure the `stable` channel is selected, and install the operator:

  ![image](https://user-images.githubusercontent.com/10012486/228285768-295ee411-6e95-4cae-b5e3-5f3e7523018f.png)

  ![image](https://user-images.githubusercontent.com/10012486/228285895-f024cc8f-7a1f-45f5-aa86-0e7048b8eb68.png)

* After the install, the following resources are created:

  **Name**	| **Kind**	| **Status** |	**API version**
  ---      | ---      | ---        | ---
  prometurbo-operator.vx.x.x | `ClusterServiceVersion`	| Created	| `operators.coreos.com/v1alpha1`
  prometurbos.charts.helm.k8s.io | `CustomResourceDefinition`	| Present	| `apiextensions.k8s.io/v1`
  prometurbo-operator | `ServiceAccount`	| Present	| `core/v1`
  prometurbo-operator.vx.x.x-xxxxxxxx | `ClusterRole`	| Created	| `rbac.authorization.k8s.io/v1`
  prometurbo-operator.vx.x.x-xxxxxxxx | `ClusterRoleBinding`	| Created	| `rbac.authorization.k8s.io/v1`
  
  Note that `ClusterRole` is created for `prometurbo-operator` such that it can have the permission to create necessary `ClusterRole`s for `prometurbo` instances.
  
### Deploy a Prometurbo instance

* The following table lists the configuration parameters for `prometurbo`:

   **Parameter**                | **Description**                       | **Default Value** 
    ----------------------------|---------------------------------------|--------------
   `image.prometurboRepository` | The `prometurbo` container image repository. | `icr.io/cpopen/turbonomic/prometurbo`
   `image.prometurboTag`        | The `prometurbo` container image tag. | release version, such as `8.8.4`
   `image.turbodifRepository`   | The `turbodif` container image repository. | `icr.io/cpopen/turbonomic/turbodif`
   `image.turbodifTag`          | The `turbodif` container image tag. | release version, such as `8.8.4`
   `image.pullPolicy`           | Specify either `IfNotPresent`, or `Always`. | `IfNotPresent`
   `targetName`                 | A unique name to identify this target. | `Prometheus`
   `targetTypeSuffix`           | A unique suffix to the `DataIngestionFramework` target type. The resulting target type becomes **DataIngestionFramework-`targetTypeSuffix`** on the UI. Do not specify `Turbonomic` as it is reserved for internal use. | `Prometheus`
   `serviceAccountName`         | The name of the `serviceAccount` used by `prometurbo` pod. | `prometurbo`
   `roleName`                   | The name of the `clusterrole` bound to the above service account. | `prometurbo`
   `roleBinding`                | The name of the `clusterrolebinding` that binds the above service account to the above cluster role. | `prometurbo-binding`
   `restAPIConfig.turbonomicCredentialsSecretName` | The name of the secret that contains Turbonomic server username and password. Required if not using cleartext username and password, and not taking the default secret name. If the secrect does not exist, `prometurbo` falls back to cleartext username and password. | `turbonomic-credentials` 
   `restAPIConfig.opsManagerUserName` | The username to login to the Turbonomic server. Required if not using secret. |
   `restAPIConfig.opsManagerPassword` | The password to login to the Turbonomic server. Required if not using secret. |
   `serverMeta.version`         | Turbonomic server version.             | release version, such as `8.8.4`
   `serverMeta.turboServer`     | Turbonomic server URL.                 |
   `args.logginglevel`          | Logging level of `prometurbo`.         | `2`
   `args.ignoreCommodityIfPresent` |  Specify whether to ignore merging commodity when a commodity of the same type already exists in the server. | `false` 

* The following is a sample `prometurbo` resource YAML file:

  ```yaml
  apiVersion: charts.helm.k8s.io/v1
  kind: Prometurbo
  metadata:
    name: prometurbo-release
    namespace: turbo
  spec:
    image:
      prometurboTag: 8.8.4
      turbodifTag: 8.8.4
      pullPolicy: Always
    restAPIConfig:
      opsManagerPassword: administrator
      opsManagerUserName: administrator
    serviceAccountName: prometurbo-xl-ember
    roleBinding: prometurbo-binding-xl-ember
    roleName: prometurbo-xl-ember
    serverMeta:
      turboServer: 'https://9.46.114.249'
      version: 8.8.4
    targetName: OCP411-FYRE-IBM
  ```

* Verify the install by making sure that `prometurbo` pods are up and running:

```console
$ oc -n turbo get po | grep prometurbo
prometurbo-operator-6ffc566f4c-lwbcn       1/1     Running   0          5d23h
prometurbo-release-744947bb94-kqc2b        2/2     Running   0          5d20h
```

### Install Custom Resource Definition

```
$ oc create -f https://raw.githubusercontent.com/turbonomic/turbo-metrics/main/config/crd/bases/metrics.turbonomic.io_prometheusquerymappings.yaml
$ oc create -f https://raw.githubusercontent.com/turbonomic/turbo-metrics/main/config/crd/bases/metrics.turbonomic.io_prometheusserverconfigs.yaml
```


