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

// The MetricConfiguration is a struct that represents the configuration for a specific type of metric.
type MetricConfiguration struct {
	// Type specifies the type of metric
	// +kubebuilder:validation:Enum=responseTime;transaction;heap;collectionTime;cacheHitRate;dbMem;cpu;memory
	Type string `json:"type"`

	// QueryConfigs is a list of QueryConfiguration structs.
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	QueryConfigs []QueryConfiguration `json:"queries"`
}

// The QueryConfiguration struct represents a specific query that will be used to collect data for the metric.
type QueryConfiguration struct {
	// Type specifies the subtype of metric, for example, "used", "capacity", or "peak".
	// +kubebuilder:validation:Enum=used;capacity;peak
	Type string `json:"type"`

	// PromQL is a string that contains the PromQL query that will be used to collect data for the metric.
	PromQL string `json:"promql"`
}
