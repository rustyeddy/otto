package device

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// MockOpener implements the Opener interface for testing
type MockOpener struct {
	openErr  error
	closeErr error
	opened   bool
}

func (m *MockOpener) Open() error {
	if m.openErr != nil {
		return m.openErr
	}
	m.opened = true
	return nil
}

func (m *MockOpener) Close() error {
	if m.closeErr != nil {
		return m.closeErr
	}
	m.opened = false
	return nil
}

func TestNewDevice(t *testing.T) {
	tests := []struct {
		name      string
		devName   string
		wantName  string
		wantState DeviceState
	}{
		{
			name:      "basic device",
			devName:   "test-device",
			wantName:  "test-device",
			wantState: StateUnknown,
		},
		{
			name:      "empty name",
			devName:   "",
			wantName:  "",
			wantState: StateUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			   d := NewDevice(tt.devName, "mqtt")
			if d.Name != tt.wantName {
				t.Errorf("NewDevice().Name() = %v, want %v", d.Name, tt.wantName)
			}
			if d.State != tt.wantState {
				t.Errorf("NewDevice().State() = %v, want %v", d.State, tt.wantState)
			}
		})
	}
}

func TestDeviceState(t *testing.T) {
	d := NewDevice("test-device", "mqtt")

	states := []DeviceState{
		StateInitializing,
		StateRunning,
		StateError,
		StateStopped,
	}

	for _, state := range states {
		d.State = state
		if got := d.State; got != state {
			t.Errorf("State() = %v, want %v", got, state)
		}
	}
}

func TestDeviceError(t *testing.T) {
       d := NewDevice("test-device", "mqtt")
       testErr := errors.New("test error")

       if got := d.Error(); got != nil {
	       t.Errorf("Initial Error() = %v, want nil", got)
       }

       d.SetError(testErr)
       if got := d.Error(); got != testErr {
	       t.Errorf("Error() = %v, want %v", got, testErr)
       }

       if got := d.State; got != StateError {
	       t.Errorf("State() after error = %v, want %v", got, StateError)
       }
}

func TestDeviceTimerLoop(t *testing.T) {
	tests := []struct {
		name    string
		period  time.Duration
		wantErr bool
	}{
		{
			name:    "valid period",
			period:  10 * time.Millisecond,
			wantErr: false,
		},
		{
			name:    "zero period",
			period:  0,
			wantErr: true,
		},
		{
			name:    "negative period",
			period:  -1 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			   d := NewDevice("test-device", "mqtt")
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			calls := 0
			err := d.TimerLoop(ctx, tt.period, func() error {
				calls++
				return nil
			})

			if tt.wantErr && err == nil {
				t.Error("TimerLoop() error = nil, want error")
			}
			if !tt.wantErr && err != context.DeadlineExceeded {
				t.Errorf("TimerLoop() error = %v, want context.DeadlineExceeded", err)
			}
			if !tt.wantErr && calls == 0 {
				t.Error("TimerLoop() made no calls to readpub")
			}
		})
	}
}

func TestDeviceJSON(t *testing.T) {
	d := NewDevice("test-device", "mqtt")
	testErr := errors.New("test error")
	d.SetError(testErr)

	data, err := d.JSON()
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	var decoded struct {
		Name   string
		State  DeviceState
		Period time.Duration
		Error  string
	}

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if decoded.Name != "test-device" {
		t.Errorf("JSON name = %v, want test-device", decoded.Name)
	}
	if decoded.State != StateError {
		t.Errorf("JSON state = %v, want %v", decoded.State, StateError)
	}
	if decoded.Error != testErr.Error() {
		t.Errorf("JSON error = %v, want %v", decoded.Error, testErr.Error())
	}
}

func TestMockConfiguration(t *testing.T) {
	if IsMock() {
		t.Error("Mock should be disabled by default")
	}

	Mock(true)
	if !IsMock() {
		t.Error("Mock should be enabled after Mock(true)")
	}

	Mock(false)
	if IsMock() {
		t.Error("Mock should be disabled after Mock(false)")
	}
}
