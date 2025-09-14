package relay

import (
	"testing"

	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/messanger"
)

func TestRelay(t *testing.T) {
	device.Mock(true)

	relay := New("relay", 5)
	if relay.Name() != "relay" {
		t.Errorf("relay expected Name (%s) got (%s)", "relay", relay.Name())
	}

	msg := messanger.NewMsg(relay.Topic(), []byte("on"), "test")
	relay.Callback(msg)

	v, err := relay.Value()
	if err != nil {
		t.Fatalf("relay.Value() got error %v", err)
	}
	if v != 1 {
		t.Errorf("relay expected (1) got (%d)", v)
	}

	msg = messanger.NewMsg(relay.Topic(), []byte("off"), "test")
	relay.Callback(msg)

	v, err = relay.Value()
	if err != nil {
		t.Fatalf("relay.Value() got error %v", err)
	}
	if v != 0 {
		t.Errorf("relay expected (0) got (%d)", v)
	}
}
