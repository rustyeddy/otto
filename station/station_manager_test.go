package station

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	utils.SetStationName("test-station")
}

func TestNewStationManager(t *testing.T) {
	sm := NewStationManager()

	assert.NotNil(t, sm, "NewStationManager() should not return nil")
	assert.NotNil(t, sm.Stations, "Stations map should be initialized")
	assert.NotNil(t, sm.Stale, "Stale map should be initialized")
	assert.NotNil(t, sm.EventQ, "EventQ channel should be initialized")
	assert.NotNil(t, sm.mu, "Mutex should be initialized")
	assert.Equal(t, 0, len(sm.Stations), "Stations map should be empty initially")
	assert.Equal(t, 0, len(sm.Stale), "Stale map should be empty initially")
}

func TestGetStationManager(t *testing.T) {
	// Reset for clean test
	stations = nil

	// Test singleton behavior
	sm1 := GetStationManager()
	sm2 := GetStationManager()

	assert.Same(t, sm1, sm2, "GetStationManager() should return the same instance")
	assert.NotNil(t, sm1, "First call should return valid instance")
	assert.NotNil(t, sm2, "Second call should return valid instance")
}

func TestStationManagerAdd(t *testing.T) {
	sm := NewStationManager()

	t.Run("Add valid station", func(t *testing.T) {
		stationID := "test-station-1"
		station, err := sm.Add(stationID)

		assert.NoError(t, err, "Add() should not return error for valid ID")
		assert.NotNil(t, station, "Add() should return a station")
		assert.Equal(t, stationID, station.ID, "Station ID should match input")
		assert.Equal(t, 1, len(sm.Stations), "Stations map should contain one station")
		assert.Same(t, station, sm.Stations[stationID], "Station should be stored in map")
	})

	t.Run("Add empty station ID", func(t *testing.T) {
		station, err := sm.Add("")

		assert.Error(t, err, "Add() should return error for empty ID")
		assert.Nil(t, station, "Add() should return nil station for empty ID")
	})

	t.Run("Add duplicate station", func(t *testing.T) {
		stationID := "duplicate-test"

		// Add first time
		station1, err1 := sm.Add(stationID)
		assert.NoError(t, err1, "First add should succeed")
		assert.NotNil(t, station1, "First add should return station")

		// Add duplicate
		station2, err2 := sm.Add(stationID)
		assert.Error(t, err2, "Duplicate add should return error")
		assert.Nil(t, station2, "Duplicate add should return nil station")
		assert.Contains(t, err2.Error(), "existing station", "Error should mention existing station")
	})
}

func TestStationManagerGet(t *testing.T) {
	sm := NewStationManager()
	stationID := "test-station-get"

	t.Run("Get existing station", func(t *testing.T) {
		// Add a station first
		addedStation, err := sm.Add(stationID)
		require.NoError(t, err)
		require.NotNil(t, addedStation)

		// Test getting existing station
		retrieved := sm.Get(stationID)
		assert.Same(t, addedStation, retrieved, "Get() should return the same station instance")
	})

	t.Run("Get non-existent station", func(t *testing.T) {
		notFound := sm.Get("non-existent-station")
		assert.Nil(t, notFound, "Get() should return nil for non-existent station")
	})

	t.Run("Get with empty ID", func(t *testing.T) {
		notFound := sm.Get("")
		assert.Nil(t, notFound, "Get() should return nil for empty ID")
	})
}

func TestStationManagerCount(t *testing.T) {
	sm := NewStationManager()

	t.Run("Empty manager", func(t *testing.T) {
		assert.Equal(t, 0, sm.Count(), "Count should be 0 for empty manager")
	})

	t.Run("Add stations and count", func(t *testing.T) {
		stationIDs := []string{"station1", "station2", "station3"}

		for i, id := range stationIDs {
			station, err := sm.Add(id)
			assert.NoError(t, err, "Add should succeed for station %s", id)
			assert.NotNil(t, station, "Station should not be nil for %s", id)
			assert.Equal(t, i+1, sm.Count(), "Count should be %d after adding %s", i+1, id)
		}

		assert.Equal(t, len(stationIDs), sm.Count(), "Final count should match number of added stations")
	})
}

func TestStationManagerUpdate(t *testing.T) {
	sm := NewStationManager()
	stationID := "update-test"

	// Add a station first
	station, err := sm.Add(stationID)
	require.NoError(t, err)
	require.NotNil(t, station)

	t.Run("Update existing station", func(t *testing.T) {
		// Create a message with station in path
		msg := messanger.NewMsg("ss/d/update-test/sensor", []byte(`{"temperature": 25.5}`), "TestStationManagerUpdate")

		initialTime := station.LastHeard
		time.Sleep(10 * time.Millisecond) // Ensure time difference

		updatedStation := sm.Update(msg)

		assert.NotNil(t, updatedStation, "Update should return a station")
		assert.Same(t, station, updatedStation, "Update should return the same station instance")
		assert.True(t, station.LastHeard.After(initialTime), "LastHeard should be updated")
	})

	t.Run("Update creates new station", func(t *testing.T) {
		newStationID := "new-station-from-update"
		msg := messanger.NewMsg(fmt.Sprintf("ss/d/%s/sensor", newStationID), []byte(`{"humidity": 60}`), "test-update-stations")
		updatedStation := sm.Update(msg)

		assert.NotNil(t, updatedStation, "Update should create and return new station")
		assert.Equal(t, newStationID, updatedStation.ID, "New station should have correct ID")
		assert.NotNil(t, sm.Get(newStationID), "New station should be in manager")
	})

	t.Run("Update with invalid path", func(t *testing.T) {
		msg := messanger.NewMsg("invalid/path/without/station", []byte(`{"data": "test"}`), "invalid-topic-test")
		updatedStation := sm.Update(msg)
		assert.Nil(t, updatedStation, "Update should return nil for invalid path")
	})
}

func TestStationManagerCallback(t *testing.T) {
	sm := NewStationManager()

	t.Run("Callback updates station", func(t *testing.T) {
		msg := messanger.NewMsg(messanger.GetTopics().Data("hello"), []byte("test-sm-callback"), "test-callback-updates-station")
		sm.Callback(msg)

		station := sm.Get(utils.StationName())
		assert.NotNil(t, station, "Station should be created via callback")
		assert.Equal(t, utils.StationName(), station.ID, "Station should have correct ID")
	})

	t.Run("Callback with malformed JSON", func(t *testing.T) {
		msg := messanger.NewMsg("data/malformed-test/hello", []byte(`{invalid json`), "test-invalid-json-and-topic")
		// This should not panic
		assert.NotPanics(t, func() {
			sm.Callback(msg)
		}, "Callback should handle malformed JSON gracefully")
	})
}

func TestStationManagerConcurrency(t *testing.T) {
	sm := NewStationManager()
	const numRoutines = 10
	const stationsPerRoutine = 5

	t.Run("Concurrent adds", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numRoutines*stationsPerRoutine)

		// Test concurrent adds
		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				for j := 0; j < stationsPerRoutine; j++ {
					stationID := fmt.Sprintf("concurrent-station-%d-%d", routineID, j)
					station, err := sm.Add(stationID)
					if err != nil {
						errors <- fmt.Errorf("failed to add station %s: %v", stationID, err)
					} else if station == nil {
						errors <- fmt.Errorf("station %s is nil", stationID)
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}

		expectedCount := numRoutines * stationsPerRoutine
		assert.Equal(t, expectedCount, sm.Count(), "All stations should be added concurrently")
	})

	t.Run("Concurrent gets", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numRoutines)

		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				stationID := fmt.Sprintf("concurrent-station-%d-0", routineID)
				station := sm.Get(stationID)
				if station == nil {
					errors <- fmt.Errorf("failed to get station %s", stationID)
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}
	})
}

func TestStationManagerServeHTTP(t *testing.T) {
	sm := NewStationManager()

	// Add some test stations
	station1, err1 := sm.Add("http-test-1")
	station2, err2 := sm.Add("http-test-2")
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, station1)
	require.NotNil(t, station2)

	t.Run("GET request", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/stations", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		sm.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Should return 200 OK")
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Should set JSON content type")

		// Parse response
		var response StationManager
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")
		assert.Equal(t, 2, len(response.Stations), "Response should contain 2 stations")
	})

	t.Run("POST request", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/stations", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		sm.ServeHTTP(rr, req)

		assert.Equal(t, 401, rr.Code, "POST should return 401 (not yet supported)")
	})

	t.Run("PUT request", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/api/stations", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		sm.ServeHTTP(rr, req)

		assert.Equal(t, 401, rr.Code, "PUT should return 401 (not yet supported)")
	})
}

func TestStationManagerCapacity(t *testing.T) {
	sm := NewStationManager()

	done := make(chan (interface{}))
	eventCount := 0

	// create a goroutine to snarf up the events
	go func() {
		for {
			select {
			case <-sm.EventQ:
				eventCount++
			case <-done:
				break
			}
		}
	}()

	// Test that we can queue events without blocking
	for i := 0; i < 10; i++ {
		event := &StationEvent{
			Type:      "test",
			Device:    "sensor1",
			StationID: fmt.Sprintf("station-%d", i),
			Value:     "test-value",
			Timestamp: time.Now(),
		}

		select {
		case sm.EventQ <- event:
			// Successfully queued
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Event queue blocked at event %d", i)
		}
	}
	done <- true

	assert.Equal(t, 10, eventCount)
}

func TestEventQProcessing(t *testing.T) {
	sm := NewStationManager()

	// Add a station first
	stationID := "event-test-station"
	station, err := sm.Add(stationID)
	require.NoError(t, err)
	require.NotNil(t, station)

	event := &StationEvent{
		Type:      "sensor-reading",
		Device:    "temperature",
		StationID: stationID,
		Value:     "25.5",
		Timestamp: time.Now(),
	}
	sm.EventQ <- event
	
	// Verify the event was queued (channel should be non-empty)
	assert.Equal(t, 1, len(sm.EventQ), "Event should be queued")
}

func TestStationManagerEdgeCases(t *testing.T) {
	sm := NewStationManager()

	t.Run("Add with nil string handling", func(t *testing.T) {
		// Test adding empty string (Go doesn't have nil strings, but empty is equivalent)
		station, err := sm.Add("")
		assert.Error(t, err, "Should not allow empty station ID")
		assert.Nil(t, station, "Should return nil station for empty ID")
		assert.Equal(t, 0, sm.Count(), "Count should remain 0 after failed add")
	})

	t.Run("Get with empty string", func(t *testing.T) {
		station := sm.Get("")
		assert.Nil(t, station, "Get with empty string should return nil")
	})

	t.Run("Update with empty message", func(t *testing.T) {
		msg := messanger.NewMsg("", []byte{}, "test-empty-data")
		station := sm.Update(msg)
		assert.Nil(t, station, "Update with empty message should return nil")
	})

	t.Run("Callback with nil data", func(t *testing.T) {
		msg := messanger.NewMsg(messanger.GetTopics().Data("hello"), nil, "test-nil-data")
		assert.NotPanics(t, func() {
			sm.Callback(msg)
		}, "Callback should handle nil data gracefully")
	})
}

func TestStationManagerStart(t *testing.T) {
	sm := NewStationManager()

	t.Run("Start initializes ticker", func(t *testing.T) {
		// Note: We can't easily test the full Start() method because it
		// registers with the server and starts a goroutine.
		// In a real implementation, you might want to add dependency injection
		// or make the ticker period configurable for testing.

		assert.NotPanics(t, func() {
			// This would normally call sm.Start() but that requires server setup
			// For now, just verify the structure is sound
			assert.NotNil(t, sm.EventQ, "EventQ should be initialized")
			assert.NotNil(t, sm.mu, "Mutex should be initialized")
		}, "Start method structure should be sound")
	})
}

func TestStationManagerReset(t *testing.T) {
	// Add some stations
	sm := GetStationManager()
	station1, _ := sm.Add("reset-test-1")
	station2, _ := sm.Add("reset-test-2")
	require.NotNil(t, station1)
	require.NotNil(t, station2)
	require.Equal(t, 2, sm.Count())

	// Reset and verify
	resetStations()
	newSm := GetStationManager()

	assert.Equal(t, 0, newSm.Count(), "Count should be 0 after reset")
	assert.Empty(t, newSm.Stations, "Stations map should be empty after reset")
	assert.Empty(t, newSm.Stale, "Stale map should be empty after reset")
}
