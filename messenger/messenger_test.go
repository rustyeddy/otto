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

// mockUnsubConn is a top-level mock used by tests that need to observe Unsub calls.
type mockUnsubConn struct {
	unsubbed []string
}

func (mc *mockUnsubConn) Connect(b, u, p string) error { return nil }
func (mc *mockUnsubConn) Close()                      {}
func (mc *mockUnsubConn) Sub(topic string, handler MsgHandler) error {
	return nil
}
func (mc *mockUnsubConn) Unsub(topics ...string) {
	mc.unsubbed = append(mc.unsubbed, topics...)
}
func (mc *mockUnsubConn) PubMsg(m *Msg) error { return nil }
func (mc *mockUnsubConn) IsConnected() bool   { return true }

// mockPubConn is a top-level mock used by tests that need to observe PubMsg calls.
type mockPubConn struct {
	last *Msg
}

func (mc *mockPubConn) Connect(b, u, p string) error { return nil }
func (mc *mockPubConn) Close()                      {}
func (mc *mockPubConn) Sub(topic string, handler MsgHandler) error {
	return nil
}
func (mc *mockPubConn) Unsub(topics ...string) {}
func (mc *mockPubConn) PubMsg(msg *Msg) error {
	mc.last = msg
	return nil
}
func (mc *mockPubConn) IsConnected() bool { return true }

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

func TestMessengerUnsubCallsConnUnsub(t *testing.T) {
	orig := GetMessenger()
	defer SetMessenger(orig)

	m := NewMessenger("none")

	if m == nil {
		t.Fatal("NewMessenger returned nil")
	}

	mock := &mockUnsubConn{}
	m.Conn = mock

	m.Unsub("a/b")
	assert.ElementsMatch(t, []string{"a/b"}, mock.unsubbed)
}

func TestNobrokerUnsubRemovesFromRoot(t *testing.T) {
    resetNodes()

    called := false
    root.insert("x/y", func(m *Msg) error {
        called = true
        return nil
    })
    if root.lookup("x/y") == nil {
        t.Fatal("expected node before unsub")
    }

    c := &nobrokerConn{}
    c.Unsub("x/y")
    if root.lookup("x/y") != nil {
        t.Fatal("expected node to be removed after Unsub")
    }
    // ensure handler was not invoked by Unsub itself
    if called {
        t.Fatal("handler should not be called by Unsub")
    }
}

func TestMessengerPubUsesConnPubMsg(t *testing.T) {
	orig := GetMessenger()
	defer SetMessenger(orig)

	m := NewMessenger("none")
	if m == nil {
		t.Fatal("NewMessenger returned nil")
	}

	mock := &mockPubConn{}
	m.Conn = mock

	payload := []byte("payload")
	if err := m.Pub("t/p", payload); err != nil {
		t.Fatalf("unexpected error from Pub: %v", err)
	}
	if mock.last == nil {
		t.Fatal("expected PubMsg to be called on conn")
	}
	assert.Equal(t, "t/p", mock.last.Topic)
	b, err := Bytes(mock.last.Data)
	assert.NoError(t, err)
	assert.Equal(t, payload, b)
}

func TestNobrokerPubMsgErrorsAndDelivery(t *testing.T) {
    resetNodes()
    c := &nobrokerConn{}

    // nil message
    if err := c.PubMsg(nil); assert.Error(t, err) {
        assert.Contains(t, err.Error(), "nil")
    }

    // no subscribers
    err := c.PubMsg(&Msg{Topic: "no/one", Data: []byte("d")})
    if assert.Error(t, err) {
        assert.Contains(t, err.Error(), "No subscribers")
    }

    // delivery to subscriber
    delivered := false
    root.insert("deliver/one", func(m *Msg) error {
        delivered = true
        return nil
    })
    if err := c.PubMsg(&Msg{Topic: "deliver/one", Data: []byte("ok")}); err != nil {
        t.Fatalf("unexpected error publishing to existing subscriber: %v", err)
    }
    if !delivered {
        t.Fatal("expected handler to be invoked for delivered message")
    }
}