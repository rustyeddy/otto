package data

import (
	"fmt"
	"time"

	"github.com/rustyeddy/otto/utils"
)

// Timeseries represents a single source of data over a time period
type Timeseries struct {
	Station   string        `json:"station"`
	Label     string        `json:"label"`
	Timestamp time.Duration `json:"start"`
	Data      []*Data       `json:"data"`
}

// NewTimeseries will start a new data timeseries with the given label
func NewTimeseries(station, label string) *Timeseries {
	return &Timeseries{
		Station:   station,
		Label:     label,
		Timestamp: utils.Timestamp(),
	}
}

// Add a new Data point to the given Timeseries
func (ts *Timeseries) Add(d any) *Data {
	dat := &Data{
		Value:     d,
		Timestamp: utils.Timestamp(),
	}
	ts.Data = append(ts.Data, dat)
	return dat
}

// Len returns the number of data points contained in this timeseries
func (ts *Timeseries) Len() int {
	return len(ts.Data)
}

// String returns a human readable string describing the data
// contained therein.
func (ts *Timeseries) String() string {
	str := fmt.Sprintf("%s[%s]", ts.Station, ts.Label)
	str = fmt.Sprintf("%-20s start: %s\n\t", str, ts.Timestamp)
	for _, d := range ts.Data {
		str += d.String()
	}
	str += "\n"
	return str
}
