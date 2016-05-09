package agent

import (
	"sync"

	"github.com/janeczku/eventbridge/config"
	"github.com/janeczku/eventbridge/eventreceiver"
	"github.com/janeczku/eventbridge/events"

	log "github.com/Sirupsen/logrus"
)

// Agent supervises receiver and plugin components.
type Agent struct {
	Config    *config.Config
	receiver  *eventreceiver.EventReceiver
	quitChan  chan struct{}
	waitGroup *sync.WaitGroup
}

// New returns an Agent struct based off the given config
func New(config *config.Config) (*Agent, error) {
	agent := &Agent{
		Config:    config,
		quitChan:  make(chan struct{}),
		waitGroup: &sync.WaitGroup{},
	}

	return agent, nil
}

// Start starts the agent, receiver and configured plugins.
func (a *Agent) Start() error {
	log.Info("Starting agent")

	receiveChan := make(chan events.Event)
	a.receiver = eventreceiver.New(a.Config.Agent, a.Config.EventKinds, receiveChan)
	if err := a.receiver.Start(); err != nil {
		return err
	}

	if err := a.startPlugins(); err != nil {
		return err
	}

	a.waitGroup.Add(1)
	go a.doWork(receiveChan)

	return nil
}

// Shutdown shutdowns event receiver and plugins and stops the agent.
func (a *Agent) Shutdown() {
	close(a.quitChan)
	a.waitGroup.Wait()

	if err := a.receiver.Stop(); err != nil {
		log.WithField("error", err).Error("Error stopping event receiver")
	}

	if err := a.stopPlugins(); err != nil {
		log.WithField("error", err).Error("Error stopping plugin runners")
	}
}

func (a *Agent) startPlugins() error {
	for _, p := range a.Config.Plugins {
		if err := p.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) stopPlugins() error {
	var err error
	for _, p := range a.Config.Plugins {
		err = p.Stop()
	}
	return err
}

func (a *Agent) doWork(input chan events.Event) {
	defer a.waitGroup.Done()
	for {
		select {
		case <-a.quitChan:
			return
		case ev := <-input:
			for _, p := range a.Config.Plugins {
				if _, ok := p.EventKinds[ev.Kind]; ok {
					p.Write(ev)
				}
			}
		}
	}
}
