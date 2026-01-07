// mqtt_mock.go provides a comprehensive mock implementation of the Eclipse Paho
// MQTT client interface for testing purposes.
//
// The mock client simulates MQTT broker behavior without requiring an actual
// network connection or broker. This enables:
//   - Fast, isolated unit tests
//   - Deterministic test behavior
//   - Testing error conditions and edge cases
//   - CI/CD environments without external dependencies
//
// Key Features:
//   - Track all published messages for assertions
//   - Simulate connection, subscription, and publish operations
//   - Inject errors to test error handling
//   - Manually trigger message delivery to handlers
//   - Thread-safe for concurrent test scenarios
//
// Example:
//
//	mock := NewMockClient()
//	mqtt := NewMQTT("test", "mock", "")
//	mqtt.SetMQTTClient(mock)
//
//	// Subscribe and trigger message
//	mqtt.Subscribe("test/topic", handler)
//	mock.SimulateMessage("test/topic", []byte("test data"))
//
//	// Verify publications
//	pubs := mock.GetPublications()
//	assert.Equal(t, 1, len(pubs))
package messanger

import (
	"errors"
	"sync"
	"time"

	gomqtt "github.com/eclipse/paho.mqtt.golang"
)

// MockToken implements the gomqtt.Token interface for testing.
// It represents the result of an asynchronous MQTT operation (connect, publish, subscribe).
// The token tracks whether the operation completed and any error that occurred.
type MockToken struct {
	err    error      // Error from the operation, if any
	waited bool       // Whether Wait() has been called
	mu     sync.Mutex // Protects concurrent access
}

// NewMockToken creates a new mock token with the specified error state.
// Pass nil for a successful operation token.
//
// Parameters:
//   - err: The error to return, or nil for success
//
// Returns a pointer to the mock token.
func NewMockToken(err error) *MockToken {
	return &MockToken{err: err}
}

// Wait blocks until the operation completes (immediately for mock).
// Returns true if the operation succeeded, false if it failed.
func (t *MockToken) Wait() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.waited = true
	return t.err == nil
}

// WaitTimeout waits for the operation to complete with a timeout.
// For the mock, this behaves the same as Wait() and returns immediately.
//
// Parameters:
//   - timeout: Maximum time to wait (ignored in mock)
//
// Returns true if the operation succeeded, false if it failed.
func (t *MockToken) WaitTimeout(timeout time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.waited = true
	return t.err == nil
}

// Error returns the error from the operation, if any.
// Returns nil if the operation succeeded.
func (t *MockToken) Error() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.err
}

// Done returns a channel that is closed when the operation completes.
// For the mock, the channel is already closed (operation is instant).
func (t *MockToken) Done() <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}

// MockMessage implements the gomqtt.Message interface for testing.
// It represents an MQTT message received by a subscriber.
type MockMessage struct {
	topic     string // The topic the message was published to
	payload   []byte // The message payload
	qos       byte   // Quality of Service level (0, 1, or 2)
	retained  bool   // Whether the message was retained by the broker
	duplicate bool   // Whether this is a duplicate delivery
	messageID uint16 // Unique message identifier
	acked     bool   // Whether the message has been acknowledged
}

// NewMockMessage creates a new mock MQTT message for testing.
// The message is created with QoS 0 and not retained.
//
// Parameters:
//   - topic: The topic the message is published to
//   - payload: The message payload as bytes
//
// Returns a pointer to the mock message.
func NewMockMessage(topic string, payload []byte) *MockMessage {
	return &MockMessage{
		topic:   topic,
		payload: payload,
		qos:     0,
	}
}

// Duplicate returns whether this is a duplicate message delivery.
func (m *MockMessage) Duplicate() bool { return m.duplicate }

// Qos returns the Quality of Service level for this message.
func (m *MockMessage) Qos() byte { return m.qos }

// Retained returns whether this message was retained by the broker.
func (m *MockMessage) Retained() bool { return m.retained }

// Topic returns the topic this message was published to.
func (m *MockMessage) Topic() string { return m.topic }

// MessageID returns the unique identifier for this message.
func (m *MockMessage) MessageID() uint16 { return m.messageID }

// Payload returns the message payload as bytes.
func (m *MockMessage) Payload() []byte { return m.payload }

// Ack marks the message as acknowledged.
func (m *MockMessage) Ack() { m.acked = true }

// MockClientOptionsReader is a simple stub implementation of the
// gomqtt.ClientOptionsReader interface for testing.
type MockClientOptionsReader struct {
	clientID string // The client ID
}

// ClientID returns the client identifier.
func (m *MockClientOptionsReader) ClientID() string { return m.clientID }

// Publication represents a message that was published through the mock client.
// Tests can inspect publications to verify that code published the right messages.
type Publication struct {
	Topic    string      // The topic published to
	Payload  interface{} // The message payload
	QoS      byte        // Quality of Service level
	Retained bool        // Whether the message should be retained
}

// Subscription represents a topic subscription made through the mock client.
// Tests can inspect subscriptions to verify that code subscribed correctly.
type Subscription struct {
	Topic   string                // The topic pattern subscribed to
	QoS     byte                  // Quality of Service level
	Handler gomqtt.MessageHandler // The callback function for messages
}

// MockClient implements the complete gomqtt.Client interface for testing.
// It simulates an MQTT client without requiring a real broker connection.
//
// The mock tracks all operations (connect, publish, subscribe) and allows
// tests to:
//   - Verify publications via GetPublications()
//   - Verify subscriptions via GetSubscriptions()
//   - Inject errors via SetConnectError(), SetPublishError(), etc.
//   - Simulate incoming messages via SimulateMessage()
//   - Control connection state
//
// Thread Safety: All methods are protected by an RWMutex for concurrent access.
//
// Example:
//
//	mock := NewMockClient()
//	mock.Connect()  // Always succeeds unless SetConnectError() was called
//	mock.Publish("test/topic", 0, false, "payload")
//	pubs := mock.GetPublications()
//	assert.Equal(t, 1, len(pubs))
type MockClient struct {
	mu         sync.RWMutex             // Protects all fields
	connected  bool                     // Current connection state
	connecting bool                     // Whether connection is in progress
	options    *MockClientOptionsReader // Client options

	// Error simulation - inject errors for testing error handling
	connectErr     error // Error to return from Connect()
	publishErr     error // Error to return from Publish()
	subscribeErr   error // Error to return from Subscribe()
	unsubscribeErr error // Error to return from Unsubscribe()

	// Call tracking - records all operations for test assertions
	publications  []Publication                    // All published messages
	subscriptions map[string]Subscription          // Active subscriptions by topic
	routes        map[string]gomqtt.MessageHandler // Message routes

	// Connection tracking
	disconnectCalled  bool // Whether Disconnect() was called
	disconnectQuiesce uint // Quiesce time passed to Disconnect()

	// Callbacks
	onConnectHandler      gomqtt.OnConnectHandler      // Called on connection
	connectionLostHandler gomqtt.ConnectionLostHandler // Called on disconnection

	LastTopic   string // Last topic published to (deprecated)
	LastMessage string // Last message published (deprecated)
}

// NewMockClient creates and initializes a new mock MQTT client.
// The client starts in a disconnected state with no subscriptions.
//
// Returns a pointer to the mock client.
//
// Example:
//
//	mock := NewMockClient()
//	mqtt := NewMQTT("test", "mock", "")
//	mqtt.SetMQTTClient(mock)
func NewMockClient() *MockClient {
	return &MockClient{
		options:       &MockClientOptionsReader{},
		subscriptions: make(map[string]Subscription),
		routes:        make(map[string]gomqtt.MessageHandler),
	}
}

// ID returns the identifier for this mock client.
// Always returns "mock" to distinguish it from real clients.
func (m *MockClient) ID() string {
	return "mock"
}

// Connection methods
func (m *MockClient) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

func (m *MockClient) IsConnectionOpen() bool {
	return m.IsConnected()
}

func (m *MockClient) Connect() gomqtt.Token {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connectErr != nil {
		return NewMockToken(m.connectErr)
	}

	m.connected = true
	m.connecting = false

	// Trigger connect callback if set
	if m.onConnectHandler != nil {
		go m.onConnectHandler(m)
	}

	return NewMockToken(nil)
}

func (m *MockClient) Disconnect(quiesce uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = false
	m.disconnectCalled = true
	m.disconnectQuiesce = quiesce
}

// Publishing methods
func (m *MockClient) Publish(topic string, qos byte, retained bool, payload interface{}) gomqtt.Token {
	m.mu.Lock()
	defer m.mu.Unlock()

	pub := Publication{
		Topic:    topic,
		Payload:  payload,
		QoS:      qos,
		Retained: retained,
	}
	m.publications = append(m.publications, pub)
	return NewMockToken(m.publishErr)
}

// Subscription methods
func (m *MockClient) Subscribe(topic string, qos byte, callback gomqtt.MessageHandler) gomqtt.Token {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.subscribeErr != nil {
		return NewMockToken(m.subscribeErr)
	}

	sub := Subscription{
		Topic:   topic,
		QoS:     qos,
		Handler: callback,
	}
	m.subscriptions[topic] = sub

	return NewMockToken(nil)
}

func (m *MockClient) SubscribeMultiple(filters map[string]byte, callback gomqtt.MessageHandler) gomqtt.Token {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.subscribeErr != nil {
		return NewMockToken(m.subscribeErr)
	}

	for topic, qos := range filters {
		sub := Subscription{
			Topic:   topic,
			QoS:     qos,
			Handler: callback,
		}
		m.subscriptions[topic] = sub
	}

	return NewMockToken(nil)
}

func (m *MockClient) Unsubscribe(topics ...string) gomqtt.Token {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, topic := range topics {
		delete(m.subscriptions, topic)
	}

	return NewMockToken(m.unsubscribeErr)
}

// Routing methods
func (m *MockClient) AddRoute(topic string, callback gomqtt.MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes[topic] = callback
}

func (m *MockClient) RemoveRoute(topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.routes, topic)
}

// Configuration methods
func (m *MockClient) OptionsReader() gomqtt.ClientOptionsReader {
	// For testing purposes, we'll panic if this is called
	// Most MQTT testing doesn't require options reader functionality
	panic("OptionsReader not implemented in mock - modify test if needed")
}

func (m *MockClient) SetOrderMatters(matter bool) {
	// No-op for mock
}

// --- Mock Control Methods ---

// SetConnectError configures the mock to return an error on Connect()
func (m *MockClient) SetConnectError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectErr = err
}

// SetPublishError configures the mock to return an error on Publish()
func (m *MockClient) SetPublishError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishErr = err
}

// SetSubscribeError configures the mock to return an error on Subscribe()
func (m *MockClient) SetSubscribeError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribeErr = err
}

// GetPublications returns all publications made through this mock client
func (m *MockClient) GetPublications() []Publication {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pubs := make([]Publication, len(m.publications))
	copy(pubs, m.publications)
	return pubs
}

// GetSubscriptions returns all active subscriptions
func (m *MockClient) GetSubscriptions() map[string]Subscription {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subs := make(map[string]Subscription)
	for k, v := range m.subscriptions {
		subs[k] = v
	}
	return subs
}

// ClearPublications clears the publication history
func (m *MockClient) ClearPublications() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publications = nil
}

// SimulateMessage triggers a message delivery to subscribed handlers
// This allows testing of message handling logic
func (m *MockClient) SimulateMessage(topic string, payload []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check subscriptions first
	if sub, ok := m.subscriptions[topic]; ok {
		msg := NewMockMessage(topic, payload)
		if sub.Handler != nil {
			go sub.Handler(m, msg)
			return nil
		}
	}

	// Check routes
	if handler, ok := m.routes[topic]; ok {
		msg := NewMockMessage(topic, payload)
		if handler != nil {
			go handler(m, msg)
			return nil
		}
	}

	return errors.New("no handler found for topic: " + topic)
}

// SimulateConnectionLost simulates a connection loss
func (m *MockClient) SimulateConnectionLost(err error) {
	m.mu.Lock()
	m.connected = false
	handler := m.connectionLostHandler
	m.mu.Unlock()

	if handler != nil {
		go handler(m, err)
	}
}

// IsDisconnectCalled returns true if Disconnect was called
func (m *MockClient) IsDisconnectCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.disconnectCalled
}

// GetLastPublication returns the most recent publication
func (m *MockClient) GetLastPublication() *Publication {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.publications) == 0 {
		return nil
	}
	return &m.publications[len(m.publications)-1]
}

// HasSubscription checks if a topic is subscribed
func (m *MockClient) HasSubscription(topic string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.subscriptions[topic]
	return exists
}

// Reset clears all mock state
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = false
	m.connecting = false
	m.disconnectCalled = false
	m.disconnectQuiesce = 0
	m.connectErr = nil
	m.publishErr = nil
	m.subscribeErr = nil
	m.unsubscribeErr = nil
	m.publications = nil
	m.subscriptions = make(map[string]Subscription)
	m.routes = make(map[string]gomqtt.MessageHandler)
}
