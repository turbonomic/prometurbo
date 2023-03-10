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

// The EntityConfiguration defines the configuration for mapping from Prometheus query result
// to a specific type of Turbonomic entity.
type EntityConfiguration struct {
	// Type specifies the type of entity.
	// This field is required and must be one of application, databaseServer, or virtualMachine.
	// +kubebuilder:validation:Enum=application;databaseServer;virtualMachine
	Type string `json:"type"`

	// HostedOnVM specifies if an entity is hosted on VM
	// If not set, the entity is assumed to be hosted on a container
	HostedOnVM bool `json:"hostedOnVM,omitempty"`

	// MetricConfigs is a list of MetricConfiguration objects that specify how to collect metrics for the entity.
	// This field is required and must contain at least one metric configuration.
	// +kubebuilder:validation:MinItems=1
	MetricConfigs []MetricConfiguration `json:"metrics"`

	// AttributeConfigs is a list of AttributeConfiguration objects that specify how to map labels into attributes
	// of the entity. This field is required and must contain at least one attribute configuration.
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	AttributeConfigs []AttributeConfiguration `json:"attributes"`
}

// The AttributeConfiguration specifies how to map labels from Prometheus metrics into attributes of an entity.
type AttributeConfiguration struct {
	// The name of the attribute
	Name string `json:"name"`

	// Label is the name of the label that contains the value for this attribute.
	// If the Matches field is not specified, the value of this label will be used as the attribute value.
	Label string `json:"label"`

	// Matches is an optional regular expression that can be used to extract a pattern from the label value and
	// use that as the attribute value.
	// +optional
	Matches string `json:"matches,omitempty"`

	// As is an optional field that specifies how to reconstruct the extracted patterns from the result of the
	// Matches field and use that as the attribute value instead. This field is only evaluated when the Matches
	// field is specified.
	// +optional
	As string `json:"as,omitempty"`

	// IsIdentifier is an optional field that specifies if this attribute should be used as the identifier of an entity.
	// There should be one and only one identifier for an entity.
	// +optional
	IsIdentifier bool `json:"isIdentifier,omitempty"`
}

// The EntityStatus represents the status of an entity in a cluster.
type EntityStatus struct {
	// Type is a string that specifies the type of entity.
	Type string `json:"type"`

	// Count is a pointer to an int32 that represents the number of entities of this type in the cluster.
	// If this field is nil, it means that the number of entities is unknown or has not been discovered yet.
	Count *int32 `json:"count"`
}
