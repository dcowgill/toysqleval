package lexer

// A utility type to help keep track of newlines, so that they can be reported
// accurately in tokens and in error messages.
type lineTracker struct {
	lex     *Lexer
	line    int
	linePos int
}

// Creates a new line tracker.
func newLineTracker(lex *Lexer) *lineTracker {
	return &lineTracker{
		lex:     lex,
		line:    lex.line,
		linePos: lex.linePos,
	}
}

// Terminates the current line, which must end at lineEndPos.
func (t *lineTracker) next(lineEndPos int) {
	t.line++
	t.linePos = lineEndPos + 1
}

// Updates the lexer to the current line and column.
func (t *lineTracker) sync() {
	t.lex.line = t.line
	t.lex.linePos = t.linePos
}
