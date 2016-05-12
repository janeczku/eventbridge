package events

import (
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

// Event is used to store information relating to a Rancher API "resource.change" event
type Event struct {
	ID            string
	Timestamp     time.Time
	Kind          EventKind
	ContainerData Container
	HostData      Host
	ServiceData   Service
	StackData     Stack
}

func New(id string, kind EventKind, resourceData map[string]interface{}) (Event, error) {
	ev := Event{
		ID:        id,
		Timestamp: time.Now().UTC(),
		Kind:      kind,
	}

	var err error
	switch kind {
	case ContainerEvent:
		err = mapstructure.WeakDecode(resourceData, &ev.ContainerData)
		parseStackServiceNames(&ev.ContainerData)
	case HostEvent:
		err = mapstructure.WeakDecode(resourceData, &ev.HostData)
	case ServiceEvent:
		err = mapstructure.WeakDecode(resourceData, &ev.ServiceData)
	case StackEvent:
		err = mapstructure.WeakDecode(resourceData, &ev.StackData)
	default:
		return Event{}, fmt.Errorf("Unknown event kind: %s", kind)
	}

	if err != nil {
		return Event{}, fmt.Errorf("Failed to decode resource data: %v", err)
	}

	return ev, nil
}

func (ev Event) GetState() InstanceState {
	var state InstanceState
	switch ev.Kind {
	case ContainerEvent:
		state = ev.ContainerData.State
	case HostEvent:
		state = ev.HostData.State
	case ServiceEvent:
		state = ev.ServiceData.State
	case StackEvent:
		state = ev.StackData.State
	}
	return state
}

func (ev Event) GetHealthState() HealthState {
	var healthState HealthState
	switch ev.Kind {
	case ContainerEvent:
		healthState = ev.ContainerData.HealthState
	case ServiceEvent:
		healthState = ev.ServiceData.HealthState
	case StackEvent:
		healthState = ev.StackData.HealthState
	}

	if len(healthState) == 0 {
		healthState = StateUnknown
	}

	return healthState
}

func (ev Event) GetName() string {
	var name string
	switch ev.Kind {
	case ContainerEvent:
		name = ev.ContainerData.Name
	case HostEvent:
		name = ev.HostData.Name
	case ServiceEvent:
		name = ev.ServiceData.Name
	case StackEvent:
		name = ev.StackData.Name
	}
	return name
}

func (ev Event) String() string {
	return fmt.Sprintf("[%s] %s '%s' is now in the '%s' state (health: '%s')",
		ev.Timestamp.Format("2006-01-02 15:04:05"), ev.Kind, ev.GetName(), ev.GetState(), ev.GetHealthState())
}

func parseStackServiceNames(container *Container) {
	parts := strings.SplitN(container.Name, "_", 3)
	if len(parts) != 3 {
		return
	}
	container.StackName = parts[0]
	container.ServiceName = parts[1]
	container.Name = parts[1] + "_" + parts[2]
}
