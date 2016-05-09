package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/janeczku/eventbridge/events"
	"github.com/janeczku/eventbridge/pluginrunner"
	"github.com/janeczku/eventbridge/plugins"

	log "github.com/Sirupsen/logrus"
	"github.com/bbangert/toml"
)

type Config struct {
	Agent      *AgentConfig
	Plugins    []*pluginrunner.PluginRunner
	EventKinds map[events.EventKind]bool
}

type AgentConfig struct {
	RancherAccessKey   string `toml:"rancher_access_key"`
	RancherSecretKey   string `toml:"rancher_secret_key"`
	RancherURL         string `toml:"rancher_url"`
	EventReceiverCount int    `toml:"event_receiver_count"`
	EventQueueLimit    int    `toml:"event_queue_limit"`
	HealthCheckPort    int    `toml:"health_check_port"`
	LogLevel           string `toml:"loglevel"`
}

// New initializes a new config object with defaults.
func New() *Config {
	c := &Config{
		Agent: &AgentConfig{
			EventReceiverCount: 5,
			EventQueueLimit:    50,
			HealthCheckPort:    10240,
			LogLevel:           "info",
		},
		Plugins:    make([]*pluginrunner.PluginRunner, 0),
		EventKinds: make(map[events.EventKind]bool),
	}
	return c
}

// LoadConfig loads the agent and plugin configs from the given file.
func (c *Config) LoadConfig(configPath string) error {
	var configFile map[string]toml.Primitive
	contents, err := replaceEnvsFile(configPath)
	if err != nil {
		return fmt.Errorf("Error loading config file: %v", err)
	}
	if _, err = toml.Decode(contents, &configFile); err != nil {
		return fmt.Errorf("Error parsing config file: %v", err)
	}

	// Agent config
	agentConfig, ok := configFile["agent"]
	if !ok {
		return fmt.Errorf("%s: missing [agent] config", configPath)
	}

	ignoreFields := map[string]interface{}{}
	err = toml.PrimitiveDecodeStrict(agentConfig, c.Agent, ignoreFields)
	if err != nil {
		return fmt.Errorf("Error parsing [agent] config: %v", err)
	}

	delete(configFile, "agent")

	// Plugin configs
	for pluginName, pluginConf := range configFile {
		if err = c.addPlugin(pluginName, pluginConf); err != nil {
			return fmt.Errorf("Error parsing [%s] config: %v", pluginName, err)
		}
	}

	for _, kinds := range plugins.PluginEventKinds {
		for _, kind := range kinds {
			if _, ok := c.EventKinds[kind]; !ok {
				c.EventKinds[kind] = true
			}
		}
	}

	return nil
}

// PluginNames returns a list of human-friendly names of all configured plugins.
func (c *Config) PluginNames() []string {
	var names []string
	for _, p := range c.Plugins {
		names = append(names, p.Plugin.Name())
	}
	return names
}

func replaceEnvsFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	contents := os.ExpandEnv(string(bytes))
	return contents, nil
}

func (c *Config) addPlugin(name string, config toml.Primitive) error {
	factory, ok := plugins.RegisteredPlugins[name]
	if !ok {
		return fmt.Errorf("Unknown plugin '%s'", name)
	}

	plugin := factory()

	if err := toml.PrimitiveDecode(config, plugin); err != nil {
		return fmt.Errorf("Could not parse config for plugin '%s': %v", name, err)
	}

	eventKinds, ok := plugins.PluginEventKinds[name]
	if !ok {
		return fmt.Errorf("No event kinds defined for plugin '%s'", name)
	}

	m := make(map[events.EventKind]bool)
	for _, kind := range eventKinds {
		m[kind] = true
	}

	runner := pluginrunner.New(name, plugin, c.Agent.EventQueueLimit, m)
	log.WithField("pluginName", name).Debug("Added plugin runner")
	c.Plugins = append(c.Plugins, runner)

	return nil
}
