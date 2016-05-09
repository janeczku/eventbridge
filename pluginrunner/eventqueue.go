package pluginrunner

import (
	"github.com/janeczku/eventbridge/events"
)

const (
	// Default size of the event queue.
	DEFAULT_QUEUE_SIZE = 50
)

// EventQueue is used for queuing events in a ring buffer.
// If the buffer is full, then the oldest events are dropped
// first.
type EventQueue struct {
	Buffer chan events.Event
	drops  int
}

// NewEventBuffer returns a new EventBuffer with the given capacity.
func NewEventQueue(size int) *EventQueue {
	return &EventQueue{
		Buffer: make(chan events.Event, size),
	}
}

// Size returns the current size of the queue.
func (eq *EventQueue) Size() int {
	return len(eq.Buffer)
}

// Drops returns the total number of dropped events.
func (eq *EventQueue) Drops() int {
	return eq.drops
}

// Add adds an event to the queue.
func (eq *EventQueue) Add(event events.Event) {
	select {
	case eq.Buffer <- event:
	default:
		eq.drops++
		<-eq.Buffer
		eq.Buffer <- event
	}
}
