//go:build ignore

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
//	// Create a messanger with embedded MQTT broker
//	msg := messanger.NewMessanger("otto")
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
	"fmt"
	"log/slog"
)

// MessangerConfig holds configuration parameters for messanger initialization.
// Currently supports broker address configuration for MQTT-based messangers.
type Config struct {
	// Broker is the address of the MQTT broker (hostname or IP)
	Broker   string `json:"broker"`
	Username string `json:"username"`
	Password string `json:"password"`
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
type MessangerIntf interface {
	// ID returns the unique identifier for this messanger instance
	ID() string

	// Connect establishes connection to the underlying messaging backend.
	// For local messangers this is a no-op. For MQTT it connects to the broker.
	Connect() error

	// Pub publishes data to the specified topic. The data can be any type that
	// can be converted to []byte (string, []byte, int, bool, float64).
	// Returns an error if publishing fails.
	Pub(topic string, msg any) error

	// Error returns the last error encountered by the messanger, if any
	Error() error

	// Close cleanly shuts down the messanger, closing connections and
	// cleaning up resources
	Close()
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

func (mb *MessangerBase) Pub(topic string, data any) error {

	b, err := Bytes(data)
	if err != nil {
		slog.Error("messanger failed to convert bytes", "error", err)
		return err
	}
	msg := NewMsg(topic, b, "otto")
	return mb.PubMsg(msg)
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
