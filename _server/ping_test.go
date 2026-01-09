package server

import (
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {

	ts := httptest.NewServer(Ping{})
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		slog.Error(err.Error())
	}
	pong, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		slog.Error(err.Error())
	}

	if string(pong) != "Pong\n" {
		slog.Error(err.Error())
	}
}
