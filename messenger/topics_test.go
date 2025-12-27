package messenger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rustyeddy/otto/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetTopics(t *testing.T) {
	topicInstance := GetTopics()
	if topicInstance == nil {
		t.Fatal("Expected Topics instance, got nil")
	}
}

func TestSetStationName(t *testing.T) {
	utils.SetStationName("TestStation")
	if utils.StationName() != "TestStation" {
		t.Errorf("Expected StationName to be 'TestStation', got '%s'", utils.StationName())
	}
}

func TestControl(t *testing.T) {
	topics := GetTopics()
	utils.SetStationName("TestStation")
	controlTopic := topics.Control("foo")
	expected := "o/c/TestStation/foo"
	assert.Equal(t, controlTopic, expected)
	assert.Equal(t, topics.Map[controlTopic], 1)
}

func TestData(t *testing.T) {
	topics := GetTopics()
	utils.SetStationName("TestStation")
	dataTopic := topics.Data("bar")
	expected := "o/d/TestStation/bar"
	if dataTopic != expected {
		t.Errorf("Expected data topic '%s', got '%s'", expected, dataTopic)
	}
	if topics.Map[dataTopic] != 1 {
		t.Errorf("Expected topic count for '%s' to be 1, got %d", dataTopic, topics.Map[dataTopic])
	}
}

func TestServeHTTP(t *testing.T) {
	topics := GetTopics()

	stationName := "TestStation"
	utils.SetStationName(stationName)
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

	assert.Equal(t, 0, decodedTopics.Map["o/c/TesstStation/foo"])
	assert.Equal(t, 0, decodedTopics.Map["o/d/TesstStation/bar"])
}
