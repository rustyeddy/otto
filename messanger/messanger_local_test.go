package messanger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMessangerLocal(t *testing.T) {
	m := NewMessangerLocal("test-id", "test/topic")
	if m.ID() != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", m.ID())
	}
	if m.topic[0] != "test/topic" {
		t.Errorf("Expected Topic 'test/topic', got '%s'", m.Topic())
	}
}

func TestSetTopic(t *testing.T) {
	m := NewMessangerLocal("test-id")
	m.SetTopic("new/topic")
	if len(m.topic) == 0 || m.topic[0] != "new/topic" {
		t.Errorf("Expected Topic 'new/topic', got '%v'", m.topic)
	}
}

func TestSubscribe(t *testing.T) {
	m := NewMessangerLocal("test-id")
	handlerCalled := false
	handler := func(msg *Msg) {
		handlerCalled = true
	}

	err := m.Subscribe("test/topic", handler)
	if err != nil {
		t.Fatalf("Subscribe returned an error: %v", err)
	}

	if len(m.subs["test/topic"]) != 1 {
		t.Errorf("Expected 1 handler for 'test/topic', got %d", len(m.subs["test/topic"]))
	}

	// Simulate publishing a message to trigger the handler
	msg := NewMsg("test/topic", []byte("test data"), "test-id")
	m.PubMsg(msg)

	if !handlerCalled {
		t.Error("Expected handler to be called, but it was not")
	}
}

func TestPubMsg(t *testing.T) {
	resetNodes()
	m := NewMessangerLocal("test-id")
	handlerCalled := false
	handler := func(msg *Msg) {
		handlerCalled = true
		if string(msg.Data) != "test data" {
			fmt.Printf("FOO %+v\n", msg)
			t.Errorf("failed")
			//t.Errorf("Expected message data 'test data', got '%+v'", msg.Data)
			println("BAR")
		}
	}

	m.Subscribe("test/topic", handler)
	msg := NewMsg("test/topic", []byte("test data"), "test-id")
	m.PubMsg(msg)

	if !handlerCalled {
		t.Error("Expected handler to be called, but it was not")
	}
}

func TestPubData(t *testing.T) {
	resetNodes()
	m := NewMessangerLocal("test-id", "test/topic")
	handlerCalled := false
	handler := func(msg *Msg) {
		handlerCalled = true
		if string(msg.Data) != "42" {
			t.Errorf("Expected message data '42', got '%s'", string(msg.Data))
		}
	}

	m.Subscribe("test/topic", handler)
	m.PubData(42)

	if !handlerCalled {
		t.Error("Expected handler to be called, but it was not")
	}
}

func TestServeHTTPLocal(t *testing.T) {
	resetNodes()
	m := NewMessangerLocal("test-id", "test/topic")
	m.PubData("test data")

	req := httptest.NewRequest(http.MethodGet, "/messanger", nil)
	w := httptest.NewRecorder()

	m.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	var decodedMessanger MessangerLocal
	err := json.NewDecoder(resp.Body).Decode(&decodedMessanger)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	fmt.Printf("decode manager: %+v\n", decodedMessanger)

	if decodedMessanger.ID() != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", decodedMessanger.ID())
	}
	if len(decodedMessanger.topic) == 0 ||
		decodedMessanger.topic[0] != "test/topic" {
		t.Errorf("Expected Topic 'test/topic', got '%v'", decodedMessanger.topic)
	}
}
