// Package device provides a framework for managing hardware devices
// with support for messaging, periodic operations, and state management.
// Devices can be controlled via MQTT messages and can publish their
// state and data periodically.
package device

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/rustyeddy/otto/messanger"
)

// DeviceState represents the current operational state of a device
type DeviceState string

const (
	StateUnknown      DeviceState = "unknown"
	StateInitializing DeviceState = "initializing"
	StateRunning      DeviceState = "running"
	StateError        DeviceState = "error"
	StateStopped      DeviceState = "stopped"
)

// Opener represents a device that can be opened and closed for communication.
type Opener interface {
	Open() error
	Close() error
}

// OnOff represents a device that can be turned on and off.
type OnOff interface {
	On() error
	Off() error
}

// Name represents a device that has a human-readable name.
type Name interface {
	Name() string
}

// mockConfig handles mock device configuration with thread safety
type mockConfig struct {
	enabled bool
	mu      sync.RWMutex
}

var mockCfg = &mockConfig{}

// Mock enables or disables mock device behavior
func Mock(mocking bool) {
	mockCfg.mu.Lock()
	defer mockCfg.mu.Unlock()
	mockCfg.enabled = mocking
}

// IsMock returns the current mock state
func IsMock() bool {
	mockCfg.mu.RLock()
	defer mockCfg.mu.RUnlock()
	return mockCfg.enabled
}

// Device represents a physical or virtual device with messaging capabilities
type Device struct {
	Name      string              // Human readable device name
	State     DeviceState         // Current device state
	Messanger messanger.Messanger // Messaging system
	Period    time.Duration       // Period for timed operations
	Val       any                 // Mock value storage
	err       error               // Last error encountered (use SetError to set)
	mu        sync.RWMutex        // Protects device state
	Opener                        // Device opening interface
}

// SetError sets the device error and updates the state to StateError
func (d *Device) SetError(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.err = err
       if err != nil {
	       d.State = StateError
       }
}

// ErrorVal returns the last error encountered by the device
func (d *Device) Error() error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.err
}

// NewDevice creates a new device with the given name
func NewDevice(name string, t string) *Device {
	return &Device{
		Name:      name,
		State:     StateUnknown,
		Messanger: messanger.NewMessanger(t, name),
	}
}

// TimerLoop runs periodic operations with context support
func (d *Device) TimerLoop(ctx context.Context, period time.Duration, readpub func() error) error {
	if period <= 0 {
		return fmt.Errorf("invalid period: %v", period)
	}

	d.Period = period
	d.State = StateRunning

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.State = StateStopped
			return ctx.Err()
		case <-ticker.C:
			if err := readpub(); err != nil {
				slog.Error("TimerLoop failed",
					"device", d.Name,
					"error", err)
							   d.err = err 
			}
		}
	}
}

// String returns the device name
func (d *Device) String() string {
	return d.Name + " (" + string(d.State) + ") ";
}

// JSON returns a JSON representation of the device
func (d *Device) JSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	j := struct {
		Name      string
		State     DeviceState
		Messanger messanger.Messanger
		Period    time.Duration
		Error     string
	}{
		Name:      d.Name,
		State:     d.State,
		Messanger: d.Messanger,
		Period:    d.Period,
		Error:     errString(d.err),
	}

	return json.Marshal(j)
}

// errString safely converts an error to a string
func errString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
