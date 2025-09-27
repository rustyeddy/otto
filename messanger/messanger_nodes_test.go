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

func TestNodeRemove(t *testing.T) {
	root := newNode("/")
	handler := func(msg *Msg) error {
		return nil
	}

	// Insert some handlers
	root.insert("test/topic", handler)
	root.insert("test/topic2", handler)
	root.insert("test/deep/nested/topic", handler)

	// Verify nodes were created
	if len(root.nodes) != 1 {
		t.Fatalf("Expected 1 top-level node, got %d", len(root.nodes))
	}
	testNode := root.nodes["test"]
	if testNode == nil {
		t.Fatal("Expected 'test' node to exist")
	}

	// Verify nested structure
	if len(testNode.nodes) != 3 { // topic, topic2, deep
		t.Fatalf("Expected 3 child nodes under 'test', got %d", len(testNode.nodes))
	}

	// Remove one topic and verify cleanup
	root.remove("test/deep/nested/topic", nil)

	// The deep nested path should be cleaned up
	deepNode := testNode.nodes["deep"]
	if deepNode != nil {
		t.Error("Expected 'deep' node to be removed after cleanup")
	}

	// But other topics should remain
	if testNode.nodes["topic"] == nil {
		t.Error("Expected 'topic' node to remain")
	}
	if testNode.nodes["topic2"] == nil {
		t.Error("Expected 'topic2' node to remain")
	}

	// Remove all remaining topics
	root.remove("test/topic", nil)
	root.remove("test/topic2", nil)

	// Now the entire test branch should be cleaned up
	if len(root.nodes) != 0 {
		t.Errorf("Expected root to have no child nodes after all topics removed, got %d", len(root.nodes))
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
