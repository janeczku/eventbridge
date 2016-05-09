package events

const (
	ContainerEvent EventKind = "container"
	HostEvent      EventKind = "host"
	ServiceEvent   EventKind = "service"
	StackEvent     EventKind = "stack"
)

const (
	// Common health states
	StateHealthy           HealthState = "healthy"
	StateUnhealthy         HealthState = "unhealthy"
	StateUpdatingHealthy   HealthState = "updating-healthy"
	StateUpdatingUnhealthy HealthState = "updating-unhealthy"
	StateReconcile         HealthState = "reconcile"
	StateInitializing      HealthState = "initializing"
	StateReinitializing    HealthState = "reinitializing"

	// Service specific health states
	StateDegraded    HealthState = "degraded"
	StateStartedOnce HealthState = "started-once"
)

const (
	// Service states
	ServiceInactive         InstanceState = "inactive"
	ServiceActivating       InstanceState = "activating"
	ServiceActive           InstanceState = "active"
	ServiceUpdatingActive   InstanceState = "updating-active"
	ServiceUpdatingInactive InstanceState = "updating-inactive"
	ServiceUpgraded

	// Container states
	ContainerStopping   InstanceState = "stopping"
	ContainerStopped    InstanceState = "stopped"
	ContainerStarting   InstanceState = "starting"
	ContainerRunning    InstanceState = "running"
	ContainerRestarting InstanceState = "restarting"

	// Host states
	HostInactive   InstanceState = "inactive"
	HostActivating InstanceState = "activating"
	HostActive     InstanceState = "active"
)
