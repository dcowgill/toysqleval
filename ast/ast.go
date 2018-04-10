package ast

import (
	"github.com/dcowgill/toysqleval/token"
)

// Node is the interface which all AST nodes must implement.
type Node interface {
	Pos() token.Pos
}

// Expr is an expression node.
type Expr interface {
	Node
}

// CreateTableStmt is a CREATE TABLE statement node.
type CreateTableStmt struct {
	StartPos token.Pos
	Table    *Ident
	Columns  []*ColumnDefinition
}

func (n *CreateTableStmt) Pos() token.Pos { return n.StartPos }

// ColumnDefinition is a table column definition node.
type ColumnDefinition struct {
	Name     *Ident
	Type     token.Kind // e.g. Integer, Varchar, etc.
	Nullable bool
	// DefaultValue Expr
}

func (n *ColumnDefinition) Pos() token.Pos { return n.Name.Pos() }

// // DataType is a column data type node.
// type DataType struct {
// 	TypePos token.Pos
// 	Type    token.Kind
// }

// func (n *DataType) Pos() token.Pos { return n.TypePos }

// SelectStmt is a SELECT statement node.
type SelectStmt struct {
	StartPos token.Pos
	Columns  []Expr
	Table    Expr
	Where    Expr
}

func (n *SelectStmt) Pos() token.Pos { return n.StartPos }

// SelectStarExpr represents the "*" SQL operator in a SELECT expression list.
type SelectStarExpr struct {
	StartPos token.Pos
}

func (n *SelectStarExpr) Pos() token.Pos { return n.StartPos }

// InsertStmt is an INSERT statement node.
type InsertStmt struct {
	StartPos token.Pos
	Table    *Ident
	Columns  []*Ident
	Values   []Expr
}

func (n *InsertStmt) Pos() token.Pos { return n.StartPos }

// UpdateStmt is an UPDATE statement node.
type UpdateStmt struct {
	StartPos token.Pos
	Table    *Ident
	Columns  []*Ident
	Values   []Expr
	Where    Expr
}

func (n *UpdateStmt) Pos() token.Pos { return n.StartPos }

// DeleteStmt is a DELETE statement node.
type DeleteStmt struct {
	StartPos token.Pos
	Table    *Ident
	Where    Expr
}

func (n *DeleteStmt) Pos() token.Pos { return n.StartPos }

// Ident is an identifier node.
type Ident struct {
	NamePos token.Pos
	Name    string
}

func (n *Ident) Pos() token.Pos { return n.NamePos }

// BinaryExpr is a binary expression node.
type BinaryExpr struct {
	Lhs Expr
	Op  token.Kind
	Rhs Expr
}

func (n *BinaryExpr) Pos() token.Pos { return n.Lhs.Pos() }

// UnaryExpr is a unary expression node.
type UnaryExpr struct {
	StartPos token.Pos
	Op       token.Kind
	Expr     Expr
}

func (n *UnaryExpr) Pos() token.Pos { return n.StartPos }

type IntegerLiteral struct {
	ValuePos token.Pos
	Value    int64
}

func (n *IntegerLiteral) Pos() token.Pos { return n.ValuePos }

type NumberLiteral struct {
	ValuePos token.Pos
	Value    float64
}

func (n *NumberLiteral) Pos() token.Pos { return n.ValuePos }

type StringLiteral struct {
	ValuePos token.Pos
	Value    string
}

func (n *StringLiteral) Pos() token.Pos { return n.ValuePos }

type BooleanLiteral struct {
	ValuePos token.Pos
	Value    bool
}

func (n *BooleanLiteral) Pos() token.Pos { return n.ValuePos }

type Null struct {
	ValuePos token.Pos
}

func (n *Null) Pos() token.Pos { return n.ValuePos }

type FunctionCall struct {
	Name *Ident
	Args []Expr
}

func (n *FunctionCall) Pos() token.Pos { return n.Name.Pos() }
