package api

type ResourceType string

const (
	Resource_Type_Reservation     ResourceType = "reservations"
	Resource_Type_Target          ResourceType = "targets"
	Resource_Type_External_Target ResourceType = "externaltargets"
)
