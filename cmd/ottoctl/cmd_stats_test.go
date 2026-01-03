package ottoctl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	tstfile := "testdata/stats.json"
	response, err := os.ReadFile(tstfile)
	assert.NoError(t, err, "failed to open")

	tst := newTstInput(statsRun, string(response))
	err = httpQuery(t, tst)
	assert.NoError(t, err)

	output := tst.buffer.String()
	assert.Contains(t, output, "Alloc")
	assert.Contains(t, output, "TotalAlloc")
}

func TestStatsEmpty(t *testing.T) {
	response := ""
	tst := newTstInput(statsRun, response)
	err := httpQuery(t, tst)
	assert.Error(t, err)
}

func TestStatsError(t *testing.T) {
	tst := newTstInput(statsRun, "")
	tst.statusCode = 500
	err := httpQuery(t, tst)
	assert.Error(t, err)

	output := tst.errbuf.String()
	assert.Contains(t, output, "500 - Internal Server Error")
}
