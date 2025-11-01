package messanger

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessangerLocal(t *testing.T) {
	m, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	if m.ID() != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", m.ID())
	}
}

// TestMessengerLocalPubWithInvalidData tests Pub with data that can't be serialized
func TestMessengerLocalPubWithInvalidData(t *testing.T) {
	defer resetNodes()

	// Create local messenger
	messenger, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	// Try to publish data that can't be serialized (unsupported type)
	type customStruct struct {
		Value string
	}
	data := customStruct{Value: "test"}

	// This should return an error due to unsupported type
	err = messenger.Pub("test/topic", data)
	assert.Error(t, err, "expected error for unsupported data type")
}

func TestMessangerLocalSubscribe(t *testing.T) {
	resetNodes() // Reset the node tree
	m, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	tests := []struct {
		name    string
		topic   string
		wantErr bool
	}{
		{"valid topic", "test/topic", false},
		{"root topic", "/", false},
		{"multi-level topic", "a/b/c/d", false},
		{"wildcard topic", "test/+/topic", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			handler := func(msg *Msg) error {
				handlerCalled = true
				return nil
			}

			err := m.Subscribe(tt.topic, handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Test publishing to the topic
				msg := NewMsg(tt.topic, []byte("test data"), "test-id")
				m.PubMsg(msg)

				if !handlerCalled {
					t.Error("Handler was not called after publishing to topic")
				}
			}
		})
	}
}

func TestMessangerLocalPub(t *testing.T) {
	resetNodes()
	m, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	tests := []struct {
		name     string
		topic    string
		data     interface{}
		wantData string
	}{
		{"string data", "test/topic", "hello", "hello"},
		{"int data", "test/topic", 42, "42"},
		{"bool data", "test/topic", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedData := ""
			handler := func(msg *Msg) error {
				receivedData = string(msg.Data)
				return nil
			}

			m.Subscribe(tt.topic, handler)
			m.Pub(tt.topic, tt.data)

			if receivedData != tt.wantData {
				t.Errorf("Pub() got = %v, want %v", receivedData, tt.wantData)
			}
		})
	}
}

func TestMessangerLocalPubMsg(t *testing.T) {
	resetNodes()
	m, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	handlerCalled := false
	expectedData := "test data"

	handler := func(msg *Msg) error {
		handlerCalled = true
		if string(msg.Data) != expectedData {
			t.Errorf("Expected message data '%s', got '%s'", expectedData, string(msg.Data))
			return fmt.Errorf("Expected message data '%s', got '%s'", expectedData, string(msg.Data))
		}
		return nil
	}

	m.Subscribe("test/topic", handler)
	msg := NewMsg("test/topic", []byte(expectedData), "test-id")
	m.PubMsg(msg)

	if !handlerCalled {
		t.Error("Handler was not called after publishing message")
	}
}

func TestMessangerLocalPubData(t *testing.T) {
	resetNodes()
	m, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	tests := []struct {
		name     string
		data     interface{}
		wantData string
	}{
		{"string data", "hello", "hello"},
		{"int data", 42, "42"},
		{"bool data", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic := "test/topic"
			receivedData := ""
			handler := func(msg *Msg) error {
				receivedData = string(msg.Data)
				return nil
			}

			m.Subscribe(topic, handler)
			m.Pub(topic, tt.data)

			if receivedData != tt.wantData {
				t.Errorf("Pub() got = %v, want %v", receivedData, tt.wantData)
			}
		})
	}
}

// TestMessangerLocalClose needs to be implemented, however we need to have
// a way to accurately identify the subscriber that needs to be removed.
func TestMessangerLocalClose(t *testing.T) {
	// Setup
	// m := NewMessangerLocal("test-id")

	// // Create some subscriptions and publish some messages
	// handlerCalled := false
	// handler := func(msg *Msg) error {
	// 	handlerCalled = true
	// 	return nil
	// }

	// // Subscribe and publish before close
	// m.Subscribe("test/topic", handler)

	// if !handlerCalled {
	// 	t.Error("Handler should have been called before Close()")
	// }

	// // Call Close
	// m.Close()

	// // Verify that after Close(), new publications don't trigger handlers
	// handlerCalled = false

	// if handlerCalled {
	// 	t.Error("Handler should not have been called after Close()")
	// }
}

// TestMessengerLocalPubWithNoTopic tests Pub when empty topic is provided
func TestMessengerLocalPubWithNoTopic(t *testing.T) {
	defer resetNodes()

	// Create local messenger
	messenger, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	// Try to publish data with empty topic - should return error
	testData := "test string"
	err = messenger.Pub("", testData)
	assert.Error(t, err, "expected err with no topic")
}

// TestMessengerLocalPubMsgWithNoSubscribers tests PubMsg when no subscribers exist
func TestMessengerLocalPubMsgWithNoSubscribers(t *testing.T) {
	defer resetNodes()

	// Create local messenger
	messenger, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	// Create message for topic with no subscribers
	msg := NewMsg("nonexistent/topic", []byte("test"), "test-id")

	// Publish message - should log info about no subscribers
	err = messenger.PubMsg(msg)
	assert.Error(t, err, "No subscribers should return an error")

	// Verify info message was logged
	if !strings.Contains(err.Error(), "No subscribers for") {
		t.Errorf("Expected 'No subscribers' log, got: %s", err.Error())
	}
}

// TestMessengerLocalPubCountsPublications tests that Published counter is incremented
func TestMessengerLocalPubCountsPublications(t *testing.T) {
	defer resetNodes()

	// Create local messenger
	messenger, err := NewMessangerLocal("test-id")
	assert.NoError(t, err)

	// Verify initial count
	if messenger.Published != 0 {
		t.Errorf("Expected initial Published count of 0, got %d", messenger.Published)
	}

	// Publish a message
	messenger.Pub("test/topic", "test message")

	// Verify counter was incremented
	if messenger.Published != 1 {
		t.Errorf("Expected Published count of 1, got %d", messenger.Published)
	}

	// Publish another message
	messenger.Pub("test/topic2", "test message 2")

	// Verify counter was incremented again
	if messenger.Published != 2 {
		t.Errorf("Expected Published count of 2, got %d", messenger.Published)
	}
}
