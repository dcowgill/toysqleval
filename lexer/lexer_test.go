package lexer

import (
	"strconv"
	"testing"

	"github.com/dcowgill/toysqleval/token"
)

// TODO: write tests!
func TestValidInputs(t *testing.T) {
	var tests = []struct {
		input  string
		tokens []token.Token
	}{}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tokens, err := lexAll(tt.input)
			if err != nil {
				t.Fatalf("lexer failed: %s", err)
			}
			if len(tokens) != len(tt.tokens) {
				t.Fatalf("lexer returned %d tokens (%+v), want %d tokens", len(tokens), tokens, len(tt.tokens))
			}
			for i, expected := range tt.tokens {
				if actual := tokens[i]; !tokEq(actual, expected) {
					t.Fatalf("token %d is %s, want %s", i, actual, expected)
				}
			}
		})
	}
}

// TODO: write tests!
func TestInvalidInput(t *testing.T) {
	var tests = []struct {
		input string
		err   string
	}{}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tokens, err := lexAll(tt.input)
			if err == nil {
				t.Fatalf("lexer succeeded, want failure; tokens are %+v", tokens)
			}
			if err.Error() != tt.err {
				t.Fatalf("lexer returned error %q, want %q", err, tt.err)
			}
		})
	}
}

// Lexes input into tokens and returns the lexer error, if any.
func lexAll(input string) ([]token.Token, error) {
	lex := New(input)
	var tokens []token.Token
	for lex.Scan() {
		tokens = append(tokens, lex.Token())
	}
	return tokens, lex.Err()
}

// Compares tokens while ignoring their positions in the input.
func tokEq(t, u token.Token) bool {
	return t.Kind == u.Kind && t.Lit == u.Lit
}
