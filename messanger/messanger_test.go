package messanger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessangerBaseServeHTTP(t *testing.T) {
	mb := NewMessangerBase("test-id")

	topic := "topic1"
	mb.subs[topic] = func(msg *Msg) error { return nil }
	mb.subs["topic2"] = func(msg *Msg) error { return nil }
	mb.subs["topic3"] = func(msg *Msg) error { return nil }
	mb.Published = 5

	req, err := http.NewRequest(http.MethodGet, "/messanger", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mb.ServeHTTP)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	expectedContentType := "application/json"
	contentType := rr.Header().Get("Content-Type")
	assert.Equal(t, contentType, expectedContentType)

	var response struct {
		ID        string
		Subs      []string
		Published int
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, response.ID, "test-id")

	expectedSubs := []string{topic, "topic2", "topic3"}
	assert.True(t, reflect.DeepEqual(response.Subs, expectedSubs),
		"expected (%v) got (%v)", expectedSubs, response.Subs)
	assert.Equal(t, response.Published, 5)
}

func TestNewMessanger(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		topic     string
		expected  string
		expectNil bool
	}{
		{
			name:     "Create none messanger",
			id:       "none",
			topic:    "topic1",
			expected: "none",
		},
		{
			name:     "Create messanger with mqtt id",
			id:       "mqtt",
			topic:    "topic1",
			expected: "mqtt",
		},
		{
			name:      "Create messanger with unknown id returns nil",
			id:        "unknown",
			topic:     "topic1",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMessanger(tt.id)
			if tt.expectNil {
				assert.Nil(t, m)
			} else {
				assert.NotNil(t, m)
				assert.Equal(t, tt.expected, m.ID())
			}
		})
	}
}

func TestGetMessanger(t *testing.T) {
	// Ensure singleton behavior - use 'none' to avoid broker conflicts
	m1 := NewMessanger("none")
	m2 := GetMessanger()

	if m1 != m2 {
		t.Errorf("expected GetMessanger to return the same instance, got different instances")
	}
}

func TestMessangerBaseError(t *testing.T) {
	mb := NewMessangerBase("test-id")
	if err := mb.Error(); err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}
