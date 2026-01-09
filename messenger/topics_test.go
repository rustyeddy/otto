package messenger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopicSchemePaths(t *testing.T) {
	t.Parallel()

	scheme := TopicScheme{Prefix: "otto"}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{name: "state", got: scheme.State("lamp"), expected: "otto/devices/lamp/state"},
		{name: "set", got: scheme.Set("lamp"), expected: "otto/devices/lamp/set"},
		{name: "event", got: scheme.Event("lamp"), expected: "otto/devices/lamp/event"},
		{name: "status", got: scheme.Status("lamp"), expected: "otto/devices/lamp/status"},
		{name: "meta", got: scheme.Meta("lamp"), expected: "otto/devices/lamp/meta"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.got)
		})
	}
}

func TestStateAs(t *testing.T) {
	t.Parallel()

	reg := NewRegistry(nil, TopicScheme{Prefix: "otto"})

	reg.stateMu.Lock()
	reg.stateAny["relay"] = true
	reg.stateMu.Unlock()

	val, ok := StateAs[bool](reg, "relay")
	require.True(t, ok)
	assert.True(t, val)

	_, ok = StateAs[int](reg, "relay")
	assert.False(t, ok)
}
