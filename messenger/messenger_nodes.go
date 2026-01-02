package messenger

import (
	"reflect"
	"strings"
)

// node represents a single node in the topic routing tree used by the local messanger.
// The tree structure enables efficient wildcard matching for MQTT-style topic patterns.
// Each node represents one segment of a topic path (e.g., "ss", "c", "station", "temp").
//
// Wildcard Support:
//   - '+' matches exactly one topic level (e.g., "ss/c/+/temp" matches "ss/c/station1/temp")
//   - '#' matches zero or more levels (e.g., "ss/c/#" matches all control topics)
//
// The tree is organized hierarchically where each level of the topic creates a new
// level in the tree. Handlers are stored at leaf nodes and are invoked when messages
// match the complete topic path from root to that node.
type node struct {
	index    string           // The topic segment this node represents
	nodes    map[string]*node // Child nodes indexed by their topic segment
	handlers []MsgHandler     // Message handlers registered at this node
}

var (
	// root is the global root node of the topic routing tree.
	// All topic subscriptions are inserted as paths from this root.
	root *node
)

func init() {
	root = &node{
		nodes: make(map[string]*node),
	}
}

// newNode creates a new routing tree node for the given topic segment.
// Each node can have multiple child nodes and multiple message handlers.
//
// Parameters:
//   - index: The topic segment this node represents (e.g., "station", "+", "#")
//
// Returns a pointer to the newly created node.
func newNode(index string) *node {
	n := &node{
		index: index,
		nodes: make(map[string]*node),
	}
	return n
}

// initNodes initializes a new empty routing tree with just a root node.
// This is used during package initialization and testing.
func initNodes() {
	root = &node{
		nodes: make(map[string]*node),
	}
}

// clearNodes removes the routing tree by setting the root to nil.
// This is primarily used in testing to clean up state between tests.
func clearNodes() {
	root = nil
}

// resetNodes clears and re-initializes the routing tree.
// This is useful in testing to ensure a clean state.
func resetNodes() {
	clearNodes()
	initNodes()
}

// insert adds a message handler to the routing tree at the path specified by the topic.
// The topic is split into segments by '/' and each segment creates or traverses a node.
// The handler is added to the final node in the path.
//
// This enables wildcard topic subscriptions:
//   - Subscribing to "ss/c/+/temp" will match "ss/c/station1/temp", "ss/c/station2/temp", etc.
//   - Subscribing to "ss/c/#" will match all topics starting with "ss/c/"
//
// Parameters:
//   - topic: The topic pattern to subscribe to (e.g., "ss/c/station/temp" or "ss/c/+/temp")
//   - mh: The message handler to invoke when messages match this topic pattern
//
// Multiple handlers can be registered for the same topic pattern; all will be invoked.
func (n *node) insert(topic string, mh MsgHandler) {
	indexes := strings.Split(topic, "/")
	pn := n
	for _, idx := range indexes {
		if nn, ex := pn.nodes[idx]; !ex {
			nn = newNode(idx)
			pn.nodes[idx] = nn
			pn = nn
		} else {
			pn = nn
		}

	}
	// The last node push the callback on the callback list
	pn.handlers = append(pn.handlers, mh)
}

// remove attempts to remove a specific handler from a topic in the routing tree.
// This function is currently unimplemented (TODO) and returns immediately.
//
// The intended behavior is to:
//  1. Navigate to the node matching the topic
//  2. Remove the specified handler from that node's handler list
//  3. Clean up empty parent nodes if they have no other handlers or children
//
// Parameters:
//   - topic: The topic pattern from which to remove the handler
//   - handler: The specific handler function to remove
func (n *node) remove(topic string, handler MsgHandler) {
	// traverse the tree and collect nodes along the path
	indexes := strings.Split(topic, "/")
	var nodes []*node
	pn := n
	nodes = append(nodes, pn)
	for _, idx := range indexes {
		nn, ex := pn.nodes[idx]
		if !ex {
			// topic path doesn't exist; nothing to remove
			return
		}
		nodes = append(nodes, nn)
		pn = nn
	}

	// target node is the last in nodes
	target := nodes[len(nodes)-1]

	if handler == nil {
		// remove all handlers for this topic
		target.handlers = nil
	} else {
		// remove only matching handler(s)
		var keep []MsgHandler
		targetPtr := reflect.ValueOf(handler).Pointer()
		for _, h := range target.handlers {
			if reflect.ValueOf(h).Pointer() != targetPtr {
				keep = append(keep, h)
			}
		}
		target.handlers = keep
	}

	// cleanup: remove empty nodes from the bottom up
	// skip root (nodes[0])
	for i := len(nodes) - 1; i > 0; i-- {
		cur := nodes[i]
		parent := nodes[i-1]
		if len(cur.handlers) == 0 && len(cur.nodes) == 0 {
			// delete from parent
			delete(parent.nodes, cur.index)
			continue
		}
		break
	}
}

// lookup finds the node in the routing tree that matches the given topic,
// taking into account MQTT wildcard matching rules.
//
// Matching Rules:
//   - Exact segment matches are preferred (e.g., "station1" matches "station1")
//   - '#' wildcard matches all remaining segments (e.g., "ss/c/#" matches "ss/c/station/temp")
//   - '+' wildcard matches exactly one segment (e.g., "ss/c/+/temp" matches "ss/c/station1/temp")
//
// The lookup traverses the tree segment by segment, checking for exact matches first,
// then '#' wildcard (which terminates the search), then '+' wildcard (which continues).
//
// Parameters:
//   - topic: The complete topic path to lookup (e.g., "ss/c/station1/temp")
//
// Returns the node containing handlers for this topic, or nil if no matching node exists.
func (n *node) lookup(topic string) *node {
	indexes := strings.Split(topic, "/")
	pn := n
	for _, idx := range indexes {

		nn, ex := pn.nodes[idx]
		if ex {
			pn = nn
			continue
		}

		nn, ex = pn.nodes["#"]
		if ex {
			return nn
		}

		nn, ex = pn.nodes["+"]
		if ex {
			// we will accept this path portion of the wildcard, but
			// must continue on
			pn = nn
			continue
		}
		return nil
	}
	return pn
}

// pub publishes a message to all handlers registered at this node.
// Each handler is invoked in sequence with the provided message.
//
// Note: Handlers are called synchronously in the current goroutine.
// If a handler needs to perform long-running operations, it should
// spawn a goroutine internally to avoid blocking other handlers.
//
// Parameters:
//   - m: The message to deliver to all registered handlers
func (n *node) pub(m *Msg) {
	for _, h := range n.handlers {
		h(m)
	}
}
