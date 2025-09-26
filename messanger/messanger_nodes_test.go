package messanger

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
