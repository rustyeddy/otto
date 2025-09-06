package messanger

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// DataManager is a map of Timeseries data that belongs to
// a specific station. The timeseries for each station are
// differentiated by the timeseries labels.
type DataManager struct {
	DataMap    map[string]map[string]*Timeseries `json:"datamap"`
	*Messanger `json:"-"`
}

var (
	dataManager *DataManager
)

func GetDataManager() *DataManager {
	if dataManager == nil {
		dataManager = NewDataManager()
	}
	return dataManager
}

// NewDataManager creates a new DataManager typically called
// by NewStation()
func NewDataManager() (dm *DataManager) {
	dm = &DataManager{
		DataMap:   make(map[string]map[string]*Timeseries),
		Messanger: NewMessanger("DataManager"),
	}
	return dm
}

// Add will add data according to station and label
func (dm *DataManager) Add(station, label string, data any) {
	stmap, ex := dm.DataMap[station]
	if !ex {
		dm.DataMap[station] = make(map[string]*Timeseries)
		stmap = dm.DataMap[station]
	}
	ts, ex := stmap[label]
	if !ex {
		stmap[label] = NewTimeseries(station, label)
	}
	ts = stmap[label]
	ts.Add(data)
}

func (dm *DataManager) Dump(w io.Writer) {
	for _, st := range dm.DataMap {
		for _, ts := range st {
			fmt.Fprint(w, ts.String())
		}
	}
}

// Callback is the callback used by the DataManager to receive
// MQTT messangers. TODO: move this call back to the stations because
// the stations will have a better understanding of the data they
// are subscribing to.
func (dm *DataManager) Callback(msg *Msg) {
	if msg.IsJSON() {
		m, err := msg.Map()
		if err != nil {
			return
		}
		for k, v := range m {
			dm.Add(msg.Station(), k, v)
		}
		return
	}

	dm.Add(msg.Station(), msg.Last(), msg.Data)
}

// ServeHTTP provides a REST interface to the config structure
func (dm DataManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		err := json.NewEncoder(w).Encode(dm)
		if err != nil {
			slog.Error("Failed to encode data:", "error", err)
			http.Error(w, "failure", 401)
		}

	case "POST", "PUT":
		http.Error(w, "Not Yet Supported", 401)
	}
}
