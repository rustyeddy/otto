package vh400

import (
	"testing"

	"github.com/rustyeddy/otto/device"
)

func TestVH400(t *testing.T) {
	device.Mock(true)

	v := New("vh400", 1)
	if v.Name() != "vh400" {
		t.Errorf("vh400 name expected (%s) got (%s)", "vh400", v.Name())
	}

	val, err := v.Read()
	if err != nil {
		t.Errorf("VH400 Read failed %s", err)
	}

	if val == 0.0 {
		t.Errorf("Expected value but got 0")
	}
}
