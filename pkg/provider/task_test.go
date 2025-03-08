package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/turbonomic/turbo-metrics/api/v1alpha1"
)

var (
	taskWithNoId           = &Task{}
	taskWithClusterIdK8sId = &Task{
		clusterId: &v1alpha1.ClusterIdentifier{
			ClusterLabels: map[string]string{
				"cluster": "clusterA",
			},
			ID: "5f2bd289",
		},
		k8sSvcId: "cda4d884-a053-4aba-8576-afa5d923e7c6",
	}
	taskWithK8sIdOnly = &Task{
		k8sSvcId: "cda4d884-a053-4aba-8576-afa5d923e7c6",
	}
	entityAttr = &EntityAttribute{
		ID: "10.254.15.158-demoapp",
		IP: "10.254.15.158",
	}
)

func TestGetMatchingAttributeWithClusterId(t *testing.T) {
	matchingAttrForAppOnVM := taskWithClusterIdK8sId.getMatchingAttribute(true, entityAttr)
	assert.Equal(t, "10.254.15.158", matchingAttrForAppOnVM)
	matchingAttrForAppOnContainer := taskWithClusterIdK8sId.getMatchingAttribute(false, entityAttr)
	assert.Equal(t, "10.254.15.158-5f2bd289", matchingAttrForAppOnContainer)
}

func TestGetMatchingAttributeWithK8sId(t *testing.T) {
	matchingAttrForAppOnVM := taskWithK8sIdOnly.getMatchingAttribute(true, entityAttr)
	assert.Equal(t, "10.254.15.158", matchingAttrForAppOnVM)
	matchingAttrForAppOnContainer := taskWithK8sIdOnly.getMatchingAttribute(false, entityAttr)
	assert.Equal(t, "10.254.15.158-cda4d884-a053-4aba-8576-afa5d923e7c6", matchingAttrForAppOnContainer)
}

func TestGetMatchingAttributeWithNoId(t *testing.T) {
	matchingAttrForAppOnVM := taskWithNoId.getMatchingAttribute(true, entityAttr)
	assert.Equal(t, "10.254.15.158", matchingAttrForAppOnVM)
	matchingAttrForAppOnContainer := taskWithNoId.getMatchingAttribute(false, entityAttr)
	assert.Equal(t, "10.254.15.158", matchingAttrForAppOnContainer)
}

func TestGetEntityIdWithClusterId(t *testing.T) {
	entityIdForAppOnVM := taskWithClusterIdK8sId.getEntityId(true, entityAttr)
	assert.Equal(t, "10.254.15.158-demoapp", entityIdForAppOnVM)
	entityIdForAppOnContainer := taskWithClusterIdK8sId.getEntityId(false, entityAttr)
	assert.Equal(t, "10.254.15.158-demoapp-5f2bd289", entityIdForAppOnContainer)
}

func TestGetEntityIdWithK8sId(t *testing.T) {
	entityIdForAppOnVM := taskWithK8sIdOnly.getEntityId(true, entityAttr)
	assert.Equal(t, "10.254.15.158-demoapp", entityIdForAppOnVM)
	entityIdForAppOnContainer := taskWithK8sIdOnly.getEntityId(false, entityAttr)
	assert.Equal(t, "10.254.15.158-demoapp-cda4d884-a053-4aba-8576-afa5d923e7c6", entityIdForAppOnContainer)
}

func TestGetEntityIdWithNoId(t *testing.T) {
	entityIdForAppOnVM := taskWithNoId.getEntityId(true, entityAttr)
	assert.Equal(t, "10.254.15.158-demoapp", entityIdForAppOnVM)
	entityIdForAppOnContainer := taskWithNoId.getEntityId(false, entityAttr)
	assert.Equal(t, "10.254.15.158-demoapp", entityIdForAppOnContainer)
}
