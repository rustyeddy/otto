package messenger

import "path"

// TopicScheme builds MQTT topic paths for devices.
type TopicScheme struct {
	Prefix string // e.g. "otto" or "home"
}

func (s TopicScheme) base(name string) string { return path.Join(s.Prefix, "devices", name) }

// State returns the MQTT topic for a device's state.
func (s TopicScheme) State(name string) string { return path.Join(s.base(name), "state") }

// Set returns the MQTT topic for a device's set command.
func (s TopicScheme) Set(name string) string { return path.Join(s.base(name), "set") }

// Event returns the MQTT topic for a device's events.
func (s TopicScheme) Event(name string) string { return path.Join(s.base(name), "event") }

// Status returns the MQTT topic for a device's status.
func (s TopicScheme) Status(name string) string { return path.Join(s.base(name), "status") }

// Meta returns the MQTT topic for a device's metadata.
func (s TopicScheme) Meta(name string) string { return path.Join(s.base(name), "meta") }
