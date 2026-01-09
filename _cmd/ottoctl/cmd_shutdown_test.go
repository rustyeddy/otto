package ottoctl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShutdown(t *testing.T) {
	tst := newTstInput(runShutdown, string(`{ "shutdown": "shutting down in 2 seconds" }`))

	err := httpQuery(t, tst)
	assert.NoError(t, err)

	output := tst.buffer.String()
	assert.Contains(t, output, "shutting down in 2 seconds")
}
