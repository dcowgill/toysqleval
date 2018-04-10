package token

import "fmt"

// Token represents a SQL lexeme.
type Token struct {
	Kind Kind
	Pos  Pos
	Lit  string
}

// String implements the Stringer interface.
func (tok Token) String() string {
	switch tok.Kind {
	case Ident:
		return fmt.Sprintf("Ident(%s)", tok.Lit)
	case NumberLiteral:
		return fmt.Sprintf("NumberLiteral(%s)", tok.Lit)
	case StringLiteral:
		return fmt.Sprintf("StringLiteral(%q)", tok.Lit)
	}
	return tok.Kind.String()
}
