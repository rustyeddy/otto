package messenger

import (
	"context"
	"fmt"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessengerNone(t *testing.T) {
	origUser := os.Getenv("MQTT_USER")
	origPass := os.Getenv("MQTT_PASS")
	t.Cleanup(func() {
		_ = os.Setenv("MQTT_USER", origUser)
		_ = os.Setenv("MQTT_PASS", origPass)
	})

	_ = os.Setenv("MQTT_USER", "u_test")
	_ = os.Setenv("MQTT_PASS", "p_test")

	m := NewMessenger("none")
	require.NotNil(t, m)

	assert.Equal(t, "otto-client", m.ID)
	assert.Equal(t, "none", m.Broker)
	// subscriptions map should be initialized
	assert.NotNil(t, m.subscriptions)
	// Conn should be nobrokerConn
	assert.IsType(t, &nobrokerConn{}, m.Conn)
	// singleton should be set
	assert.Same(t, m, GetMessenger())

	// exercise Sub and Close (should not panic)
	assert.NoError(t, m.Sub("t/test", func(_ *Msg) error { return nil }))
	assert.Len(t, m.subscriptions, 1)
	m.Close()
}

func TestNewMessengerExternal(t *testing.T) {
	m := NewMessenger("example-broker")
	assert.NotNil(t, m)
	if m == nil {
		return
	}
	assert.Equal(t, "example-broker", m.Broker)
	assert.Equal(t, External, m.BrokerType)
	// Conn should be *connMQTT
	assert.IsType(t, &connMQTT{}, m.Conn)
	assert.NotNil(t, m.subscriptions)
}

func TestNewMessengerInternal(t *testing.T) {
	m := NewMessenger("internal")
	if m == nil {
		t.Skip("NewMessenger(\"internal\") returned nil - embedded broker likely unavailable, skipping")
	}
	// If non-nil, validate builtin state and attempt cleanup.
	assert.Equal(t, Builtin, m.BrokerType)
	assert.IsType(t, &connMQTT{}, m.Conn)

	// attempt to stop embedded broker to avoid leaking in test environments
	_ = StopMQTTBroker(context.Background())
}

func TestSetGetMessenger(t *testing.T) {
	orig := GetMessenger()
	defer SetMessenger(orig)

	m := &Messenger{ID: "test-id", Broker: "test-broker"}
	SetMessenger(m)
	assert.Same(t, m, GetMessenger())
}

func TestSetMessengerNil(t *testing.T) {
	orig := GetMessenger()
	defer SetMessenger(orig)

	SetMessenger(nil)
	assert.Nil(t, GetMessenger())
}

func TestMessengerConnectNone(t *testing.T) {
	orig := GetMessenger()
	defer SetMessenger(orig)

	m := NewMessenger("none")
	assert.NotNil(t, m)
	if m == nil {
		return
	}
	assert.NoError(t, m.Connect())
}

func TestMessengerCloseNoPanic(t *testing.T) {
	orig := GetMessenger()
	defer SetMessenger(orig)

	// ensure local routing root is initialized
	resetNodes()

	m := NewMessenger("none")
	assert.NotNil(t, m)
	if m == nil {
		return
	}
	assert.NoError(t, m.Sub("t/close", func(_ *Msg) error { return nil }))

	// Close should call Unsub on the connection and not panic
	m.Close()
}

type mockConn struct {
	subs   []string
	subErr map[string]error
}

func (mc *mockConn) Connect(b string, u string, p string) error { return nil }
func (mc *mockConn) Close()                                     {}
func (mc *mockConn) Sub(topic string, handler MsgHandler) error {
	mc.subs = append(mc.subs, topic)
	if mc.subErr != nil {
		if err, ok := mc.subErr[topic]; ok {
			return err
		}
	}
	return nil
}
func (mc *mockConn) Unsub(topics ...string) {}
func (mc *mockConn) PubMsg(m *Msg) error    { return nil }
func (mc *mockConn) IsConnected() bool      { return true }

func TestSubscribeAllNoSubscriptions(t *testing.T) {
	m := NewMessenger("none")
	assert.NotNil(t, m)
	if m == nil {
		return
	}
	m.subscriptions = map[string]MsgHandler{} // ensure empty
	mock := &mockConn{}
	m.Conn = mock

	m.SubscribeAll(nil)

	assert.Len(t, mock.subs, 0)
}

func TestSubscribeAllWithSubscriptionsAndErrors(t *testing.T) {
	m := NewMessenger("none")
	assert.NotNil(t, m)
	if m == nil {
		return
	}
	mock := &mockConn{subErr: map[string]error{"bad/topic": fmt.Errorf("fail")}}
	m.Conn = mock

	m.subscriptions = map[string]MsgHandler{
		"good/topic": func(*Msg) error { return nil },
		"bad/topic":  func(*Msg) error { return nil },
	}

	m.SubscribeAll(nil)

	assert.Len(t, mock.subs, 2)
	assert.ElementsMatch(t, []string{"good/topic", "bad/topic"}, mock.subs)
}
