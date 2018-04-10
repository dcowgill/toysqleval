package token

import "fmt"

// Kind specifies the type of SQL lexical element.
type Kind uint8

const (
	Invalid Kind = iota
	And
	Boolean
	Comma
	Concat
	Create
	Delete
	Div
	Dot
	Equal
	False
	From
	GreaterThan
	GreaterThanOrEqualTo
	Ident
	Insert
	Integer
	Into
	LeftParen
	LessThan
	LessThanOrEqualTo
	Minus
	Mul
	Not
	NotEqual
	Null
	Number
	NumberLiteral
	Or
	Plus
	RightParen
	Select
	Semicolon
	Set
	StringLiteral
	Table
	Timestamp
	True
	Update
	Values
	Varchar
	Where
)

func (k Kind) Precedence() int {
	switch k {
	case Or:
		return 1
	case And:
		return 2
	case Equal, NotEqual, LessThan, LessThanOrEqualTo, GreaterThan, GreaterThanOrEqualTo:
		return 3
	case Plus, Minus, Concat:
		return 4
	case Mul, Div:
		return 5
	}
	return 0
}

// String implements the Stringer interface.
func (k Kind) String() string {
	switch k {
	case Invalid:
		return "Invalid"
	case And:
		return "AND"
	case Boolean:
		return "BOOLEAN"
	case Comma:
		return ","
	case Concat:
		return "||"
	case Create:
		return "CREATE"
	case Delete:
		return "DELETE"
	case Div:
		return "/"
	case Dot:
		return "."
	case Equal:
		return "="
	case False:
		return "FALSE"
	case From:
		return "FROM"
	case GreaterThan:
		return ">"
	case GreaterThanOrEqualTo:
		return ">="
	case Ident:
		return "Ident"
	case Insert:
		return "INSERT"
	case Integer:
		return "INTEGER"
	case Into:
		return "INTO"
	case LeftParen:
		return "("
	case LessThan:
		return "<"
	case LessThanOrEqualTo:
		return "<="
	case Minus:
		return "-"
	case Mul:
		return "*"
	case Not:
		return "NOT"
	case NotEqual:
		return "!="
	case Null:
		return "NULL"
	case Number:
		return "NUMBER"
	case NumberLiteral:
		return "NumberLiteral"
	case Or:
		return "OR"
	case Plus:
		return "+"
	case RightParen:
		return ")"
	case Select:
		return "SELECT"
	case Semicolon:
		return ";"
	case Set:
		return "SET"
	case StringLiteral:
		return "StringLiteral"
	case Table:
		return "TABLE"
	case Timestamp:
		return "TIMESTAMP"
	case True:
		return "TRUE"
	case Update:
		return "UPDATE"
	case Values:
		return "VALUES"
	case Varchar:
		return "VARCHAR"
	case Where:
		return "WHERE"
	}
	panic(fmt.Sprintf("unknown token.Kind: %d", k))
}
