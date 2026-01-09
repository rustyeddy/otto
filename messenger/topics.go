package messenger

import "path"

// I think we should add a station to the MQTT topic path

type TopicScheme struct {
	Prefix string // e.g. "otto" or "home"
}

func (s TopicScheme) base(name string) string { return path.Join(s.Prefix, "devices", name) }

func (s TopicScheme) State(name string) string  { return path.Join(s.base(name), "state") }
func (s TopicScheme) Set(name string) string    { return path.Join(s.base(name), "set") }
func (s TopicScheme) Event(name string) string  { return path.Join(s.base(name), "event") }
func (s TopicScheme) Status(name string) string { return path.Join(s.base(name), "status") }
func (s TopicScheme) Meta(name string) string   { return path.Join(s.base(name), "meta") }
