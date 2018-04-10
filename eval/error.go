package eval

import (
	"fmt"

	"github.com/dcowgill/toysqleval/ast"
	"github.com/dcowgill/toysqleval/token"
)

// Error is an error that occurs during statement evaluation.
type Error struct {
	Msg string
	Pos token.Pos
}

// Error implements the error interface.
func (err *Error) Error() string {
	return fmt.Sprintf("eval:%s: %s", err.Pos, err.Msg)
}

// Returns an evaluation error. Position is indicated by the node.
func errorf(node ast.Node, format string, args ...interface{}) error {
	return &Error{Msg: fmt.Sprintf(format, args...), Pos: node.Pos()}
}
