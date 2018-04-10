package eval

import (
	"fmt"

	"github.com/dcowgill/toysqleval/ast"
)

// Namespace is a symbol table of columns.
type namespace interface {
	// Looks up a column value by name.
	lookup(name string) Value

	// Looks up an aggregate function expression by the address of its AST node.
	aggFunc(expr *ast.FunctionCall) Value
}

// emptyNamespace represents an empty namespace.
type emptyNamespace struct{}

// lookup is part of the namespace interface.
func (ns emptyNamespace) lookup(name string) Value {
	panic(fmt.Errorf("column %q does not exist", name))
}

// aggFunc is part of the namespace interface.
func (ns emptyNamespace) aggFunc(expr *ast.FunctionCall) Value {
	panic(fmt.Sprintf("failed aggFunc lookup at %s", expr.Pos()))
}

// Represents the current row being evaluated in a select or update statement.
type currentRow struct {
	table *Table
	row   []Value
}

// lookup is part of the namespace interface.
func (ns *currentRow) lookup(name string) Value {
	if ns != nil {
		if i := ns.table.colIndex(name); i >= 0 {
			return ns.row[i]
		}
	}
	panic(fmt.Errorf("column %q does not exist", name))
}

// aggFunc is part of the namespace interface.
func (ns *currentRow) aggFunc(expr *ast.FunctionCall) Value {
	panic(fmt.Sprintf("failed aggFunc lookup at %s", expr.Pos()))
}
