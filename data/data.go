package data

import (
	"fmt"
	"time"
)

var Truncate time.Duration

type Data interface {
	Timestamp() time.Time
	Value() any
	String() string
}

// Data is an array of timestamps and values representing the same
// source of data over a period of time
type DataPoint struct {
	value     any       `json:"value"`
	timestamp time.Time `json:"time-increment"`
}

func NewData(dat any, ts time.Time) Data {
	return NewDataPoint(dat, ts)
}

func NewDataPoint(dat any, ts time.Time) *DataPoint {
	d := &DataPoint{
		value:     dat,
		timestamp: ts,
	}
	return d
}

func (d DataPoint) Timestamp() time.Time {
	return d.timestamp
}

func (d DataPoint) Value() any {
	return d.value
}

func SetTruncateValue(d time.Duration) {
	Truncate = d
}

// Return the float64 representation of the data. If the data is not
// represented by a float64 value a panic will follow
func (d DataPoint) Float() float64 {
	return d.value.(float64)
}

// Int returns the integer value of the data. If the data is not
// an integer a panic will result.
func (d DataPoint) Int() int {
	return d.value.(int)
}
func (d DataPoint) String() string {
	return fmt.Sprintf("%v, ", d.Value())
}
