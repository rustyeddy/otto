package ottoctl

import (
	"testing"

	"github.com/rustyeddy/otto/utils"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLog(t *testing.T) {
	lc := utils.DefaultLogConfig()
	jbytes := lc.JSON()

	tst := newTstInput(runLog, string(jbytes))

	err := httpQuery(t, tst)
	assert.NoError(t, err)

	output := tst.buffer.String()
	assert.Contains(t, output, "Output: stdout")
	assert.Contains(t, output, "Format: text")
	assert.Contains(t, output, "FilePath: /var/log/otto.log")
	assert.Contains(t, output, "Buffer: <nil>")
}
