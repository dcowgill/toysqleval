package parser

import (
	"fmt"

	"github.com/dcowgill/toysqleval/token"
)

// Error is an error that occurs during parsing.
type Error struct {
	Msg string
	Pos token.Pos
}

// Error implements the error interface.
func (err *Error) Error() string {
	return fmt.Sprintf("parse error at %s: %s", err.Pos, err.Msg)
}
