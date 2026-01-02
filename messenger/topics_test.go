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
	assert.NotNil(t, topicInstance)
}

func TestSetStationName(t *testing.T) {
	utils.SetStationName("TestStation")
	assert.Equal(t, "TestStation", utils.StationName())
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
	assert.Equal(t, expected, dataTopic)
	assert.Equal(t, 1, topics.Map[dataTopic])
}

func TestValidateTopic(t *testing.T) {
	tt := []struct {
		topic  string
		expect bool
	}{
		{"too/short", false},
		{"x/d/foo/bar", false},
		{"o/x/foo/bar", false},
		{"o/d//bar", false},
		{"o/d/station//", false},
		{"o/c/station/control", true},
		{"o/d/station/sensor", true},
	}

	for _, tst := range tt {
		ok := ValidateTopic(tst.topic)
		assert.Equal(t, tst.expect, ok)
	}
}

func TestControlTopic(t *testing.T) {
	oldname := utils.StationName()
	utils.SetStationName("teststation")
	ct := ControlTopic("super-bad")
	assert.Equal(t, ct, "o/c/teststation/super-bad")
	utils.SetStationName(oldname)
}

func TestDataTopic(t *testing.T) {
	oldname := utils.StationName()
	utils.SetStationName("teststation")
	ct := DataTopic("super-bad")
	assert.Equal(t, ct, "o/d/teststation/super-bad")
	utils.SetStationName(oldname)
}

func TestTopic(t *testing.T) {
	top := Topic("f/y/z")
	assert.Equal(t, top, "o/f/y/z")
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

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var decodedTopics Topics
	err := json.NewDecoder(resp.Body).Decode(&decodedTopics)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 0, decodedTopics.Map["o/c/TesstStation/foo"])
	assert.Equal(t, 0, decodedTopics.Map["o/d/TesstStation/bar"])
}
