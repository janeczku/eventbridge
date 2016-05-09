# Slack Webhook Plugin

This plugin sends notifications about events to Slack using an [Incoming Webhook](https://api.slack.com/incoming-webhooks).

## Configuration

```Toml
[slack]
  # Incoming Webhook URL (required)
  webhookurl = "https://hooks.slack.com/services/<REPLACE WITH TOKEN>"
  # Slack channel (optional)
  channel = "#alerts"
  # Icon Emoji (optional)
  icon = ":cow:"
  # Username (optional)
  username = "rancher"
```
