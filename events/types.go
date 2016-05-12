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
	Metadata    map[string]interface{}
	Fqdn        string
	Vip         string
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
	Environment      map[string]string
	Labels           map[string]string
	Metadata         map[string]interface{}
	PrimaryIpAddress string
	Ports            []string
	ImageUUID        string
	HostID           string
}

type Host struct {
	ID              string
	UUID            string
	Name            string
	Description     string
	State           InstanceState
	AgentState      string
	Labels          map[string]string
	Hostname        string
	PublicEndpoints []Endpoints
}

type Endpoints struct {
	IPAddress string
	Port      int
}
