package pluginrunner

import (
	"sync"

	"github.com/janeczku/eventbridge/events"
	"github.com/janeczku/eventbridge/plugins"

	log "github.com/Sirupsen/logrus"
)

// PluginMetrics tracks various metrics for the plugin runner.
type PluginMetrics struct {
	sync.Mutex     // not used ATM
	Pending    int // events currently queued
	Dropped    int // total events dropped from queue
	Totals     int // total events received
	Successes  int // total succesfull event writes
	Errors     int // total errored event writes
}

// PluginRunner wraps a single plugin and queues events in a buffer.
type PluginRunner struct {
	Name        string
	Plugin      plugins.Plugin
	EventKinds  map[events.EventKind]bool
	WorkerCount int
	Metrics     *PluginMetrics

	eventQueue *EventQueue
	quitChan   chan struct{}
	waitGroup  *sync.WaitGroup
}

func New(name string, plugin plugins.Plugin, queueLimit int, kinds map[events.EventKind]bool) *PluginRunner {
	if queueLimit == 0 {
		queueLimit = DEFAULT_QUEUE_SIZE
	}
	r := &PluginRunner{
		Name:        name,
		Plugin:      plugin,
		EventKinds:  kinds,
		WorkerCount: 1,
		Metrics:     new(PluginMetrics),
		eventQueue:  NewEventQueue(queueLimit),
		waitGroup:   &sync.WaitGroup{},
		quitChan:    make(chan struct{}),
	}
	return r
}

// Write adds an event to the event queue.
func (r *PluginRunner) Write(ev events.Event) {
	log.WithFields(log.Fields{
		"eventId": ev.ID,
		"plugin":  r.Name,
	}).Debug("Adding event to queue")
	r.Metrics.Totals++
	r.eventQueue.Add(ev)
}

// Start invokes the plugin's Init method and dispatches event queue routine.
func (r *PluginRunner) Start() error {
	log.WithField("plugin", r.Name).Info("Initializing plugin")
	if err := r.Plugin.Init(); err != nil {
		return err
	}

	r.waitGroup.Add(r.WorkerCount)
	for i := 0; i < r.WorkerCount; i++ {
		go r.doWork()
	}

	return nil
}

// Stop stops the event queue routine and invokes the plugin's Close method.
func (r *PluginRunner) Stop() error {
	close(r.quitChan)
	r.waitGroup.Wait()
	log.WithField("plugin", r.Name).Info("Closing plugin")
	if err := r.Plugin.Close(); err != nil {
		return err
	}

	return nil
}

// Stats returns a populated metrics object.
func (r *PluginRunner) Stats() *PluginMetrics {
	r.Metrics.Pending = r.eventQueue.Size()
	r.Metrics.Dropped = r.eventQueue.Drops()
	return r.Metrics
}

func (r *PluginRunner) doWork() {
	log.WithField("plugin", r.Name).Debug("Plugin worker started")
	defer r.waitGroup.Done()
	for {
		select {
		case <-r.quitChan:
			return
		case ev := <-r.eventQueue.Buffer:
			log.WithFields(log.Fields{
				"eventId": ev.ID,
				"plugin":  r.Name,
			}).Debug("Writing event to plugin")
			if err := r.Plugin.Process(ev); err != nil {
				r.Metrics.Errors++
				log.WithFields(log.Fields{
					"error":  err,
					"plugin": r.Name,
				}).Error("Error writing event to plugin")
			} else {
				r.Metrics.Successes++
			}
		}
	}
}
