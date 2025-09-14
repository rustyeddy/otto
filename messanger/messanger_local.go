package messanger

import (
	"log/slog"
	"sync"
)

// MessangerLocal implements a local in-memory messanger that
// can be used for testing or for simple in-process communication.
// MessangerLocal does not require any external dependencies like
// an MQTT broker.
type MessangerLocal struct {
	*MessangerBase
	sync.Mutex `json:"-"`
}

// NewMessanger with the given ID and a variable number of topics that
// it will subscribe to.
func NewMessangerLocal(id string, topics ...string) *MessangerLocal {
	m := &MessangerLocal{
		MessangerBase: NewMessangerBase(id, topics...),
	}
	return m
}

// Subscribe will literally subscribe to the provide MQTT topic with
// the specified message handler.
func (m *MessangerLocal) Subscribe(topic string, handler MsgHandler) error {
	m.subs[topic] = append(m.subs[topic], handler)
	if root == nil {
		root = newNode("/")
	}
	root.insert(topic, handler)
	return nil
}

// Close is implemented to satisfy the messanger interface
func (m *MessangerLocal) Close() {
	// TODO - need to unsubscribe handlers for this messanger
	resetNodes()
	m.MessangerBase.subs = nil
}

// Pub a message via MQTT with the given topic and value
func (m *MessangerLocal) Pub(topic string, value any) {
	m.Published++
	buf, err := Bytes(value)
	if err != nil {
		m.error = err
		return
	}

	msg := NewMsg(topic, buf, m.id)
	m.PubMsg(msg)
}

// PubMsg sends an MQTT message based on the content of the Msg structure
func (m *MessangerLocal) PubMsg(msg *Msg) {
	n := root.lookup(msg.Topic)
	if n == nil {
		slog.Info("No subscribers", "topic", msg.Topic)
		return
	}
	n.pub(msg)
}

// Publish given data to this messangers topic
func (m *MessangerLocal) PubData(data any) {
	if len(m.topic) == 0 || m.topic[0] == "" {
		slog.Error("Device.Publish failed has no Topic", "name", m.ID())
		return
	}
	buf, err := Bytes(data)
	if err != nil {
		panic(err)
	}
	msg := NewMsg(m.topic[0], buf, m.id)
	m.PubMsg(msg)
}
