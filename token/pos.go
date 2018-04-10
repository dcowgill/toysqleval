package token

import "fmt"

// Pos represents a location in the input.
type Pos struct {
	Line   int // 1-based
	Column int // 0-based
}

// String implements the Stringer interface.
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}
