package messanger

import (
	"fmt"
	"log/slog"
	"sync"
)

// MessangerLocal provides an in-process messaging implementation that routes
// messages through a local topic tree without requiring an external broker.
// It supports MQTT-style wildcard subscriptions using a tree-based routing algorithm.
//
// This implementation is useful for:
//   - Testing message handlers without needing a broker
//   - Single-process applications where distributed messaging isn't needed
//   - Development and debugging
//   - Embedded systems with limited resources
//
// Wildcard Support:
//   - '+' matches exactly one topic level (e.g., "ss/c/+/temp")
//   - '#' matches zero or more levels (e.g., "ss/c/#")
//
// Thread Safety: MessangerLocal is safe for concurrent use via embedded mutex.
//
// Example:
//
//	msg := NewMessangerLocal("local")
//	msg.Subscribe("ss/c/+/temp", func(m *Msg) error {
//	    fmt.Printf("Temperature: %s\n", m.String())
//	    return nil
//	})
//	msg.Pub("ss/c/station1/temp", "25.5")
type MessangerLocal struct {
	*MessangerBase
	sync.Mutex `json:"-"`
}

// NewMessangerLocal creates and initializes a new local in-process messanger.
// The messanger is immediately ready for use without requiring a Connect() call.
//
// Parameters:
//   - id: Unique identifier for this messanger instance
//
// Returns a pointer to the initialized MessangerLocal.
//
// Example:
//
//	msg := NewMessangerLocal("test-local")
func NewMessangerLocal(id string) *MessangerLocal {
	m := &MessangerLocal{
		MessangerBase: NewMessangerBase(id),
	}

	return m
}

// ID returns the unique identifier for this local messanger instance.
func (m *MessangerLocal) ID() string {
	return m.MessangerBase.ID()
}

// Connect is a no-op for local messangers since no external connection is needed.
// It exists to satisfy the Messanger interface.
//
// Always returns nil (success).
func (m *MessangerLocal) Connect() error {
	return nil
}

// Subscribe registers a message handler for the given topic pattern in the local
// routing tree. The topic can include MQTT-style wildcards for flexible matching.
//
// The handler is inserted into the topic routing tree and stored in the base
// subscription map for tracking.
//
// Wildcard Examples:
//   - "ss/c/+/temp" matches "ss/c/station1/temp", "ss/c/station2/temp", etc.
//   - "ss/c/#" matches all topics starting with "ss/c/"
//   - "ss/c/station/+" matches any single-level child of "ss/c/station/"
//
// Parameters:
//   - topic: The topic pattern to subscribe to (with optional wildcards)
//   - handler: The callback function to invoke when matching messages arrive
//
// Returns nil on success, or an error if subscription fails.
func (m *MessangerLocal) Subscribe(topic string, handler MsgHandler) error {
	root.insert(topic, handler)
	return m.MessangerBase.Subscribe(topic, handler)
}

// Pub publishes a value to the specified topic through the local routing tree.
// The value is converted to bytes and wrapped in a Msg structure before delivery.
//
// Supported value types: []byte, string, int, bool, float64
//
// Parameters:
//   - topic: The complete topic path to publish to (e.g., "ss/c/station/led")
//   - value: The data to publish (will be converted to []byte)
//
// Returns an error if:
//   - topic is empty
//   - no subscribers exist
//   - value cannot be converted to bytes
//   - no matching handlers are found in the routing tree
//
// Example:
//
//	msg := NewMessangerLocal("local")
//	msg.Subscribe("ss/c/+/led", ledHandler)
//	msg.Pub("ss/c/station1/led", "on")  // Delivered to ledHandler
func (m *MessangerLocal) Pub(topic string, value any) error {
	m.Published++

	if topic == "" {
		return fmt.Errorf("no topic")
	}
	if len(m.subs) == 0 {
		return fmt.Errorf("no subscribers")
	}

	b, err := Bytes(value)
	if err != nil {
		return err
	}

	msg := NewMsg(topic, b, m.id)
	return m.PubMsg(msg)
}

// PubMsg publishes a pre-constructed message through the local routing tree.
// The message is looked up in the routing tree and delivered to all matching handlers.
//
// Parameters:
//   - msg: The message to publish (must not be nil)
//
// Returns an error if:
//   - msg is nil
//   - no handlers match the message topic
//
// Example:
//
//	msg := NewMsg("ss/c/station1/led", []byte("on"), "controller")
//	messanger.PubMsg(msg)
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

// Error returns the last error encountered by this messanger, if any.
// Returns nil if no error has occurred.
func (m *MessangerLocal) Error() error {
	return m.MessangerBase.Error()
}

// Close cleanly shuts down the local messanger by removing all subscriptions
// from the routing tree and clearing local state.
//
// This method is thread-safe via the embedded mutex.
//
// After Close() is called, the messanger should not be used for publishing
// or subscribing.
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
