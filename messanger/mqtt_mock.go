// mqtt_mock.go
// Mock implementation of the paho.mqtt.golang Client interface for testing.
// Provides comprehensive mocking of MQTT operations including connection,
// publishing, subscribing, and message handling.

package messanger

import (
	"errors"
	"sync"
	"time"

	gomqtt "github.com/eclipse/paho.mqtt.golang"
)

// MockToken implements gomqtt.Token interface
type MockToken struct {
	err    error
	waited bool
	mu     sync.Mutex
}

func NewMockToken(err error) *MockToken {
	return &MockToken{err: err}
}

func (t *MockToken) Wait() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.waited = true
	return t.err == nil
}

func (t *MockToken) WaitTimeout(timeout time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.waited = true
	return t.err == nil
}

func (t *MockToken) Error() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.err
}

func (t *MockToken) Done() <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}

// MockMessage implements gomqtt.Message interface
type MockMessage struct {
	topic     string
	payload   []byte
	qos       byte
	retained  bool
	duplicate bool
	messageID uint16
	acked     bool
}

func NewMockMessage(topic string, payload []byte) *MockMessage {
	return &MockMessage{
		topic:   topic,
		payload: payload,
		qos:     0,
	}
}

func (m *MockMessage) Duplicate() bool   { return m.duplicate }
func (m *MockMessage) Qos() byte         { return m.qos }
func (m *MockMessage) Retained() bool    { return m.retained }
func (m *MockMessage) Topic() string     { return m.topic }
func (m *MockMessage) MessageID() uint16 { return m.messageID }
func (m *MockMessage) Payload() []byte   { return m.payload }
func (m *MockMessage) Ack()              { m.acked = true }

// MockClientOptionsReader is a simple stub for ClientOptionsReader
type MockClientOptionsReader struct {
	clientID string
}

func (m *MockClientOptionsReader) ClientID() string { return m.clientID }

// Publication represents a published message for testing
type Publication struct {
	Topic    string
	Payload  interface{}
	QoS      byte
	Retained bool
}

// Subscription represents a subscription for testing
type Subscription struct {
	Topic   string
	QoS     byte
	Handler gomqtt.MessageHandler
}

// MockClient implements gomqtt.Client interface
// Provides comprehensive mocking capabilities for MQTT operations
type MockClient struct {
	mu         sync.RWMutex
	connected  bool
	connecting bool
	options    *MockClientOptionsReader

	// Error simulation
	connectErr     error
	publishErr     error
	subscribeErr   error
	unsubscribeErr error

	// Call tracking
	publications  []Publication
	subscriptions map[string]Subscription
	routes        map[string]gomqtt.MessageHandler

	// Connection tracking
	disconnectCalled  bool
	disconnectQuiesce uint

	// Callbacks
	onConnectHandler      gomqtt.OnConnectHandler
	connectionLostHandler gomqtt.ConnectionLostHandler

	LastTopic   string
	LastMessage string
}

// NewMockClient creates a new mock MQTT client
func NewMockClient() *MockClient {
	return &MockClient{
		options:       &MockClientOptionsReader{},
		subscriptions: make(map[string]Subscription),
		routes:        make(map[string]gomqtt.MessageHandler),
	}
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
