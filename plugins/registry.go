package plugins

import (
	log "github.com/Sirupsen/logrus"
	"github.com/janeczku/eventbridge/events"
)

type PluginFactory func() Plugin

var RegisteredPlugins = make(map[string]PluginFactory)
var PluginEventKinds = make(map[string][]events.EventKind)

func Register(name string, kinds []events.EventKind, f PluginFactory) {
	if f == nil {
		log.WithField("pluginName", name).Fatal("PluginFactory func is nil")
	}

	if _, ok := RegisteredPlugins[name]; ok {
		log.WithField("pluginName", name).Fatal("Plugin already registered")
	}

	RegisteredPlugins[name] = f
	PluginEventKinds[name] = kinds
}

func List() []string {
	plugins := make([]string, 0, len(RegisteredPlugins))
	for name, _ := range RegisteredPlugins {
		plugins = append(plugins, name)
	}
	return plugins
}
