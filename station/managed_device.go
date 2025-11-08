package station

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/rustyeddy/otto/messanger"
)

// TimerLooper interface for devices that support periodic operations
type TimerLooper interface {
	TimerLoop(duration time.Duration, done chan any, readFunc func())
}

// ManagedDevice wraps any device and adds messaging capabilities
type ManagedDevice struct {
	Name   string
	Device any // The underlying device (from devices package)
	Topic  string
}

// NewManagedDevice creates a new managed device with messaging capabilities
func NewManagedDevice(name string, device any, topic string) *ManagedDevice {
	md := &ManagedDevice{
		Name:      name,
		Device:    device,
		Topic:     topic,
	}
	return md
}

// ID returns the device name/identifier
func (md *ManagedDevice) ID() string {
	return md.Name
}

// Subscribe sets up a subscription for this device
func (md *ManagedDevice) Subscribe(topic string, callback func(bool)) {
	messanger := messanger.GetMessanger()
	if messanger == nil {
		slog.Warn("No messanger available for device", "device", md.Name)
		return
	}

	handler := func(msg *messanger.Msg) error {
		// Parse the message and call the callback
		var val bool
		dataStr := string(msg.Data)
		if dataStr == "on" || dataStr == "true" || dataStr == "1" {
			val = true
		} else {
			val = false
		}
		callback(val)
		return nil
	}

	messanger.Subscribe(topic, handler)
}

// PubData publishes data for this device
func (md *ManagedDevice) PubData(data interface{}) {
	messanger := messanger.GetMessanger()
	if messanger == nil {
		slog.Warn("No messanger available for device", "device", md.Name)
		return
	}

	var payload string
	switch v := data.(type) {
	case string:
		payload = v
	case int:
		payload = fmt.Sprintf("%d", v)
	case float64:
		payload = fmt.Sprintf("%.2f", v)
	case bool:
		if v {
			payload = "1"
		} else {
			payload = "0"
		}
	default:
		jsonData, err := json.Marshal(data)
		if err != nil {
			slog.Error("Failed to marshal device data", "device", md.Name, "error", err)
			return
		}
		payload = string(jsonData)
	}

	messanger.Pub(md.Topic, payload)
}

// ReadPub reads the device value and publishes it
func (md *ManagedDevice) ReadPub() {
	// Try different Get method signatures
	if getter, ok := md.Device.(interface{ Get() (int, error) }); ok {
		val, err := getter.Get()
		if err != nil {
			slog.Error("Failed to read device value", "device", md.Name, "error", err)
			return
		}
		md.PubData(val)
	} else if getter, ok := md.Device.(interface{ Get() (float64, error) }); ok {
		val, err := getter.Get()
		if err != nil {
			slog.Error("Failed to read device value", "device", md.Name, "error", err)
			return
		}
		md.PubData(val)
	} else if getter, ok := md.Device.(interface{ Get() (any, error) }); ok {
		val, err := getter.Get()
		if err != nil {
			slog.Error("Failed to read device value", "device", md.Name, "error", err)
			return
		}
		md.PubData(val)
	} else {
		slog.Debug("Device does not support Get() method", "device", md.Name)
	}
}

// EventLoop provides a generic event loop for devices that support it
func (md *ManagedDevice) EventLoop(done chan any) {
	// Check if the underlying device has an EventLoop method
	if eventLooper, ok := md.Device.(interface{ EventLoop(chan any, func()) }); ok {
		eventLooper.EventLoop(done, md.ReadPub)
	} else {
		slog.Warn("Device does not support EventLoop", "device", md.Name)
	}
}

// StartTimerLoop starts a timer loop for periodic device readings
func (md *ManagedDevice) StartTimerLoop(duration time.Duration, done chan any) {
	// First check if device implements TimerLooper interface
	if timerLooper, ok := md.Device.(TimerLooper); ok {
		go timerLooper.TimerLoop(duration, done, md.ReadPub)
		return
	}

	// Otherwise, provide a generic timer loop implementation
	go md.genericTimerLoop(duration, done)
}

// genericTimerLoop provides a standard timer loop for any device
func (md *ManagedDevice) genericTimerLoop(duration time.Duration, done chan any) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			md.ReadPub()
		case <-done:
			slog.Debug("Timer loop stopped", "device", md.Name)
			return
		}
	}
}
