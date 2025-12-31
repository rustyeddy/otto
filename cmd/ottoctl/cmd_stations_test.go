package ottoctl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStations(t *testing.T) {
	tstfile := "testdata/stations.json"
	response, err := os.ReadFile(tstfile)
	assert.NoError(t, err, "failed to open")

	tst := newTstInput(stationsRun, string(response))
	err = httpQuery(t, tst)
	assert.NoError(t, err)

	output := tst.buffer.String()
	assert.Contains(t, output, "ID Hostname")
	assert.Contains(t, output, "LastHeard")
	assert.Contains(t, output, "-------------------------")
}

func TestStationsEmpty(t *testing.T) {
	response := ""
	tst := newTstInput(stationsRun, response)
	err := httpQuery(t, tst)
	assert.Error(t, err)
}

func TestStationsError(t *testing.T) {
	tst := newTstInput(stationsRun, "")
	tst.statusCode = 500
	err := httpQuery(t, tst)
	assert.Error(t, err)

	output := tst.errbuf.String()
	assert.Contains(t, output, "500 - Internal Server Error")
}
