package messanger

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

var (
	messanger Messanger
)

// Subscriber is an interface that defines a struct needs to have the
// Callback(topic string, data []byte) function defined.
type MsgHandler func(msg *Msg)

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
	if messanger != nil {
		slog.Warn("Calling new messanger when one already exists",
			"old", messanger.ID(), "new", id)
		messanger = nil
	}

	switch id {
	case "local":
		messanger = NewMessangerLocal(id, topics...)
	case "mqtt":
		messanger = NewMessangerMQTT(id, topics...)
	default:
		messanger = nil
	}
	return messanger
}

// GetMessangerInstance returns the singleton instance of MessangerBase.
// It ensures that only one instance of MessangerBase is created.
func GetMessanger() Messanger {
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
	for s := range m.subs {
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
