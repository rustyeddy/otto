package messanger

import (
	"testing"
)

func TestNewMessangerLocal(t *testing.T) {
	m := NewMessangerLocal("test-id", "test/topic")
	if m.ID() != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", m.ID())
	}
	if m.Topic() != "test/topic" {
		t.Errorf("Expected Topic 'test/topic', got '%s'", m.Topic())
	}
}

func TestMessangerLocalSetTopic(t *testing.T) {
	m := NewMessangerLocal("test-id")
	m.SetTopic("new/topic")
	if m.Topic() != "new/topic" {
		t.Errorf("Expected Topic 'new/topic', got '%s'", m.Topic())
	}
}

func TestMessangerLocalSubscribe(t *testing.T) {
	resetNodes() // Reset the node tree
	m := NewMessangerLocal("test-id")

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
			handler := func(msg *Msg) {
				handlerCalled = true
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
	m := NewMessangerLocal("test-id")

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
			handler := func(msg *Msg) {
				receivedData = string(msg.Data)
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
	m := NewMessangerLocal("test-id")
	handlerCalled := false
	expectedData := "test data"

	handler := func(msg *Msg) {
		handlerCalled = true
		if string(msg.Data) != expectedData {
			t.Errorf("Expected message data '%s', got '%s'", expectedData, string(msg.Data))
		}
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
	m := NewMessangerLocal("test-id", "test/topic")

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
			receivedData := ""
			handler := func(msg *Msg) {
				receivedData = string(msg.Data)
			}

			m.Subscribe(m.Topic(), handler)
			m.PubData(tt.data)

			if receivedData != tt.wantData {
				t.Errorf("PubData() got = %v, want %v", receivedData, tt.wantData)
			}
		})
	}
}

func TestMessangerLocalClose(t *testing.T) {
	// Setup
	m := NewMessangerLocal("test-id", "test/topic")

	// Create some subscriptions and publish some messages
	handlerCalled := false
	handler := func(msg *Msg) {
		handlerCalled = true
	}

	// Subscribe and publish before close
	m.Subscribe("test/topic", handler)
	m.PubData("test message")

	if !handlerCalled {
		t.Error("Handler should have been called before Close()")
	}

	// Call Close
	m.Close()

	// Verify that after Close(), new publications don't trigger handlers
	handlerCalled = false
	m.PubData("test message after close")

	if handlerCalled {
		t.Error("Handler should not have been called after Close()")
	}
}
