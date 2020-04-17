package data

import "time"

//Data injection framework topology
type Topology struct {
	Version    string       `json:"version"`
	Updatetime int64        `json:"updateTime"`
	Scope      string       `json:"scope"`
	Source     string       `json:"source"`
	Entities   []*DIFEntity `json:"topology"`
}

func NewTopology() *Topology {
	return &Topology{
		Version:    "v1",
		Updatetime: 0,
		Entities:   []*DIFEntity{},
		Scope:      "",
	}
}

func (r *Topology) SetUpdateTime() *Topology {
	t := time.Now()
	r.Updatetime = t.Unix()
	return r
}

func (r *Topology) AddEntity(entity *DIFEntity) {
	r.Entities = append(r.Entities, entity)
}

func (r *Topology) AddEntities(entities []*DIFEntity) {
	r.Entities = append(r.Entities, entities...)
}
