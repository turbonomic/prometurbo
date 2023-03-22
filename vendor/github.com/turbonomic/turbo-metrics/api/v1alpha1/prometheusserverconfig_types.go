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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PrometheusServerConfigSpec defines the desired state of PrometheusServerConfig
type PrometheusServerConfigSpec struct {
	// Address of the Prometheus server.
	Address string `json:"address"`

	// ClusterConfigs is an optional list of ClusterConfiguration structs that specify information about the clusters
	// that the Prometheus server should obtain metrics for.
	// If this field is not specified, the Prometheus server obtains metrics only for the cluster where the
	// Prometurbo probe is running.
	// +optional
	ClusterConfigs []ClusterConfiguration `json:"clusters,omitempty"`
}

type PrometheusServerConfigStatusType string

const (
	PrometheusServerConfigStatusOK    PrometheusServerConfigStatusType = "ok"
	PrometheusServerConfigStatusError PrometheusServerConfigStatusType = "error"
)

type PrometheusServerConfigStatusReason string

const (
	PrometheusServerConfigConnectionFailure     PrometheusServerConfigStatusReason = "ConnectionFailure"
	PrometheusServerConfigAuthenticationFailure PrometheusServerConfigStatusReason = "AuthenticationFailure"
)

// PrometheusServerConfigStatus defines the observed state of PrometheusServerConfig
type PrometheusServerConfigStatus struct {
	// +optional
	State PrometheusServerConfigStatusType `json:"state,omitempty"`

	// +optional
	Reason PrometheusServerConfigStatusReason `json:"reason,omitempty"`

	// +optional
	Message string `json:"message,omitempty"`

	// +optional
	Clusters []ClusterStatus `json:"clusters,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=psc

// PrometheusServerConfig is the Schema for the prometheusserverconfigs API
type PrometheusServerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrometheusServerConfigSpec   `json:"spec,omitempty"`
	Status PrometheusServerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PrometheusServerConfigList contains a list of PrometheusServerConfig
type PrometheusServerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrometheusServerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PrometheusServerConfig{}, &PrometheusServerConfigList{})
}
