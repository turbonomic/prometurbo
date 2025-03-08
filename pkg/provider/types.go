package provider

import (
	"regexp"

	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

const (
	Used     = "used"
	Capacity = "capacity"
)

var MetricKindToDIFMetricValKind = map[string]data.DIFMetricValKind{
	Used:     data.AVERAGE,
	Capacity: data.CAPACITY,
}

type MetricDef struct {
	MType   string
	Queries map[string]string
}

type AttributeValueDef struct {
	LabelKeys    []string // possibly multiple labels are used to construct a single value
	LabelDelim   string   // delimeter inserted between the label values to form the attribute value
	ValueMatches *regexp.Regexp
	ValueAs      string
	IsIdentifier bool
}

type EntityDef struct {
	EType         string
	HostedOnVM    bool
	AttributeDefs map[string]*AttributeValueDef
	MetricDefs    []*MetricDef
}

type EntityAttribute struct {
	ID        string
	IP        string
	Service   string
	Namespace string
	AsMap     map[string]string // all attributes extracted
}
