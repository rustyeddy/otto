package messenger

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMQTT struct {
	mu             sync.Mutex
	subs           []subCall
	subscribeCalls map[string]int
	unsubCalls     map[string]int
}

type subCall struct {
	topic   string
	qos     byte
	handler func(Message)
}

func newFakeMQTT() *fakeMQTT {
	return &fakeMQTT{
		subscribeCalls: make(map[string]int),
		unsubCalls:     make(map[string]int),
	}
}

func (f *fakeMQTT) Publish(ctx context.Context, topic string, payload []byte, retain bool, qos byte) error {
	return nil
}

func (f *fakeMQTT) Subscribe(ctx context.Context, topic string, qos byte, handler func(Message)) (func() error, error) {
	f.mu.Lock()
	f.subs = append(f.subs, subCall{topic: topic, qos: qos, handler: handler})
	f.subscribeCalls[topic]++
	f.mu.Unlock()

	return func() error {
		f.mu.Lock()
		f.unsubCalls[topic]++
		f.mu.Unlock()
		return nil
	}, nil
}

func (f *fakeMQTT) SetWill(topic string, payload []byte, retain bool, qos byte) error {
	return nil
}

func (f *fakeMQTT) snapshot() (subs []subCall, subscribeCalls map[string]int, unsubCalls map[string]int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	subs = append([]subCall(nil), f.subs...)
	subscribeCalls = make(map[string]int, len(f.subscribeCalls))
	for k, v := range f.subscribeCalls {
		subscribeCalls[k] = v
	}
	unsubCalls = make(map[string]int, len(f.unsubCalls))
	for k, v := range f.unsubCalls {
		unsubCalls[k] = v
	}
	return subs, subscribeCalls, unsubCalls
}

func TestMessengerResubscribeAllSubscribes(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	mqtt := newFakeMQTT()
	m := New(mqtt)
	m.WantSub("otto/devices/lamp/set", 1, func(Message) {})
	m.WantSub("otto/devices/lamp/state", 0, func(Message) {})

	m.ResubscribeAll(ctx)

	subs, calls, _ := mqtt.snapshot()
	require.Len(t, subs, 2)
	assert.Equal(t, 1, calls["otto/devices/lamp/set"])
	assert.Equal(t, 1, calls["otto/devices/lamp/state"])
}

func TestMessengerResubscribeAllUnsubscribesPrevious(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	mqtt := newFakeMQTT()
	m := New(mqtt)
	m.WantSub("otto/devices/lamp/set", 1, func(Message) {})
	m.WantSub("otto/devices/lamp/state", 0, func(Message) {})

	m.ResubscribeAll(ctx)
	m.ResubscribeAll(ctx)

	_, calls, unsubs := mqtt.snapshot()
	assert.Equal(t, 2, calls["otto/devices/lamp/set"])
	assert.Equal(t, 2, calls["otto/devices/lamp/state"])
	assert.Equal(t, 1, unsubs["otto/devices/lamp/set"])
	assert.Equal(t, 1, unsubs["otto/devices/lamp/state"])
}
