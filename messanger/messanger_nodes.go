package messanger

import (
	"strings"	
)

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
	println("TODO add node remove command")
	return


	// // Need to remove the child node and all parent nodes provided
	// // they have no other subscribers
	// nodes = []*node
	// pn := root
	// indexes := strings.Split(topic, "/")
	// for _, idx := range indexes {
	// 	if n, ex := pn.nodes[idx]; ex {
	// 		nodes = append(nodes, n)
	// 	}
	// }

	// nodes = slices.Reverse(nodes)
	// for i, h := range node[0].handlers {
	// 	if h == handler {
	// 		node[0].handlers = append(nodes[:i], nodes[i:])
	// 	}
	// }
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
