package eval

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dcowgill/toysqleval/ast"
	"github.com/dcowgill/toysqleval/token"
)

// Supported timestamp formats.
var timeLayouts = []string{
	"2006-01-02 15:04:05 MST",
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	time.UnixDate,
	time.RubyDate,
}

// Value is a wannabe sum type that represents a SQL value.
//
// N.B. methods on Value are not null-safe; that is, it is the responsibility of
// the caller to handle null values.
type Value interface {
	fmt.Stringer

	// Type cast operators. They panic on failure.
	toBoolean() BooleanValue
	toInteger() IntegerValue
	toNumber() NumberValue
	toString() StringValue
	toTimestamp() TimestampValue
}

type BooleanValue bool

func (v BooleanValue) toBoolean() BooleanValue     { return v }
func (v BooleanValue) toInteger() IntegerValue     { panic("cannot convert Boolean to Integer") }
func (v BooleanValue) toNumber() NumberValue       { panic("cannot convert Boolean to Number") }
func (v BooleanValue) toString() StringValue       { panic("cannot convert Boolean to String") }
func (v BooleanValue) toTimestamp() TimestampValue { panic("cannot convert Boolean to Timestamp") }

func (v BooleanValue) toInt() int64 {
	if v {
		return 1
	}
	return 0
}

func (v BooleanValue) String() string {
	if v {
		return "true"
	}
	return "false"
}

type IntegerValue int64

func (v IntegerValue) toBoolean() BooleanValue     { panic("cannot convert Integer to Boolean") }
func (v IntegerValue) toInteger() IntegerValue     { return v }
func (v IntegerValue) toNumber() NumberValue       { return NumberValue(float64(v)) }
func (v IntegerValue) toString() StringValue       { return StringValue(v.String()) }
func (v IntegerValue) toTimestamp() TimestampValue { panic("cannot convert Integer to Timestamp") }

func (v IntegerValue) String() string {
	return fmt.Sprintf("%d", int64(v))
}

type NumberValue float64

func (v NumberValue) toBoolean() BooleanValue     { panic("cannot convert Number to Boolean") }
func (v NumberValue) toInteger() IntegerValue     { return IntegerValue(int64(v)) }
func (v NumberValue) toNumber() NumberValue       { return v }
func (v NumberValue) toString() StringValue       { return StringValue(v.String()) }
func (v NumberValue) toTimestamp() TimestampValue { panic("cannot convert Number to Timestamp") }

func (v NumberValue) String() string {
	return fmt.Sprintf("%v", float64(v))
}

type StringValue string

func (v StringValue) toBoolean() BooleanValue { panic("cannot convert String to Boolean") }
func (v StringValue) toInteger() IntegerValue {
	n, err := strconv.ParseInt(string(v), 10, 64)
	if err != nil {
		panic(err)
	}
	return IntegerValue(n)
}
func (v StringValue) toNumber() NumberValue {
	n, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		panic(err)
	}
	return NumberValue(n)
}
func (v StringValue) toString() StringValue { return v }
func (v StringValue) toTimestamp() TimestampValue {
	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, string(v)); err == nil {
			return TimestampValue(t)
		}
	}
	panic(fmt.Sprintf("invalid timestamp: %s", v))
}

func (v StringValue) String() string {
	return fmt.Sprintf("%q", string(v))
}

type TimestampValue time.Time

func (v TimestampValue) toBoolean() BooleanValue     { panic("cannot convert Timestamp to Boolean") }
func (v TimestampValue) toInteger() IntegerValue     { panic("cannot convert Timestamp to Integer") }
func (v TimestampValue) toNumber() NumberValue       { panic("cannot convert Timestamp to Number") }
func (v TimestampValue) toString() StringValue       { return StringValue(v.String()) }
func (v TimestampValue) toTimestamp() TimestampValue { return v }

func (v TimestampValue) String() string {
	return time.Time(v).Format(time.RFC3339)
}

func coerce(v Value, t DataType) Value {
	switch t {
	case Boolean:
		return v.toBoolean()
	case Integer:
		return v.toInteger()
	case Number:
		return v.toNumber()
	case String:
		return v.toString()
	case Timestamp:
		return v.toTimestamp()
	}
	panic(fmt.Errorf("invalid data type: %s", t))
}

func comparisonOp(expr *ast.BinaryExpr, lhs, rhs Value) bool {
	// Any comparison involving null evaluates to false.
	if lhs == nil || rhs == nil {
		return false
	}
	// Double dispatch by type.
	switch lhs := lhs.(type) {
	case BooleanValue:
		switch rhs := rhs.(type) {
		case BooleanValue:
			return cmpInts(expr, lhs.toInt(), rhs.toInt())
		}
	case IntegerValue:
		switch rhs := rhs.(type) {
		case IntegerValue:
			return cmpInts(expr, int64(lhs), int64(rhs))
		case NumberValue:
			return cmpFloats(expr, float64(lhs), float64(rhs))
		case StringValue:
			val := rhs.toInteger()
			return cmpInts(expr, int64(lhs), int64(val))
		}
	case NumberValue:
		switch rhs := rhs.(type) {
		case IntegerValue:
			return cmpFloats(expr, float64(lhs), float64(rhs))
		case NumberValue:
			return cmpFloats(expr, float64(lhs), float64(rhs))
		case StringValue:
			val := rhs.toNumber()
			return cmpFloats(expr, float64(lhs), float64(val))
		}
	case StringValue:
		switch rhs := rhs.(type) {
		case IntegerValue:
			val := rhs.toInteger()
			return cmpInts(expr, int64(val), int64(rhs))
		case NumberValue:
			val := rhs.toNumber()
			return cmpFloats(expr, float64(val), float64(rhs))
		case StringValue:
			return cmpStrings(expr, string(lhs), string(rhs))
		case TimestampValue:
			val := lhs.toTimestamp()
			return cmpTimes(expr, time.Time(val), time.Time(rhs))
		}
	case TimestampValue:
		switch rhs := rhs.(type) {
		case StringValue:
			val := rhs.toTimestamp()
			return cmpTimes(expr, time.Time(lhs), time.Time(val))
		case TimestampValue:
			return cmpTimes(expr, time.Time(lhs), time.Time(rhs))
		}
	}
	panic(errorf(expr, "invalid comparison: %s %s %s", lhs, expr.Op, rhs))
}

func cmpInts(expr *ast.BinaryExpr, a, b int64) bool {
	switch expr.Op {
	case token.Equal:
		return a == b
	case token.NotEqual:
		return a != b
	case token.LessThan:
		return a < b
	case token.LessThanOrEqualTo:
		return a <= b
	case token.GreaterThan:
		return a > b
	case token.GreaterThanOrEqualTo:
		return a >= b
	}
	panic(errorf(expr, "invalid comparison operator: %s", expr.Op))
}

func cmpFloats(expr *ast.BinaryExpr, a, b float64) bool {
	switch expr.Op {
	case token.Equal:
		return a == b
	case token.NotEqual:
		return a != b
	case token.LessThan:
		return a < b
	case token.LessThanOrEqualTo:
		return a <= b
	case token.GreaterThan:
		return a > b
	case token.GreaterThanOrEqualTo:
		return a >= b
	}
	panic(errorf(expr, "invalid comparison operator: %s", expr.Op))
}

func cmpStrings(expr *ast.BinaryExpr, a, b string) bool {
	switch expr.Op {
	case token.Equal:
		return a == b
	case token.NotEqual:
		return a != b
	case token.LessThan:
		return a < b
	case token.LessThanOrEqualTo:
		return a <= b
	case token.GreaterThan:
		return a > b
	case token.GreaterThanOrEqualTo:
		return a >= b
	}
	panic(errorf(expr, "invalid comparison operator: %s", expr.Op))
}

func cmpTimes(expr *ast.BinaryExpr, a, b time.Time) bool {
	switch expr.Op {
	case token.Equal:
		return a.Equal(b)
	case token.NotEqual:
		return !a.Equal(b)
	case token.LessThan:
		return a.Before(b)
	case token.LessThanOrEqualTo:
		return a.Before(b) || a.Equal(b)
	case token.GreaterThan:
		return a.After(b)
	case token.GreaterThanOrEqualTo:
		return a.After(b) || a.Equal(b)
	}
	panic(errorf(expr, "invalid comparison operator: %s", expr.Op))
}

func arithmeticOp(expr *ast.BinaryExpr, lhs, rhs Value) Value {
	// Any arithmetic involving null evaluates to null.
	if lhs == nil || rhs == nil {
		return nil
	}
	// Do a rudimentary numeric tower thing.
	switch lhs := lhs.(type) {
	case IntegerValue:
		switch rhs := rhs.(type) {
		case IntegerValue:
			return arithOpInt(expr, lhs, rhs)
		case NumberValue:
			return arithOpNum(expr, lhs.toNumber(), rhs)
		case StringValue:
			return arithOpInt(expr, lhs, rhs.toInteger())
		}
	case NumberValue:
		switch rhs := rhs.(type) {
		case IntegerValue:
			return arithOpNum(expr, lhs, rhs.toNumber())
		case NumberValue:
			return arithOpNum(expr, lhs, rhs)
		case StringValue:
			return arithOpNum(expr, lhs, rhs.toNumber())
		}
	case StringValue:
		switch rhs := rhs.(type) {
		case IntegerValue:
			return arithOpInt(expr, lhs.toInteger(), rhs)
		case NumberValue:
			return arithOpNum(expr, lhs.toNumber(), rhs)
		}
	}
	// If we got here, these values cannot do arithmetic together.
	panic(errorf(expr, "invalid arithmetic expression: %s %s %s", lhs, expr.Op, rhs))
}

// Does integer arithmetic.
func arithOpInt(expr *ast.BinaryExpr, lhs, rhs IntegerValue) Value {
	x := int64(lhs)
	y := int64(rhs)
	var r int64
	switch expr.Op {
	case token.Plus:
		r = x + y
	case token.Minus:
		r = x - y
	case token.Mul:
		r = x * y
	case token.Div:
		if y == 0 {
			panic(errorf(expr, "divide by zero"))
		}
		r = x / y
	}
	return IntegerValue(r)
}

// Does floating-point arithmetic.
func arithOpNum(expr *ast.BinaryExpr, lhs, rhs NumberValue) Value {
	x := float64(lhs)
	y := float64(rhs)
	var r float64
	switch expr.Op {
	case token.Plus:
		r = x + y
	case token.Minus:
		r = x - y
	case token.Mul:
		r = x * y
	case token.Div:
		if y == 0 {
			panic(errorf(expr, "divide by zero"))
		}
		r = x / y
	}
	return NumberValue(r)
}

func unaryArithOp(expr *ast.UnaryExpr, value Value) Value {
	if value == nil {
		return nil
	}
	switch expr.Op {
	case token.Plus:
		switch value := value.(type) {
		case IntegerValue, NumberValue:
			return value
		}
	case token.Minus:
		switch value := value.(type) {
		case IntegerValue:
			return IntegerValue(-int64(value))
		case NumberValue:
			return NumberValue(-float64(value))
		}
	default:
		panic(errorf(expr, "invalid unary operator: %s", expr.Op))
	}
	panic(errorf(expr, "invalid value for unary %s: %s", expr.Op, value))
}

// Computes lhs || rhs. We're more permissive than the typical database; any two
// values that can both be coerced to strings may be concatenated.
func concatOp(expr *ast.BinaryExpr, lhs, rhs Value) Value {
	// Any concatenation involving null evaluates to null.
	if lhs == nil || rhs == nil {
		return nil
	}
	return StringValue(string(lhs.toString()) + string(rhs.toString()))
}
