package device

import (
	"fmt"
	"sync"
	"testing"
)

// mockDevice implements the Name interface for testing
type mockDevice struct {
	name string
}

func (m *mockDevice) Name() string {
	return m.name
}

func TestGetDeviceManager(t *testing.T) {
	t.Run("singleton instance", func(t *testing.T) {
		dm1 := GetDeviceManager()
		dm2 := GetDeviceManager()

		if dm1 == nil {
			t.Fatal("GetDeviceManager() returned nil")
		}
		if dm2 == nil {
			t.Fatal("GetDeviceManager() returned nil on second call")
		}
		if dm1 != dm2 {
			t.Error("GetDeviceManager() returned different instances")
		}
	})

	t.Run("initial state", func(t *testing.T) {
		dm := GetDeviceManager()
		if len(dm.devices) != 0 {
			t.Errorf("new DeviceManager should have empty devices map, got %d devices", len(dm.devices))
		}
	})
}

func TestDeviceManager_Add(t *testing.T) {
	tests := []struct {
		name    string
		device  Name
		wantErr bool
	}{
		{
			name:    "valid device",
			device:  &mockDevice{name: "test1"},
			wantErr: false,
		},
		{
			name:    "nil device",
			device:  nil,
			wantErr: true,
		},
		{
			name:    "duplicate name",
			device:  &mockDevice{name: "test1"},
			wantErr: false, // Should replace existing device
		},
	}

	dm := GetDeviceManager()
	dm.Clear() // Start with clean state

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dm.Add(tt.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.device != nil {
				if got, exists := dm.Get(tt.device.Name()); !exists || got != tt.device {
					t.Errorf("Add() failed to store device %s", tt.device.Name())
				}
			}
		})
	}
}

func TestDeviceManager_Get(t *testing.T) {
	dm := GetDeviceManager()
	dm.Clear()

	device := &mockDevice{name: "test"}
	dm.Add(device)

	tests := []struct {
		name       string
		deviceName string
		want       Name
		wantExists bool
	}{
		{
			name:       "existing device",
			deviceName: "test",
			want:       device,
			wantExists: true,
		},
		{
			name:       "non-existing device",
			deviceName: "unknown",
			want:       nil,
			wantExists: false,
		},
		{
			name:       "empty name",
			deviceName: "",
			want:       nil,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := dm.Get(tt.deviceName)
			if exists != tt.wantExists {
				t.Errorf("Get() exists = %v, want %v", exists, tt.wantExists)
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceManager_Remove(t *testing.T) {
	dm := GetDeviceManager()
	dm.Clear()

	device := &mockDevice{name: "test"}
	dm.Add(device)

	tests := []struct {
		name       string
		deviceName string
		want       bool
	}{
		{
			name:       "existing device",
			deviceName: "test",
			want:       true,
		},
		{
			name:       "non-existing device",
			deviceName: "unknown",
			want:       false,
		},
		{
			name:       "empty name",
			deviceName: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dm.Remove(tt.deviceName); got != tt.want {
				t.Errorf("Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceManager_List(t *testing.T) {
	dm := GetDeviceManager()
	dm.Clear()

	// Add some test devices
	devices := []string{"dev1", "dev2", "dev3"}
	for _, name := range devices {
		dm.Add(&mockDevice{name: name})
	}

	got := dm.List()
	if len(got) != len(devices) {
		t.Errorf("List() returned %d devices, want %d", len(got), len(devices))
	}

	// Check all devices are in the list
	for _, name := range devices {
		found := false
		for _, n := range got {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List() missing device %s", name)
		}
	}
}

func TestDeviceManager_ConcurrentAccess(t *testing.T) {
	dm := GetDeviceManager()
	dm.Clear()

	var wg sync.WaitGroup
	deviceCount := 100

	// Test concurrent adds
	wg.Add(deviceCount)
	for i := 0; i < deviceCount; i++ {
		go func(id int) {
			defer wg.Done()
			device := &mockDevice{name: fmt.Sprintf("dev%d", id)}
			dm.Add(device)
		}(i)
	}
	wg.Wait()

	// Verify all devices were added
	if len(dm.List()) != deviceCount {
		t.Errorf("Expected %d devices after concurrent adds, got %d", deviceCount, len(dm.List()))
	}
}
