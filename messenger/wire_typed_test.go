package messenger

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/rustyeddy/otto/internal/testutils"
	"github.com/rustyeddy/otto/messenger/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wireMQTT struct {
	mu        sync.Mutex
	publishes []publishCall
	publishCh chan publishCall
}

func newWireMQTT() *wireMQTT {
	return &wireMQTT{
		publishCh: make(chan publishCall, 8),
	}
}

func (m *wireMQTT) Publish(ctx context.Context, topic string, payload []byte, retain bool, qos byte) error {
	call := publishCall{topic: topic, body: payload, retain: retain, qos: qos}
	m.mu.Lock()
	m.publishes = append(m.publishes, call)
	m.mu.Unlock()
	m.publishCh <- call
	return nil
}

func (m *wireMQTT) Subscribe(ctx context.Context, topic string, qos byte, handler func(Message)) (func() error, error) {
	return func() error { return nil }, nil
}

func (m *wireMQTT) SetWill(topic string, payload []byte, retain bool, qos byte) error {
	return nil
}

func TestWireSourcePublishesAndCaches(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	mqtt := newWireMQTT()
	reg := NewRegistry(mqtt, TopicScheme{Prefix: "otto"})
	src := testutils.NewSource[int]("src", 8)
	WireSource(ctx, reg, src, codec.JSON[int]{})

	src.Set() <- 42

	select {
	case call := <-mqtt.publishCh:
		assert.Equal(t, "otto/devices/src/state", call.topic)
		var got int
		require.NoError(t, json.Unmarshal(call.body, &got))
		assert.Equal(t, 42, got)
	case <-ctx.Done():
		require.Fail(t, "publish not received")
	}

	raw, ok := reg.StateRaw("src")
	require.True(t, ok)
	assert.NotEmpty(t, raw)

	val, ok := reg.StateAny("src")
	require.True(t, ok)
	assert.Equal(t, 42, val)

	src.Close()
}

func TestWireSinkDeliversToDevice(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	mqtt := newWireMQTT()
	reg := NewRegistry(mqtt, TopicScheme{Prefix: "otto"})
	// sink := &fakeSink[int]{name: "relay", in: make(chan int, 1), events: make(chan devices.Event)}
	sink := testutils.NewSink[int]("relay", 8)

	WireSink(ctx, reg, sink, codec.JSON[int]{})

	sub, ok := reg.subs["otto/devices/relay/set"]
	require.True(t, ok)

	payload, err := json.Marshal(7)
	require.NoError(t, err)
	sub.handler(Message{Topic: "otto/devices/relay/set", Payload: payload})

	select {
	case got := <-sink.Get():
		assert.Equal(t, 7, got)
	case <-ctx.Done():
		require.Fail(t, "set not delivered")
	}
}

func TestWireSinkContextCancelReturns(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()

	mqtt := newWireMQTT()
	reg := NewRegistry(mqtt, TopicScheme{Prefix: "otto"})
	// sink := &fakeSink[int]{name: "relay", in: make(chan int), events: make(chan devices.Event)}
	sink := testutils.NewSink[int]("relay", 8)

	WireSink(ctx, reg, sink, codec.JSON[int]{})

	sub, ok := reg.subs["otto/devices/relay/set"]
	require.True(t, ok)

	payload, err := json.Marshal(9)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		sub.handler(Message{Topic: "otto/devices/relay/set", Payload: payload})
		close(done)
	}()

	waitCtx, waitCancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(waitCancel)

	select {
	case <-done:
	case <-waitCtx.Done():
		require.Fail(t, "handler did not return on canceled context")
	}
}
