package ottoctl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimers(t *testing.T) {
	tstfile := "testdata/timers.json"
	response, err := os.ReadFile(tstfile)
	assert.NoError(t, err, "failed to open")

	tst := newTstInput(timersRun, string(response))
	err = httpQuery(t, tst)
	assert.NoError(t, err)

	output := tst.buffer.String()
	assert.Contains(t, output, "heartbeat")
	assert.Contains(t, output, "sensor_poll")
}

func TestTimersEmpty(t *testing.T) {
	response := ""
	tst := newTstInput(timersRun, response)
	err := httpQuery(t, tst)
	assert.Error(t, err)
}

func TestTimersError(t *testing.T) {
	tst := newTstInput(timersRun, "")
	tst.statusCode = 500
	err := httpQuery(t, tst)
	assert.Error(t, err)

	output := tst.errbuf.String()
	assert.Contains(t, output, "500 - Internal Server Error")
}
