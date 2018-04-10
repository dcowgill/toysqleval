package eval

import (
	"fmt"

	"github.com/dcowgill/toysqleval/ast"
	"github.com/dcowgill/toysqleval/token"
)

type DataType uint8

const (
	InvalidDataType DataType = iota
	Boolean
	Integer
	Number
	String
	Timestamp
)

func (dt DataType) String() string {
	switch dt {
	case Boolean:
		return "Boolean"
	case Integer:
		return "Integer"
	case Number:
		return "Number"
	case String:
		return "String"
	case Timestamp:
		return "Timestamp"
	}
	return "Unknown"
}

// Converts a token to a corresponding data type. Panics if the token does not
// correspond to a valid data type.
func dataTypeFromToken(node ast.Node, tok token.Kind) DataType {
	switch tok {
	case token.Boolean:
		return Boolean
	case token.Integer:
		return Integer
	case token.Number:
		return Number
	case token.Varchar:
		return String
	case token.Timestamp:
		return Timestamp
	}
	panic(fmt.Sprintf("internal error: %q does not refer to a valid data type", tok))
}
