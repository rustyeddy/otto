// Package messanger provides a unified interface for publish-subscribe messaging
// in the Otto IoT framework. It supports multiple messaging backends including
// local in-process messaging and MQTT-based distributed messaging.
//
// The package implements a topic-based routing system where messages are published
// to topics and delivered to all subscribers of those topics. Topics follow the
// MQTT topic format with hierarchical paths separated by slashes (e.g., "ss/c/station/sensor").
//
// Key Components:
//   - Messanger interface: Core abstraction for all messaging implementations
//   - MessangerLocal: In-process messaging with wildcard topic support
//   - MessangerMQTT: MQTT broker-based distributed messaging
//   - Msg: Message structure containing topic, payload, and metadata
//   - Topics: Topic format validation and management
//
// Example Usage:
//
//	// Create a local messanger
//	msg := messanger.NewMessanger("local")
//
//	// Subscribe to a topic
//	msg.Subscribe("ss/c/station/+", func(m *messanger.Msg) error {
//	    fmt.Printf("Received: %s\n", m.String())
//	    return nil
//	})
//
//	// Publish a message
//	msg.Pub("ss/c/station/temp", "25.5")
package messanger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
)

// MessangerConfig holds configuration parameters for messanger initialization.
// Currently supports broker address configuration for MQTT-based messangers.
type Config struct {
	// Broker is the address of the MQTT broker (hostname or IP)
	Broker   string `json:"broker"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	// messanger holds the singleton instance of the active Messanger
	msgr          Messanger
	messangerLock sync.Mutex
	config        Config
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

// Messanger is the core interface that all messaging implementations must satisfy.
// It provides methods for connecting to a messaging backend, subscribing to topics,
// publishing messages, and managing the messanger lifecycle.
//
// Implementations include:
//   - MessangerLocal: In-process messaging with wildcard topic routing
//   - MessangerMQTT: MQTT broker-based distributed messaging
//
// Thread Safety: Implementations should be safe for concurrent use.
type Messanger interface {
	// ID returns the unique identifier for this messanger instance
	ID() string

	// Connect establishes connection to the underlying messaging backend.
	// For local messangers this is a no-op. For MQTT it connects to the broker.
	Connect() error

	// Subscribe registers a handler function to receive messages on the given topic.
	// The topic can include MQTT wildcards: '+' for single level, '#' for multi-level.
	// Returns an error if subscription fails.
	Subscribe(topic string, handler MsgHandler) error

	// Pub publishes data to the specified topic. The data can be any type that
	// can be converted to []byte (string, []byte, int, bool, float64).
	// Returns an error if publishing fails.
	Pub(topic string, msg any) error

	// PubMsg publishes a pre-constructed Msg to its embedded topic.
	// This is useful when you need more control over message metadata.
	// Returns an error if publishing fails.
	PubMsg(msg *Msg) error

	// Error returns the last error encountered by the messanger, if any
	Error() error

	// Close cleanly shuts down the messanger, closing connections and
	// cleaning up resources
	Close()
}

// NewMessanger creates a new Messanger instance based on the provided ID.
// It sets up the appropriate messaging backend and stores it as the singleton instance.
//
// Supported ID values:
//   - "none": Creates a local in-process messanger without MQTT
//   - "local": Starts an embedded MQTT broker and creates an MQTT messanger
//   - default: Creates an MQTT messanger connecting to an external broker

// The created messanger becomes the global singleton accessible via GetMessanger().
// If an invalid ID is provided, logs an error and returns nil.
//
// Example:
//
//	msg := messanger.NewMessanger("local")
//	if msg == nil {
//	    log.Fatal("Failed to create messanger")
//	}
func NewMessanger(broker string) (m Messanger) {

	switch broker {
	case "none":
		msgr = NewMessangerLocal(broker)

	case "otto":
		_, err := StartMQTTBroker(context.Background())
		if err != nil {
			slog.Error("Failed to start embedded MQTT broker", "error", err)
			return nil
		}
		msgr = NewMessangerMQTT(broker, broker)

	default:
		msgr = NewMessangerMQTT(broker, broker)
	}
	return msgr
}

// GetMessanger returns the singleton Messanger instance created by NewMessanger.
// This function is thread-safe and can be called from multiple goroutines.
// Returns nil if no messanger has been created yet.
//
// Example:
//
//	msg := messanger.GetMessanger()
//	if msg != nil {
//	    msg.Pub("ss/c/station/status", "online")
//	}
func GetMessanger() Messanger {
	messangerLock.Lock()
	defer messangerLock.Unlock()

	if msgr != nil {
		return msgr
	}

	// Take broker from the config first
	broker := config.Broker
	if broker == "" {
		// if no config look for environment variable
		broker = os.Getenv("MQTT_BROKER")
	}
	if broker == "" {
		// if no environment variable then default to the built in
		// broker
		broker = "otto"
	}

	user := config.Username
	if user == "" {
		user = os.Getenv("MQTT_USERNAME")
	}
	pass := config.Password
	if pass == "" {
		pass = os.Getenv("MQTT_PASSWORD")
	}

	var err error
	switch broker {
	case "none":
		msgr = NewMessangerLocal(broker)

	case "otto":
		shutdown, err = StartMQTTBroker(context.Background())
		if err != nil {
			slog.Error("Failed to start embedded MQTT broker", "error", err)

			// A hack if bind address is in use, skip out and just
			// use the client to bind to the already running broker
			if !strings.Contains(err.Error(), "bind: address already in use") {
				return nil
			}

			slog.Info("Assuming broker is already running, connecting to existing broker")
		}
		fallthrough

	default:
		msgr = NewMessangerMQTTWithAuth("otto", broker, user, pass)
	}

	ms := GetMsgSaver()
	ms.Saving = true

	return msgr
}

// MessangerBase provides a base implementation of the Messanger interface
// with common functionality that can be embedded in specific messanger types.
// It handles subscription tracking, published message counting, and basic
// error management.
//
// This type is typically embedded in concrete implementations like MessangerLocal
// and MessangerMQTT rather than used directly.
type MessangerBase struct {
	id    string                // Unique identifier for this messanger instance
	subs  map[string]MsgHandler // Map of topic to handler functions
	error                       // Last error encountered, if any

	Published int // Count of messages published through this messanger
}

// NewMessangerBase creates a new MessangerBase instance with the given ID.
// It initializes the subscription map and sets up the base structure.
// This is typically called by concrete messanger implementations.
//
// Parameters:
//   - id: Unique identifier for the messanger instance
//
// Returns a pointer to the initialized MessangerBase.
func NewMessangerBase(id string) *MessangerBase {
	mb := &MessangerBase{
		id:   id,
		subs: make(map[string]MsgHandler),
	}
	return mb

}

// ID returns the unique identifier of this MessangerBase instance.
func (mb *MessangerBase) ID() string {
	return mb.id
}

// Error returns the last error encountered by this messanger, if any.
// Returns nil if no error has occurred.
func (mb *MessangerBase) Error() error {
	return mb.error
}

// Subscribe stores a topic subscription locally in the base implementation.
// It maps the topic string to the handler function for later retrieval.
//
// This base implementation only stores the subscription locally. Concrete
// messanger implementations should override this method to perform actual
// subscription logic (e.g., subscribing to an MQTT broker).
//
// Parameters:
//   - topic: The topic pattern to subscribe to (may include wildcards)
//   - handler: The callback function to invoke when messages arrive on this topic
//
// Returns nil on success, or an error if subscription fails.
func (mb *MessangerBase) Subscribe(topic string, handler MsgHandler) error {
	if mb.subs == nil {
		mb.subs = make(map[string]MsgHandler)
	}
	mb.subs[topic] = handler
	return nil
}

// PubMsg publishes a pre-constructed Msg structure.
// This base implementation only increments the Published counter for metrics.
//
// Concrete messanger implementations should override this method to perform
// actual message publishing (e.g., sending via MQTT or local routing).
//
// Parameters:
//   - msg: The message to publish (must not be nil)
//
// Returns an error if msg is nil, otherwise returns nil.
func (mb *MessangerBase) PubMsg(msg *Msg) error {
	if msg == nil {
		return fmt.Errorf("nil message")
	}
	mb.Published++
	// Base implementation just counts - specific messanger types should override
	// to actually publish the message
	slog.Debug("MessangerBase.PubMsg", "topic", msg.Topic, "published_count", mb.Published)
	return nil
}

// Close cleanly shuts down the messanger.
// This base implementation is a no-op that just logs the close operation.
//
// Concrete messanger implementations should override this method to perform
// actual cleanup (e.g., closing network connections, unsubscribing, etc.).
func (mb *MessangerBase) Close() {
	// Base implementation does nothing
	slog.Debug("MessangerBase.Close", "id", mb.id)
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
func (m MessangerBase) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var subs []string
	for s := range m.subs {
		subs = append(subs, s)
	}

	mbase := struct {
		ID        string
		Subs      []string
		Published int
	}{
		ID:        m.id,
		Subs:      subs,
		Published: m.Published,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(mbase)
	if err != nil {
		slog.Error("MQTT.ServeHTTP failed to encode", "error", err)
	}
}
