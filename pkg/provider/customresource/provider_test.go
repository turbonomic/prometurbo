package customresource

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/turbonomic/turbo-metrics/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	prometheusQueryMappingTypeMeta = v1.TypeMeta{
		Kind:       "PrometheusQueryMapping",
		APIVersion: "metrics.turbonomic.io/v1alpha1",
	}
	prometheusServerConfigTypeMeta = v1.TypeMeta{
		Kind:       "PrometheusServerConfig",
		APIVersion: "metrics.turbonomic.io/v1alpha1",
	}
	istioObjeMeta = v1.ObjectMeta{
		Name:      "istio",
		Namespace: "turbo",
		UID:       "3de0da57-0381-4f8e-a61d-05b6390bdb2f",
	}
	istioSpec = v1alpha1.PrometheusQueryMappingSpec{
		EntityConfigs: []v1alpha1.EntityConfiguration{
			{
				Type: "application",
				MetricConfigs: []v1alpha1.MetricConfiguration{
					{
						Type: "responseTime",
						QueryConfigs: []v1alpha1.QueryConfiguration{
							{
								Type:   "used",
								PromQL: "'rate(istio_request_duration_milliseconds_sum{request_protocol=\"http\",response_code=\"200\",reporter=\"destination\"}[1m])/rate(istio_request_duration_milliseconds_count{}[1m]) >= 0'",
							},
						},
					},
					{
						Type: "transaction",
						QueryConfigs: []v1alpha1.QueryConfiguration{
							{
								Type:   "used",
								PromQL: "'rate(istio_requests_total{request_protocol=\"http\",response_code=\"200\",reporter=\"destination\"}[1m]) > 0'",
							},
						},
					},
					{
						Type: "responseTime",
						QueryConfigs: []v1alpha1.QueryConfiguration{
							{
								Type:   "used",
								PromQL: "'rate(istio_request_duration_milliseconds_sum{request_protocol=\"grpc\",grpc_response_status=\"0\",response_code=\"200\",reporter=\"destination\"}[1m])/rate(istio_request_duration_milliseconds_count{}[1m]) >= 0'",
							},
						},
					},
					{
						Type: "transaction",
						QueryConfigs: []v1alpha1.QueryConfiguration{
							{
								Type:   "used",
								PromQL: "'rate(istio_requests_total{request_protocol=\"grpc\",grpc_response_status=\"0\",response_code=\"200\",reporter=\"destination\"}[1m]) > 0'",
							},
						},
					},
				},
				AttributeConfigs: []v1alpha1.AttributeConfiguration{
					{
						Name:         "ip",
						Label:        "instance",
						Matches:      "\\d{1,3}(?:\\.\\d{1,3}){3}(?::\\d{1,5})??",
						IsIdentifier: true,
					},
					{
						Name:  "namespace",
						Label: "destination_service_namespace",
					},
					{
						Name:  "service",
						Label: "destination_service_name",
					},
				},
			},
		},
	}
	jmxTomcatObjMeta = v1.ObjectMeta{
		Name:      "jmx-tomcat",
		Namespace: "turbo",
		UID:       "76493649-2314-4168-b7e7-41973509a2ca",
	}
	jmxTomcatSpec = v1alpha1.PrometheusQueryMappingSpec{
		EntityConfigs: []v1alpha1.EntityConfiguration{
			{
				Type:       "application",
				HostedOnVM: true,
				MetricConfigs: []v1alpha1.MetricConfiguration{
					{
						Type: "cpu",
						QueryConfigs: []v1alpha1.QueryConfiguration{
							{
								Type:   "used",
								PromQL: "'java_lang_OperatingSystem_ProcessCpuLoad'",
							},
						},
					},
					{
						Type: "memory",
						QueryConfigs: []v1alpha1.QueryConfiguration{
							{
								Type:   "used",
								PromQL: "'java_lang_Memory_HeapMemoryUsage_used/1024'",
							},
							{
								Type:   "capacity",
								PromQL: "'java_lang_Memory_HeapMemoryUsage_max/1024'",
							},
						},
					},
					{
						Type: "collectionTime",
						QueryConfigs: []v1alpha1.QueryConfiguration{
							{
								Type:   "used",
								PromQL: "'sum without (name) (delta(java_lang_GarbageCollector_CollectionTime)[10m])/600*100'",
							},
						},
					},
					{
						Type: "responseTime",
						QueryConfigs: []v1alpha1.QueryConfiguration{
							{
								Type:   "used",
								PromQL: "'rate(Catalina_GlobalRequestProcessor_processingTime{name=~\".*http-.*\"}[3m])'",
							},
						},
					},
				},
				AttributeConfigs: []v1alpha1.AttributeConfiguration{
					{
						Name:         "ip",
						Label:        "instance",
						Matches:      "\\d{1,3}(?:\\.\\d{1,3}){3}(?::\\d{1,5})??",
						IsIdentifier: true,
					},
				},
			},
		},
	}
	singleClusterObjMeta = v1.ObjectMeta{
		Name:      "singlecluster",
		Namespace: "turbo",
		UID:       "a706ef48-4cc9-41e2-854a-cc8b55e0605a",
	}
	singleClusterSvrConfigSpec = v1alpha1.PrometheusServerConfigSpec{
		Address: "http://prometheus.istio-system:9090",
		ClusterConfigs: []v1alpha1.ClusterConfiguration{
			{
				QueryMappingSelector: &v1.LabelSelector{
					MatchExpressions: []v1.LabelSelectorRequirement{
						{
							Key:      "mapping",
							Operator: v1.LabelSelectorOpNotIn,
							Values:   []string{"jmx-tomcat"},
						},
					},
				},
			},
		},
	}
	multiClusterObjMeta = v1.ObjectMeta{
		Name:      "multicluster",
		Namespace: "turbo",
		UID:       "cf3f16d0-59a7-4792-baca-a57341bcece6",
	}
	multiClusterSvrConfigSpec = v1alpha1.PrometheusServerConfigSpec{
		Address: "https://observatorium-api-open-cluster-management-observability.apps.cluster-nbx49.com:9090",
		ClusterConfigs: []v1alpha1.ClusterConfiguration{
			{
				Identifier: v1alpha1.ClusterIdentifier{
					ClusterLabels: map[string]string{
						"cluster": "clusterA",
					},
					ID: "5f2bd289",
				},
				QueryMappingSelector: &v1.LabelSelector{
					MatchExpressions: []v1.LabelSelectorRequirement{
						{
							Key:      "mapping",
							Operator: v1.LabelSelectorOpNotIn,
							Values:   []string{"istio"},
						},
					},
				},
			},
			{
				Identifier: v1alpha1.ClusterIdentifier{
					ClusterLabels: map[string]string{
						"cluster": "clusterB",
					},
					ID: "936056e5",
				},
			},
		},
	}
)

func createIstio() v1alpha1.PrometheusQueryMapping {
	return v1alpha1.PrometheusQueryMapping{
		TypeMeta:   prometheusQueryMappingTypeMeta,
		ObjectMeta: istioObjeMeta,
		Spec:       istioSpec,
	}
}

func createJmxTomcat() v1alpha1.PrometheusQueryMapping {
	return v1alpha1.PrometheusQueryMapping{
		TypeMeta:   prometheusQueryMappingTypeMeta,
		ObjectMeta: jmxTomcatObjMeta,
		Spec:       jmxTomcatSpec,
	}
}

func createSingleSvrConfig() v1alpha1.PrometheusServerConfig {
	return v1alpha1.PrometheusServerConfig{
		TypeMeta:   prometheusServerConfigTypeMeta,
		ObjectMeta: singleClusterObjMeta,
		Spec:       singleClusterSvrConfigSpec,
	}
}

func createMultiSvrConfig() v1alpha1.PrometheusServerConfig {
	return v1alpha1.PrometheusServerConfig{
		TypeMeta:   prometheusServerConfigTypeMeta,
		ObjectMeta: multiClusterObjMeta,
		Spec:       multiClusterSvrConfigSpec,
	}
}

func TestDiscoverServerConfigsFromSameNamespace(t *testing.T) {
	serverConfigs := convertToServerConfigs(
		[]v1alpha1.PrometheusQueryMapping{createIstio(), createJmxTomcat()},
		[]v1alpha1.PrometheusServerConfig{createSingleSvrConfig(), createMultiSvrConfig()})
	spew.Dump(serverConfigs)
	assert.Equal(t, 2, len(serverConfigs))
	queryMappings1 := serverConfigs[0].clusterConfigs[0].queryMappings
	queryMappings2 := serverConfigs[1].clusterConfigs[0].queryMappings
	assert.Equal(t, 2, len(queryMappings1))
	assert.Equal(t, 2, len(queryMappings2))
	assert.Equal(t, queryMappings1[0], queryMappings2[0])
}

func TestDiscoverServerConfigsFromDifferentNamespaces(t *testing.T) {
	istio := createIstio()
	istio.Namespace = "multi"
	jmxTomcat := createJmxTomcat()
	jmxTomcat.Namespace = "single"
	multiCluster := createMultiSvrConfig()
	multiCluster.Namespace = "multi"
	singleCluster := createSingleSvrConfig()
	singleCluster.Namespace = "single"
	serverConfigs := convertToServerConfigs(
		[]v1alpha1.PrometheusQueryMapping{istio, jmxTomcat},
		[]v1alpha1.PrometheusServerConfig{singleCluster, multiCluster})
	spew.Dump(serverConfigs)
	queryMappings1 := serverConfigs[0].clusterConfigs[0].queryMappings
	queryMappings2 := serverConfigs[1].clusterConfigs[0].queryMappings
	assert.Equal(t, 1, len(queryMappings1))
	assert.Equal(t, 1, len(queryMappings2))
}

func TestDiscoverServerConfigsWithFilteredQueryMappings(t *testing.T) {
	istio := createIstio()
	istio.Labels = map[string]string{
		"mapping": "istio",
	}
	jmxTomcat := createJmxTomcat()
	jmxTomcat.Labels = map[string]string{
		"mapping": "jmx-tomcat",
	}
	multiCluster := createMultiSvrConfig()
	singleCluster := createSingleSvrConfig()
	serverConfigs := convertToServerConfigs(
		[]v1alpha1.PrometheusQueryMapping{istio, jmxTomcat},
		[]v1alpha1.PrometheusServerConfig{singleCluster, multiCluster})
	spew.Dump(serverConfigs)
	for _, svrCfg := range serverConfigs {
		for _, clusterCfg := range svrCfg.clusterConfigs {
			if clusterCfg.clusterId == nil {
				assert.Equal(t, 1, len(clusterCfg.queryMappings))
				continue
			}
			if clusterCfg.clusterId.ID == "5f2bd289" {
				assert.Equal(t, 1, len(clusterCfg.queryMappings))
				continue
			}
			assert.Equal(t, 2, len(clusterCfg.queryMappings))
		}
	}
}

func TestDiscoverServerConfigsWithEmptyClusterConfig(t *testing.T) {
	istio := createIstio()
	singleCluster := createSingleSvrConfig()
	singleCluster.Spec.ClusterConfigs = nil
	serverConfigs := convertToServerConfigs(
		[]v1alpha1.PrometheusQueryMapping{istio},
		[]v1alpha1.PrometheusServerConfig{singleCluster})
	spew.Dump(serverConfigs)
	clusterCfg := serverConfigs[0].clusterConfigs[0]
	queryMappings := clusterCfg.queryMappings
	clusterId := clusterCfg.clusterId
	assert.Nil(t, clusterId)
	assert.Equal(t, 1, len(queryMappings))
}
