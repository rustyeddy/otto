package data

import (
	"fmt"
	"time"
)

// Timeseries represents a single source of data over a time period
type Timeseries struct {
	ID    string `json:"id"`
	Datas []Data `json:"data"`
}

// NewTimeseries will start a new data timeseries with the given label
func NewTimeseries(ID string) *Timeseries {
	return &Timeseries{
		ID: ID,
	}
}

// Add a new Data point to the given Timeseries
func (ts *Timeseries) Add(d any) Data {
	dat := &DataPoint{
		value:     d,
		timestamp: time.Now(),
	}
	ts.Datas = append(ts.Datas, dat)
	return dat
}

// Add a new Data point to the given Timeseries
func (ts *Timeseries) AddTimestamp(d any, t time.Time) Data {
	dat := &DataPoint{
		value:     d,
		timestamp: t,
	}
	ts.Datas = append(ts.Datas, dat)
	return dat
}

// Len returns the number of data points contained in this timeseries
func (ts *Timeseries) Len() int {
	return len(ts.Datas)
}

func (ts *Timeseries) GetReadingsInRange(start time.Time, end time.Time) []Data {
	var series []Data
	for _, d := range ts.Datas {
		if d.Timestamp().After(end) {
			break
		}
		if d.Timestamp().After(start) {
			series = append(series, d)
		}
	}
	return series
}

// String returns a human readable string describing the data
// contained therein.
func (ts *Timeseries) String() string {
	str := fmt.Sprintf("%s: ", ts.ID)
	for _, d := range ts.Datas {
		str += d.String()
	}
	str += "\n"
	return str
}
