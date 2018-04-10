package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dcowgill/toysqleval/ast"
	"github.com/dcowgill/toysqleval/eval"
	"github.com/dcowgill/toysqleval/lexer"
	"github.com/dcowgill/toysqleval/parser"
	"github.com/dcowgill/toysqleval/pprint"
)

func main() {
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	// Read SQL from stdin.
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	// Lex/parse the SQL.
	lex := lexer.New(string(input))
	stmts, err := parser.Parse(lex)
	if err != nil {
		log.Fatal(err)
	}

	// In verbose mode, pretty-print the statements.
	if *verbose {
		pp := ast.PrettyPrinter{Writer: os.Stdout, Indent: "    "}
		for _, stmt := range stmts {
			pp.Visit(stmt)
			fmt.Println("")
		}
	}

	// Execute all the statements in the same environment.
	var env eval.Environment
	for _, stmt := range stmts {
		result, err := eval.EvalStmt(&env, stmt)
		if err != nil {
			fmt.Println(err.Error())
		}
		if result != nil {
			pprint.Table(os.Stdout, result)
		} else {
			fmt.Println("OK")
		}
	}
}
