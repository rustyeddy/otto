package vh400

// See https://vegetronix.com/Products/VH400/VH400-Piecewise-Curve
// For calculations on the VWC.  Borrowed from above website

import (
	"log"
	"log/slog"

	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/device/drivers"
	"github.com/rustyeddy/otto/messanger"
)

type VH400 struct {
	*device.Device
	drivers.AnalogPin
}

func New(name string, pin int) *VH400 {
       d := device.NewDevice(name, "mqtt")
       v := &VH400{
	       Device: d,
       }
	if device.IsMock() {
		v.AnalogPin = drivers.NewMockAnalogPin(name, pin, nil)
		return v
	}

	ads := drivers.GetADS1115()
	p, err := ads.Pin(name, pin, nil)
	if err != nil {
		slog.Error("vh400.New", "name", name, "pin", pin, "error", err)
		return nil
	}
	v.AnalogPin = p
	return v
}

func (v *VH400) Name() string {
	return v.Device.Name()
}

func (v *VH400) Read() (float64, error) {
	volts, err := v.AnalogPin.Read()
	if err != nil {
		return volts, err
	}
	vwc := vwc(volts)
	return vwc, nil
}

func (v *VH400) ReadPub() error {
	vwc, err := v.Read()
	if err != nil {
		return err
	}
	v.PubData(vwc)
	return nil
}

func (v *VH400) ReadContinousPub() error {
	v.SetTopic(messanger.GetTopics().Data("vh100/" + v.Name()))
	q := v.AnalogPin.ReadContinuous()
	go func() {
		for {
			vbytes := <-q
			volts := vbytes
			vwc := vwc(volts)
			v.PubData(vwc)
		}
	}()

	return nil
}

/*
Most curves can be approximated with linear segments of the form:

y= m*x-b,

where m is the slope of the line

The VH400's Voltage to VWC curve can be approximated with 4 segments
of the form:

VWC= m*V-b

where V is voltage.

m = (VWC2 - VWC1) / (V2-V1)

where V1 and V2 are voltages recorded at the respective VWC levels of
VWC1 and VWC2. After m is determined, the y-axis intercept coefficient
b can be found by inserting one of the end points into the equation:

b= m*v-VWC
*/
func vwc(volts float64) float64 {
	var coef float64
	var rem float64

	switch {
	case volts >= 0.0 && volts <= 1.1:
		coef = 10.0
		rem = 1.0

	case volts > 1.1 && volts <= 1.3:
		coef = 25.0
		rem = 17.5

	case volts > 1.3 && volts <= 1.82:
		coef = 48.08
		rem = 47.5

	case volts > 1.82 && volts <= 2.2:
		coef = 26.32
		rem = 7.80

	case volts > 2.2 && volts <= 3.1:
		coef = 62.5
		rem = 7.89

	default:
		log.Printf("VH400 Invalid voltage: out of range 0.0 -> 3.0 %5.2f", volts)
		return 0.0
	}
	vwc := coef*volts - rem
	return vwc
}
