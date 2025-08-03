package station

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/server"
)

// Station is the primary structure that holds an array of
// Sensors which in turn hold a timeseries of datapoints.
type Station struct {
	ID         string        `json:"id"`
	LastHeard  time.Time     `json:"last-heard"`
	Expiration time.Duration `json:"expiration"` // how long to timeout a station
	Hostname   string        `json:"hostname"`
	Local      bool          `json:"local"`
	Ifaces     []*Iface      `json:"iface"`

	*messanger.Messanger `json:"messanger"`
	device.DeviceManager `json:"devices"`

	errq   chan error
	errors []error `json:"errors"`

	time.Duration `json:"duration"`
	ticker        *time.Ticker `json:"-"`

	done chan bool  `json:"-"`
	mu   sync.Mutex `json:"-"`
}

type Iface struct {
	Name    string
	IPAddrs []net.IP
	MACAddr string
}

// NewStation creates a new Station with an ID as provided
// by the first parameter. Here we need to detect a duplicate
// station before trying to register another one.
func NewStation(id string) (st *Station) {
	st = &Station{
		ID:         id,
		Expiration: 3 * time.Minute,
		Messanger:  messanger.NewMessanger(id),
		Duration:   1 * time.Minute,
	}

	st.errq = make(chan error)
	go func() {
		for {
			select {
			case <-st.done:
				return

			case err := <-st.errq:
				st.errors = append(st.errors, err)
			}
		}
	}()

	return st
}

// Initialize the local station
func (st *Station) Init() {
	// get IP addresses
	st.GetNetwork()

	topics := messanger.GetTopics()
	topics.SetStationName(st.Hostname)
	st.Topic = topics.Data("hello")

	// start either an announcement timer or a timer to timeout
	// stale stations
	if st.Duration != 0 {
		err := st.StartTicker(st.Duration)
		if err != nil {
			st.SaveError(err)
			slog.Error("ticker failed", "error", err)
		}
	}
}

func (st *Station) SaveError(err error) {
	st.errq <- err
}

// StartTicker will cause the station timer to go off at
// st.Duration time periods to either perform an announcement
// or in the case we are a hub we will time the station out after
// station.Period * 3.
func (st *Station) StartTicker(duration time.Duration) error {
	if st.ticker != nil {
		return errors.New("Station ticker is already running")
	}
	st.ticker = time.NewTicker(duration)
	go func() {
		defer st.ticker.Stop()
		// TODO pass in done
		for {
			select {
			case <-st.ticker.C:
				st.SayHello()
			}
		}
	}()
	// just to get started
	st.SayHello()
	return nil
}

func (st *Station) SayHello() {
	jbuf, err := json.Marshal(st)
	if err != nil {
		slog.Error("Failed to encode station info: ", "error", err)
		return
	}
	st.LastHeard = time.Now()
	st.PubData(string(jbuf))
}

// GetNetwork will set the IP addresses
func (st *Station) GetNetwork() error {
	h, err := os.Hostname()
	if err != nil {
		slog.Error("Failed to determine out hostname", "error", err)
		st.errq <- err
	}
	st.Hostname = h

	ifas, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, ifa := range ifas {
		addrs, err := ifa.Addrs()
		if err != nil {
			return err
		}

		ifs := &Iface{
			Name:    ifa.Name,
			MACAddr: ifa.HardwareAddr.String(),
		}

		var ip net.IP
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			default:
				continue
			}

			if ip.IsLoopback() || ip.IsMulticast() ||
				ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
				continue
			}

			ifs.IPAddrs = append(ifs.IPAddrs, ip)
		}

		if len(ifs.IPAddrs) == 0 {
			continue
		}
		st.Ifaces = append(st.Ifaces, ifs)
	}
	// 10.11.5.3/16 10.11.78.252/16
	return nil
}

func (st *Station) Register() {
	// this needs to move
	srv := server.GetServer()
	srv.Register("/api/station/"+st.ID, st)
}

// Update() will append a new data value to the series
// of data points.
func (s *Station) Update(msg *messanger.Msg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastHeard = time.Now()
}

// Stop the station from advertising
func (st *Station) Stop() {
	if st.ticker != nil {
		st.ticker.Stop()
	}
	st.done <- true
}

// AddDevice will do what it says by placing the device with a given
// name in the stations device manager. This library is basically a
// key value store, anything supporting the Name Interface:
// i.e. Name() string.
func (s *Station) AddDevice(device device.Name) {
	s.DeviceManager.Add(device)
}

// GetDevice returns the device (anythig supporting the Name (Name()) interface)
func (s *Station) GetDevice(name string) any {
	d, _ := s.DeviceManager.Get(name)
	return d
}

// Create an endpoint for this device to be queried.
func (s Station) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(s)

	case "POST", "PUT":
		http.Error(w, "Not Yet Supported", 401)
	}
}
