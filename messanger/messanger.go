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
type MsgHandler func(msg *Msg)

// Messanger is the interface that all messangers must implement
type Messanger interface {
	ID() string
	Subscribe(topic string, handler MsgHandler) error
	SetTopic(topic string)
	Topic() string
	PubMsg(msg *Msg)
	PubData(data any)
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
	topic []string
	subs  map[string][]MsgHandler
	error

	Published int
}

func NewMessangerBase(id string, topic ...string) *MessangerBase {
	return &MessangerBase{
		id:    id,
		topic: topic,
		subs:  make(map[string][]MsgHandler),
	}
}

func (mb *MessangerBase) ID() string {
	return mb.id
}

func (mb *MessangerBase) Topic() string {
	if len(mb.topic) < 1 {
		return ""
	}
	return mb.topic[0]
}

func (mb *MessangerBase) SetTopic(topic string) {
	mb.topic = append(mb.topic, topic)
}

func (mb *MessangerBase) Error() error {
	return mb.error
}

// ServeHTTP is the REST API entry point for the messanger package
func (m MessangerBase) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var subs []string
	for s, _ := range m.subs {
		subs = append(subs, s)
	}

	mbase := struct {
		ID        string
		Topics    []string
		Subs      []string
		Published int
	}{
		ID:        m.id,
		Subs:      subs,
		Topics:    m.topic,
		Published: m.Published,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(mbase)
	if err != nil {
		slog.Error("MQTT.ServeHTTP failed to encode", "error", err)
	}
}
