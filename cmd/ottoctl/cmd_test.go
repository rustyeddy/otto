package ottoctl

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rustyeddy/otto/client"
	"github.com/spf13/cobra"
)

type tstInput struct {
	response string
	buffer   *bytes.Buffer
	errbuf   *bytes.Buffer
	cmd      *cobra.Command
	args     []string
	clifunc  func(cmd *cobra.Command, args []string) error
	statusCode int
}

func newTstInput(clifunc func(cmd *cobra.Command, args []string) error, response string) (tst *tstInput) {
	tst = &tstInput{
		response: response,
		buffer:   bytes.NewBufferString(""),
		errbuf:   bytes.NewBufferString(""),
		args:     []string{},
		cmd:      &cobra.Command{},
		clifunc:  clifunc,
	}
	return tst
}

func httpQuery(t *testing.T, tst *tstInput) (err error) {
	t.Helper()

	// create the test server that our cli command is going to
	// unwittingly connect to
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, tst.response)
	}))
	defer ts.Close()

	cli = &client.Client{
		BaseURL:    ts.URL,
		HTTPClient: ts.Client(),
	}
	tst.buffer = bytes.NewBufferString("")
	tst.errbuf = bytes.NewBufferString("")
	cmdOutput = tst.buffer
	errOutput = tst.errbuf

	tst.clifunc(tst.cmd, tst.args)
	return err
}
