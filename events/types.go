package events

type EventKind string
type HealthState string
type InstanceState string

type Stack struct {
	ID          string
	UUID        string
	Name        string
	Description string
	State       InstanceState
	HealthState HealthState
	Removed     bool
}

type Service struct {
	ID          string
	UUID        string
	Version     string
	Name        string
	Description string
	Scale       int
	State       InstanceState
	HealthState HealthState
	Removed     bool
	Metadata    map[string]interface{}
	Fqdn        string
	Vip         string
	CreateIndex int
}

type Container struct {
	ID               string
	UUID             string
	Version          string
	Name             string
	Description      string
	ServiceName      string
	StackName        string
	State            InstanceState
	HealthState      HealthState
	Removed          bool
	Environment      map[string]string
	Labels           map[string]string
	Metadata         map[string]interface{}
	PrimaryIpAddress string
	Ports            []string
	ImageUUID        string
	HostID           string
	CreateIndex      int
}

type Host struct {
	ID              string
	UUID            string
	Name            string
	Description     string
	State           InstanceState
	AgentState      string
	Removed         bool
	Labels          map[string]string
	Hostname        string
	PublicEndpoints []Endpoints
}

type Endpoints struct {
	IPAddress string
	Port      int
}
