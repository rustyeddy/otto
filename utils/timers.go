package utils

import "time"

// Ticker is a wrapper around time.Ticker it is given a name, it hold
// the duration and kept in a map indexed by name such that it is easy
// to lookup to shutdown or reset
type Ticker struct {
	Name string
	*time.Ticker
	Func func(t time.Time)
}

var (
	// Start time is the time otto started
	StartTime time.Time

	// the map with all our tickers
	tickers = make(map[string]*Ticker)
)

func init() {
	StartTime = time.Now()
}

// Timestamp returns the time.Duration since the program was started,
// useful to stamping communication messages.
func Timestamp() time.Duration {
	return time.Since(StartTime)
}

// NewTicker creates a time.Ticker with the name n that will fire
// every d time.Duration. The function f will be called every time
// ticker goes off.  The ticker can be stoped, restarted and reset
// with a different duration
func NewTicker(n string, d time.Duration, f func(t time.Time)) *Ticker {
	t := &Ticker{
		Name:   n,
		Ticker: time.NewTicker(d),
		Func:   f,
	}

	tickers[n] = t
	go func() {
		for tick := range t.Ticker.C {
			f(tick)
		}
	}()
	return t
}

// GetTickers will return the map of all ticker values.
func GetTickers() map[string]*Ticker {
	return tickers
}

// GetTicker will return the named ticker or nil if it does not exist
func GetTicker(n string) *Ticker {
	t, _ := tickers[n]
	return t
}
