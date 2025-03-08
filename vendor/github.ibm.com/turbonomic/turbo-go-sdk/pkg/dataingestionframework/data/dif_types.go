package data

import (
	"fmt"

	set "github.com/deckarep/golang-set"
)

// DIFEntity defines data ingestion framework topology entity
type DIFEntity struct {
	UID                 string                     `json:"uniqueId"`
	Type                string                     `json:"type"`
	Name                string                     `json:"name"`
	HostedOn            *DIFHostedOn               `json:"hostedOn"`
	MatchingIdentifiers *DIFMatchingIdentifiers    `json:"matchIdentifiers"`
	PartOf              []*DIFPartOf               `json:"partOf"`
	Metrics             map[string][]*DIFMetricVal `json:"metrics"`
	Controllable        bool                       `json:"controllable"`
	Cloneable           bool                       `json:"cloneable"`
	Suspendable         bool                       `json:"suspendable"`
	ProviderMustClone   bool                       `json:"providerMustClone"`
	clusterId           string
	namespace           string
	attributes          map[string]string
	partOfSet           set.Set
	hostTypeSet         set.Set
}

type DIFMatchingIdentifiers struct {
	IPAddress                    string `json:"ipAddress"`
	KubernetesFullyQualifiedName string `json:"kubernetesFullyQualifiedName"`
}

type DIFHostedOn struct {
	HostType  []DIFHostType `json:"hostType"`
	IPAddress string        `json:"ipAddress"`
	HostUuid  string        `json:"hostUuid"`
}

type DIFPartOf struct {
	ParentEntity string `json:"entity"`
	UniqueId     string `json:"uniqueId"`
	Label        string `json:"label,omitempty"`
}

func NewDIFEntity(uid, eType string) *DIFEntity {
	return &DIFEntity{
		UID:         uid,
		Type:        eType,
		Name:        uid,
		partOfSet:   set.NewSet(),
		hostTypeSet: set.NewSet(),
		Metrics:     make(map[string][]*DIFMetricVal),
	}
}

func (e *DIFEntity) WithName(name string) *DIFEntity {
	e.Name = name
	return e
}

func (e *DIFEntity) WithControllable(controllable bool) *DIFEntity {
	e.Controllable = controllable
	return e
}

func (e *DIFEntity) WithCloneable(cloneable bool) *DIFEntity {
	e.Cloneable = cloneable
	return e
}

func (e *DIFEntity) WithSuspendable(suspendable bool) *DIFEntity {
	e.Suspendable = suspendable
	return e
}

func (e *DIFEntity) WithProviderMustClone(providerMustClone bool) *DIFEntity {
	e.ProviderMustClone = providerMustClone
	return e
}

func (e *DIFEntity) WithClusterId(clusterId string) *DIFEntity {
	e.clusterId = clusterId
	return e
}

func (e *DIFEntity) GetClusterId() string {
	return e.clusterId
}

func (e *DIFEntity) WithNamespace(namespace string) *DIFEntity {
	e.namespace = namespace
	return e
}

func (e *DIFEntity) GetNamespace() string {
	return e.namespace
}

func (e *DIFEntity) WithAttributes(attrs map[string]string) *DIFEntity {
	e.attributes = attrs
	return e
}

func (e *DIFEntity) GetAttribute(key string) string {
	return e.attributes[key]
}

func (e *DIFEntity) PartOfEntity(entity, id, label string) *DIFEntity {
	if e.partOfSet.Contains(id) {
		return e
	}
	e.partOfSet.Add(id)
	e.PartOf = append(e.PartOf, &DIFPartOf{entity, id, label})
	return e
}

func (e *DIFEntity) HostedOnType(hostType DIFHostType) *DIFEntity {
	if e.hostTypeSet.Contains(hostType) {
		return e
	}
	if e.HostedOn == nil {
		e.HostedOn = &DIFHostedOn{}
	}
	e.HostedOn.HostType = append(e.HostedOn.HostType, hostType)
	e.hostTypeSet.Add(hostType)
	return e
}

func (e *DIFEntity) GetHostedOnType() []DIFHostType {
	var hostTypes []DIFHostType
	for _, hostType := range e.hostTypeSet.ToSlice() {
		hostTypes = append(hostTypes, hostType.(DIFHostType))
	}
	return hostTypes
}

func (e *DIFEntity) HostedOnIP(ip string) *DIFEntity {
	if e.HostedOn == nil {
		e.HostedOn = &DIFHostedOn{}
	}
	e.HostedOn.IPAddress = ip
	return e
}

func (e *DIFEntity) HostedOnUID(uid string) *DIFEntity {
	if e.HostedOn == nil {
		e.HostedOn = &DIFHostedOn{}
	}
	e.HostedOn.HostUuid = uid
	return e
}

// Matching creates or overwrites an existing IPAddress matcher.
// This remains for backward compatability. Since then, multiple potential matchers have
// been added. Instead, you should modify DIFMatchingIdentifiers directly.
// Deprecated
func (e *DIFEntity) Matching(id string) *DIFEntity {
	if e.MatchingIdentifiers == nil {
		e.MatchingIdentifiers = &DIFMatchingIdentifiers{
			IPAddress: id,
		}
		return e
	}
	// Overwrite
	e.MatchingIdentifiers.IPAddress = id
	return e
}

// AddMetric add a metric with certain type, kind, value and key to the DIF entity.
//
//	 This function makes it easier to add a metric of the same type (e.g., memory) but
//	 different kind (e.g., average, or capacity) to a DIF entity, because they can be
//	 discovered at different times.
//	 The DIFEntity.Metrics is a map where the key is the metric type, and the value is
//	 a list of DIFMetricVal. We need a list of DIFMetricVal to hold metrics with the same
//	 type but different keys, for example:
//		kpi: [
//			{
//				average: 123,
//				capacity: 1000,
//				key: "total_messages_in_queue"
//			},
//			{
//				average: 104.44444444444444,
//				capacity: 1000,
//				key: "total_waiting_time_in_queue"
//			}
//		],
func (e *DIFEntity) AddMetric(metricType string, kind DIFMetricValKind, value float64, key string) {
	var metricVal *DIFMetricVal
	var metricKey *string
	// Only set non-empty key
	if key != "" {
		metricKey = &key
	}
	meList, found := e.Metrics[metricType]
	if !found || len(meList) < 1 {
		// This is a new metric type, or the metric list for this type is empty
		metricVal = &DIFMetricVal{Key: metricKey}
		e.Metrics[metricType] = []*DIFMetricVal{metricVal}
	} else {
		// The metric type already exists with non-empty metric list
		for _, me := range meList {
			if sameKey(me.Key, metricKey) {
				// We found a metric with the same key (including nil key).
				// The existing metricVal.Average or metricVal.Capacity
				// will be overwritten.
				metricVal = me
				break
			}
		}
		if metricVal == nil {
			// This is a metric of the same type, but a new key.
			// Create a new metricVal and append it to the metric list.
			metricVal = &DIFMetricVal{Key: metricKey}
			e.Metrics[metricType] = append(e.Metrics[metricType], metricVal)
		}
	}
	if kind == AVERAGE {
		metricVal.Average = &value
	} else if kind == CAPACITY {
		metricVal.Capacity = &value
	}
}

func sameKey(key1 *string, key2 *string) bool {
	if key1 == key2 {
		return true
	}
	if key1 == nil || key2 == nil {
		return false
	}
	if *key1 == *key2 {
		return true
	}
	return false
}

func (e *DIFEntity) AddMetrics(metricType string, metricVals []*DIFMetricVal) {
	e.Metrics[metricType] = append(e.Metrics[metricType], metricVals...)
}

// PairwiseAggregate aggregates, in a pairwise-sum fashion, the given metric values to this entity by the given type.
// Ideally, we should compare the two timestamps and only add those in the same time neighborhood.
// For simplicity, in the following, we just assume most likely the two arrays with the same commodity key are of the
// same length and their timestamps are the same. If the given metric val array is longer than in this entity, then a
// copy of its longer portion will simply be appended.
func (e *DIFEntity) PairwiseAggregate(metricType string, metricValsToAggregate []*DIFMetricVal) {
	len2 := len(metricValsToAggregate)
	if len2 == 0 {
		// nothing to aggregate
		return
	}
	// all new values have the same key
	commodityKey := metricValsToAggregate[0].Key

	metricVals := e.Metrics[metricType]
	len1 := len(metricVals)

	i2 := 0
	for i1 := 0; i1 < len1 && i2 < len2; i1++ {
		if sameKey(metricVals[i1].Key, commodityKey) {
			metricVals[i1].Sum(metricValsToAggregate[i2])
			i2++
		}
	}
	if i2 < len2 {
		for _, val := range metricValsToAggregate[i2:] {
			metricVals = append(metricVals, val.Clone())
		}
		e.Metrics[metricType] = metricVals
	}
}

// PairwiseAggregateAll aggregates the given set of metrics, type by type, slot by slot, into this entity.
func (e *DIFEntity) PairwiseAggregateAll(metrics map[string][]*DIFMetricVal) {
	for t, m := range metrics {
		e.PairwiseAggregate(t, m)
	}
}

func (e *DIFEntity) String() string {
	s := fmt.Sprintf("%s[%s:%s]", e.Type, e.UID, e.Name)
	if e.MatchingIdentifiers != nil {
		s += fmt.Sprintf(" IP[%s]", e.MatchingIdentifiers.IPAddress)
		s += fmt.Sprintf(" KubernetesFullyQualifiedName[%s]", e.MatchingIdentifiers.KubernetesFullyQualifiedName)
	}
	if e.PartOf != nil {
		s += fmt.Sprintf(" PartOf")
		for _, partOf := range e.PartOf {
			s += fmt.Sprintf("[%s:%s]", partOf.ParentEntity, partOf.UniqueId)
		}
	}
	if e.HostedOn != nil {
		s += fmt.Sprintf(" HostedOn")
		s += fmt.Sprintf("[%s:%s]",
			e.HostedOn.HostUuid, e.HostedOn.IPAddress)
	}
	for metricName, metricList := range e.Metrics {
		for _, metric := range metricList {
			s += fmt.Sprintf(" Metric %s:[%v]", metricName, metric)
		}
	}
	return s
}

func ClearCommodityKeys(metrics map[string][]*DIFMetricVal) map[string][]*DIFMetricVal {
	for _, metricData := range metrics {
		for _, metricDataValue := range metricData {
			metricDataValue.Key = nil
		}
	}
	return metrics
}
