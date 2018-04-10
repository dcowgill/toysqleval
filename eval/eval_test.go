package eval_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/dcowgill/toysqleval/eval"
	"github.com/dcowgill/toysqleval/lexer"
	"github.com/dcowgill/toysqleval/parser"
	"github.com/dcowgill/toysqleval/pprint"
)

// Loads the input files in the "tests" subdirectory, runs each of them through
// the evaluator, and compares the result to the corresponding output file.
func TestEvalFile(t *testing.T) {
	for _, tt := range loadTestFiles() {
		t.Run(tt.name, func(t *testing.T) {
			actual := deleteTrailingWhitespace(evalFile(tt.input))
			expect := deleteTrailingWhitespace(readFile(tt.output))
			if actual != expect {
				t.Fatalf("wrong output; actual versus expected:\n%q\n%q", actual, expect)
			}
		})
	}
}

// Evaluates a SQL file and returns the output. Panics if an error occurs
// reading the file or parsing its contents; this is a test of the evaluator,
// not the parser or lexical analyzer.
//
// For SELECT statements, we print the resulting table; see the pprint package.
// For statements that do not have a result, such as INSERT and UPDATE
// statements, we print "OK\n".
// Otherwise, if the evaluator returns an error, we print it, then a newline.
//
func evalFile(filename string) string {
	lex := lexer.New(readFile(filename))
	stmts, err := parser.Parse(lex)
	must(err)
	sb := new(strings.Builder)
	var env eval.Environment
	for _, stmt := range stmts {
		result, err := eval.EvalStmt(&env, stmt)
		switch {
		case err != nil:
			fmt.Fprintln(sb, err.Error())
		case result == nil:
			fmt.Fprintln(sb, "OK")
		default:
			pprint.Table(sb, result)
		}
	}
	return sb.String()
}

// A test case represented as a pair of files: the input is SQL, and the output
// is the expected output of the evalFile function, given the input.
type testPair struct {
	name          string // name of test case
	input, output string // filenames
}

// Returns a sorted set of input/output file pairs.
func loadTestFiles() []*testPair {
	wd, err := os.Getwd() // guaranteed to be same as this file's directory
	must(err)
	wd = path.Join(wd, "tests")
	files, err := ioutil.ReadDir(wd)
	must(err)
	// Read all the files in the directory and accumulate in pairs.
	pairs := make(map[string]*testPair)
	for _, fi := range files {
		var (
			filename = fi.Name()
			absPath  = path.Join(wd, filename)
			base     = path.Base(filename)
			ext      = path.Ext(filename)
			name     = base[:len(base)-len(ext)] // remove the extension
		)
		if _, ok := pairs[name]; !ok {
			pairs[name] = &testPair{name: name}
		}
		switch ext {
		case ".in":
			pairs[name].input = absPath
		case ".out":
			pairs[name].output = absPath
		default:
			panic(fmt.Sprintf("unexpected file in test directory: %q", filename))
		}
	}
	// Verify that every input has an output and vice versa.
	// Also accumulate the test case names, then sort them.
	keys := make([]string, 0, len(pairs))
	for name, p := range pairs {
		if p.input == "" || p.output == "" {
			panic(fmt.Sprintf("test %q missing either input (%q) or output (%q)", name, p.input, p.output))
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	// Return the test cases as a list, not a map.
	answer := make([]*testPair, len(keys))
	for i, key := range keys {
		answer[i] = pairs[key]
	}
	return answer
}

// Removes trailing whitespace from the end of each line in s. This makes
// comparisons of file contents less brittle w/r/t formatting.
func deleteTrailingWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRightFunc(line, func(r rune) bool { return unicode.IsSpace(r) })
	}
	return strings.Join(lines, "\n")
}

// Reads the file contents and returns them as a string. Panics on error.
func readFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Bail out on error. Use for unexpected conditions.
func must(err error) {
	if err != nil {
		panic(err)
	}
}
