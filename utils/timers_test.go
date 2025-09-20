package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// setupTest clears the tickers map before each test to ensure clean state
func setupTest() {
	tickers = make(map[string]*Ticker)
}

// teardownTest stops all tickers and clears the map after each test
func teardownTest() {
	for _, ticker := range tickers {
		if ticker.Ticker != nil {
			ticker.Ticker.Stop()
		}
	}
	tickers = make(map[string]*Ticker)
}

func TestTimestamp(t *testing.T) {
	// Test that Timestamp returns a positive duration
	ts := Timestamp()
	if ts < 0 {
		t.Errorf("Expected positive timestamp, got %v", ts)
	}

	// Test that consecutive calls show increasing time
	time.Sleep(1 * time.Millisecond)
	ts2 := Timestamp()
	if ts2 <= ts {
		t.Errorf("Expected timestamp to increase, got %v then %v", ts, ts2)
	}
}

func TestNewTicker(t *testing.T) {
	setupTest()
	defer teardownTest()

	count := 0
	done := make(chan bool)
	start := time.Now()
	var times []time.Time
	var mu sync.Mutex

	f := func(ti time.Time) {
		mu.Lock()
		times = append(times, ti)
		count++
		if count >= 5 {
			done <- true
		}
		mu.Unlock()
	}

	name := "test-ticker"
	ticker := NewTicker(name, 2*time.Millisecond, f)

	// Test ticker properties
	if ticker.Name != name {
		t.Errorf("Expected ticker name '%s', got '%s'", name, ticker.Name)
	}

	if ticker.Func == nil {
		t.Error("Expected ticker function to be set")
	}

	// Wait for ticks
	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Ticker did not fire expected number of times within timeout")
	}

	mu.Lock()
	finalCount := count
	finalTicks := ticker.ticks
	mu.Unlock()

	if finalCount < 5 {
		t.Errorf("Expected at least 5 ticks, got %d", finalCount)
	}

	if finalTicks != finalCount {
		t.Errorf("Expected ticker.ticks (%d) to equal count (%d)", finalTicks, finalCount)
	}

	// Test timing is reasonable (allowing some variance for CI environments)
	elapsed := time.Since(start)
	expectedMin := time.Duration(finalCount-1) * 2 * time.Millisecond
	expectedMax := time.Duration(finalCount+2) * 2 * time.Millisecond
	if elapsed < expectedMin || elapsed > expectedMax {
		t.Logf("Warning: timing variance - expected between %v and %v, got %v", expectedMin, expectedMax, elapsed)
	}

	// Test that ticker was added to global map
	if len(tickers) != 1 {
		t.Errorf("Expected 1 ticker in global map, got %d", len(tickers))
	}

	storedTicker, exists := tickers[name]
	if !exists {
		t.Error("Expected ticker to be stored in global map")
	}

	if storedTicker != ticker {
		t.Error("Expected stored ticker to be the same instance")
	}
}

func TestGetTicker(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Test getting non-existent ticker
	ticker := GetTicker("non-existent")
	if ticker != nil {
		t.Error("Expected nil for non-existent ticker")
	}

	// Create a ticker and test retrieval
	name := "get-test"
	f := func(ti time.Time) {}
	createdTicker := NewTicker(name, 10*time.Millisecond, f)

	retrievedTicker := GetTicker(name)
	if retrievedTicker == nil {
		t.Error("Expected to retrieve created ticker")
	}

	if retrievedTicker != createdTicker {
		t.Error("Expected retrieved ticker to be the same instance")
	}

	if retrievedTicker.Name != name {
		t.Errorf("Expected ticker name '%s', got '%s'", name, retrievedTicker.Name)
	}
}

func TestGetTickers(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Test empty map
	allTickers := GetTickers()
	if len(allTickers) != 0 {
		t.Errorf("Expected empty ticker map, got %d tickers", len(allTickers))
	}

	// Create multiple tickers
	names := []string{"ticker1", "ticker2", "ticker3"}
	f := func(ti time.Time) {}

	for _, name := range names {
		NewTicker(name, 10*time.Millisecond, f)
	}

	allTickers = GetTickers()
	if len(allTickers) != len(names) {
		t.Errorf("Expected %d tickers, got %d", len(names), len(allTickers))
	}

	// Verify all tickers are present
	for _, name := range names {
		if ticker, exists := allTickers[name]; !exists {
			t.Errorf("Expected ticker '%s' to exist", name)
		} else if ticker.Name != name {
			t.Errorf("Expected ticker name '%s', got '%s'", name, ticker.Name)
		}
	}
}

func TestMultipleTickers(t *testing.T) {
	setupTest()
	defer teardownTest()

	const numTickers = 3
	const ticksPerTicker = 3

	counters := make(map[string]int)
	done := make(chan string, numTickers)
	var mu sync.Mutex

	// Create multiple tickers with different intervals
	for i := 0; i < numTickers; i++ {
		name := fmt.Sprintf("ticker-%d", i)
		interval := time.Duration(i+1) * 2 * time.Millisecond

		counters[name] = 0

		f := func(tickerName string) func(time.Time) {
			return func(ti time.Time) {
				mu.Lock()
				counters[tickerName]++
				if counters[tickerName] >= ticksPerTicker {
					done <- tickerName
				}
				mu.Unlock()
			}
		}(name)

		ticker := NewTicker(name, interval, f)
		if ticker == nil {
			t.Fatalf("Failed to create ticker '%s'", name)
		}
	}

	// Wait for all tickers to complete
	completed := make(map[string]bool)
	timeout := time.After(200 * time.Millisecond)

	for len(completed) < numTickers {
		select {
		case tickerName := <-done:
			completed[tickerName] = true
		case <-timeout:
			mu.Lock()
			for name, count := range counters {
				t.Logf("Ticker '%s' completed %d/%d ticks", name, count, ticksPerTicker)
			}
			mu.Unlock()
			t.Fatal("Timeout waiting for all tickers to complete")
		}
	}

	// Verify all tickers fired the expected number of times
	mu.Lock()
	for name, count := range counters {
		if count < ticksPerTicker {
			t.Errorf("Ticker '%s' expected at least %d ticks, got %d", name, ticksPerTicker, count)
		}
	}
	mu.Unlock()

	// Verify all tickers are in the global map
	allTickers := GetTickers()
	if len(allTickers) != numTickers {
		t.Errorf("Expected %d tickers in global map, got %d", numTickers, len(allTickers))
	}

	// Verify each ticker's internal state
	for i := 0; i < numTickers; i++ {
		name := fmt.Sprintf("ticker-%d", i)
		ticker := GetTicker(name)
		if ticker == nil {
			t.Errorf("Ticker '%s' not found in global map", name)
			continue
		}

		if ticker.ticks < ticksPerTicker {
			t.Errorf("Ticker '%s' internal tick count %d, expected at least %d", name, ticker.ticks, ticksPerTicker)
		}

		if ticker.lastTick.IsZero() {
			t.Errorf("Ticker '%s' lastTick should not be zero", name)
		}
	}
}

func TestTickerConcurrency(t *testing.T) {
	setupTest()
	defer teardownTest()

	const numGoroutines = 5
	const ticksPerGoroutine = 2

	var wg sync.WaitGroup
	var mu sync.Mutex
	totalTicks := 0

	f := func(ti time.Time) {
		mu.Lock()
		totalTicks++
		mu.Unlock()
	}

	// Create ticker
	ticker := NewTicker("concurrent-test", 1*time.Millisecond, f)

	// Start multiple goroutines that access ticker data
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ticksPerGoroutine; j++ {
				// Read ticker state (this should be safe)
				_ = ticker.Name
				_ = ticker.ticks
				_ = ticker.lastTick
				time.Sleep(2 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Let the ticker run a bit more
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	finalTicks := totalTicks
	mu.Unlock()

	if finalTicks < 5 {
		t.Errorf("Expected at least 5 ticks from concurrent test, got %d", finalTicks)
	}
}
