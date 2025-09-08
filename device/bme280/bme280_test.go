package bme280

import (
	"testing"
	"time"

	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/messanger"
)

func TestBME280(t *testing.T) {
	name := "bme-test"
	bus := "/dev/i2c-fake"
	addr := 0x76

	device.Mock(true)
	bme := New(name, bus, addr)
	if bme == nil {
		t.Error("Failed to create bme device")
	}

	if bme.Name() != name {
		t.Errorf("expected name (%s) got (%s)", name, bme.Name())
	}

	if bme.bus != bus {
		t.Errorf("expected bus (%s) got (%s)", bus, bme.bus)
	}
	if bme.addr != addr {
		t.Errorf("expected addr (%x) got (%x)", addr, bme.addr)
	}

	err := bme.Init()
	if err != nil {
		t.Error("failed to initialize mock BME device")
	}

	resp, err := bme.Read()
	if err != nil {
		t.Error("failed to read from mock BME device")
	}

	// XXX these values are hard coded in the bme280 device mock read
	// may change to random values or other
	if resp.Temperature == 0.0 {
		t.Errorf("Failed to read temperature")
	}
	if resp.Pressure == 0.0 {
		t.Errorf("Failed to read temperature")
	}
	if resp.Humidity == 0.0 {
		t.Errorf("Failed to read temperature")
	}

	// Set up for bme EventLoop run the loop every 200 milliseconds then
	// stop the loop
	count := 0
	topic := messanger.GetTopics().Data(name)
	bme.SetTopic(topic)
	bme.Subscribe(topic, func(msg *messanger.Msg) {
		if msg.Topic != topic {
			t.Errorf("expected topic (%s) got (%s)", topic, msg.Topic)
			return
		}
		mmm, err := msg.Map()
		if err != nil {
			t.Errorf("failed to map bme280 response %s", err)
			return
		}

		for key, val := range mmm {
			switch key {
			case "Temperature", "temperature":
				if val == 0.0 {
					t.Errorf("%s expected (rand) got (%4.2f)", key, val)
					return
				}

			case "Humidity", "humidity":
				if val == 0.0 {
					t.Errorf("%s expected (rand) got (%4.2f)", key, val)
					return
				}

			case "Pressure", "pressure":
				if val == 0.0 {
					t.Errorf("%s expected (rand) got (%4.2f)", key, val)
					return
				}

			default:
				t.Errorf("Unexpected key value %s - %v", key, val)
				return
			}
		}
		count++
	})

	done := make(chan any)
	go bme.TimerLoop(100*time.Millisecond, done, bme.ReadPub)

	select {
	case <-done:
		break

	case <-time.After(1 * time.Second):
		done <- true
	}
	if count < 5 {
		t.Errorf("Expected to recieve messanges expected (>= 5) got (%d) ", count)
	}
}
