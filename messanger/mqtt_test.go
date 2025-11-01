package messanger

import (
	"errors"
	"os"
	"testing"
	"time"

	gomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

func TestMockClient_Connect(t *testing.T) {
	mock := NewMockClient()

	// Test successful connection
	token := mock.Connect()
	if !token.Wait() {
		t.Error("expected token.Wait() to return true")
	}
	if token.Error() != nil {
		t.Errorf("expected no error, got %v", token.Error())
	}
	if !mock.IsConnected() {
		t.Error("expected client to be connected")
	}

	// Test connection error
	mock.Reset()
	mock.SetConnectError(errors.New("connection failed"))
	token = mock.Connect()
	if token.Wait() {
		t.Error("expected token.Wait() to return false on error")
	}
	if token.Error() == nil {
		t.Error("expected error from token")
	}
	if mock.IsConnected() {
		t.Error("expected client to not be connected on error")
	}
}

func TestMockClient_Publish(t *testing.T) {
	mock := NewMockClient()

	// Test successful publish
	token := mock.Publish("test/topic", 0, false, "test message")
	if !token.Wait() {
		t.Error("expected token.Wait() to return true")
	}
	if token.Error() != nil {
		t.Errorf("expected no error, got %v", token.Error())
	}

	// Check publication was recorded
	pubs := mock.GetPublications()
	if len(pubs) != 1 {
		t.Fatalf("expected 1 publication, got %d", len(pubs))
	}
	if pubs[0].Topic != "test/topic" {
		t.Errorf("expected topic 'test/topic', got %s", pubs[0].Topic)
	}
	if pubs[0].Payload != "test message" {
		t.Errorf("expected payload 'test message', got %v", pubs[0].Payload)
	}

	// Test publish error
	mock.SetPublishError(errors.New("publish failed"))
	token = mock.Publish("test/topic2", 0, false, "test message 2")
	if token.Error() == nil {
		t.Error("expected error from token")
	}
}

func TestMockClient_Subscribe(t *testing.T) {
	mock := NewMockClient()

	var receivedTopic string
	var receivedPayload []byte
	handler := func(client gomqtt.Client, msg gomqtt.Message) {
		receivedTopic = msg.Topic()
		receivedPayload = msg.Payload()
	}

	// Test successful subscription
	token := mock.Subscribe("test/topic", 0, handler)
	if !token.Wait() {
		t.Error("expected token.Wait() to return true")
	}
	if token.Error() != nil {
		t.Errorf("expected no error, got %v", token.Error())
	}

	// Check subscription was recorded
	if !mock.HasSubscription("test/topic") {
		t.Error("expected subscription to be recorded")
	}

	// Test message simulation
	err := mock.SimulateMessage("test/topic", []byte("hello world"))
	if err != nil {
		t.Errorf("expected no error from SimulateMessage, got %v", err)
	}

	// Give handler time to execute
	time.Sleep(10 * time.Millisecond)

	if receivedTopic != "test/topic" {
		t.Errorf("expected received topic 'test/topic', got %s", receivedTopic)
	}
	if string(receivedPayload) != "hello world" {
		t.Errorf("expected received payload 'hello world', got %s", string(receivedPayload))
	}

	// Test subscription error
	mock.Reset()
	mock.SetSubscribeError(errors.New("subscribe failed"))
	token = mock.Subscribe("test/topic", 0, handler)
	if token.Error() == nil {
		t.Error("expected error from token")
	}
}

func TestMockClient_Unsubscribe(t *testing.T) {
	mock := NewMockClient()

	// Subscribe first
	handler := func(client gomqtt.Client, msg gomqtt.Message) {}
	mock.Subscribe("test/topic", 0, handler)

	if !mock.HasSubscription("test/topic") {
		t.Error("expected subscription to exist")
	}

	// Unsubscribe
	token := mock.Unsubscribe("test/topic")
	if !token.Wait() {
		t.Error("expected token.Wait() to return true")
	}
	if token.Error() != nil {
		t.Errorf("expected no error, got %v", token.Error())
	}

	if mock.HasSubscription("test/topic") {
		t.Error("expected subscription to be removed")
	}
}

func TestMockClient_Disconnect(t *testing.T) {
	mock := NewMockClient()

	// Connect first
	mock.Connect()
	if !mock.IsConnected() {
		t.Error("expected client to be connected")
	}

	// Disconnect
	mock.Disconnect(250)
	if mock.IsConnected() {
		t.Error("expected client to be disconnected")
	}
	if !mock.IsDisconnectCalled() {
		t.Error("expected disconnect to be called")
	}
}

func TestMockMessage(t *testing.T) {
	msg := NewMockMessage("test/topic", []byte("test payload"))

	if msg.Topic() != "test/topic" {
		t.Errorf("expected topic 'test/topic', got %s", msg.Topic())
	}
	if string(msg.Payload()) != "test payload" {
		t.Errorf("expected payload 'test payload', got %s", string(msg.Payload()))
	}
	if msg.Qos() != 0 {
		t.Errorf("expected QoS 0, got %d", msg.Qos())
	}
	if msg.Retained() {
		t.Error("expected retained to be false")
	}
	if msg.Duplicate() {
		t.Error("expected duplicate to be false")
	}

	// Test ack
	msg.Ack()
	if !msg.acked {
		t.Error("expected message to be acked")
	}
}

func TestMockToken(t *testing.T) {
	// Test successful token
	token := NewMockToken(nil)
	if !token.Wait() {
		t.Error("expected Wait() to return true for successful token")
	}
	if !token.WaitTimeout(time.Second) {
		t.Error("expected WaitTimeout() to return true for successful token")
	}
	if token.Error() != nil {
		t.Error("expected no error for successful token")
	}

	// Test error token
	err := errors.New("test error")
	token = NewMockToken(err)
	if token.Wait() {
		t.Error("expected Wait() to return false for error token")
	}
	if token.WaitTimeout(time.Second) {
		t.Error("expected WaitTimeout() to return false for error token")
	}
	if token.Error() != err {
		t.Errorf("expected error %v, got %v", err, token.Error())
	}

	// Test Done channel
	select {
	case <-token.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("expected Done() channel to be closed")
	}
}

func TestMQTT_ConnectWithMock(t *testing.T) {
	// Test that the fixed Connect method works properly
	mock := NewMockClient()
	mqtt := NewMQTT("test-client", "localhost", "test/topic")
	mqtt.Client = mock

	// Test successful connection
	err := mqtt.Connect()
	if err != nil {
		t.Errorf("expected no error from Connect, got %v", err)
	}
	if !mqtt.IsConnected() {
		t.Error("expected MQTT to report connected")
	}
	if mqtt.Broker != "localhost" {
		t.Errorf("expected broker 'localhost', got %s", mqtt.Broker)
	}

	// Test connection with environment variable
	os.Setenv("MQTT_BROKER", "test-broker")
	defer os.Unsetenv("MQTT_BROKER")

	mqtt2 := NewMQTT("test-client-2", "localhost", "test/topic")
	mqtt2.Client = mock
	err = mqtt2.Connect()
	if err != nil {
		t.Errorf("expected no error from Connect with env broker, got %v", err)
	}
	if mqtt2.Broker != "test-broker" {
		t.Errorf("expected broker 'test-broker', got %s", mqtt2.Broker)
	}

	// Test connection error - need to reset mock first
	mock.Reset()
	mock.SetConnectError(errors.New("connection failed"))
	mqtt3 := NewMQTT("test-client-3", "localhost", "test/topic")
	mqtt3.Client = mock
	err = mqtt3.Connect()
	if err == nil {
		t.Error("expected error from Connect when mock has error")
	}
}

func TestSetMQTTClient_Fixed(t *testing.T) {
	// Test that SetMQTTClient works with global state properly
	mock := NewMockClient()

	// SetMQTTClient should work even when global mqtt is nil
	mqttClient := SetMQTTClient(mock)
	if mqttClient == nil {
		t.Fatal("SetMQTTClient returned nil")
	}
	if mqttClient.Client != mock {
		t.Error("SetMQTTClient did not set the mock client")
	}
	if mqttClient.ID() != "default" {
		t.Errorf("expected default ID 'default', got %s", mqttClient.ID())
	}
}

func TestGetMQTT_Fixed(t *testing.T) {
	// Reset global state
	mqtt = nil

	// GetMQTT should initialize if nil
	mqttClient := GetMQTT()
	if mqttClient == nil {
		t.Fatal("GetMQTT returned nil")
	}
	if mqttClient.ID() != "default" {
		t.Errorf("expected default ID 'default', got %s", mqttClient.ID())
	}
	if mqttClient.Broker != "localhost" {
		t.Errorf("expected default broker 'localhost', got %s", mqttClient.Broker)
	}

	// Second call should return same instance
	mqttClient2 := GetMQTT()
	if mqttClient != mqttClient2 {
		t.Error("GetMQTT should return same instance on subsequent calls")
	}
}

func TestMockClient_Integration(t *testing.T) {
	// Test the mock with direct instantiation to avoid global state issues
	mock := NewMockClient()

	// Create MQTT instance and inject mock directly
	mqttInstance := &MQTT{
		id:     "test-client",
		Broker: "localhost",
		Client: mock,
	}

	// Test IsConnected
	if mqttInstance.IsConnected() {
		t.Error("expected client to not be connected initially")
	}

	// Manually set connected state in mock for testing
	mock.connected = true
	if !mqttInstance.IsConnected() {
		t.Error("expected client to be connected after setting mock state")
	}

	// Test publish
	mqttInstance.Publish("test/topic", "test message")
	pubs := mock.GetPublications()
	if len(pubs) != 1 {
		t.Fatalf("expected 1 publication, got %d", len(pubs))
	}
	if pubs[0].Topic != "test/topic" {
		t.Errorf("expected topic 'test/topic', got %s", pubs[0].Topic)
	}

	// Test subscribe with message handling
	var receivedMsg *Msg
	err := mqttInstance.Subscribe("test/incoming", func(msg *Msg) error {
		receivedMsg = msg
		return nil
	})
	if err != nil {
		t.Errorf("expected no error from Subscribe, got %v", err)
	}

	// Simulate incoming message
	err = mock.SimulateMessage("test/incoming", []byte("incoming data"))
	if err != nil {
		t.Errorf("expected no error from SimulateMessage, got %v", err)
	}

	// Give handler time to execute
	time.Sleep(10 * time.Millisecond)

	if receivedMsg == nil {
		t.Fatal("expected to receive a message")
	}
	if receivedMsg.Topic != "test/incoming" {
		t.Errorf("expected topic 'test/incoming', got %s", receivedMsg.Topic)
	}
	if string(receivedMsg.Data) != "incoming data" {
		t.Errorf("expected data 'incoming data', got %s", string(receivedMsg.Data))
	}

	// Test close
	mqttInstance.Close()
	if !mock.IsDisconnectCalled() {
		t.Error("expected disconnect to be called")
	}
}

func TestMQTT_Error(t *testing.T) {
	// Test the Error() method which currently has 0% coverage
	mqtt := NewMQTT("test-client", "localhost", "test/topic")

	// Initially should have no error
	if mqtt.Error() != nil {
		t.Error("expected no error initially")
	}

	// Set an error via the error field
	testErr := errors.New("test error")
	mqtt.error = testErr

	if mqtt.Error() != testErr {
		t.Errorf("expected error %v, got %v", testErr, mqtt.Error())
	}
}

func TestMQTTPublishEdgeCases(t *testing.T) {
	// Test Publish method edge cases to improve coverage from 42.9%
	mock := NewMockClient()
	mqtt := NewMQTT("test-client", "localhost", "test/topic")
	mqtt.Client = mock

	err := mqtt.Publish("", "test")
	assert.Error(t, err, "Expected error when publishing to MQTT with no topic")
}

func TestMQTT_PublishWithNilClient(t *testing.T) {
	// Test publish with nil client (should log warning and return)
	mqtt := NewMQTT("test-client", "localhost", "test/topic")
	// Don't set Client, leaving it nil

	// This should not panic, just log a warning
	mqtt.Publish("test/topic", "test message")

	// Test that no error is set when client is nil
	if mqtt.Error() != nil {
		t.Errorf("expected no error when client is nil, got %v", mqtt.Error())
	}
}

func TestMQTT_PublishWithNullToken(t *testing.T) {
	// Test the null token case in Publish method
	mock := NewMockClient()
	mqtt := NewMQTT("test-client", "localhost", "test/topic")
	mqtt.Client = mock

	// This case is hard to test directly since the mock always returns a token
	// But we can test the token error case
	mock.SetPublishError(errors.New("publish error"))
	mqtt.Publish("test/topic", "test message")

	// Verify error was set
	if mqtt.Error() == nil {
		t.Error("expected error to be set after publish failure")
	}
}

func TestMQTT_SubscribeEdgeCases(t *testing.T) {
	// Test Subscribe method edge cases to improve coverage from 66.7%
	mqtt := NewMQTT("test-client", "localhost", "test/topic")

	// Test subscribe with nil client
	err := mqtt.Subscribe("test/topic", func(msg *Msg) error { return nil })
	if err == nil {
		t.Error("expected error when subscribing with nil client")
	}

	// Test with mock client and subscription error
	mock := NewMockClient()
	mock.SetSubscribeError(errors.New("subscribe failed"))
	mqtt.Client = mock

	err = mqtt.Subscribe("test/topic", func(msg *Msg) error { return nil })
	if err == nil {
		t.Error("expected error from Subscribe when mock has error")
	}

	// Verify error was set in MQTT instance
	if mqtt.Error() == nil {
		t.Error("expected MQTT instance error to be set")
	}
}

func TestMQTT_IsConnectedEdgeCases(t *testing.T) {
	// Test IsConnected edge cases to improve coverage from 66.7%
	mqtt := NewMQTT("test-client", "localhost", "test/topic")

	// Test with nil client
	if mqtt.IsConnected() {
		t.Error("expected false when client is nil")
	}

	// Test with mock client
	mock := NewMockClient()
	mqtt.Client = mock

	// Initially not connected
	if mqtt.IsConnected() {
		t.Error("expected false when mock client not connected")
	}

	// Set connected and test
	mock.connected = true
	if !mqtt.IsConnected() {
		t.Error("expected true when mock client is connected")
	}
}

func TestMessangerMQTT_NewAndID(t *testing.T) {
	// Test MessangerMQTT creation and ID method (currently 0% coverage)
	mqttMessanger, err := NewMessangerMQTT("mqtt-test", "localhost")
	assert.NoError(t, err)

	if mqttMessanger == nil {
		t.Fatal("NewMessangerMQTT returned nil")
	}

	if mqttMessanger.ID() != "mqtt-test" {
		t.Errorf("expected ID 'mqtt-test', got %s", mqttMessanger.ID())
	}

	// Verify it has both MessangerBase and MQTT
	if mqttMessanger.MessangerBase == nil {
		t.Error("MessangerBase should not be nil")
	}
	if mqttMessanger.MQTT == nil {
		t.Error("MQTT should not be nil")
	}
}

func TestMessangerMQTT_Subscribe(t *testing.T) {
	// Test Subscribe method (currently 0% coverage)
	mock := NewMockClient()
	mqttMessanger, err := NewMessangerMQTT("mqtt-test", "localhost")
	assert.NoError(t, err)
	mqttMessanger.MQTT.Client = mock

	var receivedMsg *Msg
	handler := func(msg *Msg) error {
		receivedMsg = msg
		return nil
	}

	err = mqttMessanger.Subscribe("test/topic", handler)
	if err != nil {
		t.Errorf("expected no error from Subscribe, got %v", err)
	}

	// Verify subscription was registered
	if !mock.HasSubscription("test/topic") {
		t.Error("expected subscription to be registered in mock")
	}

	// Test message handling
	err = mock.SimulateMessage("test/topic", []byte("test data"))
	if err != nil {
		t.Errorf("expected no error from SimulateMessage, got %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Give handler time to execute

	if receivedMsg == nil {
		t.Fatal("expected to receive a message")
	}
	if receivedMsg.Topic != "test/topic" {
		t.Errorf("expected topic 'test/topic', got %s", receivedMsg.Topic)
	}
}

func TestMessangerMQTT_Pub(t *testing.T) {
	// Test Pub method (currently 0% coverage)
	mock := NewMockClient()
	mqttMessanger, err := NewMessangerMQTT("mqtt-test", "localhost")
	assert.NoError(t, err)
	mqttMessanger.MQTT.Client = mock

	mqttMessanger.Pub("test/topic", "test message")

	// Verify publish was called
	pubs := mock.GetPublications()
	if len(pubs) != 1 {
		t.Fatalf("expected 1 publication, got %d", len(pubs))
	}
	if pubs[0].Topic != "test/topic" {
		t.Errorf("expected topic 'test/topic', got %s", pubs[0].Topic)
	}
	if pubs[0].Payload != "test message" {
		t.Errorf("expected payload 'test message', got %v", pubs[0].Payload)
	}

	// Verify Published counter was incremented
	if mqttMessanger.Published != 1 {
		t.Errorf("expected Published count 1, got %d", mqttMessanger.Published)
	}
}

func TestMessangerMQTT_PubMsg(t *testing.T) {
	// Test PubMsg method (currently 0% coverage)
	mock := NewMockClient()
	mqttMessanger, err := NewMessangerMQTT("mqtt-test", "localhost")
	assert.NoError(t, err)
	mqttMessanger.MQTT.Client = mock

	msg := NewMsg("test/topic", []byte("test data"), "test-source")
	mqttMessanger.PubMsg(msg)

	// Verify publish was called with correct data
	pubs := mock.GetPublications()
	if len(pubs) != 1 {
		t.Fatalf("expected 1 publication, got %d", len(pubs))
	}
	if pubs[0].Topic != "test/topic" {
		t.Errorf("expected topic 'test/topic', got %s", pubs[0].Topic)
	}
	// The payload should be the Data field
	if string(pubs[0].Payload.([]byte)) != "test data" {
		t.Errorf("expected payload 'test data', got %v", pubs[0].Payload)
	}
}

func TestMessangerMQTT_Error(t *testing.T) {
	// Test Error method (currently 0% coverage)
	mock := NewMockClient()
	mqttMessanger, err := NewMessangerMQTT("mqtt-test", "localhost")
	assert.NoError(t, err)
	mqttMessanger.MQTT.Client = mock

	// Initially no error
	if mqttMessanger.Error() != nil {
		t.Error("expected no error initially")
	}

	// Set error in MQTT and verify
	testErr := errors.New("mqtt error")
	mqttMessanger.MQTT.error = testErr

	if mqttMessanger.Error() != testErr {
		t.Errorf("expected error %v, got %v", testErr, mqttMessanger.Error())
	}
}

func TestMsgPrinter_MsgHandler(t *testing.T) {
	// Test MsgPrinter.MsgHandler method (currently 0% coverage)
	printer := &MsgPrinter{}
	msg := NewMsg("test/topic", []byte("test data"), "test-source")

	// This method just prints, so we can't easily test output
	// But we can ensure it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MsgHandler panicked: %v", r)
		}
	}()

	printer.MsgHandler(msg)
}

func TestMockClient_UtilityFunctions(t *testing.T) {
	// Test uncovered mock utility functions
	mock := NewMockClient()

	// Test GetSubscriptions
	handler := func(client gomqtt.Client, msg gomqtt.Message) {}
	mock.Subscribe("topic1", 0, handler)
	mock.Subscribe("topic2", 1, handler)

	subs := mock.GetSubscriptions()
	if len(subs) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subs))
	}
	if _, exists := subs["topic1"]; !exists {
		t.Error("expected topic1 in subscriptions")
	}
	if subs["topic1"].QoS != 0 {
		t.Errorf("expected QoS 0 for topic1, got %d", subs["topic1"].QoS)
	}

	// Test ClearPublications
	mock.Publish("test", 0, false, "data")
	if len(mock.GetPublications()) == 0 {
		t.Error("expected publications before clear")
	}

	mock.ClearPublications()
	if len(mock.GetPublications()) != 0 {
		t.Error("expected no publications after clear")
	}

	// Test GetLastPublication
	lastPub := mock.GetLastPublication()
	if lastPub != nil {
		t.Error("expected nil for last publication when none exist")
	}

	mock.Publish("last", 0, false, "lastdata")
	lastPub = mock.GetLastPublication()
	if lastPub == nil {
		t.Fatal("expected last publication to exist")
	}
	if lastPub.Topic != "last" {
		t.Errorf("expected last publication topic 'last', got %s", lastPub.Topic)
	}

	// Test SubscribeMultiple
	filters := map[string]byte{
		"multi1": 0,
		"multi2": 1,
	}
	token := mock.SubscribeMultiple(filters, handler)
	if !token.Wait() {
		t.Error("expected SubscribeMultiple token to succeed")
	}

	// Verify subscriptions were added
	if !mock.HasSubscription("multi1") {
		t.Error("expected multi1 subscription")
	}
	if !mock.HasSubscription("multi2") {
		t.Error("expected multi2 subscription")
	}

	// Test IsConnectionOpen
	mock.connected = false
	if mock.IsConnectionOpen() {
		t.Error("expected IsConnectionOpen to return false when not connected")
	}

	mock.connected = true
	if !mock.IsConnectionOpen() {
		t.Error("expected IsConnectionOpen to return true when connected")
	}

	// Test AddRoute and RemoveRoute
	mock.AddRoute("route1", handler)
	// No direct way to verify route was added, but ensure no panic

	mock.RemoveRoute("route1")
	// No direct way to verify route was removed, but ensure no panic

	// Test SetOrderMatters (no-op)
	mock.SetOrderMatters(true)
	mock.SetOrderMatters(false)
}

func TestMockClient_SimulateConnectionLost(t *testing.T) {
	// Test SimulateConnectionLost function
	mock := NewMockClient()
	mock.connected = true

	var lostCalled bool
	var lostErr error

	// Set a connection lost handler (this would typically be done via client options)
	mock.connectionLostHandler = func(client gomqtt.Client, err error) {
		lostCalled = true
		lostErr = err
	}

	testErr := errors.New("connection lost")
	mock.SimulateConnectionLost(testErr)

	// Give handler time to execute
	time.Sleep(10 * time.Millisecond)

	if !lostCalled {
		t.Error("expected connection lost handler to be called")
	}
	if lostErr != testErr {
		t.Errorf("expected error %v, got %v", testErr, lostErr)
	}
	if mock.IsConnected() {
		t.Error("expected client to be disconnected after connection lost")
	}
}

func TestMockClient_SimulateMessageRoutes(t *testing.T) {
	// Test SimulateMessage with routes instead of subscriptions
	mock := NewMockClient()

	var receivedTopic string
	var receivedPayload []byte

	handler := func(client gomqtt.Client, msg gomqtt.Message) {
		receivedTopic = msg.Topic()
		receivedPayload = msg.Payload()
	}

	// Add route instead of subscription
	mock.AddRoute("route/topic", handler)

	err := mock.SimulateMessage("route/topic", []byte("route data"))
	if err != nil {
		t.Errorf("expected no error from SimulateMessage with route, got %v", err)
	}

	// Give handler time to execute
	time.Sleep(10 * time.Millisecond)

	if receivedTopic != "route/topic" {
		t.Errorf("expected topic 'route/topic', got %s", receivedTopic)
	}
	if string(receivedPayload) != "route data" {
		t.Errorf("expected payload 'route data', got %s", string(receivedPayload))
	}
}

func TestMockClient_ErrorScenarios(t *testing.T) {
	// Test various error scenarios in mock
	mock := NewMockClient()

	// Test SimulateMessage with no handler
	err := mock.SimulateMessage("nonexistent", []byte("data"))
	if err == nil {
		t.Error("expected error when simulating message with no handler")
	}

	// Test subscribe with error
	mock.SetSubscribeError(errors.New("sub error"))
	token := mock.Subscribe("topic", 0, func(client gomqtt.Client, msg gomqtt.Message) {})
	if token.Error() == nil {
		t.Error("expected error from subscribe token")
	}

	// Test SubscribeMultiple with error
	filters := map[string]byte{"topic": 0}
	token = mock.SubscribeMultiple(filters, func(client gomqtt.Client, msg gomqtt.Message) {})
	if token.Error() == nil {
		t.Error("expected error from SubscribeMultiple token")
	}

	// Test unsubscribe with error
	mock.unsubscribeErr = errors.New("unsub error")
	token = mock.Unsubscribe("topic")
	if token.Error() == nil {
		t.Error("expected error from unsubscribe token")
	}
}

func TestMockClient_OptionsReader(t *testing.T) {
	// Test OptionsReader panic behavior
	mock := NewMockClient()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected OptionsReader to panic")
		}
	}()

	mock.OptionsReader()
}

func TestMockMessage_CompleteInterface(t *testing.T) {
	// Test all MockMessage methods
	msg := NewMockMessage("test/topic", []byte("test payload"))

	// Test MessageID (currently 0% coverage)
	if msg.MessageID() != 0 {
		t.Errorf("expected MessageID 0, got %d", msg.MessageID())
	}

	// Test setting various fields
	msg.qos = 2
	msg.retained = true
	msg.duplicate = true
	msg.messageID = 123

	if msg.Qos() != 2 {
		t.Errorf("expected QoS 2, got %d", msg.Qos())
	}
	if !msg.Retained() {
		t.Error("expected retained to be true")
	}
	if !msg.Duplicate() {
		t.Error("expected duplicate to be true")
	}
	if msg.MessageID() != 123 {
		t.Errorf("expected MessageID 123, got %d", msg.MessageID())
	}
}

func TestMockClientOptionsReader(t *testing.T) {
	// Test MockClientOptionsReader methods (currently 0% coverage)
	reader := &MockClientOptionsReader{clientID: "test-client"}

	if reader.ClientID() != "test-client" {
		t.Errorf("expected ClientID 'test-client', got %s", reader.ClientID())
	}
}
