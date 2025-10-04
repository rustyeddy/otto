package data

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SensorData struct {
	ID        string
	Type      string
	Value     float64
	Unit      string
	Timestamp time.Time
	Quality   string
}

type DeviceData struct {
	ID       string
	Name     string
	Type     string
	Status   string
	Location string
	Sensors  []SensorData
}

func TestSensorData(t *testing.T) {
	t.Run("Create SensorData", func(t *testing.T) {
		data := &SensorData{
			ID:        "temp-001",
			Type:      "temperature",
			Value:     23.5,
			Unit:      "celsius",
			Timestamp: time.Now(),
			Quality:   "good",
		}

		assert.Equal(t, "temp-001", data.ID)
		assert.Equal(t, "temperature", data.Type)
		assert.Equal(t, 23.5, data.Value)
		assert.Equal(t, "celsius", data.Unit)
		assert.Equal(t, "good", data.Quality)
		assert.False(t, data.Timestamp.IsZero())
	})

	t.Run("JSON Marshal/Unmarshal", func(t *testing.T) {
		original := &SensorData{
			ID:        "humid-001",
			Type:      "humidity",
			Value:     65.2,
			Unit:      "percent",
			Timestamp: time.Now().Truncate(time.Second), // Truncate for JSON precision
			Quality:   "excellent",
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "humid-001")
		assert.Contains(t, string(jsonData), "humidity")

		// Unmarshal back
		var unmarshaled SensorData
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, original.ID, unmarshaled.ID)
		assert.Equal(t, original.Type, unmarshaled.Type)
		assert.Equal(t, original.Value, unmarshaled.Value)
		assert.Equal(t, original.Unit, unmarshaled.Unit)
		assert.Equal(t, original.Quality, unmarshaled.Quality)
		assert.True(t, original.Timestamp.Equal(unmarshaled.Timestamp))
	})
}

func TestData(t *testing.T) {
	values := []any{
		1,
		2.1,
		"strung",
		`{ "int": 1,  "float": 2.0, "string": "strung" }`,
		[]byte{'A', 'B', 'C'},
	}

	SetTruncateValue(time.Nanosecond)

	var datas []Data
	for _, val := range values {
		datas = append(datas, NewData(val, time.Now()))
	}
	for _, dat := range datas {
		switch ty := dat.Value().(type) {
		case int:
			if dat.Value().(int) != 1 {
				t.Errorf("expected int value (1) got (%d)", dat.Value())
			}

		case float64:
			if dat.Value().(float64) != 2.1 {
				t.Errorf("expected int value (2.0) got (%3.1f)", dat.Value())
			}

		case string:
			str := dat.Value().(string)
			if json.Valid([]byte(str)) {
				var m map[string]interface{}
				err := json.Unmarshal([]byte(str), &m)
				if err != nil {
					t.Errorf("Failed to unmarshal data: %s", err)
				}
				continue
			}

			if str != "strung" {
				t.Errorf("expected int value (strung) got (%s)", dat.Value())
			}

		case []byte:
			if string(dat.Value().([]byte)) != "ABC" {
				t.Errorf("expected byte array (ABC) got (%s)", dat.Value())
			}

		default:
			t.Errorf("Unexpected data type %s", ty)
		}
	}

}
