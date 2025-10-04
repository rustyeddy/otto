package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeSeries(t *testing.T) {

	ts := NewTimeseries("test-station")
	assert.Equal(t, "test-station", ts.ID)
	assert.Equal(t, 0, len(ts.Datas))

	cnt := 10
	for i := 0; i < cnt; i++ {
		ts.Add(i)
	}

	assert.Equal(t, cnt, ts.Len())

	for i := 0; i < cnt; i++ {
		d := ts.Datas[i]
		d1 := d.(*DataPoint)
		dp := d1
		assert.Equal(t, i, dp.Int())
	}
}
func TestGetReadingsInRange(t *testing.T) {
	ts := NewTimeseries("test-station")

	// Add data points with specific timestamps
	now := time.Now()
	dataPoints := []struct {
		value     int
		timestamp time.Time
	}{
		{value: 1, timestamp: now.Add(-10 * time.Minute)},
		{value: 2, timestamp: now.Add(-5 * time.Minute)},
		{value: 3, timestamp: now.Add(-2 * time.Minute)},
		{value: 4, timestamp: now.Add(-1 * time.Minute)},
		{value: 5, timestamp: now},
		{value: 5, timestamp: now.Add(2 * time.Minute)},
	}

	for _, dp := range dataPoints {
		ts.AddTimestamp(dp.value, dp.timestamp)
	}

	// Define the range
	start := now.Add(-6 * time.Minute)
	end := now.Add(-1 * time.Minute)

	// Get readings in range
	readings := ts.GetReadingsInRange(start, end)

	// Verify the results
	assert.Equal(t, 3, len(readings), "Expect 2 readings")
	assert.Equal(t, 2, readings[0].(*DataPoint).Int(), "expect reading 0 to be 1")
	assert.Equal(t, 3, readings[1].(*DataPoint).Int(), "exprect reading 1 to be 2")
}
func TestTimeseriesString(t *testing.T) {
	ts := NewTimeseries("test-station")

	// Add data points
	ts.Add(1)
	ts.Add(2)
	ts.Add(3)

	// Expected string representation
	expected := "test-station: 1, 2, 3, \n"

	// Verify the string representation
	assert.Equal(t, expected, ts.String())
}
