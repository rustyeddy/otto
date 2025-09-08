package messanger

import (
	"errors"
	"testing"
)

// Mock implementation of the Messanger interface
type MockMessanger struct {
	id       string
	topic    string
	err      error
	closed   bool
	handlers map[string]MsgHandler
}

func (m *MockMessanger) ID() string {
	return m.id
}

func (m *MockMessanger) Subscribe(topic string, handler MsgHandler) error {
	if topic == "" {
		return errors.New("topic cannot be empty")
	}
	if handler == nil {
		return errors.New("handler cannot be nil")
	}
	m.handlers[topic] = handler
	return nil
}

func (m *MockMessanger) SetTopic(topic string) {
	m.topic = topic
}

func (m *MockMessanger) PubMsg(msg *Msg) {
	if handler, ok := m.handlers[m.topic]; ok {
		handler(msg)
	}
}

func (m *MockMessanger) PubData(data any) {
	// Mock implementation for PubData
}

func (m *MockMessanger) Error() error {
	return m.err
}

func (m *MockMessanger) Close() {
	m.closed = true
}

// Test cases
func TestMessanger_ID(t *testing.T) {
	mock := &MockMessanger{id: "12345"}
	if mock.ID() != "12345" {
		t.Errorf("expected ID to be '12345', got '%s'", mock.ID())
	}
}

func TestMessanger_Subscribe(t *testing.T) {
	mock := &MockMessanger{handlers: make(map[string]MsgHandler)}

	err := mock.Subscribe("test-topic", func(msg *Msg) {})
	if err != nil {
		t.Errorf("expected no error, got '%v'", err)
	}

	err = mock.Subscribe("", nil)
	if err == nil {
		t.Errorf("expected error for empty topic or nil handler, got nil")
	}
}

func TestMessanger_SetTopic(t *testing.T) {
	mock := &MockMessanger{}
	mock.SetTopic("new-topic")
	if mock.topic != "new-topic" {
		t.Errorf("expected topic to be 'new-topic', got '%s'", mock.topic)
	}
}

func TestMessanger_PubMsg(t *testing.T) {
	mock := &MockMessanger{handlers: make(map[string]MsgHandler)}
	mock.SetTopic("test-topic")
	var receivedMsg *Msg
	mock.Subscribe("test-topic", func(msg *Msg) {
		receivedMsg = msg
	})

	msg := &Msg{}
	mock.PubMsg(msg)

	if receivedMsg != msg {
		t.Errorf("expected received message to be '%v', got '%v'", msg, receivedMsg)
	}
}

func TestMessanger_Error(t *testing.T) {
	mock := &MockMessanger{err: errors.New("test error")}
	if mock.Error() == nil || mock.Error().Error() != "test error" {
		t.Errorf("expected error 'test error', got '%v'", mock.Error())
	}
}

func TestMessanger_Close(t *testing.T) {
	mock := &MockMessanger{}
	mock.Close()
	if !mock.closed {
		t.Errorf("expected messanger to be closed, but it wasn't")
	}
}
