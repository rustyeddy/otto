package messanger

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rustyeddy/otto/utils"
)

func getMsg() (*Msg, time.Time) {
	now := time.Now()
	path := "ss/d/%s/test"

	b := fmt.Sprintf("%d", 4)
	m := New(path, []byte(b), "test")
	m.Source = "be:ef:ca:fe:01"
	m.Timestamp = utils.Timestamp()

	return m, now
}

func TestStationMsg(t *testing.T) {
	topic := "ss/d/be:ef:ca:fe:01/station"
	omsg, _ := getMsg()
	omsg.Source = "be:ef:ca:fe:01"
	if omsg.Last() != "test" {
		t.Errorf("Failed to get station")
	}

	j, err := json.Marshal(omsg)
	if err != nil {
		t.Errorf("json marshal failed %+v", err)
		return
	}

	msg := New(topic, j, "test")
	if msg == nil {
		t.Error("msg topic expected but is nil")
	}

	if msg.Topic != topic {
		t.Errorf("msg topic expected (%s) got (%s)", topic, msg.Topic)
	}

	path := strings.Split(topic, "/")
	if len(path) != len(msg.Path) {
		t.Errorf("msg path len expected (%d) got (%d)", len(path), len(msg.Path))
	}

	for i := 0; i < len(path); i++ {
		if path[i] != msg.Path[i] {
			t.Errorf("msg path[%d] expected (%s) got (%s)", i, path[i], msg.Path[i])
		}
	}

	if msg.Source != "test" {
		t.Errorf("msg source expected (test) got (%s)", msg.Source)
	}

	for i := 0; i < len(j); i++ {
		if msg.Data[i] != j[i] {
			t.Errorf("msg data[%d] expected (% x) got (% x)", i, j[i], msg.Data[i])
		}
	}
}

func TestJSON(t *testing.T) {
	m, _ := getMsg()

	jstr := `{ "int": 10, "float": 12.3, "string": "45.6" }`
	m.Data = []byte(jstr)

	jbyte, err := json.Marshal(m)
	if err != nil {
		t.Errorf("Failed to marshal message %+v", m)
	}

	var m2 Msg
	err = json.Unmarshal(jbyte, &m2)
	if err != nil {
		t.Error("Failed to unmarshal message", err)
	}

	if m2.ID != m.ID || m2.Topic != m.Topic ||
		m.Timestamp != m2.Timestamp ||
		m.Source != m2.Source {
		t.Errorf("Failed to unmarshal message expected (%+v) got (%+v)", m, m2)
	}

	if len(m.Data) != len(m2.Data) {
		t.Errorf("Msg Data Len expected(%d) got (%d)", len(m.Data), len(m2.Data))
	} else {
		for i := 0; i < len(m.Data); i++ {
			if m.Data[i] != m2.Data[i] {
				t.Errorf("Messages data[%d] expected (%d) got (%d)", i, m.Data[i], m2.Data[i])
			}
		}
	}

	if len(m.Path) != len(m2.Path) {
		t.Errorf("Msg Path Len expected(%d) got (%d)", len(m.Path), len(m2.Path))
	} else {
		for i := 0; i < len(m.Path); i++ {
			if m.Path[i] != m2.Path[i] {
				t.Errorf("Messages path[%d] expected (%s) got (%s)", i, m.Path[i], m2.Path[i])
			}
		}
	}

	if !m2.IsJSON() {
		t.Error("Msg expected to be JSON but is not ")
	}

	mpp, err := m2.Map()
	if err != nil {
		t.Errorf("Msg expected map but got an error (%s)", err)
	}

	for k, v := range mpp {
		switch k {
		case "int":
			if v != 10.0 {
				t.Errorf("Expected int (%d) got (%f)", 10, v)
			}
		case "float":
			if v != 12.3 {
				t.Errorf("Expected float (%f) got (%f)", 12.3, v)
			}
		case "string":
			if v != "45.6" {
				t.Errorf("Expected string (%s) got (%s)", "45.6", k)
			}

		}
	}

	m.Data = []byte("this is not json")
	if m.IsJSON() {
		t.Errorf("JSON expected (false) got (true) ")
	}
}
