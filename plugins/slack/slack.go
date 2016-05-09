// Package slack provides a Slack notification plugin
package slack

import (
	"fmt"

	"github.com/janeczku/eventbridge/events"
	"github.com/janeczku/eventbridge/plugins"

	"github.com/huguesalary/slack-go"
)

const (
	Version = "0.0.1"
)

var acceptedEventKinds = []events.EventKind{
	events.ContainerEvent,
	events.ServiceEvent,
	events.StackEvent,
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
	msg := &slack.Message{
		Username:  s.Username,
		Channel:   s.Channel,
		IconEmoji: s.Icon,
	}
	attach := msg.NewAttachment()
	attach.Pretext = fmt.Sprintf("[%s] Resource change event", ev.Timestamp.Format("2006-01-02 15:04:05"))
	attach.Fallback = ev.String()
	attach.Color = "#8BC7FF"
	fields := map[string]string{
		"kind":        string(ev.Kind),
		"name":        ev.Name(),
		"state":       string(ev.State()),
		"healthstate": string(ev.HealthState()),
	}

	for k, v := range fields {
		attach.AddField(&slack.Field{
			Title: k,
			Value: v,
			Short: false,
		})
	}

	c := slack.NewClient(s.WebHookURL)
	return c.SendMessage(msg)
}

func (s *Slack) Name() string {
	return "Slack Webhook Plugin"
}

func (s *Slack) Close() error {
	return nil
}

func init() {
	plugins.Register("slack", acceptedEventKinds, func() plugins.Plugin {
		return NewSlack()
	})
}
