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
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "test", omsg.Last())

	j, err := json.Marshal(omsg)
	if !assert.NoError(t, err) {
		return
	}

	msg := NewMsg(topic, j, "test")
	if !assert.NotNil(t, msg) {
		return
	}
	assert.Equal(t, topic, msg.Topic)
	path := strings.Split(topic, "/")
	assert.Equal(t, path, msg.Path)
	assert.Equal(t, "test", msg.Source)
	assert.Equal(t, j, msg.Data)
}

func TestJSON(t *testing.T) {
	m, _ := getMsg()

	jstr := `{ "int": 10, "float": 12.3, "string": "45.6" }`
	m.Data = []byte(jstr)

	jbyte, err := json.Marshal(m)
	if !assert.NoError(t, err) {
		return
	}

	var m2 Msg
	if !assert.NoError(t, json.Unmarshal(jbyte, &m2)) {
		return
	}

	assert.Equal(t, m.ID, m2.ID)
	assert.Equal(t, m.Topic, m2.Topic)
	assert.Equal(t, m.Timestamp, m2.Timestamp)
	assert.Equal(t, m.Source, m2.Source)

	assert.Equal(t, m.Data, m2.Data)
	assert.Equal(t, m.Path, m2.Path)

	assert.True(t, m2.IsJSON())

	mpp, err := m2.Map()
	if !assert.NoError(t, err) {
		return
	}

	if v, ok := mpp["int"]; ok {
		assert.Equal(t, 10.0, v)
	}
	if v, ok := mpp["float"]; ok {
		assert.Equal(t, 12.3, v)
	}
	if v, ok := mpp["string"]; ok {
		assert.Equal(t, "45.6", v)
	}

	m.Data = []byte("this is not json")
	assert.False(t, m.IsJSON())
}

func TestMsg_Station(t *testing.T) {
	msg := &Msg{Path: []string{"ss", "d", "be:ef:ca:fe:01", "station"}}
	assert.Equal(t, "be:ef:ca:fe:01", msg.Station())

	msg = &Msg{Path: []string{"ss", "d"}}
	assert.Equal(t, "", msg.Station())
}

func TestMsg_Last(t *testing.T) {
	msg := &Msg{Path: []string{"ss", "d", "test"}}
	assert.Equal(t, "test", msg.Last())

	msg = &Msg{Path: []string{}}
	assert.Equal(t, "", msg.Last())
}

func TestMsg_Byte(t *testing.T) {
	data := []byte("test data")
	msg := &Msg{Data: data}
	assert.Equal(t, string(data), string(msg.Byte()))
}

func TestMsg_String(t *testing.T) {
	data := []byte("test data")
	msg := &Msg{Data: data}
	assert.Equal(t, string(data), msg.String())
}

func TestMsg_Float64(t *testing.T) {
	data := []byte("123.45")
	msg := &Msg{Data: data}
	assert.Equal(t, 123.45, msg.Float64())
}

func TestMsg_IsJSON(t *testing.T) {
	msg := &Msg{Data: []byte(`{"key": "value"}`)}
	assert.True(t, msg.IsJSON())

	msg = &Msg{Data: []byte("not json")}
	assert.False(t, msg.IsJSON())
}

func TestMsg_JSON(t *testing.T) {
	msg := &Msg{ID: 1, Topic: "test", Data: []byte("test data")}
	jsonData, err := msg.JSON()
	assert.NoError(t, err)

	var unmarshalledMsg Msg
	assert.NoError(t, json.Unmarshal(jsonData, &unmarshalledMsg))
	assert.Equal(t, msg.ID, unmarshalledMsg.ID)
	assert.Equal(t, msg.Topic, unmarshalledMsg.Topic)
}

func TestMsg_Map(t *testing.T) {
	msg := &Msg{Data: []byte(`{"key": "value"}`)}
	m, err := msg.Map()
	assert.NoError(t, err)
	assert.Equal(t, "value", m["key"])

	msg = &Msg{Data: []byte("not json")}
	_, err = msg.Map()
	assert.Error(t, err)
}

func TestNewMsg(t *testing.T) {
	topic := "test/topic"
	data := []byte("test data")
	source := "test-source"

	msg := NewMsg(topic, data, source)

	assert.True(t, msg.ID > 0)
	assert.Equal(t, topic, msg.Topic)
	assert.Equal(t, string(data), string(msg.Data))
	assert.Equal(t, source, msg.Source)
	assert.True(t, msg.Timestamp > 0)
	assert.Len(t, msg.Path, 2)
	assert.Equal(t, "test", msg.Path[0])
	assert.Equal(t, "topic", msg.Path[1])

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
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
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
		assert.Equal(t, "station1", msg.Station())
	})

	t.Run("Last", func(t *testing.T) {
		assert.Equal(t, "sensor", msg.Last())
	})

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, "42.5", msg.String())
	})

	t.Run("Float64", func(t *testing.T) {
		assert.Equal(t, 42.5, msg.Float64())
	})
}

func TestMsgJSON(t *testing.T) {
	msg := NewMsg("test/topic", []byte(`{"key":"value"}`), "test")

	t.Run("IsJSON", func(t *testing.T) {
		assert.True(t, msg.IsJSON())
	})

	t.Run("JSON", func(t *testing.T) {
		jsonBytes, err := msg.JSON()
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, json.Valid(jsonBytes))
	})

	t.Run("Map", func(t *testing.T) {
		m, err := msg.Map()
		if !assert.NoError(t, err) {
			return
		}
		if v, ok := m["key"]; assert.True(t, ok) {
			assert.Equal(t, "value", v)
		}
	})
}

func TestMsgSaver(t *testing.T) {
	saver := GetMsgSaver()

	t.Run("StartSaving", func(t *testing.T) {
		saver.StartSaving()
		assert.Equal(t, 0, len(saver.Messages))

		iid := msgid
		for i := 0; i < 10; i++ {
			_ = NewMsg("o/d/test-station/message", []byte("hello, world!"), "testing")
		}
		assert.Equal(t, 10, len(saver.Messages))
		for _, smsg := range saver.Messages {
			iid++
			assert.Equal(t, iid, smsg.ID)
		}
	})

	t.Run("StopSaving", func(t *testing.T) {
		saver.StopSaving()
		initialLen := len(saver.Messages)
		NewMsg("test/topic", []byte("test"), "test")
		assert.Equal(t, initialLen, len(saver.Messages))
	})

	t.Run("ServeHTTP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/messages", nil)
		w := httptest.NewRecorder()

		saver.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})
}

func TestMsgBool(t *testing.T) {
	tsts := []struct {
		val    string
		expect bool
	}{
		{"true", true},
		{"false", false},
		{"ok", true},
		{"on", true},
		{"1", true},
		{"0", false},
	}

	for _, tst := range tsts {
		msg := NewMsg("o/d/station/booltest", []byte(tst.val), "booltest")
		assert.Equal(t, tst.expect, msg.Bool())
	}

}

func TestMsgDump(t *testing.T) {
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

	assert.Contains(t, dump, "ID: 123")
	assert.Contains(t, dump, `Path: ["test" "topic"]`)
	assert.Contains(t, dump, `Args: ["arg1" "arg2"]`)
	assert.Contains(t, dump, "Src: test source")
	assert.Contains(t, dump, "Time: ")
	assert.Contains(t, dump, "Data: test data")
}
