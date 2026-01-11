package utils

import (
	"fmt"
	"math/rand"
	"net/http"
)

// PeriodicRandomData will collected a new random piece of data
// every period and transmit it to the given mqtt channel
type Rando struct {
	F float64
}

func NewRando() (r *Rando) {
	r = &Rando{
		F: 0.0,
	}
	return r
}

func (p Rando) Float64() float64 {
	p.F = rand.Float64()
	return p.F
}

func (p Rando) String() interface{} {
	p.F = rand.Float64()
	s := fmt.Sprintf("%f", p.F)
	return s
}

func (p Rando) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%f", p.Float64())
}
