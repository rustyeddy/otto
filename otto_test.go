package otto

import (
	"testing"
	"time"

	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/server"
	"github.com/rustyeddy/otto/station"
	"github.com/stretchr/testify/assert"
)

func TestOttO_Init(t *testing.T) {
	o := &OttO{Name: "test"}

	o.Init()

	assert.NotNil(t, o.done, "done channel should be initialized")
	assert.NotNil(t, o.Messanger, "Messanger should be initialized")
	assert.NotNil(t, o.StationManager, "StationManager should be initialized")
	assert.NotNil(t, o.Station, "Station should be initialized")
	assert.NotNil(t, o.DataManager, "DataManager should be initialized")
	assert.NotNil(t, o.Server, "Server should be initialized")
}

func TestOttO_Start(t *testing.T) {
	o := &OttO{
		Name:           "test",
		done:           make(chan any),
		StationManager: station.GetStationManager(),
		Server:         server.GetServer(),
	}

	go func() {
		time.Sleep(1 * time.Second)
		close(o.done)
	}()

	err := o.Start()
	assert.NoError(t, err, "Start should not return an error")
}

func TestOttO_Stop(t *testing.T) {
	o := &OttO{
		Name:      "test",
		done:      make(chan any),
		Messanger: messanger.NewMessanger("test", ""),
		Server:    server.GetServer(),
	}

	go func() {
		time.Sleep(1 * time.Second)
		close(o.done)
	}()

	o.Stop()
}

func TestOttO_Done(t *testing.T) {
	o := &OttO{}
	o.Init()
	done := o.Done()
	assert.NotNil(t, done, "Done channel should not be nil")
}
