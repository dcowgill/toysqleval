package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dcowgill/toysqleval/token"
)

const singleQuote = '\''

var sqlKeywords = map[string]token.Kind{
	"and":       token.And,
	"boolean":   token.Boolean,
	"create":    token.Create,
	"delete":    token.Delete,
	"false":     token.False,
	"from":      token.From,
	"insert":    token.Insert,
	"integer":   token.Integer,
	"into":      token.Into,
	"not":       token.Not,
	"null":      token.Null,
	"number":    token.Number,
	"or":        token.Or,
	"select":    token.Select,
	"set":       token.Set,
	"table":     token.Table,
	"timestamp": token.Timestamp,
	"true":      token.True,
	"update":    token.Update,
	"values":    token.Values,
	"varchar":   token.Varchar,
	"where":     token.Where,
}

// Lexer represents a SQL lexical analyzer.
type Lexer struct {
	input   []rune      // text to parse into tokens
	pos     int         // current offset in input
	line    int         // current line number; one-based counting
	linePos int         // offset in input of current line start
	tok     token.Token // one lookahead token
	err     error       // if not nil, lexer is the error state
}

// New creates a new lexer for the given input.
func New(input string) *Lexer {
	return &Lexer{input: []rune(input), line: 1}
}

// Scan advances the lexer to the next token, which can be retrieved by calling
// the Token method. Returns false if there are no more tokens or if an error
// occurs. Call the Err method to inspect the error state.
func (lex *Lexer) Scan() bool {
	if lex.err != nil {
		return false
	}
	lex.tok = token.Token{Pos: lex.getPos()}
	if !lex.skipSpace() {
		return false
	}
	r := lex.input[lex.pos]
	switch r {
	case ',':
		return lex.consumeRune(token.Comma)
	case '|':
		if lex.matchString("||", token.Concat) {
			return true
		}
	case '/':
		return lex.consumeRune(token.Div)
	case '.':
		if r1 := lex.nextRune(); unicode.IsDigit(r1) {
			return lex.consumeNumber() // e.g. ".01"
		}
		return lex.consumeRune(token.Dot)
	case '=':
		return lex.consumeRune(token.Equal)
	case '>':
		if lex.matchString(">=", token.GreaterThanOrEqualTo) {
			return true
		}
		return lex.consumeRune(token.GreaterThan)
	case '(':
		return lex.consumeRune(token.LeftParen)
	case '<':
		if lex.matchString("<=", token.LessThanOrEqualTo) {
			return true
		}
		return lex.consumeRune(token.LessThan)
	case '-':
		return lex.consumeRune(token.Minus)
	case '*':
		return lex.consumeRune(token.Mul)
	case '!':
		if lex.matchString("!=", token.NotEqual) {
			return true
		}
	case '+':
		return lex.consumeRune(token.Plus)
	case ')':
		return lex.consumeRune(token.RightParen)
	case ';':
		return lex.consumeRune(token.Semicolon)
	case singleQuote:
		return lex.consumeQuotedString()
	}

	if unicode.IsDigit(r) {
		return lex.consumeNumber()
	}

	if word := strings.ToLower(lex.nextWord()); word != "" {
		// Try to match one of the SQL keywords.
		if kind, ok := sqlKeywords[word]; ok {
			lex.setToken(kind)
			lex.pos += len(word)
			return true
		}
		// Otherwise, it's an arbitrary identifier.
		lex.setToken(token.Ident)
		lex.tok.Lit = word
		lex.pos += len(word)
		return true
	}

	lex.err = lex.errorf("unexpected character: %c", lex.input[lex.pos])
	return false
}

// Token returns the current token. Calling Scan will overwrite this value.
func (lex *Lexer) Token() token.Token {
	return lex.tok
}

// EOF reports whether the input is exhausted.
func (lex *Lexer) EOF() bool {
	return lex.pos >= len(lex.input)
}

// Err reports the lexing error, if any. Once the lexer encounters an error,
// further calls to Scan will immediately return false, doing nothing.
func (lex *Lexer) Err() error {
	return lex.err
}

// Returns the rune in the input after the current one. If there isn't enough
// input, returns zero.
func (lex *Lexer) nextRune() rune {
	if lex.pos < len(lex.input)-1 {
		return lex.input[lex.pos+1]
	}
	return 0
}

// Reports the current location in the input.
func (lex *Lexer) getPos() token.Pos {
	return token.Pos{Line: lex.line, Column: lex.pos - lex.linePos}
}

// Sets the current token and records its position.
func (lex *Lexer) setToken(kind token.Kind) {
	lex.tok.Kind = kind
	lex.tok.Pos = lex.getPos()
}

// Advances past any whitespace starting at the current position, keeping track
// of the current line number. Returns true if any input remains.
func (lex *Lexer) skipSpace() bool {
	lineTracker := newLineTracker(lex)
	for lex.pos < len(lex.input) {
		switch {
		case !unicode.IsSpace(lex.input[lex.pos]):
			lineTracker.sync()
			return true
		case lex.input[lex.pos] == '\n':
			lineTracker.next(lex.pos)
		case lex.input[lex.pos] == '\r':
			if lex.nextRune() == '\n' {
				lex.pos++
			}
			lineTracker.next(lex.pos)
		}
		lex.pos++
	}
	return false
}

// Advances one rune and sets the current token. Never fails; that is, use this
// when the token consists of exactly one rune and its kind is known a priori.
func (lex *Lexer) consumeRune(kind token.Kind) bool {
	lex.setToken(kind)
	lex.pos++
	return true
}

// Parses and advances past a number at the current position. On failure, moves
// the lexer into the error state and returns false.
func (lex *Lexer) consumeNumber() bool {
	i := lex.pos
	for i < len(lex.input) {
		switch lex.input[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', 'e':
			i++
			continue
		}
		break
	}
	if i == lex.pos {
		panic("expected a number")
	}
	lit := string(lex.input[lex.pos:i])
	if _, err := strconv.ParseFloat(lit, 64); err != nil {
		lex.err = lex.errorf("invalid numeric literal: %s", err)
		return false
	}
	lex.setToken(token.NumberLiteral)
	lex.tok.Lit = lit
	lex.pos = i
	return true
}

// Parses and advances past a single-quoted string at the current position. On
// failure, moves the lexer into the error state and returns false.
func (lex *Lexer) consumeQuotedString() bool {
	if lex.input[lex.pos] != singleQuote {
		panic("called on non single-quote char")
	}
	var runes []rune
	i := lex.pos + 1
	lineTracker := newLineTracker(lex)
	for i < len(lex.input) {
		switch lex.input[i] {
		case singleQuote:
			// A literal single quote may appear within a string as a doubled
			// single quote. Otherwise, this marks the end of the string.
			if i < len(lex.input)-1 && lex.input[i+1] == singleQuote {
				runes = append(runes, singleQuote)
				i += 2
				continue
			}
			lex.setToken(token.StringLiteral)
			lex.tok.Lit = string(runes)
			lineTracker.sync()
			lex.pos = i + 1
			return true
		case '\n':
			// Newlines are legal in SQL string literals.
			lineTracker.next(i)
		case '\r':
			// We should also permit DOS-style line endings (CRLF), since we
			// can't control the input encoding and we want sensible error
			// messages regardless of the kind of line endings in the input.
			if i < len(lex.input)-1 && lex.input[i] == '\n' {
				runes = append(runes, lex.input[i])
				i++
			}
			lineTracker.next(i)
		}
		runes = append(runes, lex.input[i])
		i++
	}
	lex.err = lex.errorf("unterminated string")
	return false
}

// Tries to match the specified string exactly. On success, advances the lexer
// position past the string and sets the current token to the given kind.
func (lex *Lexer) matchString(s string, kind token.Kind) bool {
	if nextPos := lex.pos + len(s); nextPos > len(lex.input) {
		return false // insufficient input remains
	}
	for i, ch := range s {
		if lex.input[lex.pos+i] != ch {
			return false // rune mismatch
		}
	}
	lex.setToken(kind)
	lex.pos += len(s)
	return true
}

// Returns the word at the current position, or an empty string is none exists.
// A word is defined as a sequence of runes that could form a valid SQL
// identifier. It is a runtime error to call this when the input is exhausted.
func (lex *Lexer) nextWord() string {
	if lex.pos >= len(lex.input) {
		panic("input exhausted")
	}
	if first := lex.input[lex.pos]; !unicode.IsLetter(first) && first != '_' {
		return "" // a word must begin with a letter or underscore
	}
	i := lex.pos + 1
	for i < len(lex.input) && isIdentRune(lex.input[i]) {
		i++
	}
	return string(lex.input[lex.pos:i])
}

// Reports whether r is a rune that may appear in a SQL identifier.
func isIdentRune(r rune) bool {
	switch {
	case unicode.IsLetter(r):
	case unicode.IsDigit(r):
	case r == '_':
	default:
		return false
	}
	return true
}

// Generates an error at the current line and column.
func (lex *Lexer) errorf(format string, args ...interface{}) error {
	return &Error{Msg: fmt.Sprintf(format, args...), Pos: lex.getPos()}
}
