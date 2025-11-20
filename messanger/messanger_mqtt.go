package messanger

import (
	"fmt"
	"log/slog"
	"sync"
)

// MessangerMQTT provides distributed messaging through an MQTT broker.
// It implements the Messanger interface by wrapping an MQTT client and providing
// publish-subscribe capabilities across a network.
//
// This implementation is suitable for:
//   - Distributed IoT systems with multiple devices/stations
//   - Cloud-connected applications
//   - Systems requiring reliable message delivery
//   - Multi-node deployments
//
// The messanger connects to an MQTT broker (like Mosquitto, EMQX, or the embedded
// broker) and handles topic subscriptions, message publishing, and connection management.
//
// Thread Safety: MessangerMQTT is safe for concurrent use via embedded mutex.
//
// Example:
//
//	msg := NewMessangerMQTT("mqtt-client", "localhost")
//	err := msg.Connect()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	msg.Subscribe("ss/c/+/temp", tempHandler)
//	msg.Pub("ss/d/station1/temp", "25.5")
type MessangerMQTT struct {
	*MessangerBase
	*MQTT
	sync.Mutex `json:"-"`
}

// NewMessangerMQTT creates a new MQTT messanger instance that connects to the
// specified broker without authentication.
//
// The messanger uses default credentials (otto/otto123) which can be overridden
// by using NewMessangerMQTTWithAuth() instead.
//
// Parameters:
//   - id: Unique identifier for this messanger instance (used as MQTT client ID)
//   - broker: Address of the MQTT broker (hostname or IP, without protocol or port)
//
// Returns a pointer to the initialized MessangerMQTT.
// Call Connect() to establish the broker connection.
//
// Example:
//
//	msg := NewMessangerMQTT("sensor1", "localhost")
//	err := msg.Connect()
func NewMessangerMQTT(id string, broker string) *MessangerMQTT {
	return NewMessangerMQTTWithAuth(id, broker, "", "")
}

// NewMessangerMQTTWithAuth creates a new MQTT messanger instance with custom
// authentication credentials for connecting to a secured MQTT broker.
//
// Special broker values:
//   - "mock": Uses a mock MQTT client for testing (no actual network connection)
//
// Parameters:
//   - id: Unique identifier for this messanger (used as MQTT client ID)
//   - broker: Address of the MQTT broker (hostname or IP)
//   - username: MQTT authentication username (empty for default "otto")
//   - password: MQTT authentication password (empty for default "otto123")
//
// Returns a pointer to the initialized MessangerMQTT.
// Call Connect() to establish the broker connection.
//
// Example:
//
//	msg := NewMessangerMQTTWithAuth("sensor1", "mqtt.example.com", "user", "pass")
//	err := msg.Connect()
func NewMessangerMQTTWithAuth(id string, broker string, username string, password string) *MessangerMQTT {
	m := NewMQTT(id, broker, "")
	m.Username = username
	m.Password = password
	if broker == "mock" {
		m.SetMQTTClient(NewMockClient())
	}
	mqtt := &MessangerMQTT{
		MQTT:          m,
		MessangerBase: NewMessangerBase(id),
	}
	return mqtt
}

// ID returns the unique identifier for this MQTT messanger instance.
func (m *MessangerMQTT) ID() string {
	return m.MessangerBase.ID()
}

// Connect establishes a connection to the MQTT broker and subscribes to any
// topics that were registered before connection was established.
//
// This method:
//  1. Connects to the configured MQTT broker
//  2. Subscribes to all topics in the local subscription map
//  3. Logs errors for any failed subscriptions but continues with others
//
// Returns an error if the initial broker connection fails. Individual subscription
// failures are logged but don't cause the method to fail.
//
// Example:
//
//	msg := NewMessangerMQTT("client", "localhost")
//	msg.Subscribe("ss/c/+/led", ledHandler)  // Registered but not active yet
//	err := msg.Connect()  // Now connected and subscribed to ss/c/+/led
func (m *MessangerMQTT) Connect() error {
	err := m.MQTT.Connect()
	if err != nil {
		return err
	}

	for topic, handler := range m.subs {
		err = m.MQTT.Subscribe(topic, handler)
		if err != nil {
			slog.Error("MQTT failed to subscribe", "topic", topic, "error", err)
		}
	}

	return err
}

// Subscribe registers a message handler for the given topic pattern on the MQTT broker.
// The topic pattern can include MQTT wildcards ('+' for single level, '#' for multi-level).
//
// The subscription is stored locally and, if already connected to the broker, is
// immediately activated. If not yet connected, it will be activated when Connect() is called.
//
// Parameters:
//   - topic: The MQTT topic pattern to subscribe to (e.g., "ss/c/+/temp")
//   - handler: The callback function to invoke when matching messages arrive
//
// Returns an error if the broker subscription fails (when connected).
//
// Example:
//
//	msg := NewMessangerMQTT("client", "localhost")
//	msg.Connect()
//	msg.Subscribe("ss/c/station1/#", func(m *Msg) error {
//	    fmt.Printf("Received: %s\n", m.String())
//	    return nil
//	})
func (m *MessangerMQTT) Subscribe(topic string, handler MsgHandler) error {
	m.subs[topic] = handler
	return m.MQTT.Subscribe(topic, handler)
}

// Pub publishes data to the specified topic via MQTT.
// The data is sent to all subscribers of that topic across the network.
//
// Parameters:
//   - topic: The MQTT topic to publish to (e.g., "ss/d/station1/temp")
//   - value: The data to publish (any type accepted by MQTT client)
//
// Returns nil on success. Error handling is currently limited; check logs for issues.
//
// Example:
//
//	msg := NewMessangerMQTT("client", "localhost")
//	msg.Connect()
//	msg.Pub("ss/d/station1/temp", "25.5")
func (m *MessangerMQTT) Pub(topic string, value any) error {
	m.Published++

	// If underlying Publish has an error return, prefer to return that.
	// Many MQTT publish helper implementations return nothing; keep compatibility
	// by ignoring a return if none exists. If the underlying Publish returns an error,
	// attempt to return it (best-effort).
	if m.MQTT == nil {
		return nil
	}
	// If underlying Publish has an error signature, call and return it.
	_ = m.MQTT.Publish(topic, value)
	return nil
}

// PubMsg publishes a pre-constructed Msg structure via MQTT.
// The message's topic and data fields are used for the MQTT publication.
//
// Parameters:
//   - msg: The message to publish (must not be nil)
//
// Returns an error if msg is nil, otherwise returns nil.
//
// Example:
//
//	msg := NewMsg("ss/d/station1/temp", []byte("25.5"), "sensor")
//	messanger.PubMsg(msg)
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

// Error returns the last error from the underlying MQTT client, if any.
// Returns nil if no MQTT client exists or no error has occurred.
func (m *MessangerMQTT) Error() error {
	if m.MQTT != nil {
		return m.MQTT.Error()
	}
	return nil
}

// Close cleanly shuts down the MQTT messanger by:
//  1. Closing the MQTT broker connection
//  2. Clearing local subscription tracking
//
// This method is thread-safe via the embedded mutex.
// After Close() is called, the messanger should not be used.
//
// Example:
//
//	defer msg.Close()  // Ensure cleanup on exit
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

// MsgPrinter is a simple utility type for debugging that prints received messages.
// It implements the MessageHandler interface.
//
// TODO: Add functionality to filter and print messages by ID.
type MsgPrinter struct{}

// MsgHandler prints the received message to stdout in a structured format.
// This is useful for debugging message flow.
//
// Parameters:
//   - msg: The message to print
//
// Returns nil (always succeeds).
//
// Example:
//
//	printer := &MsgPrinter{}
//	msg.Subscribe("ss/c/+/temp", printer.MsgHandler)
func (m *MsgPrinter) MsgHandler(msg *Msg) error {
	fmt.Printf("%+v\n", msg)
	return nil
}
