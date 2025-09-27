package messanger

import "strings"

type node struct {
	index    string
	nodes    map[string]*node
	handlers []MsgHandler
}

var (
	root *node
)

func init() {
	root = &node{
		nodes: make(map[string]*node),
	}
}

func newNode(index string) *node {
	n := &node{
		index: index,
		nodes: make(map[string]*node),
	}
	return n
}

func initNodes() {
	root = &node{
		nodes: make(map[string]*node),
	}
}

func clearNodes() {
	root = nil
}

func resetNodes() {
	clearNodes()
	initNodes()
}

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

func (n *node) remove(topic string, handler MsgHandler) {
	indexes := strings.Split(topic, "/")
	
	// Track the path to the target node
	path := make([]*node, 0, len(indexes)+1)
	pn := n
	path = append(path, pn)
	
	// Navigate to the target node, building the path
	for _, idx := range indexes {
		if nn, ex := pn.nodes[idx]; ex {
			pn = nn
			path = append(path, pn)
		} else {
			// Topic not found, nothing to remove
			return
		}
	}
	
	// For the Close() use case, we clear all handlers for the topic
	// This is appropriate since Close() should remove all subscriptions for this messanger instance
	targetNode := path[len(path)-1]
	targetNode.handlers = nil
	
	// Clean up empty nodes from leaf to root
	for i := len(path) - 1; i > 0; i-- {
		node := path[i]
		parent := path[i-1]
		parentIndex := indexes[i-1]
		
		// Remove node if it has no handlers and no child nodes
		if len(node.handlers) == 0 && len(node.nodes) == 0 {
			delete(parent.nodes, parentIndex)
		} else {
			// If this node is not empty, stop cleanup (parent nodes above might still be needed)
			break
		}
	}
}

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

func (n *node) pub(m *Msg) {
	for _, h := range n.handlers {
		h(m)
	}
}
