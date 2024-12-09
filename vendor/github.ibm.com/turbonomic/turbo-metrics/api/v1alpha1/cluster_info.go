/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// The ClusterConfiguration struct is used to configure the behavior of a Prometurbo probe when collecting
// metrics from a cluster.
type ClusterConfiguration struct {
	// The Identifier field is an optional field that specifies the cluster identifier for a Prometurbo probe.
	// If this field is not specified, the probe will default to the cluster where it is running.
	// +optional
	Identifier ClusterIdentifier `json:"identifier,omitempty"`

	// The QueryMappingSelector field is an optional field that specifies a label selector for PrometheusQueryMapping
	// resources. This field is of type *metav1.LabelSelector, which is a Kubernetes API type that represents
	// a label selector.
	// If the QueryMappingSelector field is not defined, it will default to all PrometheusQueryMapping resources in the
	// current namespace. If it is defined, it should be set to a valid label selector that can be used to identify
	// the desired resources.
	// +optional
	QueryMappingSelector *metav1.LabelSelector `json:"queryMappingSelector,omitempty"`
}

// The ClusterIdentifier struct is used to identify a Kubernetes cluster and provide labels
// to be used in PromQL queries when you are monitoring multiple Kubernetes clusters.
type ClusterIdentifier struct {
	// The ClusterLabels that store the labels that identify the cluster when executing PromQL queries
	// against the Prometheus server.
	// Use this field to specify different labels for each cluster.
	// These labels, if exists, will be used in PromQL queries to filter metrics from a specific cluster.
	// For example, the following labels could be used to select metrics from the "production" cluster in the
	// "us-west-2" region.
	//     clusterLabels := map[string]string {
	//         "cluster": "production",
	//         "region":  "us-west-2",
	//     }
	// +optional
	ClusterLabels map[string]string `json:"clusterLabels,omitempty"`

	// The unique ID of the cluster.
	// Get the ID by running the following command inside the cluster:
	//     kubectl -n default get svc kubernetes -ojsonpath='{.metadata.uid}'
	// The resulting output should be the Kubernetes service ID, which is a version 4 UUID.
	// For example, 5f2bd289-20b8-4c3c-be48-f5c5d8ff9c82.
	// +kubebuilder:validation:Pattern:="^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	ID string `json:"id"`
}

// The ClusterStatus struct defines the status of a cluster.
type ClusterStatus struct {
	// ID is the unique ID that identifies the cluster.
	ID string `json:"id"`

	// Entities is a list of EntityStatus objects.
	// This field is omitted if there are no entities found in the cluster.
	Entities []EntityStatus `json:"entities,omitempty"`

	// LastDiscoveryTime is a metav1.Time object that indicates when the cluster was last discovered.
	// This field is optional and can be omitted if the discovery time is not known.
	LastDiscoveryTime *metav1.Time `json:"lastDiscoveryTime,omitempty"`
}
