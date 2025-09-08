package messanger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

type MessangerLocal struct {
	id        string
	Topic     string
	Published int
	Subs      map[string][]MsgHandler
	sync.Mutex
}

// NewMessanger with the given ID and a variable number of topics that
// it will subscribe to.
func NewMessangerLocal(id string, topic ...string) *MessangerLocal {
	m := &MessangerLocal{
		id: id,
	}
	m.Subs = make(map[string][]MsgHandler)
	if len(topic) > 0 {
		m.Topic = topic[0]
	}
	return m
}

func (m *MessangerLocal) ID() string {
	return m.id
}

func (m *MessangerLocal) SetTopic(topic string) {
	m.Topic = topic
}

// Subscribe will literally subscribe to the provide MQTT topic with
// the specified message handler.
func (m *MessangerLocal) Subscribe(topic string, handler MsgHandler) error {
	m.Subs[topic] = append(m.Subs[topic], handler)
	root.insert(topic, handler)
	return nil
}

// Publish a message via MQTT with the given topic and value
// func (m *MessangerLocal) Pub(value any) {
// 	m.Published++
// 	msg := New(topic, value, m.id)
// 	m.PubMsg(msg)
// }

// PubMsg sends an MQTT message based on the content of the Msg structure
func (m *MessangerLocal) PubMsg(msg *Msg) {
	n := root.lookup(msg.Topic)
	if n == nil {
		slog.Info("No subscribers", "topic", msg.Topic)
		return
	}
	n.pub(msg)
}

// Publish given data to this messangers topic
func (m *MessangerLocal) PubData(data any) {
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

	msg := NewMsg(m.Topic, buf, m.id)
	m.PubMsg(msg)
}

// ServeHTTP is the REST API entry point for the messanger package
func (m *MessangerLocal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(m)
	if err != nil {
		slog.Error("MQTT.ServeHTTP failed to encode", "error", err)
	}
}

type node struct {
	index    string
	nodes    map[string]*node
	handlers []MsgHandler
}

var (
	root node
)

func init() {
	root = node{
		nodes: make(map[string]*node),
	}
}

func newNode(index string) *node {
	n := &node{
		index: index,
		nodes: make(map[string]*node),
	}
	return n
}

func (n *node) insert(topic string, mh MsgHandler) {
	indexes := strings.Split(topic, "/")
	pn := n
	for _, idx := range indexes {
		if nn, ex := pn.nodes[idx]; !ex {
			nn = newNode(idx)
			pn.nodes[idx] = nn
			pn = nn
		} else {
			pn = nn
		}

	}
	// The last node push the callback on the callback list
	pn.handlers = append(pn.handlers, mh)
}

func (n *node) lookup(topic string) *node {
	indexes := strings.Split(topic, "/")
	pn := n
	for _, idx := range indexes {

		nn, ex := pn.nodes[idx]
		if ex {
			pn = nn
			continue
		}

		nn, ex = pn.nodes["#"]
		if ex {
			return nn
		}

		nn, ex = pn.nodes["+"]
		if ex {
			// we will accept this path portion of the wildcard, but
			// must continue on
			pn = nn
			continue
		}
		return nil
	}
	return pn
}

func (n *node) pub(m *Msg) {
	for _, h := range n.handlers {
		h(m)
	}
}
