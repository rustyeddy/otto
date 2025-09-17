package bme280

import (
	"encoding/json"
	"testing"

	"github.com/rustyeddy/otto/device"
)

// Test constants
const (
	TestI2CBus     = "/dev/i2c-1"
	TestI2CAddress = 0x77
)

func TestBME280Creation(t *testing.T) {
	tests := []struct {
		name    string
		devName string
		bus     string
		addr    int
		wantErr bool
	}{
		{
			name:    "valid configuration",
			devName: "bme-test",
			bus:     "/dev/i2c-fake",
			addr:    0x76,
			wantErr: false,
		},
		{
			name:    "default configuration",
			devName: "bme-default",
			bus:     TestI2CBus,
			addr:    TestI2CAddress,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device.Mock(true)
			defer device.Mock(false)

			bme := New(tt.devName, tt.bus, tt.addr)
			if bme == nil {
				t.Fatal("Failed to create BME280 device")
			}

			if bme.Name() != tt.devName {
				t.Errorf("Name() = %v, want %v", bme.Name(), tt.devName)
			}

			// Only test Init if the method exists
			if err := bme.Open(); err != nil && !tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBME280Reading(t *testing.T) {
	device.Mock(true)
	defer device.Mock(false)

	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	if err := bme.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	tests := []struct {
		name           string
		minTemperature float64
		maxTemperature float64
		minHumidity    float64
		maxHumidity    float64
		minPressure    float64
		maxPressure    float64
	}{
		{
			name:           "valid ranges",
			minTemperature: -40.0,
			maxTemperature: 85.0,
			minHumidity:    0.0,
			maxHumidity:    100.0,
			minPressure:    300.0,
			maxPressure:    1100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := bme.Read()
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}

			if resp == nil {
				t.Fatal("Read() returned nil response")
			}

			if resp.Temperature < tt.minTemperature || resp.Temperature > tt.maxTemperature {
				t.Errorf("Temperature %f outside valid range [%f, %f]",
					resp.Temperature, tt.minTemperature, tt.maxTemperature)
			}

			if resp.Humidity < tt.minHumidity || resp.Humidity > tt.maxHumidity {
				t.Errorf("Humidity %f outside valid range [%f, %f]",
					resp.Humidity, tt.minHumidity, tt.maxHumidity)
			}

			if resp.Pressure < tt.minPressure || resp.Pressure > tt.maxPressure {
				t.Errorf("Pressure %f outside valid range [%f, %f]",
					resp.Pressure, tt.minPressure, tt.maxPressure)
			}
		})
	}
}

func TestBME280ReadPub(t *testing.T) {
	device.Mock(true)
	defer device.Mock(false)

	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	if err := bme.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	// Test ReadPub method
	err := bme.ReadPub()
	if err != nil {
		t.Errorf("ReadPub() error = %v", err)
	}
}

func TestBME280JSON(t *testing.T) {
	device.Mock(true)
	defer device.Mock(false)

	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	if err := bme.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	// Test JSON method if it exists
	data, err := bme.JSON()
	if err != nil {
		t.Errorf("JSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("JSON() returned empty data")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("JSON() returned invalid JSON: %v", err)
	}
}

func TestBME280String(t *testing.T) {
	device.Mock(true)
	defer device.Mock(false)

	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	str := bme.String()

	if str == "" {
		t.Error("String() returned empty string")
	}

	if str != bme.Name() {
		t.Errorf("String() = %v, want %v", str, bme.Name())
	}
}
