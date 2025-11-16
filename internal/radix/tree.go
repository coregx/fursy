package radix

import (
	"fmt"
	"strings"
)

// Tree is a radix tree (compressed trie) for HTTP route matching.
// It provides efficient path matching with support for static paths,
// named parameters (:id), and catch-all parameters (*path).
type Tree struct {
	root *node
}

// Param represents a URL parameter extracted from the path.
type Param struct {
	Key   string // Parameter name (e.g., "id" from /:id)
	Value string // Parameter value extracted from path
}

// New creates a new radix tree with an initialized root node.
func New() *Tree {
	return &Tree{
		root: &node{
			path:  "",
			nType: root,
		},
	}
}

// Insert adds a new route to the tree with the given handler.
// Returns an error if the path is invalid or conflicts with existing routes.
func (t *Tree) Insert(path string, handler interface{}) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if path[0] != '/' {
		return fmt.Errorf("path must begin with '/'")
	}

	// Validate path segments
	if err := validatePath(path); err != nil {
		return err
	}

	fullPath := path
	return t.insertNode(path, handler, t.root, fullPath)
}

// insertNode is the recursive implementation of Insert.
func (t *Tree) insertNode(path string, handler interface{}, n *node, fullPath string) error {
	// If node is root, skip its path
	if n.nType == root {
		// Special case for root path "/"
		if path == "/" {
			if n.handler != nil {
				return fmt.Errorf("route already exists: /")
			}
			n.handler = handler
			n.fullPath = "/"
			return nil
		}

		// Check for wildcard at start
		if path != "" && (path[0] == ':' || path[0] == '*') {
			return t.insertWildcard(path, handler, n, fullPath)
		}

		// Check if path contains wildcard - need to split
		wildcardIdx := findWildcardIndex(path)
		if wildcardIdx > 0 {
			// Split at the wildcard: static prefix before wildcard
			staticPart := path[:wildcardIdx]
			wildcardPart := path[wildcardIdx:]

			// Try to find existing child for static part
			if staticPart != "" {
				c := staticPart[0]
				child := n.findChild(c)
				if child != nil {
					// Continue with existing node
					return t.insertNode(staticPart+wildcardPart, handler, child, fullPath)
				}
			}

			// Create static node for the prefix (don't include trailing /)
			staticNode := &node{
				path:  staticPart,
				nType: static,
			}
			n.addChild(staticNode)

			// Recurse to insert wildcard part
			return t.insertWildcard(wildcardPart, handler, staticNode, fullPath)
		}

		// No wildcards - try to find existing child
		if path != "" {
			c := path[0]
			child := n.findChild(c)
			if child != nil {
				return t.insertNode(path, handler, child, fullPath)
			}
		}

		// Create new child for this path
		child := &node{
			path:     path,
			nType:    static,
			handler:  handler,
			fullPath: fullPath,
		}
		n.addChild(child)
		return nil
	}

	// Find longest common prefix
	i := longestCommonPrefix(path, n.path)

	// Split edge if necessary (but only if there IS a common prefix!)
	if i < len(n.path) && i > 0 {
		// Create child with remaining part of current node
		child := &node{
			path:      n.path[i:],
			wildChild: n.wildChild,
			nType:     n.nType, // ⚠️ CRITICAL: Preserve node type!
			indices:   n.indices,
			children:  n.children,
			handler:   n.handler,
			priority:  n.priority,
			fullPath:  n.fullPath,
		}

		// Update current node to hold only common prefix
		n.path = n.path[:i] // ✅ Use n.path[:i] not path[:i]
		n.wildChild = false
		n.nType = static // Common prefix is always static
		n.children = []*node{child}
		n.indices = string([]byte{child.path[0]})
		n.handler = nil
		n.fullPath = ""
	}

	// If no common prefix at all, this path doesn't belong here
	if i == 0 && n.path != "" {
		return fmt.Errorf("internal error: no common prefix between %q and %q", path, n.path)
	}

	// Consume matched prefix from path
	if i < len(path) {
		path = path[i:]

		// Check for wildcard at current position
		if path[0] == ':' || path[0] == '*' {
			return t.insertWildcard(path, handler, n, fullPath)
		}

		// Check if path contains wildcard later - need to split
		wildcardIdx := findWildcardIndex(path)
		if wildcardIdx > 0 {
			// Split: static prefix + wildcard part
			staticPart := path[:wildcardIdx]
			wildcardPart := path[wildcardIdx:]

			// Try to find existing child for static part
			c := staticPart[0]
			child := n.findChild(c)

			if child != nil {
				// Continue down the tree
				child.incrementPriority()
				return t.insertNode(staticPart+wildcardPart, handler, child, fullPath)
			}

			// Create static node, then recurse for wildcard
			staticNode := &node{
				path:  staticPart,
				nType: static,
			}
			n.addChild(staticNode)
			return t.insertWildcard(wildcardPart, handler, staticNode, fullPath)
		}

		// No wildcards - try to find existing child
		c := path[0]
		child := n.findChild(c)

		if child != nil {
			// Continue down the tree
			child.incrementPriority()
			return t.insertNode(path, handler, child, fullPath)
		}

		// Create new child
		child = &node{
			path:     path,
			nType:    static,
			handler:  handler,
			fullPath: fullPath,
		}
		n.addChild(child)
		return nil
	}

	// Path consumed, this node is the endpoint
	if n.handler != nil {
		return fmt.Errorf("route already exists: %s", fullPath)
	}

	n.handler = handler
	n.fullPath = fullPath
	return nil
}

// insertWildcard handles insertion of wildcard routes (:param or *catchAll).
func (t *Tree) insertWildcard(path string, handler interface{}, n *node, fullPath string) error {
	// Skip leading slash if present (path might be "/:id" or ":id")
	if path != "" && path[0] == '/' {
		path = path[1:]
	}

	// Now path should start with : or *
	if path == "" || (path[0] != ':' && path[0] != '*') {
		return fmt.Errorf("insertWildcard called with non-wildcard path: %s", path)
	}

	// Determine wildcard type
	var wildcardType nodeType
	var end int

	if path[0] == ':' {
		wildcardType = param
		// Find end of param name (next / or end of string)
		end = 1
		for end < len(path) && path[end] != '/' {
			end++
		}
	} else { // path[0] == '*'
		wildcardType = catchAll
		// Catch-all must be last segment
		end = len(path)
	}

	// Extract wildcard name
	wildcardName := path[1:end]
	if wildcardName == "" {
		return fmt.Errorf("wildcard name cannot be empty")
	}

	// Check if wildcard already exists
	if existingWild := n.getWildChild(); existingWild != nil {
		// Extract existing wildcard name from path
		existingName := existingWild.path[1:]
		if idx := strings.IndexByte(existingName, '/'); idx != -1 {
			existingName = existingName[:idx]
		}

		if existingName != wildcardName {
			return fmt.Errorf("conflicting wildcard names: %s vs %s", existingName, wildcardName)
		}

		// Continue with existing wildcard node
		if end < len(path) {
			return t.insertNode(path[end:], handler, existingWild, fullPath)
		}

		if existingWild.handler != nil {
			return fmt.Errorf("route already exists")
		}

		existingWild.handler = handler
		existingWild.fullPath = fullPath
		return nil
	}

	// Create wildcard node
	wildcardNode := &node{
		path:     path[:end],
		nType:    wildcardType,
		fullPath: fullPath,
	}

	n.addChild(wildcardNode)

	// If more path remains after param, check what's next
	if end < len(path) {
		if wildcardType == catchAll {
			return fmt.Errorf("catch-all must be the last segment")
		}

		remaining := path[end:]
		// Remaining path should start with / (e.g., "/posts" or "/:id")
		if remaining != "" {
			// Check if remaining part has wildcards
			wildcardIdx := findWildcardIndex(remaining)
			switch {
			case wildcardIdx == 0:
				// Starts with wildcard (e.g., "/:id" - but we already skipped the /)
				// This shouldn't happen with our slicing
				return fmt.Errorf("unexpected wildcard at position 0: %s", remaining)
			case wildcardIdx > 0:
				// Has wildcard later (e.g., "/foo/:id")
				// Split: static prefix + wildcard part
				staticPart := remaining[:wildcardIdx]
				wildcardPart := remaining[wildcardIdx:]

				// Create static node
				staticNode := &node{
					path:  staticPart,
					nType: static,
				}
				wildcardNode.addChild(staticNode)

				// Recurse for wildcard part
				return t.insertWildcard(wildcardPart, handler, staticNode, fullPath)
			default:
				// No wildcards, just static path remaining
				staticNode := &node{
					path:     remaining,
					nType:    static,
					handler:  handler,
					fullPath: fullPath,
				}
				wildcardNode.addChild(staticNode)
				return nil
			}
		}
	}

	// Wildcard is endpoint
	wildcardNode.handler = handler
	return nil
}

// Lookup finds a handler for the given path and extracts parameters.
// Returns the handler, extracted parameters, and whether a match was found.
func (t *Tree) Lookup(path string) (handler interface{}, params []Param, found bool) {
	if path == "" {
		return nil, nil, false
	}

	params = make([]Param, 0, 8) // Pre-allocate for common case
	return t.lookupNode(path, t.root, params)
}

// lookupNode is the recursive implementation of Lookup.
func (t *Tree) lookupNode(path string, n *node, params []Param) (interface{}, []Param, bool) {
	// Special handling for root node
	if n.nType == root {
		// Check for exact "/" match
		if path == "/" {
			if n.handler != nil {
				return n.handler, params, true
			}
			return nil, nil, false
		}

		// Try to find matching static child FIRST (priority over wildcards)
		if path != "" {
			c := path[0]
			if child := n.findChild(c); child != nil {
				if handler, ps, found := t.lookupNode(path, child, params); found {
					return handler, ps, true
				}
			}
		}

		// Then try wildcard child (for routes like /*path)
		if n.wildChild {
			if wildChild := n.getWildChild(); wildChild != nil {
				if handler, ps, found := t.lookupWildcard(path, wildChild, params); found {
					return handler, ps, true
				}
			}
		}

		return nil, nil, false
	}

	// Check if path matches node.path prefix
	if !strings.HasPrefix(path, n.path) {
		return nil, nil, false
	}

	// Consume matched prefix
	path = path[len(n.path):]

	// If path fully consumed, check for handler
	if path == "" {
		if n.handler != nil {
			return n.handler, params, true
		}
		return nil, nil, false
	}

	// Try static children FIRST (priority over wildcards)
	c := path[0]
	if child := n.findChild(c); child != nil {
		if handler, ps, found := t.lookupNode(path, child, params); found {
			return handler, ps, true
		}
	}

	// Then try wildcard child (param or catchAll)
	if n.wildChild {
		if wildChild := n.getWildChild(); wildChild != nil {
			if handler, ps, found := t.lookupWildcard(path, wildChild, params); found {
				return handler, ps, true
			}
		}
	}

	return nil, nil, false
}

// lookupWildcard handles lookup in wildcard nodes.
func (t *Tree) lookupWildcard(path string, n *node, params []Param) (interface{}, []Param, bool) {
	// Extract param name from node path
	paramName := n.path[1:] // Skip ':' or '*'

	if n.nType == catchAll {
		// Catch-all captures entire remaining path
		params = append(params, Param{
			Key:   paramName,
			Value: path,
		})
		if n.handler != nil {
			return n.handler, params, true
		}
		return nil, nil, false
	}

	// param type: capture until next '/' or end
	end := 0
	for end < len(path) && path[end] != '/' {
		end++
	}

	// Extract param name if it contains '/'
	if idx := strings.IndexByte(paramName, '/'); idx != -1 {
		paramName = paramName[:idx]
	}

	params = append(params, Param{
		Key:   paramName,
		Value: path[:end],
	})

	// If more path remains, continue
	if end < len(path) {
		path = path[end:]
		// Find next segment
		if len(n.children) > 0 {
			for _, child := range n.children {
				if handler, ps, found := t.lookupNode(path, child, params); found {
					return handler, ps, true
				}
			}
		}
		return nil, nil, false
	}

	// Path consumed
	if n.handler != nil {
		return n.handler, params, true
	}

	return nil, nil, false
}

// validatePath validates the path format.
func validatePath(path string) error {
	for i := 0; i < len(path); i++ {
		c := path[i]

		// Check for wildcard
		if c == ':' || c == '*' {
			// Ensure there's a name after wildcard
			if i+1 >= len(path) || path[i+1] == '/' {
				return fmt.Errorf("wildcard name cannot be empty at position %d", i)
			}

			// Catch-all must be last
			if c == '*' {
				// Find end of wildcard name
				end := i + 1
				for end < len(path) && path[end] != '/' {
					end++
				}
				// If not at end, error
				if end < len(path) {
					return fmt.Errorf("catch-all must be the last segment")
				}
			}
		}
	}

	return nil
}

// longestCommonPrefix returns the length of the longest common prefix.
func longestCommonPrefix(a, b string) int {
	i := 0
	maxLen := min(len(a), len(b))
	for i < maxLen && a[i] == b[i] {
		i++
	}
	return i
}

// findWildcardIndex finds the index of the first wildcard (: or *) in the path.
// Returns -1 if no wildcard found.
func findWildcardIndex(path string) int {
	for i := 0; i < len(path); i++ {
		if path[i] == ':' || path[i] == '*' {
			return i
		}
	}
	return -1
}
