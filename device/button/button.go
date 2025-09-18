package button

import (
	"log/slog"
	"time"

	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/device/drivers"
	"github.com/warthog618/go-gpiocdev"
)

// Button is a simple momentary GPIO Device that can be detected with
// the Button is pushed (rising edge) or when it is released (falling
// edge). Low is open, high is closed.
type Button struct {
	*device.Device
	*drivers.DigitalPin
}

// New creates a new button with the given name, offset represents the
// pin number and a series of line options. Todo reference the gpiodev
// manual for LineReq options
func New(name string, offset int, opts ...gpiocdev.LineReqOption) *Button {
	var evtQ chan gpiocdev.LineEvent
	evtQ = make(chan gpiocdev.LineEvent)
	bopts := []gpiocdev.LineReqOption{
		gpiocdev.WithPullUp,
		gpiocdev.WithDebounce(10 * time.Millisecond),
		gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
			evtQ <- evt
		}),
	}

	for _, o := range opts {
		bopts = append(bopts, o)
	}

       b := &Button{
	       Device:     device.NewDevice(name, "mqtt"),
	       DigitalPin: drivers.NewDigitalPin(name, offset, bopts...),
       }
	b.EvtQ = evtQ
	return b
}

// ReadPub will read the value of the button and publish the results.
func (b *Button) ReadPub() {
	val, err := b.Get()
	if err != nil {
		slog.Error("Failed to read buttons value: ", "error", err.Error())
		return
	}
	slog.Debug("read", "device", "button", "val", val)
	b.PubData(val)
}
