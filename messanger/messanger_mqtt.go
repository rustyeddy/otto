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

// NewMessanger with the given ID and a variable number of topics that
// it will subscribe to.
func NewMessangerMQTT(id string, topics ...string) *MessangerMQTT {
	m := &MessangerMQTT{
		MessangerBase: NewMessangerBase(id, topics...),
		MQTT:          NewMQTT(id, topics...),
	}
	return m
}

func (m *MessangerMQTT) ID() string {
	return m.MessangerBase.ID()
}

// Subscribe will literally subscribe to the provide MQTT topic with
// the specified message handler.
func (m *MessangerMQTT) Subscribe(topic string, handler MsgHandler) error {
	m.subs[topic] = append(m.subs[topic], handler)
	return m.MQTT.Subscribe(topic, handler)
}

// Publish a message via MQTT with the given topic and value
func (m *MessangerMQTT) Pub(topic string, value any) {
	m.Published++
	m.Publish(topic, value)
}

// PubMsg sends an MQTT message based on the content of the Msg structure
func (m *MessangerMQTT) PubMsg(msg *Msg) {
	m.Publish(msg.Topic, msg.Data)
}

// Publish given data to this messangers topic
func (m *MessangerMQTT) PubData(data any) {
	if len(m.topic) < 1 || m.topic[0] == "" {
		slog.Error("Device.Publish failed has no Topic", "name", m.MessangerBase.id)
		return
	}
	var buf []byte

	switch data.(type) {
	case []byte:
		buf = data.([]byte)

	case string:
		buf = []byte(data.(string))

	case int:
		str := fmt.Sprintf("%d", data.(int))
		buf = []byte(str)

	case float64:
		str := fmt.Sprintf("%5.2f", data.(float64))
		buf = []byte(str)

	default:
		slog.Error("Unknown Type: ", "topic", m.Topic, "type", fmt.Sprintf("%T", data))
	}

	msg := NewMsg(m.topic[0], buf, m.MessangerBase.id)
	m.PubMsg(msg)
}

func (m *MessangerMQTT) Error() error {
	return m.MQTT.Error()
}

// MsgPrinter will simply print a Msg that has been supplied. TODO,
// create a member function that will print messages by msg ID.
type MsgPrinter struct{}

// MsgHandler will print out the message that has been supplied.
func (m *MsgPrinter) MsgHandler(msg *Msg) {
	fmt.Printf("%+v\n", msg)
}
