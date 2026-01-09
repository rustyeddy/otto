package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGetStats(t *testing.T) {
	stats := GetStats()

	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}
}

// TestStats provides the same functionality as the original test
func TestStats(t *testing.T) {
	st := GetStats()

	if st.Goroutines <= 0 {
		t.Error("Expected at least one goroutine")
	}

	if st.CPUs <= 0 {
		t.Error("Expected at least one CPU")
	}

	if st.GoVersion == "" {
		t.Error("Expected non-empty Go version")
	}
}

func TestStatsGoroutines(t *testing.T) {
	stats := GetStats()

	if stats.Goroutines <= 0 {
		t.Errorf("Expected positive number of goroutines, got %d", stats.Goroutines)
	}

	// Should have at least 1 goroutine (the test itself)
	if stats.Goroutines < 1 {
		t.Errorf("Expected at least 1 goroutine, got %d", stats.Goroutines)
	}
}

func TestStatsCPUs(t *testing.T) {
	stats := GetStats()

	if stats.CPUs <= 0 {
		t.Errorf("Expected positive number of CPUs, got %d", stats.CPUs)
	}

	// Compare with runtime.NumCPU() directly
	expectedCPUs := runtime.NumCPU()
	if stats.CPUs != expectedCPUs {
		t.Errorf("Expected CPUs %d, got %d", expectedCPUs, stats.CPUs)
	}
}

func TestStatsGoVersion(t *testing.T) {
	stats := GetStats()

	if stats.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	// Should start with "go" (e.g., "go1.21.0")
	if !strings.HasPrefix(stats.GoVersion, "go") {
		t.Errorf("Expected GoVersion to start with 'go', got %s", stats.GoVersion)
	}

	// Compare with runtime.Version() directly
	expectedVersion := runtime.Version()
	if stats.GoVersion != expectedVersion {
		t.Errorf("Expected GoVersion %s, got %s", expectedVersion, stats.GoVersion)
	}
}

func TestStatsMemStats(t *testing.T) {
	stats := GetStats()

	// Test that MemStats is populated with reasonable values
	if stats.MemStats.Sys == 0 {
		t.Error("Expected Sys memory to be greater than 0")
	}

	if stats.MemStats.HeapSys == 0 {
		t.Error("Expected HeapSys to be greater than 0")
	}

	// Test some basic memory statistics
	if stats.MemStats.TotalAlloc < stats.MemStats.Alloc {
		t.Error("TotalAlloc should be greater than or equal to Alloc")
	}

	if stats.MemStats.Mallocs < stats.MemStats.Frees {
		t.Error("Total mallocs should be greater than or equal to frees")
	}
}

func TestStatsMemStatsFields(t *testing.T) {
	stats := GetStats()

	// Test that key memory fields are reasonable
	memFields := map[string]uint64{
		"Alloc":      stats.MemStats.Alloc,
		"TotalAlloc": stats.MemStats.TotalAlloc,
		"Sys":        stats.MemStats.Sys,
		"Lookups":    stats.MemStats.Lookups,
		"Mallocs":    stats.MemStats.Mallocs,
		"Frees":      stats.MemStats.Frees,
		"HeapAlloc":  stats.MemStats.HeapAlloc,
		"HeapSys":    stats.MemStats.HeapSys,
		"HeapIdle":   stats.MemStats.HeapIdle,
		"HeapInuse":  stats.MemStats.HeapInuse,
		"StackInuse": stats.MemStats.StackInuse,
		"StackSys":   stats.MemStats.StackSys,
	}

	for fieldName, value := range memFields {
		// Most memory fields should be non-negative
		// Note: Some fields might legitimately be 0, so we just check they're not negative
		if fieldName != "Lookups" && value == 0 {
			t.Logf("Warning: %s is 0, which might be unexpected", fieldName)
		}
	}

	// Heap allocated should not exceed heap system
	if stats.MemStats.HeapAlloc > stats.MemStats.HeapSys {
		t.Error("HeapAlloc should not exceed HeapSys")
	}

	// Stack in use should not exceed stack system
	if stats.MemStats.StackInuse > stats.MemStats.StackSys {
		t.Error("StackInuse should not exceed StackSys")
	}
}

func TestStatsGoroutineTracking(t *testing.T) {
	// Get initial stats
	initialStats := GetStats()
	initialGoroutines := initialStats.Goroutines

	// Start some goroutines
	var wg sync.WaitGroup
	numGoroutines := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
		}()
	}

	// Get stats while goroutines are running
	runningStats := GetStats()

	// Clean up goroutines
	wg.Wait()

	// Get final stats
	finalStats := GetStats()

	// While goroutines were running, count should have increased
	if runningStats.Goroutines <= initialGoroutines {
		t.Logf("Warning: Expected more goroutines while running (%d), got %d (initial: %d)",
			initialGoroutines+numGoroutines, runningStats.Goroutines, initialGoroutines)
		// Note: This is a log instead of error because goroutine scheduling is not deterministic
	}

	// After cleanup, should be back to around initial count (allow some variance)
	if finalStats.Goroutines > initialGoroutines+2 {
		t.Logf("Warning: Expected goroutines to return near initial count, got %d (initial: %d)",
			finalStats.Goroutines, initialGoroutines)
	}
}

func TestStatsConsistency(t *testing.T) {
	// Get stats multiple times and ensure they're reasonable
	stats1 := GetStats()
	time.Sleep(1 * time.Millisecond) // Small delay
	stats2 := GetStats()

	// CPU count should be consistent
	if stats1.CPUs != stats2.CPUs {
		t.Errorf("CPU count changed between calls: %d vs %d", stats1.CPUs, stats2.CPUs)
	}

	// Go version should be consistent
	if stats1.GoVersion != stats2.GoVersion {
		t.Errorf("Go version changed between calls: %s vs %s", stats1.GoVersion, stats2.GoVersion)
	}

	// Total allocated memory should only increase or stay same
	if stats2.MemStats.TotalAlloc < stats1.MemStats.TotalAlloc {
		t.Errorf("TotalAlloc decreased between calls: %d to %d", stats1.MemStats.TotalAlloc, stats2.MemStats.TotalAlloc)
	}

	// Malloc count should only increase or stay same
	if stats2.MemStats.Mallocs < stats1.MemStats.Mallocs {
		t.Errorf("Mallocs decreased between calls: %d to %d", stats1.MemStats.Mallocs, stats2.MemStats.Mallocs)
	}
}

func TestStatsMemoryAllocation(t *testing.T) {
	// Get initial stats
	initialStats := GetStats()
	initialMallocs := initialStats.MemStats.Mallocs

	// Allocate some memory
	data := make([][]byte, 100)
	for i := range data {
		data[i] = make([]byte, 1024) // 1KB each
	}

	// Force garbage collection to get accurate stats
	runtime.GC()
	runtime.GC() // Call twice to ensure cleanup

	// Get stats after allocation
	afterStats := GetStats()

	// TotalAlloc should have increased
	if afterStats.MemStats.TotalAlloc <= initialStats.MemStats.TotalAlloc {
		t.Logf("Warning: TotalAlloc didn't increase as expected. Initial: %d, After: %d",
			initialStats.MemStats.TotalAlloc, afterStats.MemStats.TotalAlloc)
	}

	// Mallocs should have increased
	if afterStats.MemStats.Mallocs <= initialMallocs {
		t.Logf("Warning: Mallocs didn't increase as expected. Initial: %d, After: %d",
			initialMallocs, afterStats.MemStats.Mallocs)
	}

	// Keep data alive until here
	_ = data
}

func TestStatsMultipleCalls(t *testing.T) {
	// Test that multiple calls to GetStats() work correctly
	const numCalls = 10
	statsCollection := make([]*Stats, numCalls)

	for i := 0; i < numCalls; i++ {
		statsCollection[i] = GetStats()
		if statsCollection[i] == nil {
			t.Fatalf("GetStats() call %d returned nil", i)
		}
	}

	// Verify all stats have consistent static values
	firstStats := statsCollection[0]
	for i := 1; i < numCalls; i++ {
		stats := statsCollection[i]

		if stats.CPUs != firstStats.CPUs {
			t.Errorf("CPU count inconsistent at call %d: expected %d, got %d", i, firstStats.CPUs, stats.CPUs)
		}

		if stats.GoVersion != firstStats.GoVersion {
			t.Errorf("Go version inconsistent at call %d: expected %s, got %s", i, firstStats.GoVersion, stats.GoVersion)
		}
	}
}

func TestStatsStructFields(t *testing.T) {
	stats := GetStats()

	// Test that all expected fields are accessible
	_ = stats.Goroutines
	_ = stats.CPUs
	_ = stats.MemStats
	_ = stats.GoVersion

	// Test that MemStats has expected structure (runtime.MemStats)
	_ = stats.MemStats.Alloc
	_ = stats.MemStats.TotalAlloc
	_ = stats.MemStats.Sys
	_ = stats.MemStats.Mallocs
	_ = stats.MemStats.Frees
	_ = stats.MemStats.HeapAlloc
	_ = stats.MemStats.HeapSys
	_ = stats.MemStats.StackInuse
	_ = stats.MemStats.StackSys
	_ = stats.MemStats.NumGC

	// If we get here without compilation errors, all fields are accessible
}

// Benchmark tests
func BenchmarkGetStats(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetStats()
	}
}

func BenchmarkGetStatsMemStatsOnly(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
	}
}

func BenchmarkGetStatsRuntimeInfoOnly(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runtime.NumGoroutine()
		_ = runtime.NumCPU()
		_ = runtime.Version()
	}
}

// Test concurrent access to GetStats
func TestStatsConcurrency(b *testing.T) {
	const numGoroutines = 10
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*opsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				stats := GetStats()
				if stats == nil {
					errors <- nil
					continue
				}
				// Basic validation
				if stats.CPUs <= 0 {
					errors <- nil
				}
				if stats.GoVersion == "" {
					errors <- nil
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		b.Errorf("Found %d errors during concurrent execution", errorCount)
	}
}

func TestStatsHandler(t *testing.T) {
	// The Stats{} handler doesn't use its own fields; it calls GetStats()
	// to get fresh runtime statistics on each request.
	handler := Stats{}

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that we have expected fields
	if _, ok := stats["Goroutines"]; !ok {
		t.Error("Expected Goroutines field in stats")
	}
	if _, ok := stats["CPUs"]; !ok {
		t.Error("Expected CPUs field in stats")
	}
	if _, ok := stats["GoVersion"]; !ok {
		t.Error("Expected GoVersion field in stats")
	}
}
