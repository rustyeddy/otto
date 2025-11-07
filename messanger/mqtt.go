package messanger

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	gomqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	mqtt *MQTT
)

// MQTT is a wrapper around the Paho MQTT Go package
// Wraps the Broker, ID and Debug variables.
type MQTT struct {
	id     string `json:"id"`
	Broker string `json:"broker"`
	Debug  bool   `json:"debug"`

	error         `json:"error"`
	gomqtt.Client `json:"-"`
}

// NewMQTT creates a new instance of the MQTT client type.
func NewMQTT(id string, broker string, topics string) *MQTT {
	mqtt = &MQTT{
		id:     id,
		Broker: broker,
	}
	return mqtt
}

// SetMQTTClient allows a mock client to be used rather than an
// actual MQTT client to allow for Mocking MQTT connections if
// desired
func (m *MQTT) SetMQTTClient(c gomqtt.Client) *MQTT {
	m.Client = c
	return m
}

// GetMQTT returns the singleton instance of the MQTT client, the
// first time it is called it will initialize the client if needed.
var GetMQTT = func() *MQTT {
	if mqtt == nil {
		mqtt = &MQTT{
			id:     "default",
			Broker: "localhost",
		}
	}
	return mqtt
}

func (m *MQTT) ID() string {
	return m.id
}

// IsConnected will tell you if the MQTT client is connected to
// the specified broker
func (m *MQTT) IsConnected() bool {
	if m.Client == nil {
		return false
	}
	return m.Client.IsConnected()
}

func (m *MQTT) Error() error {
	return m.error
}

// Connect to the MQTT broker after setting some MQTT options
// then connecting to the MQTT broker
func (m *MQTT) Connect() error {

	if m.Debug {
		gomqtt.DEBUG = log.Default()
		gomqtt.ERROR = log.Default()
	}

	if m.Broker == "" {
		m.Broker = os.Getenv("MQTT_BROKER")
	}
	if m.Broker == "" {
		m.Broker = "localhost"
	}

	url := "tcp://" + m.Broker + ":1883"
	opts := gomqtt.NewClientOptions()
	opts.AddBroker(url)
	opts.SetClientID(m.id)
	opts.SetCleanSession(true)

	// If we are testing m.Client will point to the mock client otherwise
	// in real life a new real client will be created
	if m.Client == nil {
		m.Client = gomqtt.NewClient(opts)
	}

	token := m.Client.Connect()
	token.Wait()
	if token.Error() != nil {
		slog.Error("MQTT Connect: ", "error", token.Error())
		m.error = token.Error()
		return fmt.Errorf("Failed to connect to MQTT broker %s", token.Error())
	}
	return nil
}

// Subscribe will cause messangers to the given topic to be passed along to the
// MsgHandle f
func (m *MQTT) Subscribe(topic string, f MsgHandler) error {
	if m.Client == nil {
		slog.Warn("MQTT client is not connected to a broker: ", "broker", m.Broker)
		return fmt.Errorf("MQTT Client is not connected to broker: %s", m.Broker)
	}

	var err error
	token := m.Client.Subscribe(topic, byte(0), func(c gomqtt.Client, m gomqtt.Message) {
		slog.Debug("MQTT incoming: ", "topic", m.Topic(), "payload", string(m.Payload()))
		msg := NewMsg(m.Topic(), m.Payload(), "mqtt-sub")
		f(msg)
	})

	token.Wait()
	if token.Error() != nil {
		// TODO: add routing that automatically subscribes subscribers when a
		// connection has been made
		m.error = token.Error()
		return token.Error()
	}
	return err
}

// Publish will publish a value to the given topic
func (m *MQTT) Publish(topic string, value any) error {
	var t gomqtt.Token

	if topic == "" {
		return fmt.Errorf("MQTT Publish topic is nil")
	}

	if m.Client == nil {
		return fmt.Errorf("MQTT Client is not connected to a broker")
	}

	if t = m.Client.Publish(topic, byte(0), false, value); t == nil {
		if false {
			return fmt.Errorf("MQTT Pub NULL token topic %s - value: %+v", topic, value)
		}
		return nil
	}

	t.Wait()
	if t.Error() != nil {
		m.error = t.Error()
		return fmt.Errorf("MQTT Publish token error %+v", t.Error())
	}
	return nil
}

// Close will disconnect from the MQTT broker and close the client
// connection.
func (m *MQTT) Close() {
	if m.Client != nil {
		m.Client.Disconnect(1000)
	}
}
