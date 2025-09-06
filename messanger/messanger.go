package messanger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// Subscriber is an interface that defines a struct needs to have the
// Callback(topic string, data []byte) function defined.
type MsgHandler func(msg *Msg)

// Subscriber interface defines a type that can subscribe to published messages
type Subscriber interface {
	Subscribe(topic string, handler MsgHandler) error
}

// Publisher interface allows objects to publish message to a particular
// topic as defined in the message.Msg
type Publisher interface {
	Publish(topic string, data any)
}

// PubSub is the combination of the Publisher and Subscriber
// interfaces for variout devices to use.
type PubSub interface {
	Publisher
	Subscriber

	Error() error
	Close()
}

// Messanger represents a type that can publish and subscribe to messages
type Messanger struct {
	ID        string                  `json:"id"`
	Topic     string                  `json:"topic"`
	Published int64                   `json:"published"`
	Subs      map[string][]MsgHandler `json:"-"`
	PubSub    `json"-"`

	sync.Mutex `json:"-"`
}

// NewMessanger with the given ID and a variable number of topics that
// it will subscribe to.
func NewMessanger(ID string, topic ...string) *Messanger {
	m := &Messanger{
		ID: ID,
	}
	m.PubSub = GetMQTT()
	m.Subs = make(map[string][]MsgHandler)
	if len(topic) > 0 {
		m.Topic = topic[0]
	}
	return m
}

func (m *Messanger) Name() string {
	return m.ID
}

// Subscribe will literally subscribe to the provide MQTT topic with
// the specified message handler.
func (m *Messanger) Subscribe(topic string, handler MsgHandler) error {
	m.Subs[topic] = append(m.Subs[topic], handler)
	return m.PubSub.Subscribe(topic, handler)
}

// Publish a message via MQTT with the given topic and value
func (m *Messanger) Publish(topic string, value any) {
	m.Published++
	m.PubSub.Publish(topic, value)
}

// PubMsg sends an MQTT message based on the content of the Msg structure
func (m *Messanger) PubMsg(msg *Msg) {
	m.Publish(msg.Topic, msg.Data)
}

// Publish given data to this messangers topic
func (m *Messanger) PubData(data any) {
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

	msg := New(m.Topic, buf, m.ID)
	m.PubMsg(msg)
}

// ServeHTTP is the REST API entry point for the messanger package
func (m *Messanger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
