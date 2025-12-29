package ottoctl

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rustyeddy/otto/client"
)

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

	runVersion(tst.cmd, tst.args)
	return err
}
