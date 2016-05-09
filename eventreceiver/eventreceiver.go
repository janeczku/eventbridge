package eventreceiver

import (
	"fmt"

	"github.com/janeczku/eventbridge/config"
	"github.com/janeczku/eventbridge/events"

	log "github.com/Sirupsen/logrus"
	revents "github.com/rancher/go-machine-service/events"
	"github.com/rancher/go-rancher/client"
)

var eventKindMapping = map[string]events.EventKind{
	"container":   events.ContainerEvent,
	"environment": events.StackEvent,
	"host":        events.HostEvent,
	"service":     events.ServiceEvent,
}

type EventReceiver struct {
	output      chan events.Event
	config      *config.AgentConfig
	eventKinds  map[events.EventKind]bool
	eventRouter *revents.EventRouter
}

func New(config *config.AgentConfig, eventKinds map[events.EventKind]bool, output chan events.Event) *EventReceiver {
	return &EventReceiver{
		output:     output,
		config:     config,
		eventKinds: eventKinds,
	}
}

func (r *EventReceiver) Start() error {
	log.WithField("rancherURL", r.config.RancherURL).Debug("Starting event receiver")
	eventHandlers := map[string]revents.EventHandler{
		"resource.change": r.EventHandler,
		"ping":            r.PingNoOp,
	}

	router, err := revents.NewEventRouter("", 0, r.config.RancherURL, r.config.RancherAccessKey,
		r.config.RancherSecretKey, nil, eventHandlers, "", r.config.EventReceiverCount)
	if err != nil {
		return fmt.Errorf("Could not connect to event stream: %v", err)
	}

	r.eventRouter = router

	readyChan := make(chan bool)
	errorChan := make(chan error)

	go func(c chan error) {
		err = r.eventRouter.StartWithoutCreate(readyChan)
		c <- err
	}(errorChan)

	select {
	case <-readyChan:
		break
	case err := <-errorChan:
		return fmt.Errorf("Event stream listener exited: %v", err)
	}

	return nil
}

func (r *EventReceiver) Stop() error {
	log.Debug("Stopping event receiver")
	return r.eventRouter.Stop()
}

func (r *EventReceiver) EventHandler(ev *revents.Event, cli *client.RancherClient) error {
	log.WithFields(log.Fields{
		"name":       ev.Name,
		"eventId":    ev.ID,
		"resourceId": ev.ResourceID,
		"EventKind":  ev.ResourceType,
	}).Debug("Received event")

	var kind events.EventKind
	if val, ok := eventKindMapping[ev.ResourceType]; ok {
		kind = val
	}

	_, wanted := r.eventKinds[kind]
	if len(kind) == 0 || !wanted {
		return nil
	}

	return r.transformEvent(ev, kind)
}

func (r *EventReceiver) PingNoOp(ev *revents.Event, cli *client.RancherClient) error {
	return nil
}

func (r *EventReceiver) transformEvent(ev *revents.Event, kind events.EventKind) error {
	resourceData := ev.Data["resource"].(map[string]interface{})
	newEvent, err := events.New(ev.ID, kind, resourceData)
	if err != nil {
		log.WithFields(log.Fields{
			"eventId":      ev.ID,
			"kind":         kind,
			"resourceData": resourceData,
			"error":        err,
		}).Error("Failed to transform event")
		return nil
	}

	log.WithFields(log.Fields{
		"eventID": newEvent.ID,
		"kind":    newEvent.Kind,
	}).Debug("Transformed event")

	r.output <- newEvent

	return nil
}
