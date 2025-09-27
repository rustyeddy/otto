package station

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/utils"
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
	*device.DeviceManager `json:"devices"`

	errq   chan error
	errors []error `json:"errors"`

	time.Duration `json:"duration"`
	ticker        *time.Ticker `json:"-"`

	done   chan bool          `json:"-"`
	cancel context.CancelFunc `json:"-"`
	mu     sync.RWMutex       `json:"-"`

	// Add metrics
	Metrics *StationMetrics `json:"metrics"`

	// Internal generic device store for tests and loose coupling
	devicesMu sync.RWMutex `json:"-"`
	devices   map[string]any
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

// Helper function to detect test mode
func isTestMode() bool {
	// Detect "go test" by checking executable name and -test.* flags
	base := filepath.Base(os.Args[0])
	if strings.HasSuffix(base, ".test") {
		return true
	}
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-test.") {
			return true
		}
	}
	return false
}

// NoopMessanger satisfies messanger.Messanger without doing any work.
// This is used during tests to prevent external connections.
type NoopMessanger struct {
	topic string
}

func (n *NoopMessanger) ID() string { return "noop" }
func (n *NoopMessanger) Subscribe(topic string, handler messanger.MsgHandler) error {
	// no-op
	return nil
}
func (n *NoopMessanger) SetTopic(topic string) { n.topic = topic }
func (n *NoopMessanger) Topic() string         { return n.topic }

// PubMsg no-op implementation
func (n *NoopMessanger) PubMsg(msg *messanger.Msg) error {
	// no-op
	return nil
}

// PubData no-op implementation
func (n *NoopMessanger) PubData(_ any) error {
	// no-op
	return nil
}
func (n *NoopMessanger) Error() error { return nil }
func (n *NoopMessanger) Close()       {}

// NewStation creates a new Station with an ID as provided
// by the first parameter. Here we need to detect a duplicate
// station before trying to register another one.
func newStation(id string) (*Station, error) {
	if id == "" {
		return nil, errors.New("station ID cannot be empty")
	}

	// Check for duplicate stations using the station manager
	sm := GetStationManager()
	if existingStation := sm.Get(id); existingStation != nil {
		return nil, fmt.Errorf("station with ID %s already exists", id)
	}

	st := &Station{
		ID:         id,
		Expiration: 3 * time.Minute,
		Duration:   1 * time.Minute,
		errq:       make(chan error, 10),
		done:       make(chan bool, 1),
		Metrics:    NewStationMetrics(),
		devices:    make(map[string]any),
	}

	// Use a no-op messanger when running tests to avoid external deps
	if isTestMode() {
		st.Messanger = &NoopMessanger{}
	} else {
		// Use the workspace-wide topic name from utils for station topics
		topic := utils.StationName()
		if topic == "" {
			// fallback to previous pattern if utils not configured
			topic = "otto/stations"
		}
		messanger.NewMessanger("local", topic+"/"+id)
		st.Messanger = messanger.GetMessanger()
	}

	go st.errorHandler()
	return st, nil
}

// Initialize the local station
func (st *Station) Init() {
	// get IP addresses
	_ = st.GetNetwork()

	// set hostname
	if hn, err := os.Hostname(); err == nil {
		st.Hostname = hn
	}

	// Update network metrics
	if st.Metrics != nil {
		ifCount := len(st.Ifaces)
		ipCount := 0
		for _, ifc := range st.Ifaces {
			ipCount += len(ifc.IPAddrs)
		}
		st.Metrics.UpdateNetworkMetrics(ifCount, ipCount)
	}

	// mark last heard now
	st.mu.Lock()
	st.LastHeard = time.Now()
	st.mu.Unlock()
}

func (st *Station) SaveError(err error) {
	if err == nil {
		return
	}
	select {
	case st.errq <- err:
	default:
		// channel full, record directly
		if st.Metrics != nil {
			st.Metrics.RecordError()
		}
	}
}

// StartTicker will cause the station timer to go off at
// st.Duration time periods to either perform an announcement
// or in the case we are a hub we will time the station out after
// station.Period * 3.
func (st *Station) StartTicker(duration time.Duration) error {
	if duration <= 0 {
		duration = st.Duration
		if duration <= 0 {
			duration = time.Minute
		}
	}

	// Add context support for clean cancellation
	ctx, cancel := context.WithCancel(context.Background())
	st.cancel = cancel

	// create ticker
	st.ticker = time.NewTicker(duration)

	// just to get started
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-st.done:
				return
			case <-st.ticker.C:
				st.SayHello()
			}
		}
	}()

	return nil
}

func (st *Station) SayHello() {
	// publish a simple hello payload if available
	if st.Messanger != nil {
		payload := map[string]any{
			"type": "hello",
			"id":   st.ID,
			"ts":   time.Now().UTC(),
		}
		if err := st.Messanger.PubData(payload); err != nil {
			// record the error for metrics / diagnostics
			st.SaveError(fmt.Errorf("SayHello publish failed: %w", err))
		}
	}

	// Record metrics
	if st.Metrics != nil {
		st.Metrics.RecordAnnouncement()
	}

	st.mu.Lock()
	st.LastHeard = time.Now()
	st.mu.Unlock()
}

// GetNetwork will set the IP addresses
func (st *Station) GetNetwork() error {
	var ifaces []*Iface

	sysIfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, sysIf := range sysIfaces {
		// skip down interfaces
		if sysIf.Flags&net.FlagUp == 0 {
			continue
		}
		// create iface
		ifc := &Iface{
			Name:    sysIf.Name,
			MACAddr: sysIf.HardwareAddr.String(),
		}
		addrs, _ := sysIf.Addrs()
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			ifc.IPAddrs = append(ifc.IPAddrs, ip)
		}

		ifaces = append(ifaces, ifc)
	}

	st.mu.Lock()
	st.Ifaces = ifaces
	st.mu.Unlock()

	// update network metrics
	if st.Metrics != nil {
		ifCount := len(ifaces)
		ipCount := 0
		for _, ifc := range ifaces {
			ipCount += len(ifc.IPAddrs)
		}
		st.Metrics.UpdateNetworkMetrics(ifCount, ipCount)
	}

	return nil
}

// Update() will append a new data value to the series
// of data points.
func (s *Station) Update(msg *messanger.Msg) {
	// Record metrics
	if s.Metrics != nil && msg != nil {
		var size uint64
		if msg.Data != nil {
			size = uint64(len(msg.Data))
		}
		s.Metrics.RecordMessageReceived(size)
	}

	// Update last heard
	s.mu.Lock()
	s.LastHeard = time.Now()
	s.mu.Unlock()

	// TODO: Actually process the message data
}

// Stop the station from advertising
func (st *Station) Stop() {
	// Cancel context if available
	if st.cancel != nil {
		st.cancel()
	}

	// Stop ticker
	if st.ticker != nil {
		st.ticker.Stop()
	}

	// Close channels and cleanup
	select {
	case st.done <- true:
	default:
	}

	// Safely close messanger
	if st.Messanger != nil {
		st.Messanger.Close()
	}
}

// AddDevice will do what it says by placing the device with a given
// name in the stations device manager. This library is basically a
// key value store, anything supporting the Name Interface:
// i.e. Name() string.
func (s *Station) AddDevice(d device.Name) {
	if d == nil {
		return
	}
	name := d.Name()

	// store generically
	s.devicesMu.Lock()
	if s.devices == nil {
		s.devices = make(map[string]any)
	}
	s.devices[name] = d
	devCount := len(s.devices)
	s.devicesMu.Unlock()

	// Update device metrics
	if s.Metrics != nil {
		s.Metrics.UpdateDeviceMetrics(devCount, devCount, s.Metrics.DeviceErrorCount)
	}
	// TODO: Track active vs total
}

// GetDevice returns the device (anythig supporting the Name (Name()) interface)
func (s *Station) GetDevice(name string) any {
	s.devicesMu.RLock()
	defer s.devicesMu.RUnlock()
	return s.devices[name]
}

// Create an endpoint for this device to be queried.
func (s Station) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}

func (st *Station) IsHealthy() bool {
	st.mu.RLock()
	last := st.LastHeard
	exp := st.Expiration
	st.mu.RUnlock()

	healthy := true
	if exp > 0 {
		healthy = time.Since(last) < exp
	}

	// record health check
	if st.Metrics != nil {
		st.Metrics.RecordHealthCheck(healthy)
	}
	return healthy
}

// GetMetricsEndpoint provides an HTTP endpoint for metrics
func (st *Station) GetMetricsEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		m := st.Metrics.GetMetrics()
		_ = json.NewEncoder(w).Encode(m)
	}
}

// errorHandler consumes the error queue and records metrics
func (st *Station) errorHandler() {
	for err := range st.errq {
		if err == nil {
			continue
		}
		st.mu.Lock()
		st.errors = append(st.errors, err)
		st.mu.Unlock()

		if st.Metrics != nil {
			st.Metrics.RecordError()
		}
	}
}
