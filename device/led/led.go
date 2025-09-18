package led

import (
	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/device/drivers"
	"github.com/rustyeddy/otto/messanger"
	"github.com/warthog618/go-gpiocdev"
)

type LED struct {
	*device.Device
	*drivers.DigitalPin
}

func New(name string, offset int) *LED {
       led := &LED{
	       Device: device.NewDevice(name, "mqtt"),
       }
	g := drivers.GetGPIO()
	led.DigitalPin = g.Pin(name, offset, gpiocdev.AsOutput(0))
	return led
}

func (l *LED) Callback(msg *messanger.Msg) {
	switch msg.String() {
	case "off", "OFF", "Off", "0":
		l.Off()

	case "on", "ON", "On", "1":
		l.On()
	}
	return
}
