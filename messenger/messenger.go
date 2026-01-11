package messenger

import (
	"context"
	"log/slog"
	"sync"
)

type subSpec struct {
	topic   string
	qos     byte
	handler func(Message)
}

// Messenger manages desired MQTT subscriptions.
type Messenger struct {
	MQTT MQTT

	mu            sync.RWMutex
	subscriptions map[string]subSpec // key=topic
	unsubs        map[string]func() error
}

// New returns a Messenger for the provided MQTT client.
func New(mqtt MQTT) *Messenger {
	return &Messenger{
		MQTT:          mqtt,
		subscriptions: map[string]subSpec{},
		unsubs:        map[string]func() error{},
	}
}

// WantSub registers a subscription that should always be active.
func (m *Messenger) WantSub(topic string, qos byte, handler func(Message)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscriptions[topic] = subSpec{topic: topic, qos: qos, handler: handler}
}

// ResubscribeAll applies desired subscriptions on connect and reconnect.
func (m *Messenger) ResubscribeAll(ctx context.Context) {
	slog.Info("MQTT connected; (re)subscribing", "count", len(m.subscriptions))

	// Snapshot desired subs under read lock, then do network work unlocked.
	m.mu.RLock()
	subs := make([]subSpec, 0, len(m.subscriptions))
	for _, s := range m.subscriptions {
		subs = append(subs, s)
	}
	m.mu.RUnlock()

	for _, s := range subs {
		// If we already had an active sub, try to unsubscribe first (optional).
		m.mu.Lock()
		if u, ok := m.unsubs[s.topic]; ok && u != nil {
			_ = u()
			delete(m.unsubs, s.topic)
		}
		m.mu.Unlock()

		unsub, err := m.MQTT.Subscribe(ctx, s.topic, s.qos, s.handler)
		if err != nil {
			slog.Error("MQTT subscribe failed", "topic", s.topic, "error", err)
			continue
		}

		m.mu.Lock()
		m.unsubs[s.topic] = unsub
		m.mu.Unlock()

		slog.Info("MQTT subscribed", "topic", s.topic, "qos", s.qos)
	}
}
