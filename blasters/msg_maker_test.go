package blasters

import (
	"encoding/json"
	"testing"

	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/utils"
)

func TestMsgMaker(t *testing.T) {
	wd := WeatherData{}
	utils.SetStationName("test-station")
	msg := wd.NewMsg(messanger.GetTopics().Data("weather"))
	exp := []string{"ss", "d", "test-station", "weather"}
	for i := 0; i < len(exp); i++ {
		if msg.Path[i] != exp[i] {
			t.Errorf("expected path index[%d] to be (%s) got (%s)", i, exp[i], msg.Path[i])
		}
	}

	if msg.Source != "weather-data" {
		t.Errorf("expected path to be (weather-data) got (%s)", msg.Source)
	}

	var j WeatherData
	err := json.Unmarshal(msg.Byte(), &j)
	if err != nil {
		t.Errorf("failed to unmarshal message %s", err)
	}

	if j.Tempc == 0 {
		t.Errorf("tempc is not random got 0")
	}

	if j.Humidity == 0 {
		t.Errorf("humidity is not random got 0")
	}

}
