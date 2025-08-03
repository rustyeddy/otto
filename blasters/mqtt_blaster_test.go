package blasters

import (
	"fmt"
	"testing"
	"time"

	"github.com/rustyeddy/otto/messanger"
)

var (
	blasters *MQTTBlasters
)

func init() {
	// contrary to popular testing practices we are going to
	// pre-register our end points here because the builtin go server
	// mux does not support unregistering handlers.  So we will just
	// share them between testso
	blasters = NewMQTTBlasters(5)
}

func TestBlasters(t *testing.T) {
	if blasters.Count != 5 {
		t.Errorf("expected (5) blasters got (%d)", blasters.Count)
	}

	if blasters.Running {
		t.Error("expected running to be (false) got (true)")
	}

	if blasters.Wait != 2000 {
		t.Errorf("expected wait to be (2000) got (%d)", blasters.Wait)
	}

	for i := 0; i < 5; i++ {
		b := blasters.Blasters[i]
		stid := fmt.Sprintf("station-%d", i)
		if b.ID != stid {
			t.Errorf("expected station id (%s) got (%s)", stid, b.ID)
		}

		topic := fmt.Sprintf("ss/d/%s/temphum", stid)
		if b.Topic != topic {
			t.Errorf("expected topic (%s) got (%s)", topic, b.Topic)
		}
	}
}

func TestBlasting(t *testing.T) {
	c := messanger.GetMockClient()
	m := messanger.SetMQTTClient(c)
	m.Connect()

	for _, bl := range blasters.Blasters {
		m.Subscribe(bl.Topic, messanger.GetDataManager().Callback)
	}

	go blasters.Blast()
	time.Sleep(2 * time.Second)

	for _, bl := range blasters.Blasters {
		if bl == nil {
			t.Error("Explected a blaster got (nil)")
		}
	}

	blasters.Stop()
}
