package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRando(t *testing.T) {
	r := NewRando()
	v := r.Float64()
	if v <= 0.0 {
		t.Errorf("Expected a float value got (%f2.1)", v)
	}

	str := r.String()
	if str == "" {
		t.Errorf("expected a float value got (%s)", r.String())
	}
}

func TestRandHTTP(t *testing.T) {
	ts := httptest.NewServer(NewRando())
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}

	cbuf, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	var f float64
	i, err := fmt.Sscanf(string(cbuf), "%f", &f)
	if err != nil || i != 1 {
		t.Errorf("Expected err (nil) got (%s) also expected i = (1) got (%d)", err, i)
	}

	if f < 0 || f > 1.0 {
		t.Errorf("Expected float point (> 0 && < 1) got (%f) ", f)
	}
}
