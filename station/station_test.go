package station

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func StationCreation(count int) []string {
	ids := []string{
		"127.0.0.1",
		"127.0.0.2",
		"127.0.0.3",
		"127.0.0.4",
		"127.0.0.5",
	}

	sm := NewStationManager()
	for _, id := range ids {
		sm.Add(id)
	}
	return ids
}

func TestStation(t *testing.T) {
	resetStations() // Clean state for test

	localip := "127.0.0.1"
	st, err := NewStation(localip)
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	if st.ID != localip {
		t.Errorf("IP expecting (%s) got (%s)", localip, st.ID)
	}
}

func TestStationManager(t *testing.T) {
	resetStations() // Clean state for test

	count := 5
	sm := NewStationManager()
	sids := StationCreation(count)
	for _, id := range sids {
		sm.Add(id)
	}

	if sm.Count() != len(sids) {
		t.Errorf("Station Manager count got (%d) expected (%d)",
			sm.Count(), len(sids))
	}

	for _, id := range sids {
		st := sm.Get(id)
		if st == nil {
			t.Errorf("Get station expected (%s) got nil", id)
			continue
		}
		if st.ID != id {
			t.Errorf("Get station expected (%s) got (%s)", id, st.ID)
		}
	}
}

func TestStationJSON(t *testing.T) {
	resetStations() // Clean state for test

	st, err := NewStation("aa:bb:cc:dd:ee:11")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	st.LastHeard = time.Now()

	j, err := json.Marshal(st)
	if err != nil {
		t.Errorf("Marshal Station failed: %+v", err)
		return
	}

	var station Station
	err = json.Unmarshal(j, &station)
	if err != nil {
		t.Errorf("Unmarshal Station failed: %+v", err)
		return
	}

	if station.ID != st.ID {
		t.Errorf("Expected ID %s, got %s", st.ID, station.ID)
	}
}

func TestRecordHealthCheck(t *testing.T) {
	metrics := NewStationMetrics()

	// Test recording healthy check
	metrics.RecordHealthCheck(true)

	// Get current metrics to check values
	currentMetrics := metrics.GetMetrics()

	// Verify metrics were updated
	if currentMetrics.HealthCheckCount != 1 {
		t.Errorf("Expected HealthCheckCount to be 1, got %d", currentMetrics.HealthCheckCount)
	}
	if currentMetrics.HealthyChecks != 1 {
		t.Errorf("Expected HealthyChecks to be 1, got %d", currentMetrics.HealthyChecks)
	}
	if currentMetrics.UnhealthyChecks != 0 {
		t.Errorf("Expected UnhealthyChecks to be 0, got %d", currentMetrics.UnhealthyChecks)
	}

	// Test recording unhealthy check
	metrics.RecordHealthCheck(false)

	// Get updated metrics
	currentMetrics = metrics.GetMetrics()

	// Verify metrics were updated
	if currentMetrics.HealthCheckCount != 2 {
		t.Errorf("Expected HealthCheckCount to be 2, got %d", currentMetrics.HealthCheckCount)
	}
	if currentMetrics.HealthyChecks != 1 {
		t.Errorf("Expected HealthyChecks to be 1, got %d", currentMetrics.HealthyChecks)
	}
	if currentMetrics.UnhealthyChecks != 1 {
		t.Errorf("Expected UnhealthyChecks to be 1, got %d", currentMetrics.UnhealthyChecks)
	}
}

func TestRecordHealthCheckConcurrency(t *testing.T) {
	metrics := NewStationMetrics()

	// Test concurrent health check recording
	const numRoutines = 100
	const checksPerRoutine = 10

	var wg sync.WaitGroup

	// Start goroutines that record healthy checks
	for i := 0; i < numRoutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < checksPerRoutine; j++ {
				metrics.RecordHealthCheck(true)
			}
		}()
	}

	// Start goroutines that record unhealthy checks
	for i := 0; i < numRoutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < checksPerRoutine; j++ {
				metrics.RecordHealthCheck(false)
			}
		}()
	}

	wg.Wait()

	// Get final metrics
	finalMetrics := metrics.GetMetrics()

	// Verify final counts
	expectedTotal := uint64(numRoutines * checksPerRoutine)
	expectedHealthy := uint64(numRoutines / 2 * checksPerRoutine)
	expectedUnhealthy := uint64(numRoutines / 2 * checksPerRoutine)

	if finalMetrics.HealthCheckCount != expectedTotal {
		t.Errorf("Expected HealthCheckCount to be %d, got %d", expectedTotal, finalMetrics.HealthCheckCount)
	}
	if finalMetrics.HealthyChecks != expectedHealthy {
		t.Errorf("Expected HealthyChecks to be %d, got %d", expectedHealthy, finalMetrics.HealthyChecks)
	}
	if finalMetrics.UnhealthyChecks != expectedUnhealthy {
		t.Errorf("Expected UnhealthyChecks to be %d, got %d", expectedUnhealthy, finalMetrics.UnhealthyChecks)
	}
}

func TestRecordHealthCheckSequence(t *testing.T) {
	metrics := NewStationMetrics()

	// Test a sequence of health checks
	healthChecks := []bool{true, true, false, true, false, false, true}

	expectedHealthy := uint64(0)
	expectedUnhealthy := uint64(0)

	for i, healthy := range healthChecks {
		metrics.RecordHealthCheck(healthy)

		if healthy {
			expectedHealthy++
		} else {
			expectedUnhealthy++
		}

		// Get current metrics
		currentMetrics := metrics.GetMetrics()

		// Verify counts after each check
		expectedTotal := uint64(i + 1)
		if currentMetrics.HealthCheckCount != expectedTotal {
			t.Errorf("After check %d: Expected HealthCheckCount to be %d, got %d", i, expectedTotal, currentMetrics.HealthCheckCount)
		}
		if currentMetrics.HealthyChecks != expectedHealthy {
			t.Errorf("After check %d: Expected HealthyChecks to be %d, got %d", i, expectedHealthy, currentMetrics.HealthyChecks)
		}
		if currentMetrics.UnhealthyChecks != expectedUnhealthy {
			t.Errorf("After check %d: Expected UnhealthyChecks to be %d, got %d", i, expectedUnhealthy, currentMetrics.UnhealthyChecks)
		}
	}
}

func TestRecordHealthCheckHealthScore(t *testing.T) {
	metrics := NewStationMetrics()

	// Record some health checks
	metrics.RecordHealthCheck(true)  // 1/1 = 100%
	metrics.RecordHealthCheck(true)  // 2/2 = 100%
	metrics.RecordHealthCheck(false) // 2/3 = 66.67%
	metrics.RecordHealthCheck(true)  // 3/4 = 75%

	// Get metrics (which calls UpdateMetrics internally)
	currentMetrics := metrics.GetMetrics()

	expectedScore := float64(3) / float64(4) * 100 // 75%
	if currentMetrics.HealthScore != expectedScore {
		t.Errorf("Expected HealthScore to be %f, got %f", expectedScore, currentMetrics.HealthScore)
	}
}

func TestRecordHealthCheckZeroState(t *testing.T) {
	metrics := NewStationMetrics()

	// Get initial metrics
	initialMetrics := metrics.GetMetrics()

	// Verify initial state
	if initialMetrics.HealthCheckCount != 0 {
		t.Errorf("Expected initial HealthCheckCount to be 0, got %d", initialMetrics.HealthCheckCount)
	}
	if initialMetrics.HealthyChecks != 0 {
		t.Errorf("Expected initial HealthyChecks to be 0, got %d", initialMetrics.HealthyChecks)
	}
	if initialMetrics.UnhealthyChecks != 0 {
		t.Errorf("Expected initial UnhealthyChecks to be 0, got %d", initialMetrics.UnhealthyChecks)
	}
	if initialMetrics.HealthScore != 0 {
		t.Errorf("Expected initial HealthScore to be 0, got %f", initialMetrics.HealthScore)
	}
}

func TestStationMetrics(t *testing.T) {
	station, err := NewStation("test-station")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	// Test various metric recording functions
	station.Metrics.RecordAnnouncement()
	station.Metrics.RecordMessageSent(100)
	station.Metrics.RecordMessageReceived(200)
	station.Metrics.RecordError()
	station.Metrics.RecordResponseTime(50 * time.Millisecond)

	// Get current metrics
	metrics := station.Metrics.GetMetrics()

	// Verify metrics were recorded
	if metrics.AnnouncementsSent != 1 {
		t.Errorf("Expected AnnouncementsSent to be 1, got %d", metrics.AnnouncementsSent)
	}
	if metrics.MessagesSent != 1 {
		t.Errorf("Expected MessagesSent to be 1, got %d", metrics.MessagesSent)
	}
	if metrics.MessagesSentBytes != 100 {
		t.Errorf("Expected MessagesSentBytes to be 100, got %d", metrics.MessagesSentBytes)
	}
	if metrics.MessagesReceived != 1 {
		t.Errorf("Expected MessagesReceived to be 1, got %d", metrics.MessagesReceived)
	}
	if metrics.MessagesReceivedBytes != 200 {
		t.Errorf("Expected MessagesReceivedBytes to be 200, got %d", metrics.MessagesReceivedBytes)
	}
	if metrics.ErrorCount != 1 {
		t.Errorf("Expected ErrorCount to be 1, got %d", metrics.ErrorCount)
	}
}
