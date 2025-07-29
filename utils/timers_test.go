package utils

import (
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	count := 0

	done := make(chan bool)
	start := time.Now()
	var times []time.Time
	f := func(ti time.Time) {
		times = append(times, ti)
		count++
		if count > 9 {
			done <- true
		}
	}

	name := "alpha"
	ticker := NewTicker(name, 1*time.Millisecond, f)

	if len(tickers) != 1 {
		t.Errorf("Expected 1 timer got (%d)", len(tickers))
	}

	t1, ex := tickers[name]
	if ex == false || t1 == nil {
		t.Error("Expected to find timer alpha but did not")
	}

	if ticker != t1 {
		t.Error("Expected ticker t to == ticker t1 but did not")
	}

	<-done

	if count != 10 {
		t.Errorf("Expected count (10) got (%d)", count)
	}

	elapsed := time.Now().Sub(start)
	if elapsed.Milliseconds() >= 11 {
		t.Errorf("Expected duration (<11ms) got (%d)", elapsed.Milliseconds())
	}
}
