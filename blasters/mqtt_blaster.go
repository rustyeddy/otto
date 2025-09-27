package blasters

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/station"
)

// Blaster is a virtual station that will spew messages to a given
// topic to be used for testing.
type Blaster struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	*station.Station
	Topic string
}

// MQTTBlasters is a collection of blaster to be used for testing
// multiple stations
type MQTTBlasters struct {
	Count    int
	Blasters []*Blaster
	Running  bool
	Wait     int
}

// NewMQTTBlasters will create a count number of blasters ready to
// blast MQTT messages
func NewMQTTBlasters(count int) *MQTTBlasters {
	mb := &MQTTBlasters{
		Count:   count,
		Running: false,
		Wait:    2000,
	}
	
	sm := station.GetStationManager()
	mb.Blasters = make([]*Blaster, mb.Count)
	for i := 0; i < mb.Count; i++ {
		id := fmt.Sprintf("station-%d", i)
		topic := fmt.Sprintf("ss/d/%s/temphum", id)

		st, err := sm.Add(id)
		if err != nil {
		 	panic(err)
		}
		mb.Blasters[i] = &Blaster{
			Topic:   topic,
			Station: st,
		}
	}
	return mb
}

// Blast will start the configured blasters to start blasting. TODO
// add a function that can be used to generate packets based on
// various configurations. TODO: allow the replay of a captured
// Msg stream
func (mb *MQTTBlasters) Blast() error {

	msgr := messanger.NewMessangerMQTT("blaster")
	wd := &WeatherData{}

	// now start blasting
	mb.Running = true
	for mb.Running {
		for i := 0; i < mb.Count; i++ {
			b := mb.Blasters[i]
			msg := wd.NewMsg(b.Topic)
			msgr.PubMsg(msg)
		}
		time.Sleep(time.Duration(mb.Wait) * time.Millisecond)
	}
	slog.Info("MQTT Blaster has stopped")
	return nil
}

// (No Start method for Blaster, logic is handled by MQTTBlasters)

// Stop will cause the blasters to stop blasting.
func (mb *MQTTBlasters) Stop() {
	mb.Running = false
}

// (Removed NewMQTTBlaster, not needed)
