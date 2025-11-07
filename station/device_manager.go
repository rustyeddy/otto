package station

import (
	"sync"
)

type DeviceManager struct {
	// Internal generic device store for tests and loose coupling
	devices   map[string]any
	Metrics		*DeviceMetrics
	mu sync.RWMutex `json:"-"`
}

type DeviceMetrics struct {
	// Device metrics
	DeviceCount      int    `json:"device_count"`
	ActiveDevices    int    `json:"active_devices"`
	DeviceErrorCount uint64 `json:"device_error_count"`
	mu sync.RWMutex `json:"-"`
}

func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make(map[string]any),
		Metrics: &DeviceMetrics{},
	}
}

// AddDevice will do what it says by placing the device with a given
// name in the stations device manager. This library is basically a
// key value store, anything supporting the ID Interface:
// i.e. ID() string.
func (dm *DeviceManager) Register(d interface{ ID() string }) {
	if d == nil {
		return
	}
	name := d.ID()

	// store generically
	dm.mu.Lock()
	dm.devices[name] = d
	devCount := len(dm.devices)
	dm.mu.Unlock()

	// Update device metrics
	if dm.Metrics != nil {
		dm.Metrics.UpdateMetrics(devCount, devCount, dm.Metrics.DeviceErrorCount)
	}
	// TODO: Track active vs total
}

func (dm *DeviceManager) Remove(id string) {
	d := dm.Get(id)
	if d == nil {
		return
	}
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	delete(dm.devices, id)
}

// GetDevice returns the device (anythig supporting the Name (Name()) interface)
func (dm *DeviceManager) Get(name string) any {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.devices[name]
}

// UpdateDeviceMetrics updates device-related metrics
func (dm *DeviceMetrics) UpdateMetrics(deviceCount, activeDevices int, deviceErrors uint64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.DeviceCount = deviceCount
	dm.ActiveDevices = activeDevices
	dm.DeviceErrorCount = deviceErrors
}

