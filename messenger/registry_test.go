package messenger

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type registryMQTT struct {
	mu        sync.Mutex
	publishes []publishCall
	wills     []publishCall

	subscribeCalls map[string]int
	unsubCalls     map[string]int
}

type publishCall struct {
	topic  string
	body   []byte
	retain bool
	qos    byte
}

func newRegistryMQTT() *registryMQTT {
	return &registryMQTT{
		subscribeCalls: make(map[string]int),
		unsubCalls:     make(map[string]int),
	}
}

func (m *registryMQTT) Publish(ctx context.Context, topic string, payload []byte, retain bool, qos byte) error {
	m.mu.Lock()
	m.publishes = append(m.publishes, publishCall{topic: topic, body: payload, retain: retain, qos: qos})
	m.mu.Unlock()
	return nil
}

func (m *registryMQTT) Subscribe(ctx context.Context, topic string, qos byte, handler func(Message)) (func() error, error) {
	m.mu.Lock()
	m.subscribeCalls[topic]++
	m.mu.Unlock()

	return func() error {
		m.mu.Lock()
		m.unsubCalls[topic]++
		m.mu.Unlock()
		return nil
	}, nil
}

func (m *registryMQTT) SetWill(topic string, payload []byte, retain bool, qos byte) error {
	m.mu.Lock()
	m.wills = append(m.wills, publishCall{topic: topic, body: payload, retain: retain, qos: qos})
	m.mu.Unlock()
	return nil
}

func (m *registryMQTT) snapshot() (publishes []publishCall, wills []publishCall, subs map[string]int, unsubs map[string]int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	publishes = append([]publishCall(nil), m.publishes...)
	wills = append([]publishCall(nil), m.wills...)

	subs = make(map[string]int, len(m.subscribeCalls))
	for k, v := range m.subscribeCalls {
		subs[k] = v
	}
	unsubs = make(map[string]int, len(m.unsubCalls))
	for k, v := range m.unsubCalls {
		unsubs[k] = v
	}
	return publishes, wills, subs, unsubs
}

type fakeDevice struct {
	name       string
	desc       devices.Descriptor
	events     chan devices.Event
	run        func(ctx context.Context) error
	descriptor bool
}

func (d *fakeDevice) Name() string { return d.name }
func (d *fakeDevice) Run(ctx context.Context) error {
	if d.run != nil {
		return d.run(ctx)
	}
	<-ctx.Done()
	return nil
}
func (d *fakeDevice) Events() <-chan devices.Event { return d.events }
func (d *fakeDevice) Descriptor() devices.Descriptor {
	if d.descriptor {
		return d.desc
	}
	return devices.Descriptor{}
}

func TestRegistryResubscribeAll(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	mqtt := newRegistryMQTT()
	reg := NewRegistry(mqtt, TopicScheme{Prefix: "otto"})
	reg.WantSub("otto/devices/lamp/set", 1, func(Message) {})
	reg.WantSub("otto/devices/lamp/state", 0, func(Message) {})

	reg.ResubscribeAll(ctx)
	reg.ResubscribeAll(ctx)

	_, _, subs, unsubs := mqtt.snapshot()
	assert.Equal(t, 2, subs["otto/devices/lamp/set"])
	assert.Equal(t, 2, subs["otto/devices/lamp/state"])
	assert.Equal(t, 1, unsubs["otto/devices/lamp/set"])
	assert.Equal(t, 1, unsubs["otto/devices/lamp/state"])
}

func TestRegistryRunReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	mqtt := newRegistryMQTT()
	reg := NewRegistry(mqtt, TopicScheme{Prefix: "otto"})

	events := make(chan devices.Event)
	close(events)

	dev := &fakeDevice{
		name:       "lamp",
		desc:       devices.Descriptor{Name: "lamp", Kind: "switch", ValueType: "bool", Access: devices.ReadWrite},
		events:     events,
		descriptor: true,
		run: func(context.Context) error {
			return errors.New("boom")
		},
	}
	reg.Add(dev)

	err := reg.Run(ctx)
	require.Error(t, err)

	publishes, wills, _, _ := mqtt.snapshot()
	require.Len(t, wills, 1)
	assert.Equal(t, "otto/devices/lamp/status", wills[0].topic)

	var status StatusPayload
	var meta MetaPayload
	var gotStatus bool
	var gotMeta bool

	for _, call := range publishes {
		switch call.topic {
		case "otto/devices/lamp/status":
			require.NoError(t, json.Unmarshal(call.body, &status))
			assert.Equal(t, "online", status.Status)
			gotStatus = true
		case "otto/devices/lamp/meta":
			require.NoError(t, json.Unmarshal(call.body, &meta))
			assert.Equal(t, "lamp", meta.Name)
			gotMeta = true
		}
	}

	assert.True(t, gotStatus)
	assert.True(t, gotMeta)
}

func TestRegistryRunGracefulShutdownPublishesOffline(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	waitCtx, waitCancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(waitCancel)

	mqtt := newRegistryMQTT()
	reg := NewRegistry(mqtt, TopicScheme{Prefix: "otto"})

	started := make(chan struct{})
	events := make(chan devices.Event)

	dev := &fakeDevice{
		name:   "sensor",
		events: events,
		run: func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			return nil
		},
	}
	reg.Add(dev)

	done := make(chan error, 1)
	go func() { done <- reg.Run(ctx) }()

	select {
	case <-started:
	case <-waitCtx.Done():
		require.Fail(t, "device did not start")
	}

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-waitCtx.Done():
		require.Fail(t, "registry did not exit after cancel")
	}

	publishes, _, _, _ := mqtt.snapshot()
	var gotOnline bool
	var gotOffline bool
	for _, call := range publishes {
		if call.topic != "otto/devices/sensor/status" {
			continue
		}
		var status StatusPayload
		require.NoError(t, json.Unmarshal(call.body, &status))
		if status.Status == "online" {
			gotOnline = true
		}
		if status.Status == "offline" {
			gotOffline = true
		}
	}

	assert.True(t, gotOnline)
	assert.True(t, gotOffline)
}
