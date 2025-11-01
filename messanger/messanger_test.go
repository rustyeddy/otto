package messanger

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	err := mock.Subscribe("test-topic", func(msg *Msg) error { return nil })
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
	mock.Subscribe("test-topic", func(msg *Msg) error {
		receivedMsg = msg
		return nil
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
func TestMessangerBaseTopic(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		expected  string
		expectErr bool
	}{
		{
			name:      "No topics",
			topic:     "",
			expected:  "",
			expectErr: false,
		},
		{
			name:      "Single topic",
			topic:     "topic1",
			expected:  "topic1",
			expectErr: false,
		},
		{
			name:      "Single topic",
			topic:     "topic1",
			expected:  "topic1",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb, err := NewMessangerBase("test-id", tt.topic)
			if tt.expectErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			if got := mb.Topic(); got != tt.expected {
				t.Errorf("expected topic '%s', got '%s'", tt.expected, got)
			}
		})
	}
}
func TestMessangerBase_ServeHTTP(t *testing.T) {
	mb, err := NewMessangerBase("test-id", "topic1")
	require.NoError(t, err)

	mb.subs["topic1"] = func(msg *Msg) error { return nil }
	mb.Published = 5

	req, err := http.NewRequest(http.MethodGet, "/messanger", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mb.ServeHTTP)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}

	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("expected Content-Type '%s', got '%s'", expectedContentType, contentType)
	}

	var response struct {
		ID        string
		Topic     string
		Subs      []string
		Published int
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got '%s'", response.ID)
	}

	expectedTopics := "topic1"
	if !reflect.DeepEqual(response.Topic, expectedTopics) {
		t.Errorf("expected topics %v, got %v", expectedTopics, response.Topic)
	}

	expectedSubs := []string{"topic1"}
	if !reflect.DeepEqual(response.Subs, expectedSubs) {
		t.Errorf("expected subs %v, got %v", expectedSubs, response.Subs)
	}

	if response.Published != 5 {
		t.Errorf("expected Published 5, got %d", response.Published)
	}
}
func TestNewMessanger(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		topic    string
		expected string
	}{
		{
			name:     "Create local messanger",
			id:       "local",
			topic:    "topic1",
			expected: "local",
		},
		{
			name:     "Create messanger with mqtt id",
			id:       "mqtt",
			topic:    "topic1",
			expected: "mqtt",
		},
		{
			name:     "Create messanger with custom id",
			id:       "unknown",
			topic:    "topic1",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMessanger(tt.id, tt.topic)
			assert.NoError(t, err)
			assert.NotNil(t, m)
			assert.Equal(t, tt.expected, m.ID())
		})
	}
}

func TestGetMessanger(t *testing.T) {
	// Ensure singleton behavior
	m1, err := NewMessanger("local", "topic1")
	assert.NoError(t, err)
	m2 := GetMessanger()

	if m1 != m2 {
		t.Errorf("expected GetMessanger to return the same instance, got different instances")
	}
}

func TestMessangerBaseError(t *testing.T) {
	mb, err := NewMessangerBase("test-id", "test-topic")
	assert.NoError(t, err)
	if err := mb.Error(); err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}
