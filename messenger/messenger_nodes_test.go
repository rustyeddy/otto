package messenger

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNode(t *testing.T) {
	n := newNode("test")
	require.Equal(t, "test", n.index)
	require.Len(t, n.nodes, 0)
	require.Len(t, n.handlers, 0)
}

func TestRemoveNoopLeavesHandler(t *testing.T) {
	resetNodes()

	handlerCalled := false
	handler := func(m *Msg) error {
		handlerCalled = true
		return nil
	}

	root.insert("test/remove", handler)
	// remove the handler and ensure the node is cleaned up
	root.remove("test/remove", handler)

	n := root.lookup("test/remove")
	require.Nil(t, n)

	// publishing should not find a node to deliver to
	// (handler should not be called)
	if root.lookup("test/remove") != nil {
		root.lookup("test/remove").pub(&Msg{Topic: "test/remove"})
	}
	require.False(t, handlerCalled)
}

func TestRemoveNoPanicWhenMissing(t *testing.T) {
	resetNodes()

	// Removing a non-existent topic/handler should not panic or break state.
	root.remove("non/existent/topic", nil)

	require.NotNil(t, root)
}

func TestInitNodes(t *testing.T) {
	clearNodes()
	require.Nil(t, root)

	initNodes()
	require.NotNil(t, root)
	require.Len(t, root.nodes, 0)

	// Verify the new root is usable
	called := false
	h := func(m *Msg) error {
		called = true
		return nil
	}
	root.insert("a/b", h)

	n := root.lookup("a/b")
	require.NotNil(t, n)
	n.pub(&Msg{Topic: "a/b"})
	require.True(t, called)
}

func TestInsertAndLookup(t *testing.T) {
	root := newNode("/")
	handlerCalled := false
	handler := func(msg *Msg) error {
		handlerCalled = true
		return nil
	}

	// Insert a topic with a handler
	root.insert("test/topic", handler)

	// Lookup the topic
	node := root.lookup("test/topic")
	require.NotNil(t, node)

	// Verify the handler is present
	require.Len(t, node.handlers, 1)

	// Trigger the handler
	msg := &Msg{Topic: "test/topic"}
	node.pub(msg)

	require.True(t, handlerCalled)
}

func TestWildcardLookup(t *testing.T) {
	root := newNode("/")
	handlerCalled := false
	handler := func(msg *Msg) error {
		handlerCalled = true
		return nil
	}

	// Insert a wildcard topic
	root.insert("test/+/wildcard", handler)

	// Lookup a matching topic
	node := root.lookup("test/something/wildcard")
	require.NotNil(t, node)
	require.Len(t, node.handlers, 1)

	// Trigger the handler
	msg := &Msg{Topic: "test/something/wildcard"}
	node.pub(msg)
	require.True(t, handlerCalled)
}

func TestMultiLevelWildcardLookup(t *testing.T) {
	root := newNode("/")
	handlerCalled := false
	handler := func(msg *Msg) error {
		handlerCalled = true
		return nil
	}

	// Insert a multi-level wildcard topic
	root.insert("test/#", handler)

	// Lookup a matching topic
	node := root.lookup("test/any/number/of/levels")
	require.NotNil(t, node)
	require.Len(t, node.handlers, 1)

	// Trigger the handler
	msg := &Msg{Topic: "test/any/number/of/levels"}
	node.pub(msg)
	require.True(t, handlerCalled)
}

func TestClearNodes(t *testing.T) {
	root = newNode("/")
	require.NotNil(t, root)

	clearNodes()
	require.Nil(t, root)
}

func TestResetNodes(t *testing.T) {
	// prepare a populated root
	root = newNode("/")
	root.insert("foo/bar", func(m *Msg) error { return nil })

	require.NotNil(t, root.lookup("foo/bar"))

	// reset and validate
	resetNodes()
	require.NotNil(t, root)
	require.Len(t, root.nodes, 0)
	// previous entries should be gone
	require.Nil(t, root.lookup("foo/bar"))

	// ensure new root is usable
	called := false
	h := func(m *Msg) error {
		called = true
		return nil
	}
	root.insert("new/topic", h)
	n := root.lookup("new/topic")
	require.NotNil(t, n)
	n.pub(&Msg{Topic: "new/topic"})
	require.True(t, called)
}

func TestServeHTTPGetRootReturnsSuccessfulResponse(t *testing.T) {
	m := NewMessenger("none")
	if m == nil {
		t.Fatal("NewMessenger returned nil")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Ensure ServeHTTP does not panic and returns a successful status code.
	m.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.True(t, res.StatusCode >= 200 && res.StatusCode < 400, "expected 2xx/3xx status, got %d, body: %s", res.StatusCode, string(body))
	assert.NotNil(t, body, "expected non-nil response body")
}

func TestServeHTTPUnknownPathStillHandled(t *testing.T) {
	m := NewMessenger("none")
	if m == nil {
		t.Fatal("NewMessenger returned nil")
	}

	req := httptest.NewRequest(http.MethodGet, "/nonexistent/path", nil)
	w := httptest.NewRecorder()

	m.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	// Handler should respond without panic; accept any non-5xx status.
	assert.False(t, res.StatusCode >= 500, "unexpected server error status %d, body: %s", res.StatusCode, string(body))
	assert.NotNil(t, body)
}

func TestConnMQTTConnectEmptyBroker(t *testing.T) {
	c := &connMQTT{}
	err := c.Connect("", "", "")
	assert.Error(t, err, "expected error when connecting with empty broker address")
}

func TestConnMQTTConnectUnreachableHost(t *testing.T) {
	c := &connMQTT{}
	broker := "127.0.0.1:65535"
	err := c.Connect(broker, "user", "pass")
	if err == nil {
		t.Skipf("unexpectedly connected to broker at %s in test environment; skipping", broker)
	}
	assert.Error(t, err, "expected error when connecting to unreachable broker")
}
