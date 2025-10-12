// Package messanger provides an interface and implementation for a messaging system
// that supports subscribing to topics, publishing messages, and handling errors.
// It includes a base implementation and allows for different types of messangers
// such as "local" and "mqtt".

// MsgHandler defines the function signature for handling messages.
// It is used as a callback for subscribers to process incoming messages.

// Messanger is the interface that all messanger implementations must adhere to.
// It provides methods for subscribing to topics, publishing messages, setting topics,
// retrieving the current topic, handling errors, and closing the messanger.

// NewMessanger creates a new instance of a Messanger based on the provided ID.
// Supported IDs include "local" and "mqtt". If a messanger already exists, it logs
// a warning and replaces the existing instance with the new one.

// GetMessanger returns the singleton instance of the Messanger. It ensures that
// only one instance of the Messanger exists at any given time.

// MessangerBase is a base implementation of the Messanger interface. It provides
// common functionality such as managing topics, subscriptions, and published message
// counts. It can be extended by specific messanger implementations.

// NewMessangerBase creates a new instance of MessangerBase with the given ID and topics.
// It initializes the subscription map and sets the provided topics.

// ID returns the unique identifier of the MessangerBase instance.

// Topic returns the first topic in the list of topics managed by the MessangerBase.
// If no topics are set, it returns an empty string.

// SetTopic appends a new topic to the list of topics managed by the MessangerBase.

// Error returns the current error state of the MessangerBase instance.

// ServeHTTP is the REST API entry point for the messanger package. It provides
// information about the messanger instance, including its ID, topics, subscriptions,
// and the number of published messages. The response is returned in JSON format.
package messanger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

var (
	messanger     Messanger
	messangerLock sync.Mutex
	once          sync.Once
)

// Subscriber is an interface that defines a struct needs to have the
// Callback(topic string, data []byte) function defined.
type MsgHandler func(msg *Msg) error

type MessageHandler interface {
	HandleMsg() func(msg *Msg) error
}

// Messanger is the interface that all messangers must implement
type Messanger interface {
	ID() string
	Subscribe(topic string, handler MsgHandler) error
	SetTopic(topic string)
	Topic() string

	// Publish methods should return an error when something goes wrong.
	Pub(topic string, msg any) error
	PubMsg(msg *Msg) error
	PubData(data any) error

	Error() error
	Close()
}

func NewMessanger(id string, topics ...string) Messanger {
	switch id {
	case "local":
		messanger = NewMessangerLocal(id, topics...)
	case "mqtt":
		messanger = NewMessangerMQTT(id, topics...)
	default:
		messanger = nil
	}
	if messanger != nil && messanger.ID() != id {
		slog.Warn("Messanger already initialized with a different ID",
			"existing", messanger.ID(), "requested", id)
	}
	return messanger
}

// GetMessangerInstance returns the singleton instance of MessangerBase.
// It ensures that only one instance of MessangerBase is created.
func GetMessanger() Messanger {
	messangerLock.Lock()
	defer messangerLock.Unlock()
	return messanger
}

// MessangerBase
type MessangerBase struct {
	id    string
	topic string
	subs  map[string]MsgHandler
	error

	Published int
}

func NewMessangerBase(id string, topic ...string) *MessangerBase {
	return &MessangerBase{
		id:    id,
		topic: topic,
		subs:  make(map[string]MsgHandler),
	}
}

func (mb *MessangerBase) ID() string {
	return mb.id
}

func (mb *MessangerBase) Topic() string {
	return mb.topic
}

func (mb *MessangerBase) SetTopic(topic string) {
	mb.topic = topic
}

func (mb *MessangerBase) Error() error {
	return mb.error
}

// Subscribe stores the subscription locally in the base implementation.
// Specific messanger implementations should override this method to handle
// actual subscription logic (e.g., MQTT broker subscription).
func (mb *MessangerBase) Subscribe(topic string, handler MsgHandler) error {
	if mb.subs == nil {
		mb.subs = make(map[string]MsgHandler)
	}
	mb.subs[topic] = handler
	return nil
}

// PubMsg publishes a pre-formatted Msg structure.
// This base implementation only increments the Published counter.
// Specific messanger implementations should override this method to handle
// actual message publishing.
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

// PubData publishes data to the messanger's default topic.
// It handles various data types by converting them to a byte array before publishing.
// If no topic is set, it returns an error.
func (mb *MessangerBase) PubData(data any) error {
	if mb.topic == "" {
		return fmt.Errorf("no topic set")
	}

	// Convert data to bytes
	bytes, err := Bytes(data)
	if err != nil {
		return fmt.Errorf("failed to convert data to bytes: %w", err)
	}

	// Create and publish message
	msg := NewMsg(mb.topic, bytes, mb.id)
	return mb.PubMsg(msg)
}

// Close is implemented to satisfy the messanger interface.
// Base implementation does nothing - specific messanger implementations
// should override this method to handle cleanup (e.g., closing connections).
func (mb *MessangerBase) Close() {
	// Base implementation does nothing
	slog.Debug("MessangerBase.Close", "id", mb.id)
}

// ServeHTTP is the REST API entry point for the messanger package
func (m MessangerBase) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var subs []string
	for s := range m.subs {
		subs = append(subs, s)
	}

	mbase := struct {
		ID        string
		Topic     string
		Subs      []string
		Published int
	}{
		ID:        m.id,
		Subs:      subs,
		Topic:    m.topic,
		Published: m.Published,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(mbase)
	if err != nil {
		slog.Error("MQTT.ServeHTTP failed to encode", "error", err)
	}
}
