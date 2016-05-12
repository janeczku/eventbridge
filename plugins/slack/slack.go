// Package slack provides a Slack notification plugin
package slack

import (
	"fmt"
	"time"

	"github.com/janeczku/eventbridge/events"
	"github.com/janeczku/eventbridge/plugins"

	log "github.com/Sirupsen/logrus"
	"github.com/huguesalary/slack-go"
	"github.com/juju/ratelimit"
)

const (
	Version = "0.0.1"
)

var defaultMsgColor = "#CFCDC9"

// throttle to 10 messages/min, burst 10
var rateLimiter = ratelimit.NewBucket(5*time.Second, 10)

var eventKinds = []events.EventKind{
	events.ContainerEvent,
	events.ServiceEvent,
}

// resource states that trigger a Slack notification and the color to use
var states = map[events.InstanceState]string{
	events.ServiceInactive:  "#CFCDC9",
	events.ServiceActive:    "#99CC99",
	events.ContainerStopped: "#CFCDC9",
	events.ContainerRunning: "#99CC99",
}

// message color according to the health state
var healthStates = map[events.HealthState]string{
	events.StateHealthy:   "#99CC99",
	events.StateUnhealthy: "#F2777A",
	events.StateDegraded:  "#F2777A",
}

type Slack struct {
	WebHookURL string
	Channel    string
	Icon       string
	Username   string
}

func NewSlack() *Slack {
	return &Slack{
		Icon:     ":mega:",
		Username: "rancher-eventbridge",
	}
}

func (s *Slack) Init() error {
	if s.WebHookURL == "" {
		return fmt.Errorf("Slack plugin requires the 'webhookurl' configuration parameter")
	}
	return nil
}

func (s *Slack) Process(ev events.Event) error {
	if !filterByState(ev.GetState()) {
		return nil
	}

	if avail := rateLimiter.TakeAvailable(1); avail == 0 {
		log.WithField("plugin", "slack").Warn("Dropping event. Rate limit exceeded.")
		return nil
	}

	msg := &slack.Message{
		Username:  s.Username,
		Channel:   s.Channel,
		IconEmoji: s.Icon,
	}
	attach := msg.NewAttachment()
	attach.Pretext = "Rancher resource change event"
	attach.Text = fmt.Sprintf("%s `%s` @`%s`", string(ev.Kind), ev.GetName(),
		ev.Timestamp.Format("2006-01-02 15:04:05"))
	attach.Fallback = ev.String()
	attach.Color = getMessageColor(ev.GetState(), ev.GetHealthState())
	attach.MarkdownIn = []string{"text", "fields"}
	fields := map[string]string{
		"State":  fmt.Sprintf("`%s`", ev.GetState()),
		"Health": fmt.Sprintf("`%s`", ev.GetHealthState()),
	}

	for k, v := range fields {
		attach.AddField(&slack.Field{
			Title: k,
			Value: v,
			Short: true,
		})
	}

	c := slack.NewClient(s.WebHookURL)
	return c.SendMessage(msg)
}

func filterByState(state events.InstanceState) bool {
	for s, _ := range states {
		if s == state {
			return true
		}
	}
	return false
}

func getMessageColor(state events.InstanceState, health events.HealthState) (color string) {
	if state == events.ServiceInactive || state == events.ContainerStopped {
		if _, ok := states[state]; ok {
			color = states[state]
		}
	}

	if state == events.ServiceActive || state == events.ContainerRunning {
		if _, ok := healthStates[health]; ok {
			color = healthStates[health]
		} else if _, ok := states[state]; ok {
			color = states[state]
		}
	}

	if len(color) == 0 {
		color = defaultMsgColor

	}

	return
}

func (s *Slack) Name() string {
	return "Slack Webhook Plugin"
}

func (s *Slack) Close() error {
	return nil
}

func init() {
	plugins.Register("slack", eventKinds, func() plugins.Plugin {
		return NewSlack()
	})
}
