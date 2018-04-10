package lexer

import (
	"fmt"

	"github.com/dcowgill/toysqleval/token"
)

// Error is an error that occurs during lexical analysis.
type Error struct {
	Msg string
	Pos token.Pos
}

// Error implements the error interface.
func (err *Error) Error() string {
	return fmt.Sprintf("lexer:%s: %s", err.Pos, err.Msg)
}
