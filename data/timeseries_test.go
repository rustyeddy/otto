package data

import "testing"

func TestTimeSeries(t *testing.T) {

	ts := NewTimeseries("test-station", "test-int")
	if ts.Label != "test-int" {
		t.Errorf("Timeseries label expected (test-int) got (%s)", ts.Label)
	}

	if ts.Station != "test-station" {
		t.Errorf("Timeseries label expected (test-int) got (%s)", ts.Label)
	}

	if len(ts.Data) != 0 {
		t.Errorf("Timeseries.Data expected len (0) got (%d)", len(ts.Data))
	}

	cnt := 10
	for i := 0; i < cnt; i++ {
		ts.Add(i)
	}

	if ts.Len() != cnt {
		t.Errorf("Timeseries.Len expected (%d) got (%d)", cnt, ts.Len())
	}

	for i := 0; i < cnt; i++ {
		d := ts.Data[i]
		if d.Int() != i {
			t.Errorf("Timeseries data expected (%d) got (%+v)", cnt, d.Int())
		}
	}
}
