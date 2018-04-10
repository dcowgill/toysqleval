package eval

import (
	"math"
	"strings"

	"github.com/dcowgill/toysqleval/ast"
	"github.com/dcowgill/toysqleval/token"
)

// Describes an aggregate SQL function.
type aggFunc interface {
	ast.Expr

	step(ns namespace)
	finalize() Value
}

// Implements the COUNT function.
type countAggFunc struct {
	expr   ast.Expr
	isStar bool // true if COUNT(*)
	count  int
}

func newCountAggFunc(node ast.Node, args []ast.Expr) aggFunc {
	validateArgCount(node, "count", args, 1)
	_, isStar := args[0].(*ast.SelectStarExpr)
	return &countAggFunc{expr: args[0], isStar: isStar}
}

func (fn *countAggFunc) Pos() token.Pos { return fn.expr.Pos() }

func (fn *countAggFunc) step(ns namespace) {
	if fn.isStar || evalExpr(ns, fn.expr) != nil {
		fn.count++
	}
}

func (fn *countAggFunc) finalize() Value {
	return IntegerValue(fn.count)
}

// Implements both the MIN and MAX aggregate functions.
type minMaxAggFunc struct {
	expr    ast.Expr
	fval    float64
	ival    int64
	isFloat bool
	isMin   bool
}

func newMinAggFunc(node ast.Node, args []ast.Expr) aggFunc {
	validateArgCount(node, "min", args, 1)
	return &minMaxAggFunc{expr: args[0], ival: math.MaxInt64, isMin: true}
}

func newMaxAggFunc(node ast.Node, args []ast.Expr) aggFunc {
	validateArgCount(node, "max", args, 1)
	return &minMaxAggFunc{expr: args[0], ival: math.MinInt64}
}

func (fn *minMaxAggFunc) Pos() token.Pos { return fn.expr.Pos() }

func (fn *minMaxAggFunc) setInt(n int64) {
	if fn.isMin {
		if n < fn.ival {
			fn.ival = n
		}
	} else {
		if n > fn.ival {
			fn.ival = n
		}
	}
}

func (fn *minMaxAggFunc) setFloat(n float64) {
	if fn.isMin {
		if fn.fval < n {
			fn.fval = n
		}
	} else {
		if fn.fval > n {
			fn.fval = n
		}
	}
}

func (fn *minMaxAggFunc) step(ns namespace) {
	switch value := evalExpr(ns, fn.expr).(type) {
	case IntegerValue:
		if fn.isFloat {
			fn.setFloat(float64(value))
		} else {
			fn.setInt(int64(value))
		}
	case NumberValue:
		if !fn.isFloat {
			fn.isFloat = true
			fn.fval = float64(fn.ival)
		}
		fn.setFloat(float64(value))
	case nil:
		return
	default:
		panic(errorf(fn.expr, "cannot compute MIN of %s", value))
	}
}

func (fn *minMaxAggFunc) finalize() Value {
	switch {
	case fn.isFloat:
		return NumberValue(fn.fval)
	default:
		return IntegerValue(fn.ival)
	}
}

// Implements the SUM function.
type sumAggFunc struct {
	expr    ast.Expr
	fsum    float64
	isum    int64
	isFloat bool
}

func newSumAggFunc(node ast.Node, args []ast.Expr) aggFunc {
	validateArgCount(node, "sum", args, 1)
	return &sumAggFunc{expr: args[0]}
}

func (fn *sumAggFunc) Pos() token.Pos { return fn.expr.Pos() }

func (fn *sumAggFunc) step(ns namespace) {
	switch value := evalExpr(ns, fn.expr).(type) {
	case IntegerValue:
		if fn.isFloat {
			fn.fsum += float64(value)
		} else {
			fn.isum += int64(value)
		}
	case NumberValue:
		if !fn.isFloat {
			fn.isFloat = true
			fn.fsum += float64(fn.isum)
		}
		fn.fsum += float64(value)
	case nil:
		return
	default:
		panic(errorf(fn.expr, "cannot compute SUM of %s", value))
	}
}

func (fn *sumAggFunc) finalize() Value {
	switch {
	case fn.isFloat:
		return NumberValue(fn.fsum)
	default:
		return IntegerValue(fn.isum)
	}
}

// A function that creates new aggregator functions.
type aggFuncConstructor func(ast.Node, []ast.Expr) aggFunc

var builtinAggFuncs = map[string]aggFuncConstructor{
	"count": newCountAggFunc,
	"min":   newMinAggFunc,
	"max":   newMaxAggFunc,
	"sum":   newSumAggFunc,
}

// Reports whether a name refers to a builtin aggregate func.
func isAggFunc(name string) bool {
	_, found := builtinAggFuncs[name]
	return found
}

// Panics if a function was given the wrong number of arguments.
func validateArgCount(node ast.Node, fname string, args []ast.Expr, expected int) {
	if len(args) != expected {
		panic(errorf(node, "wrong number of arguments to %s: got %d, want %d",
			strings.ToUpper(fname), len(args), expected))
	}
}

// Rewrites an expression, replacing each aggregate function call with an
// aggFunc that will actually compute it; the aggFunc values are also
// accumulated in the rewriter itself. The original AST is not modified.
type aggFuncRewriter struct {
	funcs []aggFunc
}

func (st *aggFuncRewriter) rewrite(expr ast.Expr) ast.Expr {
	// We only need to handle a subset of all nodes that have children, since we
	// know a priori that we are rewriting expressions, not statements.
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{Lhs: st.rewrite(expr.Lhs), Op: expr.Op, Rhs: st.rewrite(expr.Rhs)}
	case *ast.UnaryExpr:
		return &ast.UnaryExpr{StartPos: expr.StartPos, Expr: st.rewrite(expr.Expr), Op: expr.Op}
	case *ast.FunctionCall:
		funcName := expr.Name.Name
		if constructor := builtinAggFuncs[funcName]; constructor != nil {
			newNode := constructor(expr, expr.Args)
			st.funcs = append(st.funcs, newNode)
			return newNode
		} else {
			// For now, other kinds of functions are not supported.
			panic(errorf(expr, "unknown function: %s", funcName))
		}
	}
	return expr // rewrite unnecessary
}

// Recursively searches an AST for an aggregate function.
func containsAggFunc(node ast.Node) bool {
	var fn ast.WalkFunc
	found := false
	fn = func(node ast.Node) ast.WalkFunc {
		if found {
			return nil // prune search
		}
		if node, ok := node.(*ast.FunctionCall); ok {
			if isAggFunc(node.Name.Name) {
				found = true
				return nil // prune search
			}
		}
		return fn
	}
	ast.Walk(node, fn)
	return found
}

// Verifies the following: (1) no argument of an aggregate contains a nested
// call to an aggregate function; (2) no column identifier exists outside of
// an aggregate function. N.B. only call this function on the projections of a
// SELECT that contains one or more aggregate functions.
func validateAggExpr(node ast.Node) {
	var fn ast.WalkFunc
	fn = func(node ast.Node) ast.WalkFunc {
		switch node := node.(type) {
		case *ast.FunctionCall:
			if isAggFunc(node.Name.Name) {
				for _, arg := range node.Args {
					if containsAggFunc(arg) {
						panic(errorf(arg, "aggregate function calls cannot be nested"))
					}
				}
				return nil // prune search
			}
		case *ast.Ident:
			panic(errorf(node, "column %q must appear in the GROUP BY clause "+
				"or be used in an aggregate function", node.Name))
		}
		return fn
	}
	ast.Walk(node, fn)
}
