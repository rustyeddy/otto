package device

import (
	"testing"
)

// Mock Name implementation for testing
type mockDevice struct {
	name string
}

func (m *mockDevice) Name() string {
	return m.name
}

func TestGetDeviceManager_Singleton(t *testing.T) {
	dm1 := GetDeviceManager()
	dm2 := GetDeviceManager()

	if dm1 != dm2 {
		t.Error("GetDeviceManager should return the same instance (singleton)")
	}
}

func TestAddAndGetDevice(t *testing.T) {
	dm := &DeviceManager{}
	dev := &mockDevice{name: "testDevice"}

	dm.Add(dev)
	got, exists := dm.Get("testDevice")

	if !exists {
		t.Error("Device should exist after being added")
	}
	if got.Name() != "testDevice" {
		t.Errorf("Expected device name 'testDevice', got '%s'", got.Name())
	}
}

func TestAddDeviceNilMap(t *testing.T) {
	dm := &DeviceManager{Devices: nil}
	dev := &mockDevice{name: "foo"}

	dm.Add(dev)

	got, exists := dm.Get("foo")
	if !exists || got.Name() != "foo" {
		t.Error("Device not added correctly when Devices map was nil")
	}
}

func TestGetNonexistentDevice(t *testing.T) {
	dm := &DeviceManager{}
	_, exists := dm.Get("doesnotexist")
	if exists {
		t.Error("Should not find a nonexistent device")
	}
}
