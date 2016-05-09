package plugins

import (
	"github.com/janeczku/eventbridge/events"
)

// Plugin accepts and processes events.
type Plugin interface {
	// Init is called once on application start
	Init() error
	// Close is called once on application exit
	Close() error
	// Name returns the human-friendly name of the plugin
	Name() string
	// Process accepts an event for processing
	Process(ev events.Event) error
}
