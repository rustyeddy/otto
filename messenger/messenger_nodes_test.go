package messenger

import (
	"testing"
)

func TestNewNode(t *testing.T) {
	n := newNode("test")
	if n.index != "test" {
		t.Errorf("Expected index 'test', got '%s'", n.index)
	}
	if len(n.nodes) != 0 {
		t.Errorf("Expected nodes map to be empty, got %d elements", len(n.nodes))
	}
	if len(n.handlers) != 0 {
		t.Errorf("Expected handlers to be empty, got %d elements", len(n.handlers))
	}
}

func TestRemoveNoopLeavesHandler(t *testing.T) {
	resetNodes()

	handlerCalled := false
	handler := func(m *Msg) error {
		handlerCalled = true
		return nil
	}

	root.insert("test/remove", handler)
	// current remove is a noop; ensure it doesn't remove the handler
	root.remove("test/remove", handler)

	n := root.lookup("test/remove")
	if n == nil {
		t.Fatal("expected node for 'test/remove', got nil")
	}
	if len(n.handlers) != 1 {
		t.Fatalf("expected 1 handler after remove noop, got %d", len(n.handlers))
	}

	n.pub(&Msg{Topic: "test/remove"})
	if !handlerCalled {
		t.Error("expected handler to be called after remove noop, but it was not")
	}
}

func TestRemoveNoPanicWhenMissing(t *testing.T) {
	resetNodes()

	// Removing a non-existent topic/handler should not panic or break state.
	root.remove("non/existent/topic", nil)

	if root == nil {
		t.Fatal("expected root to remain initialized after remove on missing topic")
	}
}

func TestInitNodes(t *testing.T) {
	clearNodes()
	if root != nil {
		t.Fatal("expected root to be nil after clearNodes")
	}

	initNodes()
	if root == nil {
		t.Fatal("expected root to be initialized after initNodes")
	}
	if len(root.nodes) != 0 {
		t.Fatalf("expected root.nodes to be empty after initNodes, got %d", len(root.nodes))
	}

	// Verify the new root is usable
	called := false
	h := func(m *Msg) error {
		called = true
		return nil
	}
	root.insert("a/b", h)

	n := root.lookup("a/b")
	if n == nil {
		t.Fatal("expected node for 'a/b' after insert, got nil")
	}
	n.pub(&Msg{Topic: "a/b"})
	if !called {
		t.Error("expected handler to be called after insert on initNodes root")
	}
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
	if node == nil {
		t.Fatal("Expected to find node for 'test/topic', got nil")
	}

	// Verify the handler is present
	if len(node.handlers) != 1 {
		t.Fatalf("Expected 1 handler, got %d", len(node.handlers))
	}

	// Trigger the handler
	msg := &Msg{Topic: "test/topic"}
	node.pub(msg)

	if !handlerCalled {
		t.Error("Expected handler to be called, but it was not")
	}
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
	if node == nil {
		t.Fatal("Expected to find node for 'test/something/wildcard', got nil")
	}

	// Verify the handler is present
	if len(node.handlers) != 1 {
		t.Fatalf("Expected 1 handler, got %d", len(node.handlers))
	}

	// Trigger the handler
	msg := &Msg{Topic: "test/something/wildcard"}
	node.pub(msg)

	if !handlerCalled {
		t.Error("Expected handler to be called, but it was not")
	}
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
	if node == nil {
		t.Fatal("Expected to find node for 'test/any/number/of/levels', got nil")
	}

	// Verify the handler is present
	if len(node.handlers) != 1 {
		t.Fatalf("Expected 1 handler, got %d", len(node.handlers))
	}

	// Trigger the handler
	msg := &Msg{Topic: "test/any/number/of/levels"}
	node.pub(msg)

	if !handlerCalled {
		t.Error("Expected handler to be called, but it was not")
	}
}

func TestClearNodes(t *testing.T) {
	root = newNode("/")
	if root == nil {
		t.Fatal("Expected root to be initialized, got nil")
	}

	clearNodes()
	if root != nil {
		t.Fatal("Expected root to be cleared, but it was not")
	}
}
