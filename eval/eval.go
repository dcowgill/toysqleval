package eval

import (
	"github.com/dcowgill/toysqleval/ast"
	"github.com/dcowgill/toysqleval/token"
)

// EvalStmt evaluates a statement and returns the result.
func EvalStmt(env *Environment, stmt ast.Node) (table *Table, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = errorf(stmt, "%s", r)
			}
		}
	}()
	switch stmt := stmt.(type) {
	case *ast.CreateTableStmt:
		evalCreateTableStmt(env, stmt)
		return
	case *ast.SelectStmt:
		table = evalSelectStmt(env, stmt)
		return
	case *ast.InsertStmt:
		evalInsertStmt(env, stmt)
		return
	case *ast.UpdateStmt:
		evalUpdateStmt(env, stmt)
		return
	case *ast.DeleteStmt:
		evalDeleteStmt(env, stmt)
		return
	}
	return nil, errorf(stmt, "cannot evaluate non-statement %t", stmt)
}

// Evaluates a create table statement.
func evalCreateTableStmt(env *Environment, stmt *ast.CreateTableStmt) {
	table := &Table{Name: stmt.Table.Name}
	for _, col := range stmt.Columns {
		table.Columns = append(table.Columns, &Column{
			Name:     col.Name.Name,
			Type:     dataTypeFromToken(col, col.Type),
			Nullable: col.Nullable,
		})
	}
	if err := env.CreateTable(table); err != nil {
		panic(err)
	}
}

// Evaluates a select statement.
func evalSelectStmt(env *Environment, stmt *ast.SelectStmt) *Table {
	tableExpr, ok := stmt.Table.(*ast.Ident)
	if !ok {
		panic(errorf(stmt, "table subexpressions not supported"))
	}
	table := env.lookupTable(tableExpr.Name)

	// Expand "select *".
	projection := make([]ast.Expr, 0, len(stmt.Columns))
	for _, expr := range stmt.Columns {
		if _, ok := expr.(*ast.SelectStarExpr); ok {
			for _, col := range table.Columns {
				projection = append(projection, &ast.Ident{NamePos: expr.Pos(), Name: col.Name})
			}
		} else {
			projection = append(projection, expr)
		}
	}

	// Different code path for selects with aggregate functions.
	for _, expr := range projection {
		if containsAggFunc(expr) {
			return evalAggregateSelectStmt(env, stmt, table, projection)
		}
	}

	// Generate the result set: select first, then project.
	var results []Row
	for _, row := range table.Data {
		ns := &currentRow{table, row}
		if stmt.Where != nil {
			if !bool(evalExpr(ns, stmt.Where).toBoolean()) {
				continue // row does not match
			}
		}
		result := make(Row, len(projection))
		for i, expr := range projection {
			result[i] = evalExpr(ns, expr)
		}
		results = append(results, result)
	}

	// Create column names for the result set.
	meta := make([]*Column, len(projection))
	for i, expr := range projection {
		name := "?"
		if ident, ok := expr.(*ast.Ident); ok {
			name = ident.Name
		}
		meta[i] = &Column{Name: name}
	}

	return &Table{Columns: meta, Data: results}
}

// Evaluates a select statement whose projection includes one or more aggregate
// functions, but does not have a GROUP BY clause.
func evalAggregateSelectStmt(env *Environment, stmt *ast.SelectStmt, table *Table, projection []ast.Expr) *Table {
	// Rewrite each projection, accumulating aggFuncs.
	var rewriter aggFuncRewriter
	for i, expr := range projection {
		validateAggExpr(expr)
		projection[i] = rewriter.rewrite(expr)
	}
	// Pass every row that satisfies the where clause to every aggregate.
	matched := false
	for _, row := range table.Data {
		ns := &currentRow{table, row}
		if stmt.Where != nil {
			if !bool(evalExpr(ns, stmt.Where).toBoolean()) {
				continue
			}
		}
		for _, fn := range rewriter.funcs {
			fn.step(ns)
		}
		matched = true
	}
	// Build the result row using the output of the aggregate functions. If we
	// did not match a single row, however, return an empty result set.
	var data []Row
	if matched {
		result := make(Row, len(projection))
		for i, expr := range projection {
			result[i] = evalExpr(emptyNamespace{}, expr)
		}
		data = []Row{result}
	}
	// Create the table metadata and return the result.
	meta := make([]*Column, len(projection))
	for i := range projection {
		meta[i] = &Column{Name: "?"}
	}
	return &Table{Columns: meta, Data: data}
}

// Evaluates an insert statement.
func evalInsertStmt(env *Environment, stmt *ast.InsertStmt) {
	table := env.lookupTable(stmt.Table.Name)
	names := make([]string, len(stmt.Columns))
	for i, col := range stmt.Columns {
		names[i] = col.Name
	}
	values := make([]Value, len(stmt.Values))
	for i, expr := range stmt.Values {
		values[i] = evalExpr(emptyNamespace{}, expr) // no symbol table here
	}
	table.insert(names, values)
}

// Evaluates an update statement.
func evalUpdateStmt(env *Environment, stmt *ast.UpdateStmt) {
	table := env.lookupTable(stmt.Table.Name)
	for _, row := range table.Data {
		// First step: select.
		ns := &currentRow{table, row}
		if stmt.Where != nil {
			if !bool(evalExpr(ns, stmt.Where).toBoolean()) {
				continue
			}
		}
		// Second step: apply the SET clause.
		for i, name := range stmt.Columns {
			n := table.colIndex(name.Name)
			if n < 0 {
				panic(errorf(name, "column %q of relation %q does not exist", name.Name, table.Name))
			}
			row[n] = evalExpr(ns, stmt.Values[i])
		}
	}
}

// Evaluates a delete statement.
func evalDeleteStmt(env *Environment, stmt *ast.DeleteStmt) {
	table := env.lookupTable(stmt.Table.Name)
	var newData []Row
	for _, row := range table.Data {
		ns := &currentRow{table, row}
		if stmt.Where != nil {
			if !bool(evalExpr(ns, stmt.Where).toBoolean()) {
				newData = append(newData, row)
			}
		}
	}
	table.Data = newData
}

// Evaluates an expression.
func evalExpr(ns namespace, expr ast.Expr) Value {
	switch expr := expr.(type) {
	case *ast.Ident:
		return ns.lookup(expr.Name)
	case *ast.IntegerLiteral:
		return IntegerValue(expr.Value)
	case *ast.NumberLiteral:
		v := NumberValue(expr.Value)
		return v
	case *ast.StringLiteral:
		return StringValue(expr.Value)
	case *ast.BooleanLiteral:
		return BooleanValue(expr.Value)
	case *ast.Null:
		return nil
	case *ast.BinaryExpr:
		return evalBinaryExpr(ns, expr)
	case *ast.UnaryExpr:
		return evalUnaryExpr(ns, expr)
	case *ast.FunctionCall:
		panic(errorf(expr, "non-aggregate functions are not implemented"))
	case aggFunc:
		return expr.finalize()
	}
	panic(errorf(expr, "cannot evaluate expression of type %T", expr))
}

// Evaluates a binary expression.
func evalBinaryExpr(ns namespace, expr *ast.BinaryExpr) Value {
	lhs := evalExpr(ns, expr.Lhs)
	rhs := evalExpr(ns, expr.Rhs)
	switch expr.Op {
	case token.And, token.Or:
		return BooleanValue(logicalBooleanOp(expr, lhs, rhs))
	case token.Equal, token.GreaterThan, token.GreaterThanOrEqualTo,
		token.LessThan, token.LessThanOrEqualTo, token.NotEqual:
		return BooleanValue(comparisonOp(expr, lhs, rhs))
	case token.Plus, token.Minus, token.Mul, token.Div:
		return arithmeticOp(expr, lhs, rhs)
	case token.Concat:
		return concatOp(expr, lhs, rhs)
	}
	panic(errorf(expr, "invalid binary operator: %s", expr.Op))
}

// Computes (lhs && rhs) or (lhs || rhs), depending on op.
func logicalBooleanOp(expr *ast.BinaryExpr, lhs, rhs Value) bool {
	x := bool(lhs.toBoolean())
	y := bool(rhs.toBoolean())
	switch expr.Op {
	case token.And:
		return x && y
	case token.Or:
		return x || y
	}
	panic(errorf(expr, "invalid logical boolean op: %s", expr.Op))
}

// Evaluates a unary expression.
func evalUnaryExpr(ns namespace, expr *ast.UnaryExpr) Value {
	value := evalExpr(ns, expr.Expr)
	switch expr.Op {
	case token.Plus, token.Minus:
		return unaryArithOp(expr, value)
	}
	panic(errorf(expr, "invalid unary operator: %s", expr.Op))
}
