package station

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/rustyeddy/otto/messenger"
)

// Station is the primary structure that holds an array of
// Sensors which in turn hold a timeseries of datapoints.
type Station struct {
	ID       string   `json:"id"`
	Hostname string   `json:"hostname"`
	Local    bool     `json:"local"`
	Ifaces   []*Iface `json:"iface"`
	Version  string   `json:"version"`

	Devices *DeviceManager  `json:"device-manager"`
	Metrics *StationMetrics `json:"-"` // `json:"metrics"`

	LastHeard     time.Time     `json:"last-heard"`
	Expiration    time.Duration `json:"expiration"` // how long to timeout a station
	time.Duration `json:"duration"`
	ticker        *time.Ticker `json:"-"`

	errq   chan error
	errors []error `json:"-"`

	done   chan bool          `json:"-"`
	cancel context.CancelFunc `json:"-"`
	mu     sync.RWMutex       `json:"-"`
}

type Iface struct {
	Name    string
	IPAddrs []netip.Addr
	MACAddr string
}

type StationConfig struct {
	AnnouncementInterval time.Duration
	Timeout              time.Duration
	MaxErrors            int
	MessengerType        string
}

// NewStation creates a new Station with an ID as provided
// by the first parameter. Here we need to detect a duplicate
// station before trying to register another one.
func NewStation(id string) (*Station, error) {
	if id == "" {
		return nil, errors.New("station ID cannot be empty")
	}

	st := &Station{
		ID:         id,
		Expiration: 3 * time.Minute,
		Duration:   1 * time.Minute,
		errq:       make(chan error, 10),
		done:       make(chan bool, 1),
		Metrics:    NewStationMetrics(),
		Devices:    NewDeviceManager(),
		Version:    version,
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
	defer st.mu.Unlock()
	st.LastHeard = time.Now()
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

	slog.Info("starting station ticker", "station", st.ID, "duration", st.Duration)

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
	slog.Info("sending hello message", "station", st.ID)

	// XXX: Record metrics - this needs to only be run if
	// there is a broker listening. That is moved under the
	// msgr == nil clause.
	if st.Metrics != nil {
		st.Metrics.RecordAnnouncement()
	}

	st.mu.Lock()
	st.LastHeard = time.Now()
	st.mu.Unlock()

	// publish a simple hello payload if available
	msgr := messenger.GetMessenger()
	if msgr == nil {
		return
	}

	pbytes, err := json.Marshal(st)
	if err != nil {
		slog.Error("hello failed to map payload into json", "station", st.ID, "error", err)
		return
	}

	// Use explicit topic for hello messages
	topic := messenger.Topic("d/" + st.Hostname + "/hello")
	if err := msgr.Pub(topic, pbytes); err != nil {
		// record the error for metrics / diagnostics
		st.SaveError(fmt.Errorf("SayHello publish failed: %w", err))
	}
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
			ifc.IPAddrs = append(ifc.IPAddrs, netip.MustParseAddr(ip.String()))
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
func (s *Station) Update(msg *messenger.Msg) {
	slog.Info("updating station", "id", s.ID)
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
		// Encode a pointer to avoid passing the struct by value
		_ = json.NewEncoder(w).Encode(&m)
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

// Create an endpoint for this device to be queried.
func (s *Station) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}
