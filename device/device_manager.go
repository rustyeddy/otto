package device

import (
	"fmt"
	"sync"
)

// DeviceManager handles the registration and retrieval of devices.
// It ensures thread-safe access to the device collection.
type DeviceManager struct {
	devices map[string]Name `json:"devices"`
	mu      sync.RWMutex    `json:"-"`
}

var (
	stationName string = "station"
	devices     *DeviceManager
	once        sync.Once
)

// GetDeviceManager returns the singleton instance of DeviceManager.
// It ensures thread-safe initialization and access to the device manager.
func GetDeviceManager() *DeviceManager {
	once.Do(func() {
		devices = &DeviceManager{
			devices: make(map[string]Name),
		}
	})
	return devices
}

// Add registers a new device with the manager.
// If a device with the same name exists, it will be replaced.
func (dm *DeviceManager) Add(d Name) error {
	if d == nil {
		return fmt.Errorf("cannot add nil device")
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.devices[d.Name()] = d
	return nil
}

// Get retrieves a device by name.
// Returns the device and true if found, nil and false otherwise.
func (dm *DeviceManager) Get(name string) (Name, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	d, exists := dm.devices[name]
	return d, exists
}

// Remove removes a device from the manager.
// Returns true if the device was removed, false if it didn't exist.
func (dm *DeviceManager) Remove(name string) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.devices[name]; exists {
		delete(dm.devices, name)
		return true
	}
	return false
}

// List returns a slice of all registered device names.
func (dm *DeviceManager) List() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	names := make([]string, 0, len(dm.devices))
	for name := range dm.devices {
		names = append(names, name)
	}
	return names
}

// Clear removes all devices from the manager.
func (dm *DeviceManager) Clear() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.devices = make(map[string]Name)
}
