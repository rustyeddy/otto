package station

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	messanger.Messanger  `json:"-"`
	device.DeviceManager `json:"devices"`

	errq   chan error
	errors []error `json:"errors"`

	time.Duration `json:"duration"`
	ticker        *time.Ticker `json:"-"`

	done   chan bool          `json:"-"`
	cancel context.CancelFunc `json:"-"`
	mu     sync.RWMutex       `json:"-"`

	// Add metrics
	Metrics *StationMetrics `json:"metrics"`
}

type Iface struct {
	Name    string
	IPAddrs []net.IP
	MACAddr string
}

type StationConfig struct {
	AnnouncementInterval time.Duration
	Timeout              time.Duration
	MaxErrors            int
	MessangerType        string
}

// NewStation creates a new Station with an ID as provided
// by the first parameter. Here we need to detect a duplicate
// station before trying to register another one.
func NewStation(id string) (*Station, error) {
	if id == "" {
		return nil, errors.New("station ID cannot be empty")
	}

	// Check for duplicate stations
	if existingStation := stations.Get(id); existingStation != nil {
		return nil, fmt.Errorf("station with ID %s already exists", id)
	}

	st := &Station{
		ID:         id,
		Expiration: 3 * time.Minute,
		Messanger:  messanger.NewMessangerMQTT(id),
		Duration:   1 * time.Minute,
		errq:       make(chan error, 10),
		done:       make(chan bool, 1),
		Metrics:    NewStationMetrics(),
	}

	go st.errorHandler()
	return st, nil
}

func (st *Station) errorHandler() {
	for {
		select {
		case <-st.done:
			return
		case err := <-st.errq:
			st.mu.Lock()
			// Limit error slice size
			if len(st.errors) > 100 {
				st.errors = st.errors[1:]
			}
			st.errors = append(st.errors, err)
			st.mu.Unlock()

			// Record error in metrics
			st.Metrics.RecordError()
		}
	}
}

// Initialize the local station
func (st *Station) Init() {
	// get IP addresses
	st.GetNetwork()

	topics := messanger.GetTopics()
	topics.SetStationName(st.Hostname)
	st.SetTopic(topics.Data("hello"))

	// Update network metrics
	interfaceCount := len(st.Ifaces)
	ipCount := 0
	for _, iface := range st.Ifaces {
		ipCount += len(iface.IPAddrs)
	}
	st.Metrics.UpdateNetworkMetrics(interfaceCount, ipCount)

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

	// Add context support for clean cancellation
	ctx, cancel := context.WithCancel(context.Background())
	st.cancel = cancel // Store cancel function

	go func() {
		defer st.ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
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
	start := time.Now()

	jbuf, err := json.Marshal(st)
	if err != nil {
		slog.Error("Failed to encode station info: ", "error", err)
		st.Metrics.RecordError()
		return
	}

	st.LastHeard = time.Now()
	st.PubData(string(jbuf))

	// Record metrics
	st.Metrics.RecordAnnouncement()
	st.Metrics.RecordMessageSent(uint64(len(jbuf)))
	st.Metrics.RecordResponseTime(time.Since(start))
}

// GetNetwork will set the IP addresses
func (st *Station) GetNetwork() error {
	h, err := os.Hostname()
	if err != nil {
		slog.Error("Failed to determine hostname", "error", err)
		st.Metrics.RecordError()
		return fmt.Errorf("hostname detection failed: %w", err) // Return instead of just logging
	}
	st.Hostname = h

	ifas, err := net.Interfaces()
	if err != nil {
		st.Metrics.RecordError()
		return err
	}

	for _, ifa := range ifas {
		addrs, err := ifa.Addrs()
		if err != nil {
			st.Metrics.RecordError()
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
	start := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastHeard = time.Now()

	// Record metrics
	s.Metrics.RecordAnnouncementReceived()
	s.Metrics.RecordMessageReceived(uint64(len(msg.Data)))
	s.Metrics.RecordResponseTime(time.Since(start))

	// TODO: Actually process the message data
}

// Stop the station from advertising
func (st *Station) Stop() {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.ticker != nil {
		st.ticker.Stop()
		st.ticker = nil
	}

	// Close channels and cleanup
	select {
	case st.done <- true:
	default:
	}

	if st.Messanger != nil {
		st.Messanger.Close()
	}
}

// AddDevice will do what it says by placing the device with a given
// name in the stations device manager. This library is basically a
// key value store, anything supporting the Name Interface:
// i.e. Name() string.
func (s *Station) AddDevice(device device.Name) {
	s.DeviceManager.Add(device)

	// Update device metrics
	deviceList := s.DeviceManager.List()
	s.Metrics.UpdateDeviceMetrics(len(deviceList), len(deviceList), 0) // TODO: Track active vs total
}

// GetDevice returns the device (anythig supporting the Name (Name()) interface)
func (s *Station) GetDevice(name string) any {
	d, _ := s.DeviceManager.Get(name)
	return d
}

// Create an endpoint for this device to be queried.
func (s Station) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.mu.RLock() // Read lock for concurrent access
		defer s.mu.RUnlock()

		if err := json.NewEncoder(w).Encode(s); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode station", "error", err)
			s.Metrics.RecordError()
		} else {
			s.Metrics.RecordResponseTime(time.Since(start))
		}

	case http.MethodPost, http.MethodPut:
		http.Error(w, "Method not implemented", http.StatusNotImplemented)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (st *Station) IsHealthy() bool {
	healthy := time.Since(st.LastHeard) < st.Expiration
	st.Metrics.RecordHealthCheck(healthy)
	return healthy
}

// GetMetricsEndpoint provides an HTTP endpoint for metrics
func (st *Station) GetMetricsEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		metrics := st.Metrics.GetMetrics()
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
			slog.Error("Failed to encode metrics", "error", err)
		}
	}
}
