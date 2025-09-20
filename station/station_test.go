package station

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewStation(t *testing.T) {
	resetStations()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid station creation",
			id:      "test-station",
			wantErr: false,
		},
		{
			name:    "empty ID should fail",
			id:      "",
			wantErr: true,
		},
		{
			name:    "duplicate ID should fail",
			id:      "test-station", // Same as first test
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			station, err := NewStation(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if station.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, station.ID)
			}

			if station.Metrics == nil {
				t.Error("Metrics should be initialized")
			}
		})
	}
}

func TestStationInit(t *testing.T) {
	resetStations()

	station, err := NewStation("init-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	// Test initialization
	station.Init()

	// Verify network information was gathered
	if station.Hostname == "" {
		t.Error("Hostname should be set after Init()")
	}

	// Verify metrics were updated
	metrics := station.Metrics.GetMetrics()
	if metrics.NetworkInterfaceCount == 0 && metrics.IPAddressCount == 0 {
		// This might be expected in test environment
		t.Log("No network interfaces found (expected in test environment)")
	}
}

func TestStationSayHello(t *testing.T) {
	resetStations()

	station, err := NewStation("hello-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	initialAnnouncements := station.Metrics.GetMetrics().AnnouncementsSent

	// Test saying hello
	station.SayHello()

	// Verify metrics were updated
	metrics := station.Metrics.GetMetrics()
	if metrics.AnnouncementsSent != initialAnnouncements+1 {
		t.Errorf("Expected announcements to increase by 1, got %d",
			metrics.AnnouncementsSent-initialAnnouncements)
	}

	if station.LastHeard.IsZero() {
		t.Error("LastHeard should be updated after SayHello()")
	}
}

func TestStationTicker(t *testing.T) {
	resetStations()

	station, err := NewStation("ticker-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	// Test starting ticker
	err = station.StartTicker(50 * time.Millisecond)
	if err != nil {
		t.Errorf("Failed to start ticker: %v", err)
	}

	// Wait for a few ticks
	time.Sleep(200 * time.Millisecond)

	// Stop the station
	station.Stop()

	// Verify announcements were sent
	metrics := station.Metrics.GetMetrics()
	if metrics.AnnouncementsSent == 0 {
		t.Error("Expected some announcements to be sent")
	}
}

func TestStationHealthCheck(t *testing.T) {
	resetStations()

	station, err := NewStation("health-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	// Set recent LastHeard
	station.LastHeard = time.Now()

	// Should be healthy
	if !station.IsHealthy() {
		t.Error("Station should be healthy with recent LastHeard")
	}

	// Set old LastHeard
	station.LastHeard = time.Now().Add(-5 * time.Minute)

	// Should be unhealthy
	if station.IsHealthy() {
		t.Error("Station should be unhealthy with old LastHeard")
	}
}

func TestStationDeviceManagement(t *testing.T) {
	resetStations()

	station, err := NewStation("device-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	// Create a mock device
	mockDevice := &MockDevice{name: "test-device"}

	// Add device
	station.AddDevice(mockDevice)

	// Verify device was added
	retrieved := station.GetDevice("test-device")
	if retrieved != mockDevice {
		t.Error("Device was not properly added or retrieved")
	}

	// Verify metrics were updated
	metrics := station.Metrics.GetMetrics()
	if metrics.DeviceCount != 1 {
		t.Errorf("Expected device count 1, got %d", metrics.DeviceCount)
	}
}

func TestStationHTTPHandler(t *testing.T) {
	resetStations()

	station, err := NewStation("http-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	// Test GET request
	req, err := http.NewRequest("GET", "/api/station/http-test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	station.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Parse response
	var response Station
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response.ID != "http-test" {
		t.Errorf("Expected ID 'http-test', got '%s'", response.ID)
	}
}

func TestStationMetrics(t *testing.T) {
	resetStations()

	station, err := NewStation("metrics-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	// Test various metric operations
	station.Metrics.RecordAnnouncement()
	station.Metrics.RecordMessageSent(100)
	station.Metrics.RecordMessageReceived(200)
	station.Metrics.RecordError()
	station.Metrics.RecordResponseTime(50 * time.Millisecond)
	station.Metrics.RecordHealthCheck(true)

	// Get metrics
	metrics := station.Metrics.GetMetrics()

	// Verify all metrics were recorded
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"AnnouncementsSent", metrics.AnnouncementsSent, uint64(1)},
		{"MessagesSent", metrics.MessagesSent, uint64(1)},
		{"MessagesSentBytes", metrics.MessagesSentBytes, uint64(100)},
		{"MessagesReceived", metrics.MessagesReceived, uint64(1)},
		{"MessagesReceivedBytes", metrics.MessagesReceivedBytes, uint64(200)},
		{"ErrorCount", metrics.ErrorCount, uint64(1)},
		{"HealthCheckCount", metrics.HealthCheckCount, uint64(1)},
		{"HealthyChecks", metrics.HealthyChecks, uint64(1)},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
		}
	}
}

func TestStationConcurrency(t *testing.T) {
	resetStations()

	const numRoutines = 50
	const operationsPerRoutine = 100

	station, err := NewStation("concurrency-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	var wg sync.WaitGroup

	// Test concurrent metric updates
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				station.Metrics.RecordAnnouncement()
				station.Metrics.RecordMessageSent(uint64(j))
				station.Metrics.RecordHealthCheck(j%2 == 0)
			}
		}()
	}

	wg.Wait()

	// Verify final counts
	metrics := station.Metrics.GetMetrics()
	expectedCount := uint64(numRoutines * operationsPerRoutine)

	if metrics.AnnouncementsSent != expectedCount {
		t.Errorf("Expected %d announcements, got %d",
			expectedCount, metrics.AnnouncementsSent)
	}

	if metrics.HealthCheckCount != expectedCount {
		t.Errorf("Expected %d health checks, got %d",
			expectedCount, metrics.HealthCheckCount)
	}
}

// Mock device for testing
type MockDevice struct {
	name string
}

func (m *MockDevice) Name() string {
	return m.name
}

func TestStationJSON(t *testing.T) {
	resetStations()

	station, err := NewStation("json-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	station.LastHeard = time.Now()
	station.Hostname = "test-hostname"

	// Test JSON marshaling
	data, err := json.Marshal(station)
	if err != nil {
		t.Errorf("Failed to marshal station: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Station
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal station: %v", err)
	}

	// Verify critical fields
	if unmarshaled.ID != station.ID {
		t.Errorf("Expected ID %s, got %s", station.ID, unmarshaled.ID)
	}

	if unmarshaled.Hostname != station.Hostname {
		t.Errorf("Expected hostname %s, got %s", station.Hostname, unmarshaled.Hostname)
	}
}

func TestStationErrorHandling(t *testing.T) {
	resetStations()

	station, err := NewStation("error-test")
	if err != nil {
		t.Fatalf("Failed to create station: %v", err)
	}

	// Test error handling
	testError := fmt.Errorf("test error")
	station.SaveError(testError)

	// Give some time for error handler to process
	time.Sleep(10 * time.Millisecond)

	// Verify error was recorded in metrics
	metrics := station.Metrics.GetMetrics()
	if metrics.ErrorCount == 0 {
		t.Error("Error should have been recorded in metrics")
	}
}
