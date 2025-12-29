package ottoctl

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type tstInput struct {
	response string
	buffer   *bytes.Buffer
	errbuf   *bytes.Buffer
	cmd      *cobra.Command
	args     []string
}

func TestVersion(t *testing.T) {
	tstfile := "testdata/version.json"
	response, err := os.ReadFile(tstfile)
	assert.NoError(t, err, "failed to open")

	tst := &tstInput{
		response: string(response),
		buffer:   bytes.NewBufferString(""),
		errbuf:   bytes.NewBufferString(""),
		args:     []string{},
		cmd:      &cobra.Command{},
	}

	err = httpQuery(t, tst)
	assert.NoError(t, err)

	output := tst.buffer.String()
	assert.Equal(t, "0.0.12\n", output)
}
