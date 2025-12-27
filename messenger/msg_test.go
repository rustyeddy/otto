package messenger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rustyeddy/otto/utils"
)

func getMsg() (*Msg, time.Time) {
	now := time.Now()
	path := "ss/d/%s/test"

	b := fmt.Sprintf("%d", 4)
	m := NewMsg(path, []byte(b), "test")
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

	msg := NewMsg(topic, j, "test")
	if msg == nil {
		t.Fatal("msg topic expected but is nil")
		return // This is redundant after t.Fatal but makes static analyzer happy
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

func TestMsg_Station(t *testing.T) {
	msg := &Msg{Path: []string{"ss", "d", "be:ef:ca:fe:01", "station"}}
	if msg.Station() != "be:ef:ca:fe:01" {
		t.Errorf("expected station 'station', got '%s'", msg.Station())
	}

	msg = &Msg{Path: []string{"ss", "d"}}
	if msg.Station() != "" {
		t.Errorf("expected station '', got '%s'", msg.Station())
	}
}

func TestMsg_Last(t *testing.T) {
	msg := &Msg{Path: []string{"ss", "d", "test"}}
	if msg.Last() != "test" {
		t.Errorf("expected last 'test', got '%s'", msg.Last())
	}

	msg = &Msg{Path: []string{}}
	if msg.Last() != "" {
		t.Errorf("expected last '', got '%s'", msg.Last())
	}
}

func TestMsg_Byte(t *testing.T) {
	data := []byte("test data")
	msg := &Msg{Data: data}
	if string(msg.Byte()) != string(data) {
		t.Errorf("expected byte data '%s', got '%s'", string(data), string(msg.Byte()))
	}
}

func TestMsg_String(t *testing.T) {
	data := []byte("test data")
	msg := &Msg{Data: data}
	if msg.String() != string(data) {
		t.Errorf("expected string data '%s', got '%s'", string(data), msg.String())
	}
}

func TestMsg_Float64(t *testing.T) {
	data := []byte("123.45")
	msg := &Msg{Data: data}
	if msg.Float64() != 123.45 {
		t.Errorf("expected float64 data '123.45', got '%f'", msg.Float64())
	}
}

func TestMsg_IsJSON(t *testing.T) {
	msg := &Msg{Data: []byte(`{"key": "value"}`)}
	if !msg.IsJSON() {
		t.Errorf("expected IsJSON to return true, got false")
	}

	msg = &Msg{Data: []byte("not json")}
	if msg.IsJSON() {
		t.Errorf("expected IsJSON to return false, got true")
	}
}

func TestMsg_JSON(t *testing.T) {
	msg := &Msg{ID: 1, Topic: "test", Data: []byte("test data")}
	jsonData, err := msg.JSON()
	if err != nil {
		t.Errorf("expected no error, got '%v'", err)
	}

	var unmarshalledMsg Msg
	err = json.Unmarshal(jsonData, &unmarshalledMsg)
	if err != nil {
		t.Errorf("expected no error during unmarshal, got '%v'", err)
	}

	if unmarshalledMsg.ID != msg.ID || unmarshalledMsg.Topic != msg.Topic {
		t.Errorf("expected unmarshalled message to match original, got '%v'", unmarshalledMsg)
	}
}

func TestMsg_Map(t *testing.T) {
	msg := &Msg{Data: []byte(`{"key": "value"}`)}
	m, err := msg.Map()
	if err != nil {
		t.Errorf("expected no error, got '%v'", err)
	}

	if m["key"] != "value" {
		t.Errorf("expected map value 'value', got '%v'", m["key"])
	}

	msg = &Msg{Data: []byte("not json")}
	_, err = msg.Map()
	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
}

func TestNewMsg(t *testing.T) {
	topic := "test/topic"
	data := []byte("test data")
	source := "test-source"

	msg := NewMsg(topic, data, source)

	if msg.ID <= 0 {
		t.Error("Expected positive message ID")
	}
	if msg.Topic != topic {
		t.Errorf("Expected topic %s, got %s", topic, msg.Topic)
	}
	if string(msg.Data) != string(data) {
		t.Errorf("Expected data %s, got %s", string(data), string(msg.Data))
	}
	if msg.Source != source {
		t.Errorf("Expected source %s, got %s", source, msg.Source)
	}
	if msg.Timestamp <= 0 {
		t.Error("Expected positive timestamp")
	}
	if len(msg.Path) != 2 || msg.Path[0] != "test" || msg.Path[1] != "topic" {
		t.Errorf("expected path ['test', 'topic'], got '%v'", msg.Path)
	}

}

func TestBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{"byte slice", []byte("test"), "test", false},
		{"string", "test", "test", false},
		{"integer", 42, "42", false},
		{"boolean true", true, "true", false},
		{"boolean false", false, "false", false},
		{"float64", 3.14, " 3.14", false},
		{"invalid type", struct{}{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Bytes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != tt.want {
				t.Errorf("Bytes() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestMsgMethods(t *testing.T) {
	msg := &Msg{
		ID:        1,
		Topic:     "ss/d/station1/sensor",
		Path:      []string{"ss", "d", "station1", "sensor"},
		Data:      []byte("42.5"),
		Source:    "test",
		Timestamp: time.Hour,
	}

	t.Run("Station", func(t *testing.T) {
		if got := msg.Station(); got != "station1" {
			t.Errorf("Station() = %v, want %v", got, "station1")
		}
	})

	t.Run("Last", func(t *testing.T) {
		if got := msg.Last(); got != "sensor" {
			t.Errorf("Last() = %v, want %v", got, "sensor")
		}
	})

	t.Run("String", func(t *testing.T) {
		if got := msg.String(); got != "42.5" {
			t.Errorf("String() = %v, want %v", got, "42.5")
		}
	})

	t.Run("Float64", func(t *testing.T) {
		if got := msg.Float64(); got != 42.5 {
			t.Errorf("Float64() = %v, want %v", got, 42.5)
		}
	})
}

func TestMsgJSON(t *testing.T) {
	msg := NewMsg("test/topic", []byte(`{"key":"value"}`), "test")

	t.Run("IsJSON", func(t *testing.T) {
		if !msg.IsJSON() {
			t.Error("IsJSON() should return true for valid JSON")
		}
	})

	t.Run("JSON", func(t *testing.T) {
		jsonBytes, err := msg.JSON()
		if err != nil {
			t.Errorf("JSON() error = %v", err)
		}
		if !json.Valid(jsonBytes) {
			t.Error("JSON() should return valid JSON")
		}
	})

	t.Run("Map", func(t *testing.T) {
		m, err := msg.Map()
		if err != nil {
			t.Errorf("Map() error = %v", err)
		}
		if v, ok := m["key"]; !ok || v != "value" {
			t.Errorf("Map() = %v, want map[key:value]", m)
		}
	})
}

func TestMsgSaver(t *testing.T) {
	saver := GetMsgSaver()

	t.Run("StartSaving", func(t *testing.T) {
		saver.StartSaving()
		if len(saver.Messages) != 0 {
			t.Error("Message should be saved when saving is enabled")
		}
	})

	t.Run("StopSaving", func(t *testing.T) {
		saver.StopSaving()
		initialLen := len(saver.Messages)
		NewMsg("test/topic", []byte("test"), "test")
		if len(saver.Messages) != initialLen {
			t.Error("Message should not be saved when saving is disabled")
		}
	})

	t.Run("ServeHTTP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/messages", nil)
		w := httptest.NewRecorder()

		saver.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("ServeHTTP() status = %v, want %v", resp.StatusCode, http.StatusOK)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("ServeHTTP() content-type = %v, want application/json", ct)
		}
	})
}
func TestMsg_Dump(t *testing.T) {
	msg := &Msg{
		ID:        123,
		Topic:     "test/topic",
		Path:      []string{"test", "topic"},
		Args:      []string{"arg1", "arg2"},
		Data:      []byte("test data"),
		Source:    "test source",
		Timestamp: 123456789,
	}

	dump := msg.Dump()

	if !strings.Contains(dump, "ID: 123") {
		t.Errorf("Dump() missing ID, got: %s", dump)
	}
	if !strings.Contains(dump, `Path: ["test" "topic"]`) {
		t.Errorf("Dump() missing Path, got: %s", dump)
	}
	if !strings.Contains(dump, `Args: ["arg1" "arg2"]`) {
		t.Errorf("Dump() missing Args, got: %s", dump)
	}
	if !strings.Contains(dump, "Src: test source") {
		t.Errorf("Dump() missing Source, got: %s", dump)
	}
	if !strings.Contains(dump, "Time: ") {
		t.Errorf("Dump() missing Timestamp, got: %s", dump)
	}
	if !strings.Contains(dump, "Data: test data") {
		t.Errorf("Dump() missing Data, got: %s", dump)
	}
}
