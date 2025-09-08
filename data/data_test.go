package data

import (
	"encoding/json"
	"testing"
	"time"
)

func TestData(t *testing.T) {
	start := time.Now()
	values := []any{
		1,
		2.1,
		"strung",
		`{ "int": 1,  "float": 2.0, "string": "strung" }`,
		[]byte{'A', 'B', 'C'},
	}

	SetTruncateValue(time.Nanosecond)

	var datas []*Data
	for _, val := range values {
		datas = append(datas, NewData(val, time.Since(start)))
	}
	for _, dat := range datas {
		switch ty := dat.Value.(type) {
		case int:
			if dat.Value.(int) != 1 {
				t.Errorf("expected int value (1) got (%d)", dat.Value)
			}

		case float64:
			if dat.Value.(float64) != 2.1 {
				t.Errorf("expected int value (2.0) got (%3.1f)", dat.Value)
			}

		case string:
			str := dat.Value.(string)
			if json.Valid([]byte(str)) {
				var m map[string]interface{}
				err := json.Unmarshal([]byte(str), &m)
				if err != nil {
					t.Errorf("Failed to unmarshal data: %s", err)
				}
				continue
			}

			if str != "strung" {
				t.Errorf("expected int value (strung) got (%s)", dat.Value)
			}

		case []byte:
			if string(dat.Value.([]byte)) != "ABC" {
				t.Errorf("expected byte array (ABC) got (%s)", dat.Value)
			}

		default:
			t.Errorf("Unexpected data type %s", ty)
		}
	}

}
