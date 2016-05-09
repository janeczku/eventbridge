package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/janeczku/eventbridge/agent"
	"github.com/janeczku/eventbridge/config"
	"github.com/janeczku/eventbridge/healthcheck"
	"github.com/janeczku/eventbridge/plugins"
)

var (
	Version   = "dev"
	GitCommit = "HEAD"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	app := cli.NewApp()
	app.Name = "Rancher Eventbridge"
	app.Usage = "Plugin-driven Rancher event processor"
	app.Version = fmt.Sprintf("%s (%s)", Version, GitCommit)
	app.Action = runApp
	app.Commands = []cli.Command{
		{
			Name:  "plugins",
			Usage: "list the available plugins",
			Action: func(c *cli.Context) error {
				fmt.Println("Available Plugins:")
				for _, name := range plugins.List() {
					fmt.Printf("- %s\n", name)
				}
				return nil
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "configuration file to load",
		},
		cli.StringFlag{
			Name:  "loglevel",
			Value: "info",
			Usage: "set the default loglevel (debug|info|warn|error)",
		},
	}

	app.Run(os.Args)
}

func runApp(c *cli.Context) error {
	if len(c.String("config")) == 0 {
		log.Fatalln("'--config' flag is required")
	}

	conf := config.New()
	err := conf.LoadConfig(c.String("config"))
	if err != nil {
		log.Fatal(err)
	}

	if len(conf.Plugins) == 0 {
		log.Fatal("Error: No plugins configured")
	}

	if c.IsSet("loglevel") {
		conf.Agent.LogLevel = c.String("loglevel")
	}

	log.Infof("Starting Eventbridge version %s (%s)", Version, GitCommit)
	log.Infof("Plugins active: %s", strings.Join(conf.PluginNames(), " | "))

	if level, err := log.ParseLevel(conf.Agent.LogLevel); err == nil {
		log.WithField("logLevel", conf.Agent.LogLevel).Info("Setting log level")
		log.SetLevel(level)
	}

	a, err := agent.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	err = a.Start()
	if err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	errorChan := make(chan error)
	go func(c chan error) {
		err := healthcheck.StartHealthCheck(conf.Agent.HealthCheckPort)
		c <- err
	}(errorChan)

	select {
	case e := <-errorChan:
		log.Errorf("Healthcheck exited with error: %v", e)
		break
	case s := <-signalChan:
		log.Infof("Application exit requested by signal: %s", s.String())
		break
	}

	a.Shutdown()
	return nil
}
