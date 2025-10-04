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

func TestDataPointTimestamp(t *testing.T) {
	t.Run("Returns correct timestamp", func(t *testing.T) {
		now := time.Now()
		dp := NewDataPoint("test value", now)

		assert.Equal(t, now, dp.Timestamp())
	})

	t.Run("Timestamp with zero time", func(t *testing.T) {
		zeroTime := time.Time{}
		dp := NewDataPoint(42, zeroTime)

		assert.Equal(t, zeroTime, dp.Timestamp())
		assert.True(t, dp.Timestamp().IsZero())
	})

	t.Run("Timestamp precision", func(t *testing.T) {
		// Test with nanosecond precision
		precise := time.Now().Add(123 * time.Nanosecond)
		dp := NewDataPoint("precise", precise)

		assert.Equal(t, precise, dp.Timestamp())
		assert.Equal(t, precise.Nanosecond(), dp.Timestamp().Nanosecond())
	})

	t.Run("Timestamp immutability", func(t *testing.T) {
		original := time.Now()
		dp := NewDataPoint("immutable test", original)

		// Get timestamp multiple times
		ts1 := dp.Timestamp()
		ts2 := dp.Timestamp()

		assert.Equal(t, original, ts1)
		assert.Equal(t, original, ts2)
		assert.Equal(t, ts1, ts2)

		// Modify the returned time (shouldn't affect internal timestamp)
		modified := ts1.Add(1 * time.Hour)
		assert.NotEqual(t, modified, dp.Timestamp())
		assert.Equal(t, original, dp.Timestamp())
	})
}

func TestNewDataPoint(t *testing.T) {
	t.Run("Create with string value", func(t *testing.T) {
		now := time.Now()
		dp := NewDataPoint("test string", now)

		require.NotNil(t, dp)
		assert.Equal(t, "test string", dp.Value())
		assert.Equal(t, now, dp.Timestamp())
	})

	t.Run("Create with numeric value", func(t *testing.T) {
		now := time.Now()
		dp := NewDataPoint(42.5, now)

		require.NotNil(t, dp)
		assert.Equal(t, 42.5, dp.Value())
		assert.Equal(t, now, dp.Timestamp())
	})

	t.Run("Create with nil value", func(t *testing.T) {
		now := time.Now()
		dp := NewDataPoint(nil, now)

		require.NotNil(t, dp)
		assert.Nil(t, dp.Value())
		assert.Equal(t, now, dp.Timestamp())
	})

	t.Run("Create with complex value", func(t *testing.T) {
		now := time.Now()
		complexValue := map[string]interface{}{
			"temperature": 23.5,
			"humidity":    65.2,
			"status":      "active",
		}
		dp := NewDataPoint(complexValue, now)

		require.NotNil(t, dp)
		assert.Equal(t, complexValue, dp.Value())
		assert.Equal(t, now, dp.Timestamp())
	})
}

func TestDataPoint_Value(t *testing.T) {
	t.Run("String value", func(t *testing.T) {
		dp := NewDataPoint("hello world", time.Now())
		assert.Equal(t, "hello world", dp.Value())
	})

	t.Run("Integer value", func(t *testing.T) {
		dp := NewDataPoint(42, time.Now())
		assert.Equal(t, 42, dp.Value())
	})

	t.Run("Float value", func(t *testing.T) {
		dp := NewDataPoint(3.14159, time.Now())
		assert.Equal(t, 3.14159, dp.Value())
	})

	t.Run("Boolean value", func(t *testing.T) {
		dp := NewDataPoint(true, time.Now())
		assert.Equal(t, true, dp.Value())
	})

	t.Run("Slice value", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		dp := NewDataPoint(slice, time.Now())
		assert.Equal(t, slice, dp.Value())
	})
}

func TestDataPoint_Float(t *testing.T) {
	t.Run("Valid float64 value", func(t *testing.T) {
		dp := NewDataPoint(42.5, time.Now())
		assert.Equal(t, 42.5, dp.Float())
	})

	t.Run("Integer converted to float64", func(t *testing.T) {
		// Note: This will panic because int != float64
		dp := NewDataPoint(42, time.Now())
		assert.Panics(t, func() {
			dp.Float()
		}, "Should panic when trying to convert int to float64")
	})

	t.Run("String value causes panic", func(t *testing.T) {
		dp := NewDataPoint("not a float", time.Now())
		assert.Panics(t, func() {
			dp.Float()
		}, "Should panic when trying to convert string to float64")
	})

	t.Run("Nil value causes panic", func(t *testing.T) {
		dp := NewDataPoint(nil, time.Now())
		assert.Panics(t, func() {
			dp.Float()
		}, "Should panic when trying to convert nil to float64")
	})

	t.Run("Zero float value", func(t *testing.T) {
		dp := NewDataPoint(0.0, time.Now())
		assert.Equal(t, 0.0, dp.Float())
	})

	t.Run("Negative float value", func(t *testing.T) {
		dp := NewDataPoint(-123.456, time.Now())
		assert.Equal(t, -123.456, dp.Float())
	})
}

func TestDataPointInt(t *testing.T) {
	t.Run("Valid integer value", func(t *testing.T) {
		dp := NewDataPoint(42, time.Now())
		assert.Equal(t, 42, dp.Int())
	})

	t.Run("Zero integer value", func(t *testing.T) {
		dp := NewDataPoint(0, time.Now())
		assert.Equal(t, 0, dp.Int())
	})

	t.Run("Negative integer value", func(t *testing.T) {
		dp := NewDataPoint(-100, time.Now())
		assert.Equal(t, -100, dp.Int())
	})

	t.Run("Float64 value causes panic", func(t *testing.T) {
		dp := NewDataPoint(42.5, time.Now())
		assert.Panics(t, func() {
			dp.Int()
		}, "Should panic when trying to convert float64 to int")
	})

	t.Run("String value causes panic", func(t *testing.T) {
		dp := NewDataPoint("42", time.Now())
		assert.Panics(t, func() {
			dp.Int()
		}, "Should panic when trying to convert string to int")
	})

	t.Run("Nil value causes panic", func(t *testing.T) {
		dp := NewDataPoint(nil, time.Now())
		assert.Panics(t, func() {
			dp.Int()
		}, "Should panic when trying to convert nil to int")
	})
}

func TestDataPointString(t *testing.T) {
	t.Run("String representation of string value", func(t *testing.T) {
		dp := NewDataPoint("hello", time.Now())
		result := dp.String()

		assert.Contains(t, result, "hello")
	})

	t.Run("String representation of integer value", func(t *testing.T) {
		dp := NewDataPoint(42, time.Now())
		result := dp.String()

		assert.Contains(t, result, "42")
	})

	t.Run("String representation of float value", func(t *testing.T) {
		dp := NewDataPoint(3.14, time.Now())
		result := dp.String()

		assert.Contains(t, result, "3.14")
	})

	t.Run("String representation of nil value", func(t *testing.T) {
		dp := NewDataPoint(nil, time.Now())
		result := dp.String()

		assert.Contains(t, result, "<nil>")
	})
}

func TestSetTruncateValue(t *testing.T) {
	t.Run("Set truncate value", func(t *testing.T) {
		original := Truncate
		defer func() { Truncate = original }() // Restore original value

		newValue := 5 * time.Minute
		SetTruncateValue(newValue)

		assert.Equal(t, newValue, Truncate)
	})

	t.Run("Set zero truncate value", func(t *testing.T) {
		original := Truncate
		defer func() { Truncate = original }()

		SetTruncateValue(0)
		assert.Equal(t, time.Duration(0), Truncate)
	})

	t.Run("Set negative truncate value", func(t *testing.T) {
		original := Truncate
		defer func() { Truncate = original }()

		negative := -1 * time.Hour
		SetTruncateValue(negative)
		assert.Equal(t, negative, Truncate)
	})

	t.Run("Multiple sets", func(t *testing.T) {
		original := Truncate
		defer func() { Truncate = original }()

		values := []time.Duration{
			1 * time.Second,
			1 * time.Minute,
			1 * time.Hour,
			24 * time.Hour,
		}

		for _, v := range values {
			SetTruncateValue(v)
			assert.Equal(t, v, Truncate)
		}
	})
}

func TestDataInterface(t *testing.T) {
	t.Run("DataPoint implements Data interface", func(t *testing.T) {
		var d Data
		dp := NewDataPoint("interface test", time.Now())
		d = dp

		// Should be able to call interface methods
		assert.NotNil(t, d.Value())
		assert.False(t, d.Timestamp().IsZero())
		assert.NotEmpty(t, d.String())
	})

	t.Run("Interface method consistency", func(t *testing.T) {
		dp := NewDataPoint("consistency", time.Now())
		var d Data = dp

		// Direct calls vs interface calls should be identical
		assert.Equal(t, dp.Value(), d.Value())
		assert.Equal(t, dp.Timestamp(), d.Timestamp())
		assert.Equal(t, dp.String(), d.String())
	})
}

func TestDataPointEdgeCases(t *testing.T) {
	t.Run("Very large integer", func(t *testing.T) {
		largeInt := int(1<<31 - 1) // Max int32
		dp := NewDataPoint(largeInt, time.Now())

		assert.Equal(t, largeInt, dp.Value())
		assert.Equal(t, largeInt, dp.Int())
	})

	t.Run("Very small float", func(t *testing.T) {
		smallFloat := 1e-10
		dp := NewDataPoint(smallFloat, time.Now())

		assert.Equal(t, smallFloat, dp.Value())
		assert.Equal(t, smallFloat, dp.Float())
	})

	t.Run("Very large float", func(t *testing.T) {
		largeFloat := 1e10
		dp := NewDataPoint(largeFloat, time.Now())

		assert.Equal(t, largeFloat, dp.Value())
		assert.Equal(t, largeFloat, dp.Float())
	})

	t.Run("Empty string", func(t *testing.T) {
		dp := NewDataPoint("", time.Now())

		assert.Equal(t, "", dp.Value())
		result := dp.String()
		assert.Contains(t, result, "")
	})

	t.Run("Unicode string", func(t *testing.T) {
		unicode := "Hello 世界 🌍"
		dp := NewDataPoint(unicode, time.Now())

		assert.Equal(t, unicode, dp.Value())
		assert.Contains(t, dp.String(), unicode)
	})
}
func TestNewData(t *testing.T) {
	t.Run("Create Data with string value", func(t *testing.T) {
		now := time.Now()
		data := NewData("test string", now)
		require.NotNil(t, data)

		// Verify it implements Data interface
		assert.Equal(t, "test string", data.Value())
		assert.Equal(t, now, data.Timestamp())
		assert.NotEmpty(t, data.String())
	})

	t.Run("Create Data with numeric value", func(t *testing.T) {
		now := time.Now()
		data := NewData(42.5, now)

		require.NotNil(t, data)
		assert.Equal(t, 42.5, data.Value())
		assert.Equal(t, now, data.Timestamp())
	})

	t.Run("Create Data with nil value", func(t *testing.T) {
		now := time.Now()
		data := NewData(nil, now)

		require.NotNil(t, data)
		assert.Nil(t, data.Value())
		assert.Equal(t, now, data.Timestamp())
	})

	t.Run("Create Data with complex value", func(t *testing.T) {
		now := time.Now()
		complexValue := map[string]interface{}{
			"temperature": 23.5,
			"humidity":    65.2,
			"status":      "active",
		}
		data := NewData(complexValue, now)

		require.NotNil(t, data)
		assert.Equal(t, complexValue, data.Value())
		assert.Equal(t, now, data.Timestamp())
	})

	t.Run("NewData interface compliance", func(t *testing.T) {
		var d Data
		d = NewData(123, time.Now())

		// Should be able to use as Data interface
		assert.Equal(t, 123, d.Value())
		assert.False(t, d.Timestamp().IsZero())
		assert.Contains(t, d.String(), "123")
	})
}
