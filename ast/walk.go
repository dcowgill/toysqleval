package ast

// WalkFunc is a function that can be passed to Walk.
type WalkFunc func(Node) WalkFunc

// Walk traverses an AST in depth-first order: it starts by calling fn(node),
// unless node is nil, in which case Walk returns immediately. If the function
// returned by fn(node) is not nil, Walk is invoked recursively with that
// function for each of the non-nil children of node.
func Walk(node Node, fn WalkFunc) {
	if node == nil {
		return
	}
	if fn = fn(node); fn == nil {
		return
	}

	switch node := node.(type) {
	case *SelectStmt:
		for _, child := range node.Columns {
			Walk(child, fn)
		}
		Walk(node.Table, fn)
		Walk(node.Where, fn)

	case *InsertStmt:
		Walk(node.Table, fn)
		for _, child := range node.Columns {
			Walk(child, fn)
		}
		for _, child := range node.Values {
			Walk(child, fn)
		}

	case *UpdateStmt:
		Walk(node.Table, fn)
		for i, child := range node.Columns {
			Walk(child, fn)
			Walk(node.Values[i], fn)
		}
		Walk(node.Where, fn)

	case *DeleteStmt:
		Walk(node.Table, fn)
		Walk(node.Where, fn)

	case *BinaryExpr:
		Walk(node.Lhs, fn)
		Walk(node.Rhs, fn)

	case *UnaryExpr:
		Walk(node.Expr, fn)

	case *FunctionCall:
		Walk(node.Name, fn)
		for _, arg := range node.Args {
			Walk(arg, fn)
		}
	}
}
