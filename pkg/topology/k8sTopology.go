package topology

import (
	"fmt"
	"regexp"
	"strings"

	set "github.com/deckarep/golang-set"
	"github.com/golang/glog"

	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

const (
	fqnDelim            = "/"
	appSuffix           = "app"
	containerSpecSuffix = "spec"
	workloadSuffix      = "workload"
)

var (
	deploymentNameFromPodNameRegexp, _  = regexp.Compile("^(.*)-[a-f0-9]{7,10}-[a-z0-9]{5}")
	daemonSetNameFromPodNameRegexp, _   = regexp.Compile("^(.*)-[a-f0-9]{5}")
	statefulSetNameFromPodNameRegexp, _ = regexp.Compile("^(.*)-[1-9]*[0-9]$")
	zeroValue                           = float64(0)
	oneValue                            = float64(1)
	maxValue                            = float64(10000000)
	zeroGPURequestMetrics               = []*data.DIFMetricVal{
		{
			Average:  &zeroValue,
			Capacity: &oneValue,
		},
	}
	fullGPURequestMetrics = []*data.DIFMetricVal{
		{
			Average:  &oneValue,
			Capacity: &oneValue,
		},
	}
)

type k8sTopologyHelper struct {
	entityMapById      map[string]*data.DIFEntity
	containerSetBySpec map[string]set.Set
	output             []*data.DIFEntity
}

// BuildK8sEntities constructs more k8s entities from the given metrics.
// Currently, only Nvidia GPU metrics from dcgm exporter is supported.
func BuildK8sEntities(entities []*data.DIFEntity) []*data.DIFEntity {
	helper := k8sTopologyHelper{}
	helper.entityMapById = make(map[string]*data.DIFEntity)
	helper.containerSetBySpec = make(map[string]set.Set)

	for _, entity := range entities {
		switch entity.Type {
		default:
			helper.output = append(helper.output, entity)
		case "nvidiaGPU":
			// Special handling for nvidiaGPU entities
			helper.handleNvidiaGPU(entity)
		}
	}
	helper.averageContainerSpecMetrics()
	return helper.output
}

// handleNvidiaGPU constructs a chain of k8s entities for each collected Nvidia GPU entity.
//
// The nvidiaGPU entity is a fake entity created to facilitate the aggregation of GPU and GPUMem metrics for container
// entities. Each nvidiaGPU entity represents a GPU or GPUMem metric point for a GPU device exported by the
// nvida-dcgm-exporter. Multiple metric points are then grouped by specific labels that can uniquely identify individual
// containers and aggregated for the individual containers.
//
// The aggregation is performed under the assumption that:
//   - A GPU device cannot be shared by multiple containers. That is, each container has exclusive access to the GPU
//     device assigned to it.
//   - There is only one container per pod
//   - If a pod has a workload controller, the workload controller name can be derived from the pod name
//   - MIG (Multi-Instance GPU, either mixed or single mode) is not enabled.
//
// For example, even though the following four metrics represent the GPU metrics for device 5, 4, 3 and 2 respectively,
// they are all associated with the same container: "fmaas-internal/llama-2-70b-inference-server-6794fd8b4c-bcrnx/server".
// After the aggregation, this container is using 4 GPUCores out of 4 GPUCores, i.e., 100% utilization.
//
//	{
//	  "metric": {
//	    "DCGM_FI_DRIVER_VERSION": "525.60.13",
//	    "UUID": "GPU-c06ed294-3e32-9ef7-9015-19cf243b02e8",
//	    "__name__": "DCGM_FI_DEV_GPU_UTIL",
//	    ...
//	    "device": "nvidia5",
//	    "endpoint": "gpu-metrics",
//	    "exported_container": "server",
//	    "exported_namespace": "fmaas-internal",
//	    "exported_pod": "llama-2-70b-inference-server-6794fd8b4c-bcrnx",
//	    "gpu": "5",
//	    "instance": "192.168.36.50:9400",
//	    "job": "nvidia-dcgm-exporter",
//	    "modelName": "NVIDIA A100-SXM4-80GB",
//	    ...
//	  },
//	  "value": [
//	    1700518935.082,
//	    "100"
//	  ]
//	},
//	{
//	  "metric": {
//	    "DCGM_FI_DRIVER_VERSION": "525.60.13",
//	    "UUID": "GPU-df363f22-8f9c-7d3d-fa08-2d18c13fcce5",
//	    "__name__": "DCGM_FI_DEV_GPU_UTIL",
//	    ...
//	    "device": "nvidia4",
//	    "endpoint": "gpu-metrics",
//	    "exported_container": "server",
//	    "exported_namespace": "fmaas-internal",
//	    "exported_pod": "llama-2-70b-inference-server-6794fd8b4c-bcrnx",
//	    "gpu": "4",
//	    "instance": "192.168.36.50:9400",
//	    "job": "nvidia-dcgm-exporter",
//	    "modelName": "NVIDIA A100-SXM4-80GB",
//	    ...
//	  },
//	  "value": [
//	    1700518935.082,
//	    "100"
//	  ]
//	},
//	{
//	  "metric": {
//	    "DCGM_FI_DRIVER_VERSION": "525.60.13",
//	    "UUID": "GPU-0fa3ef39-6c50-73d9-0190-35288436e917",
//	    "__name__": "DCGM_FI_DEV_GPU_UTIL",
//	    ...
//	    "device": "nvidia3",
//	    "endpoint": "gpu-metrics",
//	    "exported_container": "server",
//	    "exported_namespace": "fmaas-internal",
//	    "exported_pod": "llama-2-70b-inference-server-6794fd8b4c-bcrnx",
//	    "gpu": "3",
//	    "instance": "192.168.36.50:9400",
//	    "job": "nvidia-dcgm-exporter",
//	    "modelName": "NVIDIA A100-SXM4-80GB",
//	    ...
//	  },
//	  "value": [
//	    1700518935.082,
//	    "100"
//	  ]
//	},
//	{
//	  "metric": {
//	    "DCGM_FI_DRIVER_VERSION": "525.60.13",
//	    "UUID": "GPU-6f63ad6b-c135-78c2-8e65-2b2b8b2a59ad",
//	    "__name__": "DCGM_FI_DEV_GPU_UTIL",
//	    ...
//	    "device": "nvidia2",
//	    "endpoint": "gpu-metrics",
//	    "exported_container": "server",
//	    "exported_namespace": "fmaas-internal",
//	    "exported_pod": "llama-2-70b-inference-server-6794fd8b4c-bcrnx",
//	    "gpu": "2",
//	    "instance": "192.168.36.50:9400",
//	    "job": "nvidia-dcgm-exporter",
//	    "modelName": "NVIDIA A100-SXM4-80GB",
//	    ...
//	  },
//	  "value": [
//	    1700518935.082,
//	    "100"
//	  ]
//	},
//
// From the container pod name, we deduce a workload name for the pod based on some predefined matching rules.
// If the workload cannot be deduced, then only the node and the cluster entities are built. Otherwise, the whole
// chain will be built around the container which includes, application component, container, pod,
// container spec, workload controller and namespace. The same metric aggregations are done for all these entities as
// those done for the containers.
func (k *k8sTopologyHelper) handleNvidiaGPU(entity *data.DIFEntity) {
	// Attributes needed to identify the entities to be built
	// See config/samples/metrics_v1alpha1_nvidia-dcgm-exporter.yaml in https://github.ibm.com/turbonomic/turbo-metrics
	// repository for sample mappings between labels and attributes
	// --------------------------------------------------------------------------------
	// label                    | attribute      | note
	// --------------------------------------------------------------------------------
	// exported_container       | container      |
	// --------------------------------------------------------------------------------
	// exported_pod             | pod            |
	// --------------------------------------------------------------------------------
	// Hostname                 | nodeName       |
	// --------------------------------------------------------------------------------
	// modelName                | gpuModel       |
	// --------------------------------------------------------------------------------
	// modelName                | gpuMemSizeGb   | Extracted from modelName
	//                          |                | such as [NVIDIA A100-PCIE-40GB]
	//                          |                | to be used as gpuMem capacity
	// --------------------------------------------------------------------------------
	namespaceName := entity.GetNamespace()
	containerName := entity.GetAttribute("container")
	podName := entity.GetAttribute("pod")
	nodeName := entity.GetAttribute("nodeName")
	gpuModel := entity.GetAttribute("gpuModel")

	for _, gpuMetric := range entity.Metrics["gpu"] {
		gpuCapacity := 1.0 // 1 GPU
		gpuMetric.Capacity = &gpuCapacity
	}

	// Derive the GPU workload name using the pod name pattern
	workloadName, err := extractWorkloadNameFromPodName(podName)
	if err != nil {
		glog.Infof(err.Error())
		return
	}

	var metricsWithGpuRequest = make(map[string][]*data.DIFMetricVal)
	for kind, v := range entity.Metrics {
		metricsWithGpuRequest[kind] = v
	}
	if podName == "" {
		// This GPU is not associated with any pod, but it should still be counted towards node and cluster capacity
		metricsWithGpuRequest["gpuRequest"] = zeroGPURequestMetrics
	} else {
		metricsWithGpuRequest["gpuRequest"] = fullGPURequestMetrics
	}

	// Add cluster commodity to allow pods moving within the cluster nodes of the same GPU model
	clusterId := entity.GetClusterId()
	clusterCommodityKey := clusterId + "/" + gpuModel
	clusterCommodityMetrics := map[string][]*data.DIFMetricVal{
		"cluster": getClusterCommodity(clusterCommodityKey),
	}

	// Construct node entity
	nodeId := clusterId + fqnDelim + nodeName
	nodeEntity := k.getOrBuildEntity("virtualMachine", nodeName, nodeId, nodeId,
		metricsWithGpuRequest, clusterCommodityMetrics).
		WithControllable(true).
		WithCloneable(true).
		WithSuspendable(true)

	// Construct cluster entity
	clusterEntity := k.getOrBuildEntity("containerPlatformCluster", clusterId, clusterId, clusterId,
		metricsWithGpuRequest, map[string][]*data.DIFMetricVal{})

	if workloadName == "" {
		// No need to do the workload related topology if there's no workload on this GPU
		// But we still need to connect node with cluster
		clusterEntity.PartOfEntity("virtualMachine", nodeId, "")
		return
	}

	// Construct ID and stitching attribute
	namespaceId := strings.Join([]string{clusterId, namespaceName}, fqnDelim)
	baseName := namespaceId + fqnDelim + workloadName
	workloadFqn := baseName
	workloadId := baseName + fqnDelim + workloadSuffix
	specFqn := baseName + fqnDelim + containerName
	specId := specFqn + fqnDelim + containerSpecSuffix
	specName := workloadName + fqnDelim + containerName
	podId := strings.Join([]string{clusterId, namespaceName, podName}, fqnDelim)
	containerId := podId + fqnDelim + containerName
	appId := containerId + fqnDelim + appSuffix

	k.getOrBuildEntity("application", containerId, containerId, appId, entity.Metrics, map[string][]*data.DIFMetricVal{}).
		WithProviderMustClone(true)
	containerEntity := k.getOrBuildEntity("container", containerId, containerId, containerId, metricsWithGpuRequest, map[string][]*data.DIFMetricVal{}).
		WithControllable(true).
		WithCloneable(true).
		WithSuspendable(true).
		WithProviderMustClone(true)
	podEntity := k.getOrBuildEntity("containerPod", podName, podId, podId, metricsWithGpuRequest, clusterCommodityMetrics).
		WithControllable(true).
		WithCloneable(true).
		WithSuspendable(true)
	specEntity := k.getOrBuildEntity("containerSpec", specName, specFqn, specId, metricsWithGpuRequest, map[string][]*data.DIFMetricVal{})
	workloadEntity := k.getOrBuildEntity("workloadController", workloadName, workloadFqn, workloadId, entity.Metrics, map[string][]*data.DIFMetricVal{}).
		WithControllable(true)
	namespaceEntity := k.getOrBuildEntity("namespace", namespaceName, namespaceId, namespaceId, metricsWithGpuRequest, map[string][]*data.DIFMetricVal{})

	// Connect entities
	containerEntity.PartOfEntity("application", appId, "")
	podEntity.PartOfEntity("container", containerId, "")
	nodeEntity.PartOfEntity("containerPod", podId, "")
	specEntity.PartOfEntity("container", containerId, "")
	workloadEntity.PartOfEntity("containerSpec", specId, "").
		PartOfEntity("containerPod", podId, "")
	namespaceEntity.PartOfEntity("workloadController", workloadId, "")
	clusterEntity.PartOfEntity("namespace", namespaceId, "").
		PartOfEntity("virtualMachine", nodeId, "")

	k.addToContainerSetBySpec(containerId, specId)
}

// getOrBuildEntity creates a DIF entity if it has not been created, and then aggregates its metrics
func (k *k8sTopologyHelper) getOrBuildEntity(entityType, name, fqn, id string,
	metricsToAggregate, metricsToAdd map[string][]*data.DIFMetricVal) *data.DIFEntity {
	entity, exists := k.entityMapById[id]
	if !exists {
		entity = data.NewDIFEntity(id, entityType).WithName(name)
		entity.MatchingIdentifiers = &data.DIFMatchingIdentifiers{KubernetesFullyQualifiedName: fqn}
		for k, v := range metricsToAdd {
			entity.AddMetrics(k, v)
		}
		k.entityMapById[id] = entity
		k.output = append(k.output, entity)
	}
	entity.PairwiseAggregateAll(metricsToAggregate)
	return entity
}

// extractWorkloadNameFromPodName derives the GPU workload name using the pod name pattern.
// The workload name and the namespace form the basis naming for all the entities to be constructed.
func extractWorkloadNameFromPodName(podName string) (string, error) {
	if podName == "" {
		return "", nil
	}
	if matches := deploymentNameFromPodNameRegexp.FindStringSubmatch(podName); matches != nil {
		return matches[1], nil
	}
	if matches := daemonSetNameFromPodNameRegexp.FindStringSubmatch(podName); matches != nil {
		return matches[1], nil
	}
	if matches := statefulSetNameFromPodNameRegexp.FindStringSubmatch(podName); matches != nil {
		return matches[1], nil
	}
	return "", fmt.Errorf("unable to extract the workload name from this pod name %s; ignoring this pod", podName)
}

func getClusterCommodity(key string) []*data.DIFMetricVal {
	unit := data.COUNT
	resizable := false
	metric := data.DIFMetricVal{
		Average:   &oneValue,
		Min:       &oneValue,
		Max:       &oneValue,
		Capacity:  &maxValue,
		Unit:      &unit,
		Key:       &key,
		Resizable: &resizable,
	}
	return []*data.DIFMetricVal{&metric}
}

// addToContainerSetBySpec adds the given container id into the container set by the given container spec.
// This container set is used later to aggregate container spec metrics.
func (k *k8sTopologyHelper) addToContainerSetBySpec(containerId, specId string) {
	containerSet, exists := k.containerSetBySpec[specId]
	if !exists {
		containerSet = set.NewSet()
		k.containerSetBySpec[specId] = containerSet
	}
	containerSet.Add(containerId)
}

// averageContainerSpecMetrics averages the containerSpec metrics.
func (k *k8sTopologyHelper) averageContainerSpecMetrics() {
	for containerSpecId, containerSet := range k.containerSetBySpec {
		replicas := float64(containerSet.Cardinality())
		if containerSpecEntity, exists := k.entityMapById[containerSpecId]; exists {
			for _, metrics := range containerSpecEntity.Metrics {
				for _, val := range metrics {
					newAvg := *val.Average / replicas
					val.Average = &newAvg
					newCap := *val.Capacity / replicas
					val.Capacity = &newCap
				}
			}
		}
	}
}
