package drivers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/messanger"
	"github.com/warthog618/go-gpiocdev"
)

// GPIO is used to initialize the GPIO and pins on a raspberry pi
type GPIO struct {
	Chipname string              `json:"chipname"`
	pins     map[int]*DigitalPin `json:"pins"`
	Mock     bool                `json:"mock"`
}

var (
	gpio *GPIO
)

// GetGPIO returns the GPIO singleton for the Raspberry PI
func GetGPIO() *GPIO {
	if gpio != nil {
		return gpio
	}

	gpio = &GPIO{
		// Chipname: "gpiochip4", // raspberry pi-5
		Chipname: "gpiochip0", // raspberry pi zero
		Mock:     device.IsMock(),
	}
	gpio.pins = make(map[int]*DigitalPin)
	for _, pin := range gpio.pins {
		if err := pin.Init(); err != nil {
			slog.Error("Error initializing pin ", "offset", pin.offset)
		}
	}
	return gpio
}

// Pin initializes the given GPIO pin, name and mode
func (gpio *GPIO) Pin(name string, offset int, opts ...gpiocdev.LineReqOption) *DigitalPin {

	var dopts []gpiocdev.LineReqOption
	for _, o := range opts {
		dopts = append(dopts, o.(gpiocdev.LineReqOption))
	}
	p := &DigitalPin{
		name:   name,
		offset: offset,
		opts:   dopts,
	}

	if gpio.pins == nil {
		gpio.pins = make(map[int]*DigitalPin)
	}
	gpio.pins[offset] = p
	if err := p.Init(); err != nil {
		slog.Error(err.Error(), "name", name, "offset", offset)
	}
	return p
}

// Shutdown resets the GPIO line allowing use by another program
func (gpio *GPIO) Close() error {
	for _, p := range gpio.pins {
		p.Reconfigure(gpiocdev.AsInput)
		p.Close()
	}
	gpio.pins = nil
	return nil
}

// String returns the string representation of the GPIO
func (gpio *GPIO) String() string {
	str := ""
	for _, pin := range gpio.pins {
		str += pin.String()
	}
	return str
}

// JSON returns the JSON representation of the GPIO
func (gpio *GPIO) JSON() (j []byte, err error) {
	j, err = json.Marshal(gpio)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling GPIO %s", err)
	}
	return j, nil
}

// Line interface is used to emulate a GPIO pin as
// implemented by the go-gpiocdev package
type Line interface {
	Close() error
	Offset() int
	SetValue(int) error
	Reconfigure(...gpiocdev.LineConfigOption) error
	Value() (int, error)
}

type DigitalPin struct {
	name string
	opts []gpiocdev.LineReqOption
	Line

	offset int
	val    int
	mock   bool

	gpiocdev.EventHandler `json:"event-handler"`
	EvtQ                  chan gpiocdev.LineEvent
}

func NewDigitalPin(name string, offset int, opts ...gpiocdev.LineReqOption) *DigitalPin {
	gpio := GetGPIO()
	return gpio.Pin(name, offset, opts...)
}

func (p *DigitalPin) PinName() string {
	return p.name
}

// Init the pin from the offset and mode
func (p *DigitalPin) Init() error {

	gpio := GetGPIO()
	if gpio.Mock {
		line := GetMockLine(p.offset, p.opts...)
		p.mock = true
		p.Line = line
		return nil
	}

	line, err := gpiocdev.RequestLine(gpio.Chipname, p.offset, p.opts...)
	if err != nil {
		return err
	}
	p.Line = line
	return nil
}

func (p *DigitalPin) SetOpts(opts ...gpiocdev.LineReqOption) {
	p.opts = append(p.opts, opts...)
}

// String returns a string representation of the GPIO pin
func (pin *DigitalPin) String() string {
	v, err := pin.Value()
	if err != nil {
		slog.Error("Failed getting the value of ", "pin", pin.offset, "error", err)
	}
	str := fmt.Sprintf("%d: %d\n", pin.offset, v)
	return str
}

// Get returns the value of the pin, an error is returned if
// the GPIO value fails. Note: you can Get() the value of an
// input pin so no direction checks are done
func (pin *DigitalPin) Get() (int, error) {
	if pin.Line == nil {
		return 0, fmt.Errorf("GPIO not active")
	}
	val, err := pin.Line.Value()
	if err != nil {
		slog.Error("Failed to read PIN", "name", pin.name)
	}
	return val, err
}

// Set the value of the pin. Note: you can NOT set the value
// of an input pin, so we will check it and return an error.
// This maybe worthy of making it a panic!
func (pin *DigitalPin) Set(v int) error {
	if pin.Line == nil {
		return fmt.Errorf("GPIO not active")
	}
	pin.val = v
	return pin.Line.SetValue(v)
}

// On sets the value of the pin to 1
func (pin *DigitalPin) On() error {
	return pin.Set(1)
}

// Off sets the value of the pin to 0
func (pin *DigitalPin) Off() error {
	return pin.Set(0)
}

// Toggle with flip the value of the pin from 1 to 0 or 0 to 1
func (pin *DigitalPin) Toggle() error {
	val, err := pin.Get()
	if err != nil {
		return err
	}

	if val == 0 {
		val = 1
	} else {
		val = 0
	}
	return pin.Set(val)
}

// Callback is the default callback for pins if they are
// registered with the MQTT.Subscribe() function
func (pin DigitalPin) Callback(msg *messanger.Msg) {
	cmd := msg.String()
	switch cmd {
	case "on", "1":
		pin.On()

	case "off", "0":
		pin.Off()

	case "toggle":
		pin.Toggle()
	}
	return
}

func (d *DigitalPin) EventLoop(done chan any, readpub func()) {
	running := true
	for running {
		select {
		case evt := <-d.EvtQ:
			evtype := "falling"

			switch evt.Type {
			case gpiocdev.LineEventFallingEdge:
				evtype = "falling"

			case gpiocdev.LineEventRisingEdge:
				evtype = "raising"

			default:
				slog.Warn("GPIO EventLoop - unknown event type ", "type", evt.Type)
				continue
			}

			slog.Info("GPIO edge", "device", d.PinName(), "direction", evtype,
				"seqno", evt.Seqno, "lineseq", evt.LineSeqno)

			readpub()

		case <-done:
			running = false
		}
	}
}

func (d *DigitalPin) Close() error {
	close(d.EvtQ)
	return d.Line.Close()

}

// MockGPIO fakes the Line interface on computers that don't
// actually have GPIO pins mostly for mocking tests
type MockLine struct {
	offset                int `json:"offset"`
	Val                   int `json:"val"`
	gpiocdev.EventHandler `json:"event-handler"`
	start                 time.Time
}

func GetMockLine(offset int, opts ...gpiocdev.LineReqOption) *MockLine {
	m := &MockLine{
		offset: offset,
		start:  time.Now(),
	}
	for _, opt := range opts {
		switch v := opt.(type) {
		case gpiocdev.OutputOption:
			m.Val = v[0]

		case gpiocdev.EventHandler:
			m.EventHandler = opt.(gpiocdev.EventHandler)
			mqtt := messanger.GetMQTT()
			mqtt.Subscribe(messanger.GetTopics().Control("mock/"+strconv.Itoa(m.Offset())), m.Callback)

		default:
			// slog.Debug("MockLine does not record", "optType", v)
		}
	}

	return m
}

func (m MockLine) Close() error {
	return nil
}

func (m MockLine) Offset() int {
	return m.offset
}

func (m *MockLine) SetValue(val int) error {
	m.Val = val
	return nil
}

func (m MockLine) Reconfigure(...gpiocdev.LineConfigOption) error {
	return nil
}

func (m MockLine) Value() (int, error) {
	return m.Val, nil
}

var seqno uint32

func getSeqno() uint32 {
	seqno += 1
	return seqno
}

func (m *MockLine) Callback(msg *messanger.Msg) {
	str := msg.String()
	switch str {
	case "on":
		fallthrough

	case "1":
		m.MockHWInput(1)

	case "off":
		fallthrough

	case "0":
		m.MockHWInput(0)

	default:
		return
	}
	return
}

func (d *DigitalPin) MockHWInput(v int) {
	m := d.Line.(*MockLine)
	m.MockHWInput(v)
}

func (m *MockLine) MockHWInput(v int) {
	m.Val = v

	t := gpiocdev.LineEventRisingEdge
	if v == 0 {
		t = gpiocdev.LineEventFallingEdge
	}

	seq := getSeqno()
	if m.EventHandler != nil {
		evt := gpiocdev.LineEvent{
			Offset:    m.Offset(),
			Timestamp: time.Since(m.start),
			Type:      t,
			Seqno:     seq,
			LineSeqno: seq,
		}
		m.EventHandler(evt)
	}
}
