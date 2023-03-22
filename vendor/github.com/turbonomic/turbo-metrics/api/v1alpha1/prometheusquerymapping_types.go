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

// PrometheusQueryMappingSpec defines the desired state of PrometheusQueryMapping
type PrometheusQueryMappingSpec struct {
	// EntityConfigs specifies how Turbonomic entities can be mapped from Prometheus
	// query result
	// +kubebuilder:validation:MinItems:=1
	EntityConfigs []EntityConfiguration `json:"entities"`
}

type PrometheusQueryMappingStatusType string

const (
	PrometheusQueryMappingStatusOK    PrometheusQueryMappingStatusType = "ok"
	PrometheusQueryMappingStatusError PrometheusQueryMappingStatusType = "error"
)

type PrometheusQueryMappingStatusReason string

const (
	PrometheusQueryMappingInvalidPromQLSyntax        PrometheusQueryMappingStatusReason = "InvalidPromQLSyntax"
	PrometheusQueryMappingInvalidMetricDefinition    PrometheusQueryMappingStatusReason = "InvalidMetricDefinition"
	PrometheusQueryMappingInvalidAttributeDefinition PrometheusQueryMappingStatusReason = "InvalidAttributeDefinition"
)

// PrometheusQueryMappingStatus defines the observed state of PrometheusQueryMapping
type PrometheusQueryMappingStatus struct {
	// +optional
	State PrometheusQueryMappingStatusType `json:"state,omitempty"`
	// +optional
	Reason PrometheusQueryMappingStatusReason `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=pqm

// PrometheusQueryMapping is the Schema for the prometheusquerymappings API
type PrometheusQueryMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrometheusQueryMappingSpec   `json:"spec,omitempty"`
	Status PrometheusQueryMappingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PrometheusQueryMappingList contains a list of PrometheusQueryMapping
type PrometheusQueryMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrometheusQueryMapping `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PrometheusQueryMapping{}, &PrometheusQueryMappingList{})
}
