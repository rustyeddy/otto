package messenger

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusPayloadJSON(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	payload := StatusPayload{
		Status: "online",
		Time:   ts,
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))

	assert.Equal(t, "online", got["status"])
	assert.Equal(t, ts.Format(time.RFC3339Nano), got["time"])
}

func TestMetaPayloadJSONOmitEmpty(t *testing.T) {
	t.Parallel()

	payload := MetaPayload{
		Name:      "lamp",
		Kind:      "switch",
		ValueType: "bool",
		Access:    "rw",
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))

	assert.Equal(t, "lamp", got["name"])
	assert.Equal(t, "switch", got["kind"])
	assert.Equal(t, "bool", got["value_type"])
	assert.Equal(t, "rw", got["access"])
	assert.NotContains(t, got, "unit")
	assert.NotContains(t, got, "min")
	assert.NotContains(t, got, "max")
	assert.NotContains(t, got, "tags")
	assert.NotContains(t, got, "attrs")
}

func TestMetaPayloadJSONFull(t *testing.T) {
	t.Parallel()

	min := 1.5
	max := 9.5
	payload := MetaPayload{
		Name:      "temp",
		Kind:      "sensor",
		ValueType: "float",
		Access:    "ro",
		Unit:      "C",
		Min:       &min,
		Max:       &max,
		Tags:      []string{"indoor", "calibrated"},
		Attrs: map[string]string{
			"vendor": "acme",
		},
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))

	assert.Equal(t, "temp", got["name"])
	assert.Equal(t, "sensor", got["kind"])
	assert.Equal(t, "float", got["value_type"])
	assert.Equal(t, "ro", got["access"])
	assert.Equal(t, "C", got["unit"])
	assert.Equal(t, min, got["min"])
	assert.Equal(t, max, got["max"])
	assert.Equal(t, []any{"indoor", "calibrated"}, got["tags"])
	assert.Equal(t, map[string]any{"vendor": "acme"}, got["attrs"])
}
