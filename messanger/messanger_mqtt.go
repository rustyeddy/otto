package messanger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// Messanger represents a type that can publish and subscribe to messages
type MessangerMQTT struct {
	id    string `json:"id"`
	Topic string `json:"topic"`

	Published int64                   `json:"published"`
	Subs      map[string][]MsgHandler `json:"-"`

	*MQTT
	sync.Mutex `json:"-"`
}

// NewMessanger with the given ID and a variable number of topics that
// it will subscribe to.
func NewMessangerMQTT(id string, topic ...string) *MessangerMQTT {
	m := &MessangerMQTT{
		id: id,
	}
	m.MQTT = GetMQTT()
	m.Subs = make(map[string][]MsgHandler)
	if len(topic) > 0 {
		m.Topic = topic[0]
	}
	return m
}

func (m *MessangerMQTT) ID() string {
	return m.id
}

func (m *MessangerMQTT) SetTopic(topic string) {
	m.Topic = topic
}

// Subscribe will literally subscribe to the provide MQTT topic with
// the specified message handler.
func (m *MessangerMQTT) Subscribe(topic string, handler MsgHandler) error {
	m.Subs[topic] = append(m.Subs[topic], handler)
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
	if m.Topic == "" {
		slog.Error("Device.Publish failed has no Topic", "name", m.ID)
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

	msg := NewMsg(m.Topic, buf, m.id)
	m.PubMsg(msg)
}

// ServeHTTP is the REST API entry point for the messanger package
func (m *MessangerMQTT) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(m)
	if err != nil {
		slog.Error("MQTT.ServeHTTP failed to encode", "error", err)
	}
}

// MsgPrinter will simply print a Msg that has been supplied. TODO,
// create a member function that will print messages by msg ID.
type MsgPrinter struct{}

// MsgHandler will print out the message that has been supplied.
func (m *MsgPrinter) MsgHandler(msg *Msg) {
	fmt.Printf("%+v\n", msg)
}
