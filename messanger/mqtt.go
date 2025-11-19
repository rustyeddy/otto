package messanger

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	gomqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	// mqtt holds the singleton MQTT client instance
	mqtt *MQTT
)

// MQTT wraps the Eclipse Paho MQTT Go client with Otto-specific configuration
// and error handling. It provides a simplified interface for connecting to
// MQTT brokers and publishing/subscribing to topics.
//
// The wrapper handles:
//   - Connection management with automatic reconnection
//   - Authentication (username/password)
//   - Debug logging when enabled
//   - Error tracking
//   - Mock client injection for testing
//
// Default Configuration:
//   - Port: 1883 (standard MQTT)
//   - Protocol: TCP
//   - Username: "otto"
//   - Password: "otto123"
//   - Clean Session: true
//
// Example:
//
//	client := NewMQTT("sensor1", "mqtt.example.com", "")
//	err := client.Connect()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	client.Publish("sensors/temp", "25.5")
type MQTT struct {
	id       string `json:"id"`       // Unique client identifier
	Broker   string `json:"broker"`   // Broker hostname or IP (no protocol/port)
	Username string `json:"username"` // MQTT authentication username
	Password string `json:"password"` // MQTT authentication password
	Debug    bool   `json:"debug"`    // Enable Paho MQTT client debug logging

	error         `json:"error"` // Last error encountered
	gomqtt.Client `json:"-"`     // Embedded Paho MQTT client
}

// NewMQTT creates a new MQTT client instance with default credentials.
// The client is not connected until Connect() is called.
//
// Parameters:
//   - id: Unique client identifier (used for MQTT client ID)
//   - broker: Broker hostname or IP address (without tcp:// prefix or port)
//   - topics: Deprecated/unused parameter, kept for compatibility
//
// Returns a pointer to the initialized MQTT client.
//
// Example:
//
//	client := NewMQTT("sensor1", "localhost", "")
//	err := client.Connect()
func NewMQTT(id string, broker string, topics string) *MQTT {
	mqtt = &MQTT{
		id:       id,
		Broker:   broker,
		Username: "otto",
		Password: "otto123",
	}
	return mqtt
}

// SetMQTTClient injects a custom or mock MQTT client for testing purposes.
// This allows unit tests to use a mock client instead of connecting to a real broker.
//
// Parameters:
//   - c: The client to use (typically a MockClient for testing)
//
// Returns the MQTT instance for method chaining.
//
// Example:
//
//	mockClient := NewMockClient()
//	mqtt.SetMQTTClient(mockClient)
func (m *MQTT) SetMQTTClient(c gomqtt.Client) *MQTT {
	m.Client = c
	return m
}

// GetMQTT returns the singleton MQTT client instance, creating it with
// default settings if it doesn't exist yet.
//
// Default settings:
//   - ID: "default"
//   - Broker: "localhost"
//
// Returns a pointer to the MQTT client.
//
// Note: This is a variable (not a const function) to allow test code to
// override it if needed.
var GetMQTT = func() *MQTT {
	if mqtt == nil {
		mqtt = &MQTT{
			id:     "default",
			Broker: "localhost",
		}
	}
	return mqtt
}

// ID returns the unique identifier for this MQTT client.
func (m *MQTT) ID() string {
	return m.id
}

// IsConnected checks whether the MQTT client is currently connected to the broker.
//
// Returns true if connected, false otherwise.
//
// Example:
//
//	if !client.IsConnected() {
//	    client.Connect()
//	}
func (m *MQTT) IsConnected() bool {
	if m.Client == nil {
		return false
	}
	return m.Client.IsConnected()
}

// Error returns the last error encountered by the MQTT client.
// Returns nil if no error has occurred.
func (m *MQTT) Error() error {
	return m.error
}

// Connect establishes a connection to the configured MQTT broker.
// It configures the MQTT client with the appropriate options and attempts
// to connect, waiting for the operation to complete.
//
// Configuration precedence:
//  1. Explicitly set values on the MQTT struct
//  2. MQTT_BROKER environment variable (for broker address)
//  3. Default values ("localhost", "otto", "otto123")
//
// Connection parameters:
//   - URL format: tcp://broker:1883
//   - Clean Session: true (start fresh on each connection)
//   - QoS: 0 (fire and forget)
//
// If Debug is enabled, MQTT client debug and error logs are written to the default logger.
//
// Returns an error if the connection fails, nil on success.
//
// Example:
//
//	client := NewMQTT("sensor1", "mqtt.example.com", "")
//	err := client.Connect()
//	if err != nil {
//	    log.Fatalf("Failed to connect: %v", err)
//	}
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
	// Set username and password if provided
	if m.Username == "" {
		m.Username = "otto"
	}
	if m.Password == "" {
		m.Password = "otto123"
	}

	url := "tcp://" + m.Broker + ":1883"
	opts := gomqtt.NewClientOptions()
	opts.AddBroker(url)
	opts.SetClientID(m.id)
	opts.SetCleanSession(true)
	opts.SetUsername(m.Username)
	opts.SetPassword(m.Password)
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

// Subscribe registers a message handler for the specified MQTT topic pattern.
// When messages matching the topic are received, they are converted to Otto's
// Msg format and passed to the handler function.
//
// The subscription uses QoS 0 (at most once delivery).
//
// MQTT Wildcard Support:
//   - '+' matches exactly one topic level (e.g., "sensors/+/temp")
//   - '#' matches zero or more levels (e.g., "sensors/#")
//
// Parameters:
//   - topic: The MQTT topic pattern to subscribe to
//   - f: The handler function to invoke for each received message
//
// Returns an error if:
//   - The client is not connected
//   - The subscription operation fails
//
// Example:
//
//	client.Subscribe("sensors/+/temp", func(msg *Msg) error {
//	    fmt.Printf("Temp: %s\n", msg.String())
//	    return nil
//	})
//
// TODO: Add automatic re-subscription when reconnecting to the broker.
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

// Publish sends a message to the specified MQTT topic.
// The message is sent with QoS 0 (at most once) and is not retained by the broker.
//
// Parameters:
//   - topic: The MQTT topic to publish to (must not be empty)
//   - value: The message payload (any type accepted by the MQTT client)
//
// Returns an error if:
//   - The topic is empty
//   - The client is not connected
//   - The publish operation fails
//
// Example:
//
//	err := client.Publish("sensors/temp", "25.5")
//	if err != nil {
//	    log.Printf("Publish failed: %v", err)
//	}
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

// Close gracefully disconnects from the MQTT broker.
// It waits up to 1000ms for in-flight messages to complete before disconnecting.
//
// After Close() is called, the client should not be used for further operations
// without reconnecting via Connect().
//
// Example:
//
//	defer client.Close()  // Ensure cleanup on exit
func (m *MQTT) Close() {
	if m.Client != nil {
		m.Client.Disconnect(1000)
	}
}
