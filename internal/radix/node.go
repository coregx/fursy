// Package radix implements a radix tree (compressed trie) for HTTP routing.
package radix

// nodeType identifies the type of node in the radix tree.
type nodeType uint8

const (
	static   nodeType = iota // static path segment (e.g., /users)
	root                     // root node (/)
	param                    // named parameter (e.g., /:id)
	catchAll                 // catch-all parameter (e.g., /*path)
)

// node represents a node in the radix tree.
// Each node stores a path segment and may have children and a handler.
type node struct {
	// path stores the path segment for this node.
	// For compressed edges, this contains the full compressed prefix.
	path string

	// nType identifies the node type (static, root, param, catchAll).
	nType nodeType

	// wildChild indicates if this node has a wildcard child (param or catchAll).
	// This allows fast checking without iterating children.
	wildChild bool

	// indices stores the first byte of each child's path for fast lookup.
	// Example: if children have paths "users", "posts", "admin",
	// indices would be "upa" for O(1) child selection.
	indices string

	// children holds pointers to child nodes.
	// Ordered by priority (hot paths first) for performance.
	children []*node

	// handler stores the handler for this route if this node is an endpoint.
	// nil if this node is not a route endpoint.
	handler interface{}

	// priority represents traversal frequency, used for child reordering.
	// Higher priority children are checked first for better cache locality.
	priority uint32

	// fullPath stores the complete route path for this endpoint.
	// Used for debugging and error messages.
	fullPath string
}

// incrementPriority increases the priority of this node.
// Called during route insertion to track "hotness".
func (n *node) incrementPriority() {
	n.priority++
}

// addChild adds a child node and updates the indices string.
func (n *node) addChild(child *node) {
	if child.path == "" {
		return
	}

	// Add first character of child's path to indices
	c := child.path[0]
	n.indices += string(c)
	n.children = append(n.children, child)

	// If child is param or catchAll, mark wildChild
	if child.nType == param || child.nType == catchAll {
		n.wildChild = true
	}
}

// findChild finds a child node by the first character of its path.
// Returns the child node, or nil if not found.
func (n *node) findChild(c byte) *node {
	// Fast path: check indices string
	for i := 0; i < len(n.indices); i++ {
		if n.indices[i] == c {
			return n.children[i]
		}
	}
	return nil
}

// getWildChild returns the wildcard child (param or catchAll) if exists.
func (n *node) getWildChild() *node {
	if !n.wildChild {
		return nil
	}

	// Find param or catchAll child
	for _, child := range n.children {
		if child.nType == param || child.nType == catchAll {
			return child
		}
	}
	return nil
}
