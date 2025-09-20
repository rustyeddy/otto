package utils

import (
	"fmt"
	"sync"
	"testing"
)

func TestStationInit(t *testing.T) {
	// Test that the station name is initialized to "station" by default
	// We need to reset to initial state first
	SetStation("station")

	if Station() != "station" {
		t.Errorf("Expected default station name 'station', got '%s'", Station())
	}
}

func TestSetStation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal name", "test-station", "test-station"},
		{"empty string", "", ""},
		{"single character", "x", "x"},
		{"with spaces", "my station", "my station"},
		{"with special chars", "station-1@home", "station-1@home"},
		{"very long name", "this-is-a-very-long-station-name-that-could-be-used-in-some-scenarios", "this-is-a-very-long-station-name-that-could-be-used-in-some-scenarios"},
		{"unicode characters", "测试站点", "测试站点"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetStation(tt.input)
			result := Station()
			if result != tt.expected {
				t.Errorf("SetStation(%q): expected %q, got %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestStation(t *testing.T) {
	// Test that Station returns the correct value after setting
	testNames := []string{"station1", "station2", "final-station"}

	for _, name := range testNames {
		SetStation(name)
		result := Station()
		if result != name {
			t.Errorf("After SetStation(%q), Station() returned %q", name, result)
		}
	}
}

func TestStationSequence(t *testing.T) {
	// Test multiple sequential operations
	originalName := Station()

	// Set and verify multiple times
	names := []string{"first", "second", "third", ""}

	for i, name := range names {
		SetStation(name)
		if Station() != name {
			t.Errorf("Step %d: expected station name %q, got %q", i+1, name, Station())
		}
	}

	// Restore original name
	SetStation(originalName)
}

func TestStationConcurrency(t *testing.T) {
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	errors := make(chan string, numGoroutines*numOperations)

	// Test concurrent reads and writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Mix of reads and writes
				if j%2 == 0 {
					// Write operation
					name := fmt.Sprintf("station-%d-%d", id, j)
					SetStation(name)
				} else {
					// Read operation
					name := Station()
					if name == "" {
						// This is actually valid, but we'll track it
					}
					// We can't assert specific values due to race conditions,
					// but we can ensure no panics occur
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check if any errors occurred
	for err := range errors {
		t.Error(err)
	}

	// Ensure the station name is still accessible after concurrent operations
	finalName := Station()
	if finalName == "" {
		// This is technically valid but unexpected in normal usage
		t.Log("Station name is empty after concurrent operations")
	}

	// Reset to a known state
	SetStation("station")
}

func TestStationImmutability(t *testing.T) {
	// Test that returned string cannot affect internal state
	SetStation("original")

	name1 := Station()
	name2 := Station()

	if name1 != name2 {
		t.Errorf("Multiple calls to Station() returned different values: %q vs %q", name1, name2)
	}

	if name1 != "original" {
		t.Errorf("Expected station name 'original', got %q", name1)
	}

	// Verify the value is still correct
	if Station() != "original" {
		t.Errorf("Station name changed unexpectedly to %q", Station())
	}
}

func TestStationStatePreservation(t *testing.T) {
	// Test that station name is preserved across multiple function calls
	testName := "persistent-station"
	SetStation(testName)

	// Call Station multiple times
	for i := 0; i < 10; i++ {
		if Station() != testName {
			t.Errorf("Call %d: expected station name %q, got %q", i+1, testName, Station())
		}
	}
}

func TestStationNilSafety(t *testing.T) {
	// Test behavior with potential edge cases
	// Go strings are safe, but let's test empty and whitespace
	testCases := []string{
		"",
		" ",
		"\t",
		"\n",
		"\r\n",
		"   whitespace   ",
	}

	for _, testCase := range testCases {
		SetStation(testCase)
		result := Station()
		if result != testCase {
			t.Errorf("SetStation(%q) -> Station() = %q, expected %q", testCase, result, testCase)
		}
	}
}

func BenchmarkSetStation(b *testing.B) {
	name := "benchmark-station"
	for i := 0; i < b.N; i++ {
		SetStation(name)
	}
}

func BenchmarkStation(b *testing.B) {
	SetStation("benchmark-station")
	for i := 0; i < b.N; i++ {
		_ = Station()
	}
}

func BenchmarkStationConcurrent(b *testing.B) {
	SetStation("concurrent-benchmark")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Station()
		}
	})
}
