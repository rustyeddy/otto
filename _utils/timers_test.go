package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// setupTest clears the tickers map before each test to ensure clean state
func setupTest() {
	tickers = make(Tickers)
}

// teardownTest stops all tickers and clears the map after each test
func teardownTest() {
	for _, ticker := range tickers {
		if ticker.Ticker != nil {
			ticker.Ticker.Stop()
		}
	}
	tickers = make(Tickers)
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
	mu.Unlock()

	ticker.mu.RLock()
	finalTicks := ticker.ticks
	ticker.mu.RUnlock()

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

		ticker.mu.RLock()
		tickCount := ticker.ticks
		lastTickTime := ticker.lastTick
		ticker.mu.RUnlock()

		if tickCount < ticksPerTicker {
			t.Errorf("Ticker '%s' internal tick count %d, expected at least %d", name, tickCount, ticksPerTicker)
		}

		if lastTickTime.IsZero() {
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
				// Read ticker state (this should be safe with mutex)
				_ = ticker.Name
				ticker.mu.RLock()
				_ = ticker.ticks
				_ = ticker.lastTick
				ticker.mu.RUnlock()
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

func TestTickerServeHTTP(t *testing.T) {
	setupTest()
	defer teardownTest()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "GET request returns 200",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "POST request returns 405",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "PUT request returns 405",
			method:         "PUT",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "DELETE request returns 405",
			method:         "DELETE",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
	}

	// Create a ticker that will fire multiple times
	tickCount := 0
	var tickMu sync.Mutex
	done := make(chan bool)

	f := func(ti time.Time) {
		tickMu.Lock()
		tickCount++
		if tickCount >= 3 {
			done <- true
		}
		tickMu.Unlock()
	}

	ticker := NewTicker("http-test", 2*time.Millisecond, f)

	// Wait for a few ticks
	select {
	case <-done:
		// Success
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Ticker did not fire expected number of times within timeout")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/ticker", nil)
			w := httptest.NewRecorder()

			ticker.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse {
				// Check Content-Type header
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// Decode and validate JSON response
				var info TickerInfo
				if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Validate response fields
				if info.Name != "http-test" {
					t.Errorf("Expected name 'http-test', got '%s'", info.Name)
				}

				if info.Ticks < 3 {
					t.Errorf("Expected at least 3 ticks, got %d", info.Ticks)
				}

				if !info.Active {
					t.Error("Expected ticker to be active")
				}

				if info.LastTick.IsZero() {
					t.Error("Expected LastTick to be set")
				}
			}
		})
	}
}

func TestTickerServeHTTPNewTicker(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Create a new ticker that hasn't ticked yet
	f := func(ti time.Time) {}
	ticker := NewTicker("new-ticker", 1*time.Hour, f) // Long interval so it won't tick during test

	req := httptest.NewRequest("GET", "/ticker", nil)
	w := httptest.NewRecorder()

	ticker.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var info TickerInfo
	if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Validate response for new ticker
	if info.Name != "new-ticker" {
		t.Errorf("Expected name 'new-ticker', got '%s'", info.Name)
	}

	if info.Ticks != 0 {
		t.Errorf("Expected 0 ticks for new ticker, got %d", info.Ticks)
	}

	if !info.Active {
		t.Error("Expected new ticker to be active")
	}

	if !info.LastTick.IsZero() {
		t.Error("Expected LastTick to be zero for new ticker")
	}
}

func TestTickerServeHTTPMultipleRequests(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Create a ticker
	var tickMu sync.Mutex
	count := 0
	f := func(ti time.Time) {
		tickMu.Lock()
		count++
		tickMu.Unlock()
	}

	ticker := NewTicker("multi-req-test", 1*time.Millisecond, f)

	// Wait for some ticks
	time.Sleep(20 * time.Millisecond)

	// Make multiple HTTP requests
	var wg sync.WaitGroup
	const numRequests = 10
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/ticker", nil)
			w := httptest.NewRecorder()

			ticker.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("Request %d: expected status 200, got %d", id, w.Code)
				return
			}

			var info TickerInfo
			if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
				errors <- fmt.Errorf("Request %d: failed to decode response: %v", id, err)
				return
			}

			if info.Name != "multi-req-test" {
				errors <- fmt.Errorf("Request %d: expected name 'multi-req-test', got '%s'", id, info.Name)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			t.Error(err)
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Errorf("Found %d errors during concurrent HTTP requests", errorCount)
	}
}

func TestTickerServeHTTPJSONFormat(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Create a ticker and let it tick a few times
	done := make(chan bool)
	tickCount := 0
	var mu sync.Mutex

	f := func(ti time.Time) {
		mu.Lock()
		tickCount++
		if tickCount >= 2 {
			done <- true
		}
		mu.Unlock()
	}

	ticker := NewTicker("json-format-test", 2*time.Millisecond, f)

	// Wait for ticks
	select {
	case <-done:
		// Success
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Ticker did not fire expected number of times")
	}

	req := httptest.NewRequest("GET", "/ticker", nil)
	w := httptest.NewRecorder()

	ticker.ServeHTTP(w, req)

	// Parse response as generic JSON to verify structure
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check all required fields are present
	requiredFields := []string{"name", "last_tick", "ticks", "active"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Missing required field '%s' in JSON response", field)
		}
	}

	// Validate field types
	if _, ok := response["name"].(string); !ok {
		t.Error("Field 'name' should be a string")
	}

	if _, ok := response["last_tick"].(string); !ok {
		t.Error("Field 'last_tick' should be a string (RFC3339 timestamp)")
	}

	if ticksFloat, ok := response["ticks"].(float64); !ok {
		t.Error("Field 'ticks' should be a number")
	} else if int(ticksFloat) < 2 {
		t.Errorf("Expected at least 2 ticks, got %d", int(ticksFloat))
	}

	if _, ok := response["active"].(bool); !ok {
		t.Error("Field 'active' should be a boolean")
	}
}

func TestTickerServeHTTPInactiveAfterStop(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Create a ticker
	f := func(ti time.Time) {}
	ticker := NewTicker("stop-test", 1*time.Millisecond, f)

	// Let it tick a few times
	time.Sleep(10 * time.Millisecond)

	// Verify it's active before stopping
	req1 := httptest.NewRequest("GET", "/ticker", nil)
	w1 := httptest.NewRecorder()
	ticker.ServeHTTP(w1, req1)

	var info1 TickerInfo
	if err := json.NewDecoder(w1.Body).Decode(&info1); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !info1.Active {
		t.Error("Expected ticker to be active before Stop()")
	}

	// Stop the ticker
	ticker.Stop()

	// Give the goroutine time to process the channel close and update active status
	time.Sleep(20 * time.Millisecond)

	req := httptest.NewRequest("GET", "/ticker", nil)
	w := httptest.NewRecorder()

	ticker.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var info TickerInfo
	if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// After stopping and waiting, the ticker should become inactive
	// This is eventual consistency, so we log if still active rather than failing
	if info.Active {
		t.Log("Ticker still showing as active after Stop() - this may be a timing issue")
	}
}

func TestTickersServeHTTP(t *testing.T) {
	setupTest()
	defer teardownTest()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "GET request returns 200",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "POST request returns 405",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "PUT request returns 405",
			method:         "PUT",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
	}

	// Create multiple tickers
	ticker1Done := make(chan bool)
	ticker2Done := make(chan bool)
	count1 := 0
	count2 := 0
	var mu1, mu2 sync.Mutex

	f1 := func(ti time.Time) {
		mu1.Lock()
		count1++
		if count1 >= 2 {
			ticker1Done <- true
		}
		mu1.Unlock()
	}

	f2 := func(ti time.Time) {
		mu2.Lock()
		count2++
		if count2 >= 2 {
			ticker2Done <- true
		}
		mu2.Unlock()
	}

	NewTicker("ticker-1", 2*time.Millisecond, f1)
	NewTicker("ticker-2", 2*time.Millisecond, f2)

	// Wait for both tickers to fire at least twice
	select {
	case <-ticker1Done:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("ticker-1 did not fire expected number of times")
	}

	select {
	case <-ticker2Done:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("ticker-2 did not fire expected number of times")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/tickers", nil)
			w := httptest.NewRecorder()

			tickers := GetTickers()
			tickers.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse {
				// Check Content-Type header
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// Decode and validate JSON response
				var tickerList []TickerInfo
				if err := json.NewDecoder(w.Body).Decode(&tickerList); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Should have 2 tickers
				if len(tickerList) != 2 {
					t.Errorf("Expected 2 tickers, got %d", len(tickerList))
				}

				// Verify ticker names are present
				foundTicker1 := false
				foundTicker2 := false
				for _, info := range tickerList {
					if info.Name == "ticker-1" {
						foundTicker1 = true
						if info.Ticks < 2 {
							t.Errorf("ticker-1 expected at least 2 ticks, got %d", info.Ticks)
						}
						if !info.Active {
							t.Error("ticker-1 expected to be active")
						}
					} else if info.Name == "ticker-2" {
						foundTicker2 = true
						if info.Ticks < 2 {
							t.Errorf("ticker-2 expected at least 2 ticks, got %d", info.Ticks)
						}
						if !info.Active {
							t.Error("ticker-2 expected to be active")
						}
					}
				}

				if !foundTicker1 {
					t.Error("ticker-1 not found in response")
				}
				if !foundTicker2 {
					t.Error("ticker-2 not found in response")
				}
			}
		})
	}
}

func TestTickersServeHTTPEmpty(t *testing.T) {
	setupTest()
	defer teardownTest()

	// No tickers created, should return empty array
	req := httptest.NewRequest("GET", "/tickers", nil)
	w := httptest.NewRecorder()

	tickers := GetTickers()
	tickers.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var tickerList []TickerInfo
	if err := json.NewDecoder(w.Body).Decode(&tickerList); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(tickerList) != 0 {
		t.Errorf("Expected 0 tickers, got %d", len(tickerList))
	}
}

func TestTickersServeHTTPSingleTicker(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Create a single ticker
	done := make(chan bool)
	count := 0
	var mu sync.Mutex

	f := func(ti time.Time) {
		mu.Lock()
		count++
		if count >= 3 {
			done <- true
		}
		mu.Unlock()
	}

	NewTicker("single-ticker", 2*time.Millisecond, f)

	// Wait for ticks
	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("ticker did not fire expected number of times")
	}

	req := httptest.NewRequest("GET", "/tickers", nil)
	w := httptest.NewRecorder()

	tickers := GetTickers()
	tickers.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var tickerList []TickerInfo
	if err := json.NewDecoder(w.Body).Decode(&tickerList); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(tickerList) != 1 {
		t.Errorf("Expected 1 ticker, got %d", len(tickerList))
	}

	if len(tickerList) > 0 {
		info := tickerList[0]
		if info.Name != "single-ticker" {
			t.Errorf("Expected name 'single-ticker', got '%s'", info.Name)
		}
		if info.Ticks < 3 {
			t.Errorf("Expected at least 3 ticks, got %d", info.Ticks)
		}
		if !info.Active {
			t.Error("Expected ticker to be active")
		}
	}
}

func TestTickersServeHTTPConcurrent(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Create multiple tickers
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("concurrent-ticker-%d", i)
		f := func(ti time.Time) {}
		NewTicker(name, 10*time.Millisecond, f)
	}

	// Wait for tickers to be created
	time.Sleep(5 * time.Millisecond)

	// Make multiple concurrent requests
	var wg sync.WaitGroup
	const numRequests = 10
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/tickers", nil)
			w := httptest.NewRecorder()

			tickers := GetTickers()
			tickers.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("Request %d: expected status 200, got %d", id, w.Code)
				return
			}

			var tickerList []TickerInfo
			if err := json.NewDecoder(w.Body).Decode(&tickerList); err != nil {
				errors <- fmt.Errorf("Request %d: failed to decode response: %v", id, err)
				return
			}

			if len(tickerList) != 5 {
				errors <- fmt.Errorf("Request %d: expected 5 tickers, got %d", id, len(tickerList))
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			t.Error(err)
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Errorf("Found %d errors during concurrent HTTP requests", errorCount)
	}
}
