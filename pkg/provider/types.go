package provider

import (
	"regexp"

	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

const (
	Used     = "used"
	Capacity = "capacity"
	Scope    = "scope"
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
	LabelKey     string
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
}
