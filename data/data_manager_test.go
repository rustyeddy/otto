package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rustyeddy/otto/messanger"
)

type InvValues struct {
	Feet   int `json:"feet"`
	Inches int `json:"inches"`
}

type FloatValues struct {
	Temperature float64 `json:"temperature"`
	Pressure    float64 `json:"pressure"`
	Humidity    float64 `json:"humidity"`
}

func TestNewDataManager(t *testing.T) {
	dm := GetDataManager()
	if len(dm.DataMap) != 0 {
		t.Errorf("Datamanger map not empty expected(0) got (%d)", len(dm.DataMap))
	}
}

func TestCallbackInts(t *testing.T) {
	data := []byte(fmt.Sprintf(`{ "int": 10 }`))
	path := "ss/d/station1/test"

	dm := GetDataManager()
	msg := messanger.New(path, data, "data-manager-test")
	dm.Callback(msg)

}

func TestConfigHTTP(t *testing.T) {
	dataManager = nil
	dm := GetDataManager()
	ts := httptest.NewServer(dm)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}

	dbuf, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	var datas DataManager
	err = json.Unmarshal(dbuf, &datas)
	if err != nil {
		t.Error(err)
	}
}
