package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Ticker is a wrapper around time.Ticker it is given a name, it holds
// the duration and kept in a map indexed by name such that it is easy
// to lookup to shutdown or reset
type Ticker struct {
	Name string
	*time.Ticker
	Func func(t time.Time)

	mu       sync.RWMutex
	lastTick time.Time
	ticks    int
	active   bool
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
// ticker goes off.  The ticker can be stopped, restarted and reset
// with a different duration
func NewTicker(n string, d time.Duration, f func(t time.Time)) *Ticker {
	t := &Ticker{
		Name:   n,
		Ticker: time.NewTicker(d),
		Func:   f,
		active: true,
	}

	tickers[n] = t
	go func() {
		for tick := range t.Ticker.C {
			t.mu.Lock()
			t.lastTick = time.Now()
			t.ticks++
			t.mu.Unlock()
			f(tick)
		}
		t.mu.Lock()
		t.active = false
		t.mu.Unlock()
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

// TickerInfo holds the JSON-serializable ticker information
type TickerInfo struct {
	Name     string    `json:"name"`
	LastTick time.Time `json:"last_tick"`
	Ticks    int       `json:"ticks"`
	Active   bool      `json:"active"`
}

// ServeHTTP implements http.Handler to return ticker information as JSON.
// It returns the ticker's name, last tick time, total tick count, and active status.
func (t *Ticker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	t.mu.RLock()
	info := TickerInfo{
		Name:     t.Name,
		LastTick: t.lastTick,
		Ticks:    t.ticks,
		Active:   t.active,
	}
	t.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(info); err != nil {
		slog.Error("Failed to encode ticker info", "error", err, "ticker", t.Name)
		return
	}
}
