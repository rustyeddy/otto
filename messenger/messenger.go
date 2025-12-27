package messenger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	gomqtt "github.com/eclipse/paho.mqtt.golang"
)

type BrokerType int

const (
	None BrokerType = 0
	Builtin
	External
)

// MsgHandler is a callback function type for handling incoming messages.
// Subscribers provide a MsgHandler function that will be invoked when
// a message is received on a subscribed topic. The handler receives a
// pointer to the Msg and should return an error if message processing fails.
type MsgHandler func(msg *Msg) error

// MessageHandler is an interface for types that can handle messages.
// This provides an alternative to the MsgHandler function type for
// implementing message handling logic as methods on types.
type MessageHandler interface {
	HandleMsg(msg *Msg) error
}

type Conn interface {
	Connect(broker string, user string, pass string) error
	Close()
	Sub(string, MsgHandler) error
	Unsub(...string)
	PubMsg(*Msg) error
	IsConnected() bool
}

// Messenger
type Messenger struct {
	ID        string
	Broker    string
	Username  string
	Password  string
	Prefix    string
	Published int

	BrokerType
	Conn
	subscriptions map[string]MsgHandler
}

var (
	msgr *Messenger
)

// NewMessanger creates a new Messanger instance based on the provided ID.
// It sets up the appropriate messaging backend and stores it as the singleton instance.
//
// Supported ID values:
//   - "none": Creates a local in-process messanger without MQTT
//   - "otto": Starts an embedded MQTT broker and creates an MQTT messanger
//   - default: Creates an MQTT messanger connecting to an external broker
func NewMessenger(broker string) *Messenger {
	msgr = &Messenger{}
	msgr.ID = "otto-client"
	msgr.Broker = broker
	msgr.Username = os.Getenv("MQTT_USER")
	msgr.Password = os.Getenv("MQTT_PASS")
	msgr.Prefix = "o/"
	msgr.subscriptions = make(map[string]MsgHandler)

	switch broker {
	case "none":
		msgr.Conn = &nobrokerConn{}

	case "internal":
		// make sure the internal broker has been started
		_, err := StartMQTTBroker(context.Background())
		if err != nil {
			slog.Error("Failed to start embedded MQTT broker", "error", err)
			return nil
		}
		msgr.BrokerType = Builtin
		msgr.Conn = &connMQTT{}

	default:
		msgr.BrokerType = External
		msgr.Conn = &connMQTT{}
	}
	return msgr
}

func SetMessenger(m *Messenger) {
	msgr = m
}

func GetMessenger() *Messenger {
	return msgr
}

func (m *Messenger) Connect() error {
	err := m.Conn.Connect(m.Broker, m.Username, m.Password)
	if err == nil {
		for topic, handler := range m.subscriptions {
			err = m.Conn.Sub(topic, handler)
			if err != nil {
				slog.Error("MQTT failed to subscribe", "topic", topic, "error", err)
			}
		}
	}
	return err
}

func (m *Messenger) Close() {
	// remove the handler from the root node
	var topics []string
	for t := range m.subscriptions {
		topics = append(topics, t)
	}
	if m.Conn != nil {
		m.Conn.Unsub(topics...)
	}
}

func (m *Messenger) Sub(topic string, handler MsgHandler) error {
	if m.subscriptions == nil {
		m.subscriptions = make(map[string]MsgHandler)
	}
	m.subscriptions[topic] = handler
	return m.Conn.Sub(topic, handler)
}

func (m *Messenger) Unsub(topic string) {
	m.Conn.Unsub(topic)
}

// Pub takes a topic string and data, wraps those parameters into
// a Msg struct and call PubMsg(msg)
func (m *Messenger) Pub(topic string, data any) error {
	b, err := Bytes(data)
	if err != nil {
		slog.Error("messenger failed to convert bytes", "error", err)
		return err
	}
	msg := NewMsg(topic, b, "otto")
	return m.PubMsg(msg)
}

// Pub takes a topic string and data, wraps those parameters into
// a Msg struct and call PubMsg(msg)
func (m *Messenger) PubMsg(msg *Msg) error {
	return m.Conn.PubMsg(msg)
}

// ServeHTTP implements http.Handler to provide a REST API endpoint for
// inspecting messanger state. It returns a JSON response containing the
// messanger ID, list of subscribed topics, and count of published messages.
//
// Response format:
//
//	{
//	  "ID": "messanger-id",
//	  "Subs": ["topic1", "topic2"],
//	  "Published": 42
//	}
//
// This is useful for debugging and monitoring the messanger's state.
func (m *Messenger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var subs []string
	for s := range m.subscriptions {
		subs = append(subs, s)
	}

	mbase := struct {
		ID        string
		Subs      []string
		Published int
	}{
		ID:        m.ID,
		Subs:      subs,
		Published: m.Published,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(mbase)
	if err != nil {
		slog.Error("MQTT.ServeHTTP failed to encode", "error", err)
	}
}

type connMQTT struct {
	Debug         bool
	gomqtt.Client // Embedded Paho MQTT client
}

func (m *connMQTT) Connect(b string, u string, p string) error {
	if m.Debug {
		gomqtt.DEBUG = log.Default()
		gomqtt.ERROR = log.Default()
	}
	url := "tcp://" + b + ":1883"
	opts := gomqtt.NewClientOptions()
	opts.AddBroker(url)
	opts.SetClientID("o++o")
	opts.SetCleanSession(true)
	opts.SetUsername(u)
	opts.SetPassword(p)
	opts.SetCleanSession(true)

	// If we are testing m.Client will point to the mock client otherwise
	// in real life a new real client will be created
	if m.Client == nil {
		m.Client = gomqtt.NewClient(opts)
	}

	token := m.Client.Connect()
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("Failed to connect to MQTT broker %s", token.Error())
	}
	return nil
}

func (m *connMQTT) IsConnected() bool {
	return m.Client.IsConnected()
}

func (m *connMQTT) Close() {
	if m.Client != nil {
		m.Client.Disconnect(1000)
	}
	return
}

func (m *connMQTT) Sub(topic string, f MsgHandler) error {
	if m.Client == nil {
		return fmt.Errorf("MQTT Client is not connected to a broker")
	}

	var err error
	token := m.Client.Subscribe(topic, byte(0), func(c gomqtt.Client, m gomqtt.Message) {
		slog.Debug("MQTT incoming: ", "topic", m.Topic(), "payload", string(m.Payload()))
		msg := NewMsg(m.Topic(), m.Payload(), "mqtt-sub")
		f(msg)
	})

	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	return err

}

func (m *connMQTT) Unsub(topics ...string) {
	var token gomqtt.Token
	if m.Client == nil {
		return
	}

	if token = m.Unsubscribe(topics...); token.Wait() && token.Error() != nil {
		slog.Error("Unsubscribe error: ", "error", token.Error())
	}
}

func (m *connMQTT) PubMsg(msg *Msg) error {
	if m.Client == nil {
		return fmt.Errorf("MQTT Client is not connected to a broker")
	}

	val, err := Bytes(msg.Data)
	if err != nil {
		return fmt.Errorf("MQTT failed to convert msg to bytes: %+v", err)
	}

	var t gomqtt.Token
	if t = m.Client.Publish(msg.Topic, byte(0), false, val); t == nil {
		if false {
			return fmt.Errorf("MQTT Pub NULL token topic %s - value: %+v", msg.Topic, val)
		}
		return nil
	}

	t.Wait()
	if t.Error() != nil {
		return fmt.Errorf("MQTT Publish token error %+v", t.Error())
	}
	return nil
}

type nobrokerConn struct {
	root *node
}

func (c *nobrokerConn) Connect(b string, u string, p string) error {
	return nil
}

func (c *nobrokerConn) IsConnected() bool {
	return true
}

func (c *nobrokerConn) Close() {
}

func (c *nobrokerConn) Sub(topic string, handler MsgHandler) error {
	root.insert(topic, handler)
	return nil
}

func (c *nobrokerConn) Unsub(topics ...string) {
	for _, t := range topics {
		root.remove(t, nil)
	}
}

func (c *nobrokerConn) PubMsg(msg *Msg) error {
	if msg == nil {
		return fmt.Errorf("nil message")
	}

	// look up local routing table to pass message along to subscribers
	n := root.lookup(msg.Topic)
	if n == nil {
		return fmt.Errorf("No subscribers for %s\n", msg.Topic)
	}
	n.pub(msg)
	return nil
}
