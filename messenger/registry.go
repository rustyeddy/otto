package messenger

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/rustyeddy/devices"
)

// Logger is the minimal logging interface used by Registry.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Registry wires devices to MQTT topics and keeps a small state cache.
type Registry struct {
	MQTT   MQTT
	Topics TopicScheme
	Log    Logger

	// Defaults (override after NewRegistry if desired)
	QoSState    byte
	QoSSet      byte
	QoSEvent    byte
	QoSStatus   byte
	RetainState bool
	RetainMeta  bool

	// Command delivery guard (prevents wedging MQTT callback path)
	CommandTimeout time.Duration

	// Internal
	mu sync.RWMutex

	devs []devices.Device

	// Desired subscriptions (topic -> spec)
	subs map[string]subSpec

	// Active unsubscribers (topic -> unsub)
	unsubs map[string]func() error

	// ---- State cache ----
	stateMu sync.RWMutex

	// raw retained payloads (JSON bytes as published)
	stateRaw map[string][]byte

	// decoded state cache (optional, populated by WireSource)
	stateAny map[string]any
}

// NewRegistry builds a Registry with defaults set for QoS and retention.
func NewRegistry(m MQTT, topics TopicScheme) *Registry {
	return &Registry{
		MQTT:           m,
		Topics:         topics,
		Log:            slog.Default(),
		QoSState:       0,
		QoSSet:         1,
		QoSEvent:       0,
		QoSStatus:      1,
		RetainState:    true,
		RetainMeta:     true,
		CommandTimeout: 2 * time.Second,

		subs:     map[string]subSpec{},
		unsubs:   map[string]func() error{},
		stateRaw: make(map[string][]byte),
		stateAny: make(map[string]any),
	}
}

// Add appends a device to the registry.
func (r *Registry) Add(dev devices.Device) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devs = append(r.devs, dev)
}

// WantSub registers a subscription that should be active whenever MQTT is connected.
// Registry will apply these on every connect/reconnect.
func (r *Registry) WantSub(topic string, qos byte, handler func(Message)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subs[topic] = subSpec{topic: topic, qos: qos, handler: handler}
}

// ResubscribeAll applies all desired subscriptions (call on connect and reconnect).
func (r *Registry) ResubscribeAll(ctx context.Context) {
	r.mu.RLock()
	subs := make([]subSpec, 0, len(r.subs))
	for _, s := range r.subs {
		subs = append(subs, s)
	}
	r.mu.RUnlock()

	r.Log.Info("MQTT connected; (re)subscribing", "count", len(subs))

	for _, s := range subs {
		// Unsubscribe previous if any (best effort)
		r.mu.Lock()
		if u, ok := r.unsubs[s.topic]; ok && u != nil {
			_ = u()
			delete(r.unsubs, s.topic)
		}
		r.mu.Unlock()

		unsub, err := r.MQTT.Subscribe(ctx, s.topic, s.qos, s.handler)
		if err != nil {
			r.Log.Error("MQTT subscribe failed", "topic", s.topic, "error", err)
			continue
		}
		r.mu.Lock()
		r.unsubs[s.topic] = unsub
		r.mu.Unlock()

		r.Log.Info("MQTT subscribed", "topic", s.topic, "qos", s.qos)
	}
}

func (r *Registry) publishStatus(ctx context.Context, name, status string) {
	b, _ := json.Marshal(StatusPayload{Status: status, Time: time.Now()})
	_ = r.MQTT.Publish(ctx, r.Topics.Status(name), b, true, r.QoSStatus)
}

func (r *Registry) publishMeta(ctx context.Context, dev devices.Device) {
	d, ok := dev.(interface{ Descriptor() devices.Descriptor })
	if !ok {
		return
	}
	desc := d.Descriptor()
	mp := MetaPayload{
		Name:      desc.Name,
		Kind:      desc.Kind,
		ValueType: desc.ValueType,
		Access:    string(desc.Access),
		Unit:      desc.Unit,
		Min:       desc.Min,
		Max:       desc.Max,
		Tags:      desc.Tags,
		Attrs:     desc.Attributes,
	}
	b, err := json.Marshal(mp)
	if err != nil {
		r.Log.Warn("meta marshal failed", "device", dev.Name(), "error", err)
		return
	}
	_ = r.MQTT.Publish(ctx, r.Topics.Meta(dev.Name()), b, r.RetainMeta, r.QoSStatus)
}

func (r *Registry) wireEvents(ctx context.Context, dev devices.Device) {
	name := dev.Name()

	go func() {
		for {
			select {
			case evt, ok := <-dev.Events():
				if !ok {
					return
				}

				// JSON-friendly event payload
				wire := map[string]any{
					"device": evt.Device,
					"kind":   evt.Kind,
					"time":   evt.Time,
					"msg":    evt.Msg,
					"meta":   evt.Meta,
				}
				if evt.Err != nil {
					wire["err"] = evt.Err.Error()
				}

				b, err := json.Marshal(wire)
				if err != nil {
					continue
				}
				_ = r.MQTT.Publish(ctx, r.Topics.Event(name), b, false, r.QoSEvent)

			case <-ctx.Done():
				return
			}
		}
	}()
}

// Run starts device goroutines, wires events, and publishes status/meta.
// For reconnect-resubscribe to work, your MQTT adapter must call ResubscribeAll on connect.
func (r *Registry) Run(ctx context.Context) error {
	// Snapshot devices
	r.mu.RLock()
	devs := append([]devices.Device(nil), r.devs...)
	r.mu.RUnlock()

	// Prepare LWT + birth/meta and wire events
	for _, dev := range devs {
		name := dev.Name()

		// Set LWT (offline retained)
		offline, _ := json.Marshal(StatusPayload{Status: "offline", Time: time.Now()})
		_ = r.MQTT.SetWill(r.Topics.Status(name), offline, true, r.QoSStatus)

		// Birth online
		r.publishStatus(ctx, name, "online")

		// Meta retained (optional)
		r.publishMeta(ctx, dev)

		// Events -> MQTT
		r.wireEvents(ctx, dev)
	}

	// Start devices
	var wg sync.WaitGroup
	errCh := make(chan error, len(devs))

	for _, dev := range devs {
		wg.Add(1)
		go func(d devices.Device) {
			defer wg.Done()
			if err := d.Run(ctx); err != nil {
				errCh <- err
			}
		}(dev)
	}

	select {
	case err := <-errCh:
		// first fatal error
		return err
	case <-ctx.Done():
		// graceful shutdown
	}

	// Best effort: unsubscribe
	r.mu.Lock()
	for topic, u := range r.unsubs {
		_ = u()
		delete(r.unsubs, topic)
	}
	r.mu.Unlock()

	wg.Wait()

	// Publish offline retained
	for _, dev := range devs {
		r.publishStatus(context.Background(), dev.Name(), "offline")
	}

	return nil
}

// StateRaw returns the last published state payload for a device.
func (r *Registry) StateRaw(name string) ([]byte, bool) {
	r.stateMu.RLock()
	defer r.stateMu.RUnlock()
	b, ok := r.stateRaw[name]
	return b, ok
}

// StateAny returns the last decoded state value, if known.
func (r *Registry) StateAny(name string) (any, bool) {
	r.stateMu.RLock()
	defer r.stateMu.RUnlock()
	v, ok := r.stateAny[name]
	return v, ok
}

// StateAs returns the last decoded state as a concrete type.
func StateAs[T any](r *Registry, name string) (T, bool) {
	var zero T
	v, ok := r.StateAny(name)
	if !ok {
		return zero, false
	}
	tv, ok := v.(T)
	if !ok {
		return zero, false
	}
	return tv, true
}
