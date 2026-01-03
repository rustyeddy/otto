package station

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/rustyeddy/otto/messenger"
	"github.com/rustyeddy/otto/server"
	"github.com/rustyeddy/otto/utils"
)

// StationManager keeps track of all the stations we have seen
type StationManager struct {
	Stations map[string]*Station `json:"stations"`
	Stale    map[string]*Station `json:"stale"`
	EventQ   chan *StationEvent  `json:"-"`

	ticker *utils.Ticker `json:"-"`
	mu     *sync.Mutex   `json:"-"`
}

type StationEvent struct {
	Type      string `json:"type"`
	Device    string `json:"device"`
	StationID string `json:"stationid"`
	Value     string `json:"value"`
	Timestamp time.Time
}

var (
	stations *StationManager
	version  string
)

func GetStationManager() *StationManager {
	if stations == nil {
		stations = NewStationManager()
	}
	return stations
}

func resetStations() {
	stations = nil
	stations = NewStationManager()
}

func NewStationManager() (sm *StationManager) {
	sm = &StationManager{}
	sm.Stations = make(map[string]*Station)
	sm.Stale = make(map[string]*Station)
	sm.mu = new(sync.Mutex)
	// Buffer the event channel to avoid blocking producers and to satisfy tests
	sm.EventQ = make(chan *StationEvent, 100)
	return sm
}

func SetVersion(v string) {
	version = v
}

func (sm *StationManager) HandleMsg(msg *messenger.Msg) error {
	slog.Info("station manager recieved message on", "topic", msg.Topic)
	sm.Update(msg)
	return nil
}

func (sm *StationManager) Start() {

	srv := server.GetServer()
	srv.Register("/api/stations", sm)

	msgr := messenger.GetMessenger()
	msgr.Sub("o/d/+/hello", sm.HandleMsg)

	// Start a ticker to clean up stale entries
	quit := make(chan struct{})
	sm.ticker = utils.NewTicker("station-manager", 10*time.Second, func(t time.Time) {
		for id, st := range sm.Stations {

			// Don't try to timeout the local station
			if st.Local {
				continue
			}

			// Do not timeout stations with a duration of 0
			if st.Expiration == 0 {
				slog.Info("Station %s expiration == 0 do not timeout", "id", id)
				continue
			}

			// Timeout a station if we have not heard from it in 3
			// timeframes.
			st.mu.Lock()

			expires := st.LastHeard.Add(st.Expiration)
			if time.Until(expires) < 0 {
				sm.mu.Lock()
				slog.Info("Station has timed out", "station", id)
				sm.Stale[id] = st
				delete(sm.Stations, id)
				sm.mu.Unlock()
			}
			st.mu.Unlock()
		}
	})

	go func() {
		for {
			select {
			case ev := <-sm.EventQ:
				slog.Info("Station Event", "event", ev)
				st := sm.Get(ev.StationID)
				if st == nil {
					slog.Warn("Station Event could not find station", "station", ev.StationID)
					continue
				}

			case <-quit:
				sm.ticker.Stop()
				return
			}
		}
	}()
}

func (sm *StationManager) Stop() {
	for _, st := range sm.Stations {
		st.cancel()
	}
}

func (sm *StationManager) Get(stid string) *Station {
	sm.mu.Lock()
	st, _ := sm.Stations[stid]
	sm.mu.Unlock()
	return st
}

func (sm *StationManager) Add(st string) (station *Station, err error) {
	if sm.Get(st) != nil {
		return nil, fmt.Errorf("Error adding an existing station")
	}
	station, err = NewStation(st)
	if err != nil {
		return nil, err
	}

	slog.Info("station manager adding new station", "ID", st)
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.Stations[st] = station
	return station, nil
}

func (sm *StationManager) Update(msg *messenger.Msg) (st *Station) {
	var err error

	if !msg.Valid {
		slog.Error("Update with an invalid topic", "topic", msg.Topic)
		return nil
	}

	stid := msg.Station()
	if stid == "" {
		slog.Error("Msg path does not include staionId", "path", msg.Path)
		return nil
	}

	st = sm.Get(stid)
	if st == nil {
		if st, err = sm.Add(stid); err != nil {
			slog.Error("Station Manager failed to create new station", "stationid", stid)
			return nil
		}
	}

	// data := msg.Data
	st.Update(msg)
	return st
}

// Count returns the number of stations managed by this server
func (sm *StationManager) Count() int {
	return len(sm.Stations)
}

// ServeHTTP will handle all REST requests by clients returning an array
// of summarized stations.
func (sm StationManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		var stations []*Station
		for _, st := range sm.Stations {
			stations = append(stations, st)
		}
		json.NewEncoder(w).Encode(stations)

	case "POST", "PUT":
		http.Error(w, "Not Yet Supported", 401)
	}
}
