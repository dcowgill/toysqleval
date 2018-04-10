package ast

import (
	"fmt"
	"io"
	"strings"
)

// PrettyPrinter is a utility for printing an AST with indentation.
type PrettyPrinter struct {
	Writer io.Writer
	Indent string
	depth  int
}

// Prints a line with indentation. Appends a newline.
func (pp *PrettyPrinter) printf(format string, args ...interface{}) {
	fmt.Fprint(pp.Writer, strings.Repeat(pp.Indent, pp.depth))
	fmt.Fprintf(pp.Writer, format+"\n", args...)
}

// Visit pretty-prints the AST starting at n.
func (pp *PrettyPrinter) Visit(n Node) {
	pp.depth++

	switch n := n.(type) {
	case *CreateTableStmt:
		pp.printf("CREATE TABLE")
		pp.Visit(n.Table)
		for _, child := range n.Columns {
			pp.Visit(child)
		}

	case *ColumnDefinition:
		nullable := "Y"
		if n.Nullable {
			nullable = "N"
		}
		pp.printf("%q type=%-7s nullable=%s default=%s", n.Name.Name, n.Type, nullable, "NULL")

	case *SelectStmt:
		pp.printf("SELECT")
		for _, child := range n.Columns {
			pp.Visit(child)
		}
		pp.printf("FROM")
		pp.Visit(n.Table)
		if n.Where != nil {
			pp.printf("WHERE")
			pp.Visit(n.Where)
		}

	case *InsertStmt:
		pp.printf("INSERT INTO")
		pp.Visit(n.Table)
		pp.printf("COLUMNS")
		for _, child := range n.Columns {
			pp.Visit(child)
		}
		pp.printf("VALUES")
		for _, child := range n.Values {
			pp.Visit(child)
		}

	case *UpdateStmt:
		pp.printf("UPDATE")
		pp.Visit(n.Table)
		pp.printf("SET")
		for i, child := range n.Columns {
			pp.Visit(child)
			pp.printf("=")
			pp.Visit(n.Values[i])
		}
		if n.Where != nil {
			pp.printf("WHERE")
			pp.Visit(n.Where)
		}

	case *DeleteStmt:
		pp.printf("DELETE FROM")
		pp.Visit(n.Table)
		if n.Where != nil {
			pp.printf("WHERE")
			pp.Visit(n.Where)
		}

	case *Ident:
		pp.printf("Ident(%s)", n.Name)

	case *BinaryExpr:
		pp.printf("BinaryExpr(%s)", n.Op)
		pp.Visit(n.Lhs)
		pp.Visit(n.Rhs)

	case *UnaryExpr:
		pp.printf("UnaryExpr(%s)", n.Op)
		pp.Visit(n.Expr)

	case *SelectStarExpr:
		pp.printf("*")

	case *IntegerLiteral:
		pp.printf("Integer(%d)", n.Value)

	case *NumberLiteral:
		pp.printf("Number(%v)", n.Value)

	case *StringLiteral:
		pp.printf("String(%q)", n.Value)

	case *BooleanLiteral:
		pp.printf("Boolean(%v)", n.Value)

	case *Null:
		pp.printf("NULL")

	case *FunctionCall:
		pp.printf("FunctionCall")
		pp.Visit(n.Name)
		for _, arg := range n.Args {
			pp.Visit(arg)
		}

	default:
		panic(fmt.Sprintf("unknown node type: %t", n))
	}

	pp.depth--
}
