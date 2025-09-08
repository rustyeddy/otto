package device

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/rustyeddy/otto/messanger"
)

// Opener interface defines the Open() function for
type Opener interface {
	Open() error
}

type OnOff interface {
	On()
	Off()
}

var mock bool

func Mock(mocking bool) {
	mock = mocking
}

func IsMock() bool {
	return mock
}

type Name interface {
	Name() string
}

type Device struct {
	// Name of the device human readable
	name string

	// Suffix to be appended to the base topic for mqtt publications
	messanger.Messanger

	// Period for repititive timed tasks like collecting and
	// publishing data
	period time.Duration

	// EventQ for devices that are interupt driven
	// EvtQ chan gpiocdev.LineEvent

	// for mocking
	val any

	// Last Error encountered
	Error error

	Opener
}

// NewDevice creates a new device with the given name
func NewDevice(name string) *Device {
	d := &Device{
		name:      name,
		Messanger: messanger.NewMessangerMQTT(name),
	}
	return d
}

func (d Device) Name() string {
	return d.name
}

func (d *Device) TimerLoop(period time.Duration, done chan any, readpub func() error) {
	// No need to loop if we don't have a ticker period
	d.period = period
	if d.period <= 0 {
		return
	}
	ticker := time.NewTicker(d.period)

	running := true
	for running {
		select {
		case <-ticker.C:
			err := readpub()
			if err != nil {
				slog.Error("TimerLoop failed to readpub", "device", d.Name, "error", err)
			}

		case <-done:
			running = false
		}
	}
}

func (d *Device) String() string {
	return d.Name()
}

func (d *Device) JSON() ([]byte, error) {
	j := struct {
		Name string
		messanger.Messanger
		Period time.Duration
		Err    error
	}{
		Name:      d.name,
		Messanger: d.Messanger,
		Period:    d.period,
		Err:       d.Error,
	}

	jbytes, err := json.Marshal(j)
	return jbytes, err
}
