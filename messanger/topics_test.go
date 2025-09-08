package messanger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTopics(t *testing.T) {
	topicInstance := GetTopics()
	if topicInstance == nil {
		t.Fatal("Expected Topics instance, got nil")
	}
}

func TestSetStationName(t *testing.T) {
	topics := GetTopics()
	topics.SetStationName("TestStation")
	if topics.StationName != "TestStation" {
		t.Errorf("Expected StationName to be 'TestStation', got '%s'", topics.StationName)
	}
}

func TestControl(t *testing.T) {
	topics := GetTopics()
	topics.SetStationName("TestStation")
	controlTopic := topics.Control("foo")
	expected := "ss/c/TestStation/foo"
	if controlTopic != expected {
		t.Errorf("Expected control topic '%s', got '%s'", expected, controlTopic)
	}
	if topics.Topicmap[controlTopic] != 1 {
		t.Errorf("Expected topic count for '%s' to be 1, got %d", controlTopic, topics.Topicmap[controlTopic])
	}
}

func TestData(t *testing.T) {
	topics := GetTopics()
	topics.SetStationName("TestStation")
	dataTopic := topics.Data("bar")
	expected := "ss/d/TestStation/bar"
	if dataTopic != expected {
		t.Errorf("Expected data topic '%s', got '%s'", expected, dataTopic)
	}
	if topics.Topicmap[dataTopic] != 1 {
		t.Errorf("Expected topic count for '%s' to be 1, got %d", dataTopic, topics.Topicmap[dataTopic])
	}
}

func TestServeHTTP(t *testing.T) {
	topics := GetTopics()
	topics.SetStationName("TestStation")
	topics.Control("foo")
	topics.Data("bar")

	req := httptest.NewRequest(http.MethodGet, "/topics", nil)
	w := httptest.NewRecorder()

	topics.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	var decodedTopics Topics
	err := json.NewDecoder(resp.Body).Decode(&decodedTopics)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if decodedTopics.StationName != "TestStation" {
		t.Errorf("Expected StationName to be 'TestStation', got '%s'", decodedTopics.StationName)
	}

	if len(decodedTopics.Topicmap) != 2 {
		t.Errorf("Expected 2 topics in Topicmap, got %d", len(decodedTopics.Topicmap))
	}
}
