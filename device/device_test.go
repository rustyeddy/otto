package device

import (
	"testing"
)

func TestMock(t *testing.T) {
	mocks := []bool{true, false}

	if mock != false {
		t.Error("Expected mock (false) got (true)")
	}

	for _, b := range mocks {
		Mock(b)
		if IsMock() != b {
			t.Errorf("Expected mock (%t) got (%t)", b, mock)
		}
	}
}

func TestNewDevice(t *testing.T) {
	name := "test-device"
	dev := NewDevice(name)
	if dev.Name() != name {
		t.Errorf("Expected test device name (%s) got (%s)", name, dev.Name())
	}

	str := dev.String()
	if str == "" {
		t.Errorf("Expected device string to not be null but it was")
	}

	jbytes, err := dev.JSON()
	if err != nil {
		t.Error("Error creating JSON from device: ", err)
	}

	if len(jbytes) == 0 {
		t.Errorf("Expected JSON bytes but got (0)")
	}
}

func TestDevicePeriod(t *testing.T) {
	panic("test-device-period")
}

func TestDeviceMessanger(t *testing.T) {
	panic("test-device-messanger")
}

func TestDeviceString(t *testing.T) {
	panic("test-device-string")
}
