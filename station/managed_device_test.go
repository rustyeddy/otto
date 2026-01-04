package station

import (
	"testing"

	"github.com/rustyeddy/otto/messenger"
	"github.com/stretchr/testify/assert"
)

func TestManagedDevice(t *testing.T) {
	called := false
	topic := "o/d/station/device"
	dev := tstdevice{name: "device"}
	md := &ManagedDevice{
		Name:   "device",
		Device: dev,
		Topic:  topic,
	}

	messenger.SetMessenger(nil)
	md.Subscribe(topic, func(msg *messenger.Msg) error {
		called = true
		return nil
	})
	assert.False(t, called)

	// this will print a warning
	md.PubData(called)

	msgr := messenger.NewMessenger("none")
	assert.NotNil(t, msgr)
	md.Subscribe(topic, func(msg *messenger.Msg) error {
		called = true
		return nil
	})

	var data any
	data = "true"
	md.PubData(data)
	assert.True(t, called)

	data = "false"
	md.PubData(data)
	assert.True(t, called)
	assert.Equal(t, "false", data.(string))
}

func TestPubData(t *testing.T) {
	topic := "o/d/station/device"
	dev := tstdevice{name: "device"}
	md := &ManagedDevice{
		Name:   "device",
		Device: dev,
		Topic:  topic,
	}

	dat := ""
	msgr := messenger.GetMessenger()
	msgr.Sub(topic, func(msg *messenger.Msg) error {
		dat = string(msg.Data)
		return nil
	})

	type vers struct {
		Version string `json:"version"`
	}
	ver := vers{Version: "1.2.3"}

	tsts := []struct {
		input  any
		expect string
	}{
		{"some-string", "some-string"},
		{1, "1"},
		{123.45, "123.45"},
		{true, "1"},
		{false, "0"},
		{ver, `{"version":"1.2.3"}`},
	}

	for _, tst := range tsts {
		md.PubData(tst.input)
		assert.Equal(t, tst.expect, dat)
		dat = ""
	}

	md.PubData(1)
	assert.Equal(t, "1", dat)

	md.PubData(123.34)
	assert.Equal(t, "123.34", dat)

}

type tstdevint struct {
}

func (tst *tstdevint) Get() (int, error) { return 1, nil }

func TestReadPubInt(t *testing.T) {
	topic := "o/d/station/device"
	dev := tstdevint{}
	md := &ManagedDevice{
		Name:   "devint",
		Device: dev,
		Topic:  topic,
	}
	md.ReadPub()
}
