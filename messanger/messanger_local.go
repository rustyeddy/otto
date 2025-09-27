package messanger

import (
	"fmt"
	"log/slog"
	"sync"
)

// MessangerLocal is a simple in-process messanger useful for tests and local usage.
// It implements the Messanger interface and uses MessangerBase for common behavior.
type MessangerLocal struct {
	*MessangerBase
	sync.Mutex `json:"-"`
}

// NewMessangerLocal creates a new local messanger instance.
func NewMessangerLocal(id string, topics ...string) *MessangerLocal {
	return &MessangerLocal{
		MessangerBase: NewMessangerBase(id, topics...),
	}
}

func (m *MessangerLocal) ID() string {
	return m.MessangerBase.ID()
}

// Subscribe stores subscription handlers locally (base implementation already handles it).
func (m *MessangerLocal) Subscribe(topic string, handler MsgHandler) error {
	root.insert(topic, handler)
	return m.MessangerBase.Subscribe(topic, handler)
}

// Pub publishes a value to an explicit topic using the local messanger.
// Returns an error if publishing fails (local impl always succeeds).
func (m *MessangerLocal) Pub(topic string, value any) error {
	m.Published++

	if topic == "" {
		return fmt.Errorf("No topic")
	}
	if len(m.subs) == 0 {
		return fmt.Errorf("No subscribers")
	}

	b, err := Bytes(value)
	if err != nil {
		return err
	}

	msg := NewMsg(topic, b, m.id)
	return m.PubMsg(msg)
}

// PubData publishes arbitrary data to the default topic of this messanger.
// Returns error if no topic is configured or conversion fails.
func (m *MessangerLocal) PubData(data any) error {
	if len(m.topic) == 0 {
		return fmt.Errorf("no topic set")
	}
	topic := m.topic[0]
	bytes, err := Bytes(data)
	if err != nil {
		m.error = fmt.Errorf("failed to convert data to bytes: %w", err)
		return m.error
	}
	msg := NewMsg(topic, bytes, m.id)
	return m.PubMsg(msg)
}

// PubMsg publishes a pre-built Msg. Returns error on failure.
func (m *MessangerLocal) PubMsg(msg *Msg) error {
	m.Published++

	if msg == nil {
		return fmt.Errorf("nil message")
	}

	// look up local routing table to pass message along to subscribers
	n := root.lookup(msg.Topic)
	if n == nil {
		return fmt.Errorf("No subscribers for %s\n", msg.Topic)
	}

	n.pub(msg)
	return nil
}

func (m *MessangerLocal) Error() error {
	return m.MessangerBase.Error()
}

func (m *MessangerLocal) Close() {
	// clear subscriptions
	m.Lock()
	defer m.Unlock()

	// remove the handler from the root node
	for t, h := range m.subs {
		root.remove(t, h)
	}
	slog.Debug("MessangerLocal.Close", "id", m.ID())
}

// helper: convert common types to []byte as used by local pub dispatch.
// This helper is internal and will call the workspace Bytes() when possible.
// If Bytes() is not available, keep a minimal fallback implementation.
func toBytesUnchecked(v any) []byte {
	// try using Bytes helper if available in this package
	if b, err := Bytes(v); err == nil {
		return b
	}
	// best-effort fallback
	switch x := v.(type) {
	case []byte:
		return x
	case string:
		return []byte(x)
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}
