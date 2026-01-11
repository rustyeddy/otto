package ottoctl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	tstfile := "testdata/version.json"
	response, err := os.ReadFile(tstfile)
	assert.NoError(t, err, "failed to open")

	tst := newTstInput(runVersion, string(response))
	err = httpQuery(t, tst)
	assert.NoError(t, err)

	output := tst.buffer.String()
	assert.Equal(t, "0.0.12\n", output)
}
