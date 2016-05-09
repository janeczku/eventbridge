# Eventbridge

Eventbridge is a plugin-driven event processor for [Rancher](http://github.com/rancher/rancher).
It connects to a Rancher server's event stream and listens for events related to resource changes (stacks/services/containers/hosts).
Events are then passed to the configured plugins for processing/forwarding to a third-party.

`Work in progress`

## Supported Plugins

Currently implemented:

* [slack](https://github.com/janeczku/eventbridge/tree/master/plugins/slack)
