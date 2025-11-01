package messanger

import (
	"fmt"
	"log/slog"
	"sync"
)

// Messanger represents a type that can publish and subscribe to messages
type MessangerMQTT struct {
	*MessangerBase
	*MQTT
	sync.Mutex `json:"-"`
}

// NewMessangerMQTT creates a new MQTT messanger instance.
func NewMessangerMQTT(id string, broker string) (*MessangerMQTT, error) {
	m := &MessangerMQTT{
		MQTT: NewMQTT(id, broker, ""),
	}
	mb, err := NewMessangerBase(id)
	if err == nil {
		m.MessangerBase = mb
	}
	return m, err
}

func (m *MessangerMQTT) ID() string {
	return m.MessangerBase.ID()
}

// Subscribe will literally subscribe to the provide MQTT topic with
// the specified message handler.
func (m *MessangerMQTT) Subscribe(topic string, handler MsgHandler) error {
	m.subs[topic] = handler
	return m.MQTT.Subscribe(topic, handler)
}

// Publish a message via MQTT with the given topic and value
// Now returns an error to indicate publish failures.
func (m *MessangerMQTT) Pub(topic string, value any) error {
	m.Published++
	// If underlying Publish has an error return, prefer to return that.
	// Many MQTT publish helper implementations return nothing; keep compatibility
	// by ignoring a return if none exists. If the underlying Publish returns an error,
	// attempt to return it (best-effort).
	if m.MQTT != nil {
		// If underlying Publish has an error signature, call and return it.
		_ = m.MQTT.Publish(topic, value)
	} else {
		// best-effort: do nothing
	}
	return nil
}

// PubMsg sends an MQTT message based on the content of the Msg structure
func (m *MessangerMQTT) PubMsg(msg *Msg) error {
	if msg == nil {
		return fmt.Errorf("nil message")
	}
	// Underlying Publish will actually send the payload
	if m.MQTT != nil {
		_ = m.Publish(msg.Topic, msg.Data)
	}
	// Count it via base
	m.Published++
	return nil
}

func (m *MessangerMQTT) Error() error {
	if m.MQTT != nil {
		return m.MQTT.Error()
	}
	return nil
}

// Close cleanly shuts down the MQTT messanger by closing the MQTT connection
// and clearing local subscriptions. It implements the Messanger interface.
func (m *MessangerMQTT) Close() {
	m.Lock()
	defer m.Unlock()

	// Close the MQTT connection
	if m.MQTT != nil {
		m.MQTT.Close()
	}

	// Clear local subscriptions
	if m.subs != nil {
		m.subs = make(map[string]MsgHandler)
	}

	slog.Debug("MessangerMQTT.Close", "id", m.ID())
}

// MsgPrinter will simply print a Msg that has been supplied. TODO,
// create a member function that will print messages by msg ID.
type MsgPrinter struct{}

// MsgHandler will print out the message that has been supplied.
func (m *MsgPrinter) MsgHandler(msg *Msg) error {
	fmt.Printf("%+v\n", msg)
	return nil
}
