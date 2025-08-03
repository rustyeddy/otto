package devices

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/rustyeddy/otto/messanger"
	"github.com/warthog618/go-gpiocdev"
)

type BaseDevice struct {
	// Name of the device human readable
	name string

	// Suffix to be appended to the base topic for mqtt publications
	pub string

	// Subscription this device will listen to
	subs []string

	// Period for repititive timed tasks like collecting and
	// publishing data
	period time.Duration

	// EventQ for devices that are interupt driven
	evtQ chan gpiocdev.LineEvent
}

// NewDevice creates a new device with the given name
func NewDevice(name string) *BaseDevice {
	d := &BaseDevice{
		name: name,
	}
	return d
}

func (d BaseDevice) Name() string {
	return d.name
}

// AddPub adds a publication
func (d *BaseDevice) AddPub(p string) {
	d.pub = p
}

func (d *BaseDevice) GetPub() string {
	return d.pub
}

func (d *BaseDevice) Publish(data any) {
	if d.pub == "" {
		slog.Error("BaseDevice.Publish failed has no pub", "name", d.Name)
		return
	}
	var buf []byte
	switch data.(type) {
	case []byte:
		buf = data.([]byte)

	case string:
		buf = []byte(data.(string))

	default:
		panic("unknown type: " + fmt.Sprintf("%T", data))
	}

	msg := messanger.New(d.pub, buf, d.name)
	m := messanger.GetMQTT()
	m.PublishMsg(msg)
}

func (d *BaseDevice) GetSubs() []string {
	return d.subs
}

func (d *BaseDevice) Subscribe(topic string, f func(*messanger.Msg)) {
	d.subs = append(d.subs, topic)
	m := messanger.GetMQTT()
	m.Subscribe(topic, f)
}

func (d *BaseDevice) String() string {
	return d.Name() + ": todo finish String()"
}

func (d *BaseDevice) JSON() []byte {
	panic("todo finish BaseDevice.JSON()")
}

func (d *BaseDevice) Shutdown() {
	// XXX = this is not right
	// GetGPIO().Shutdown()
}

// func (d *BaseDevice) EvtQ(evt gpiocdev.LineEvent) {
// 	d.evtQ <- evt
// }

// func (d *BaseDevice) EventLoop(done chan any, readpub func() error) {
// 	running := true
// 	for running {
// 		select {
// 		case evt := <-d.evtQ:
// 			evtype := "falling"
// 			switch evt.Type {
// 			case gpiocdev.LineEventFallingEdge:
// 				evtype = "falling"

// 			case gpiocdev.LineEventRisingEdge:
// 				evtype = "raising"

// 			default:
// 				slog.Warn("Unknown event type ", "type", evt.Type)
// 				continue
// 			}

// 			slog.Info("GPIO edge", "device", d.Name, "direction", evtype,
// 				"seqno", evt.Seqno, "lineseq", evt.LineSeqno)

// 			err := readpub()
// 			if err != nil {
// 				slog.Error("Failed to read and publish", "device", d.Name, "error", err)
// 			}

// 		case <-done:
// 			running = false
// 		}
// 	}
// }

func (d *BaseDevice) TimerLoop(period time.Duration, done chan any, readpub func() error) {
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
