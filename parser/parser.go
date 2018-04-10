package parser

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/dcowgill/toysqleval/ast"
	"github.com/dcowgill/toysqleval/lexer"
	"github.com/dcowgill/toysqleval/token"
)

// TODO: tuple expressions, as in (1,2,3)
// TODO: select from parenthesized table subexpression
// TODO: IS and NOT, as in "WHERE x IS NOT NULL"

// Maintains the parser state.
type parser struct {
	lex *lexer.Lexer
}

// Shorthands for accessing the current lexical token.
func (p *parser) tok() token.Token { return p.lex.Token() }
func (p *parser) kind() token.Kind { return p.tok().Kind }
func (p *parser) pos() token.Pos   { return p.tok().Pos }

func Parse(lex *lexer.Lexer) (nodes []ast.Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e2, ok := r.(error); ok {
				err = e2
				return
			}
			panic(r)
		}
	}()
	if !lex.Scan() {
		return nil, nil // no input
	}
	if err := lex.Err(); err != nil {
		return nil, err // lexer error
	}
	p := parser{lex: lex}
	return p.parseStmtList(), nil
}

// Parses a semicolon-delimited list of statements.
func (p *parser) parseStmtList() []ast.Node {
	var stmts []ast.Node
	for p.kind() != token.Invalid {
		stmts = append(stmts, p.parseStmt())
		p.match(token.Semicolon)
	}
	return stmts
}

func (p *parser) parseStmt() ast.Node {
	switch p.kind() {
	case token.Create:
		return p.parseCreateTableStmt()
	case token.Select:
		return p.parseSelectStmt()
	case token.Insert:
		return p.parseInsertStmt()
	case token.Update:
		return p.parseUpdateStmt()
	case token.Delete:
		return p.parseDeleteStmt()
	}
	p.expected(token.Select, token.Insert, token.Update, token.Delete)
	return nil // not reached
}

// Parses a create table statement.
func (p *parser) parseCreateTableStmt() *ast.CreateTableStmt {
	start := p.match(token.Create)
	p.match(token.Table)
	table := p.parseIdent()
	columns := p.parseColumnDefinitions()
	return &ast.CreateTableStmt{StartPos: start.Pos, Table: table, Columns: columns}
}

// Parses a column definition in a create table statement.
func (p *parser) parseColumnDefinitions() []*ast.ColumnDefinition {
	var columns []*ast.ColumnDefinition
	p.match(token.LeftParen)
	for {
		name := p.parseIdent()

		var dataType token.Kind
		switch p.kind() {
		case token.Boolean, token.Integer, token.Number, token.Varchar, token.Timestamp:
			tok := p.next()
			dataType = tok.Kind
		default:
			p.expected(token.Boolean, token.Integer, token.Number, token.Varchar, token.Timestamp)
		}

		nullable := true
		if p.kind() == token.Not {
			p.skip(token.Not)
			p.match(token.Null)
			nullable = false
		} else if p.kind() == token.Null {
			p.skip(token.Null)
		}

		columns = append(columns, &ast.ColumnDefinition{
			Name:     name,
			Type:     dataType,
			Nullable: nullable,
		})
		if p.kind() != token.Comma {
			break
		}
		p.skip(token.Comma)
	}
	p.match(token.RightParen)
	return columns
}

// Parses a select statement.
func (p *parser) parseSelectStmt() *ast.SelectStmt {
	start := p.match(token.Select)
	columns := p.parseSelectExprList()
	p.match(token.From)
	table := p.parseIdent()
	var where ast.Expr
	if p.kind() == token.Where {
		p.skip(token.Where)
		where = p.parseExpr()
	}
	return &ast.SelectStmt{StartPos: start.Pos, Columns: columns, Table: table, Where: where}
}

// Parses an insert statement.
func (p *parser) parseInsertStmt() *ast.InsertStmt {
	start := p.match(token.Insert)
	p.match(token.Into)
	table := p.parseIdent()
	var columns []*ast.Ident
	if p.kind() == token.LeftParen {
		p.skip(token.LeftParen)
		columns = append(columns, p.parseIdent())
		for p.kind() == token.Comma {
			p.skip(token.Comma)
			columns = append(columns, p.parseIdent())
		}
		p.match(token.RightParen)
	}
	p.match(token.Values)
	p.match(token.LeftParen)
	values := p.parseExprList()
	p.match(token.RightParen)
	return &ast.InsertStmt{StartPos: start.Pos, Table: table, Columns: columns, Values: values}
}

// Parses an update statement.
func (p *parser) parseUpdateStmt() *ast.UpdateStmt {
	start := p.match(token.Update)
	table := p.parseIdent()
	p.match(token.Set)
	// N.B. the homework example has parentheses around the SET clause, which
	// AFAICT is not valid SQL, but since it's in the example, we'll support it.
	closeParen := false
	if p.kind() == token.LeftParen {
		p.skip(token.LeftParen)
		closeParen = true
	}
	var (
		columns []*ast.Ident
		values  []ast.Expr
	)
	columns = append(columns, p.parseIdent())
	p.match(token.Equal)
	values = append(values, p.parseExpr())
	for p.kind() == token.Comma {
		p.skip(token.Comma)
		columns = append(columns, p.parseIdent())
		p.match(token.Equal)
		values = append(values, p.parseExpr())
	}
	// See N.B. above.
	if closeParen {
		p.match(token.RightParen)
	}
	var where ast.Expr
	if p.kind() == token.Where {
		p.skip(token.Where)
		where = p.parseExpr()
	}
	return &ast.UpdateStmt{StartPos: start.Pos, Table: table, Columns: columns, Values: values, Where: where}
}

// Parses a delete statement.
func (p *parser) parseDeleteStmt() *ast.DeleteStmt {
	start := p.match(token.Delete)
	p.match(token.From)
	table := p.parseIdent()
	var where ast.Expr
	if p.kind() == token.Where {
		p.skip(token.Where)
		where = p.parseExpr()
	}
	return &ast.DeleteStmt{StartPos: start.Pos, Table: table, Where: where}
}

// Parses a comma-separated list of expressions in the projection of a SELECT
// statement. The difference between this and parseExprList is that
// parseSelectExprList interprets the "*" token to mean all column names.
func (p *parser) parseSelectExprList() []ast.Expr {
	exprs := []ast.Expr{p.parseExprOrStar()}
	for p.kind() == token.Comma {
		p.skip(token.Comma)
		exprs = append(exprs, p.parseExprOrStar())
	}
	return exprs
}

// Parses a single expression or a "*" operator.
func (p *parser) parseExprOrStar() ast.Expr {
	if p.kind() == token.Mul {
		p.skip(token.Mul)
		return &ast.SelectStarExpr{}
	}
	return p.parseBinaryExpr(1)
}

// Parses a comma-separated list of expressions.
func (p *parser) parseExprList() []ast.Expr {
	exprs := []ast.Expr{p.parseExpr()}
	for p.kind() == token.Comma {
		p.skip(token.Comma)
		exprs = append(exprs, p.parseExpr())
	}
	return exprs
}

// Parses a single expression.
func (p *parser) parseExpr() ast.Expr {
	return p.parseBinaryExpr(1)
}

// Parses an identifier expression.
func (p *parser) parseIdent() *ast.Ident {
	tok := p.match(token.Ident)
	return &ast.Ident{NamePos: tok.Pos, Name: tok.Lit}
}

// Parses a binary expression.
func (p *parser) parseBinaryExpr(minPrec int) ast.Expr {
	expr := p.parseUnaryExpr()
	for {
		op := p.kind()
		opPrec := op.Precedence()
		if opPrec < minPrec {
			return expr
		}
		p.skip(op)
		rhs := p.parseBinaryExpr(opPrec + 1)
		expr = &ast.BinaryExpr{Lhs: expr, Op: op, Rhs: rhs}
	}
}

// Parses a unary expression.
func (p *parser) parseUnaryExpr() ast.Expr {
	switch p.kind() {
	case token.LeftParen:
		p.skip(token.LeftParen)
		e := p.parseExpr()
		p.match(token.RightParen)
		return e
	case token.Ident:
		tok := p.next()
		if p.kind() == token.LeftParen {
			p.skip(token.LeftParen)
			funcName := &ast.Ident{NamePos: tok.Pos, Name: tok.Lit}
			funcArgs := p.parseSelectExprList()
			p.match(token.RightParen)
			return &ast.FunctionCall{Name: funcName, Args: funcArgs}
		}
		return &ast.Ident{NamePos: tok.Pos, Name: tok.Lit}
	case token.Plus, token.Minus:
		tok := p.next()
		return &ast.UnaryExpr{StartPos: tok.Pos, Op: tok.Kind, Expr: p.parseExpr()}
	case token.NumberLiteral:
		return p.parseNumberLiteral()
	case token.StringLiteral:
		tok := p.next()
		return &ast.StringLiteral{ValuePos: tok.Pos, Value: tok.Lit}
	case token.True:
		tok := p.next()
		return &ast.BooleanLiteral{ValuePos: tok.Pos, Value: true}
	case token.False:
		tok := p.next()
		return &ast.BooleanLiteral{ValuePos: tok.Pos, Value: false}
	case token.Null:
		tok := p.next()
		return &ast.Null{ValuePos: tok.Pos}
	}
	p.expected(token.LeftParen, token.Ident, token.Null, token.NumberLiteral, token.StringLiteral)
	return nil // can't get here
}

// Parses a number from a string, as either an int64 or a float64.
func (p *parser) parseNumberLiteral() ast.Expr {
	tok := p.match(token.NumberLiteral)
	for _, ch := range tok.Lit {
		if !unicode.IsDigit(ch) {
			n, _ := strconv.ParseFloat(tok.Lit, 64)
			return &ast.NumberLiteral{ValuePos: tok.Pos, Value: n}
		}
	}
	n, _ := strconv.ParseInt(tok.Lit, 10, 64)
	return &ast.IntegerLiteral{ValuePos: tok.Pos, Value: n}
}

// Advances past the specified token. It is a runtime error to call this method
// when the current token does *not* have the specified kind.
func (p *parser) skip(k token.Kind) {
	if p.kind() != k {
		p.errorf("tried to skip(%s) but token is %s", k, p.kind())
	}
	p.next()
}

// Tries to advances past a token of the specified kind.
func (p *parser) match(k token.Kind) token.Token {
	if p.kind() != k {
		p.expected(k)
	}
	return p.next()
}

// Raises a parse error indicating that one of the specified kinds of tokens was
// expected, but the current token does not match.
func (p *parser) expected(kinds ...token.Kind) {
	p.errorf("current token is %q, want one of %+v", p.kind().String(), kinds)
}

// Moves to the next token in the stream.
func (p *parser) next() token.Token {
	tok := p.lex.Token()
	p.lex.Scan()
	if err := p.lex.Err(); err != nil {
		panic(err)
	}
	return tok
}

// Panics with an error at the current position.
func (p *parser) errorf(format string, args ...interface{}) {
	panic(&Error{Msg: fmt.Sprintf(format, args...), Pos: p.pos()})
}
