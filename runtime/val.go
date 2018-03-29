package runtime

import (
	"context"
	"fmt"
)

// The TypeError is raised if an invalid type is used for a specific action.
type TypeError string

// Error interface implementation.
func (te TypeError) Error() string {
	return string(te)
}

// Create a new TypeError.
func NewTypeError(t1, t2, op string) TypeError {
	if t2 != "" {
		return TypeError(fmt.Sprintf("type error: %s not allowed with types %s and %s", op, t1, t2))
	}
	return TypeError(fmt.Sprintf("type error: %s not allowed with type %s", op, t1))
}

// Converter declares the required methods to convert a value
// to any one of the supported types (except Object and Func).
type Converter interface {
	Int(context.Context) int64
	Float(context.Context) float64
	String(context.Context) string
	Bool(context.Context) bool
	Native(context.Context) interface{}
}

// Arithmetic defines the methods required to compute all
// the supported arithmetic operations.
type Arithmetic interface {
	Add(context.Context, Val, Val) Val
	Sub(context.Context, Val, Val) Val
	Mul(context.Context, Val, Val) Val
	Div(context.Context, Val, Val) Val
	Mod(context.Context, Val, Val) Val
	Unm(context.Context, Val) Val
}

// The default, standard agora arithmetic implementation.
type defaultArithmetic struct{}

func (ar defaultArithmetic) binaryOp(ctx context.Context, l, r Val, op string, allowStrings bool) Val {
	lt, rt := Type(l), Type(r)
	mm := "__" + op
	if lt == "number" && rt == "number" {
		// Two numbers, standard arithmetic operation
		switch op {
		case "add":
			return Number(l.Float(ctx) + r.Float(ctx))
		case "sub":
			return Number(l.Float(ctx) - r.Float(ctx))
		case "mul":
			return Number(l.Float(ctx) * r.Float(ctx))
		case "div":
			return Number(l.Float(ctx) / r.Float(ctx))
		case "mod":
			return Number(l.Int(ctx) % r.Int(ctx))
		}
	} else if allowStrings && lt == "string" && rt == "string" {
		// Two strings
		switch op {
		case "add":
			return String(l.String(ctx) + r.String(ctx))
		}
	} else if lt == "object" {
		// If left operand is an object with a meta-method
		lo := l.(Object)
		if v, ok := lo.callMetaMethod(ctx, mm, r, Bool(true)); ok {
			return v
		}
	}
	// Last chance: if right operand is an object with a meta-method
	if rt == "object" {
		ro := r.(Object)
		if v, ok := ro.callMetaMethod(ctx, mm, l, Bool(false)); ok {
			return v
		}
	}
	panic(NewTypeError(lt, rt, op))
}

func (ar defaultArithmetic) Add(ctx context.Context, l, r Val) Val {
	return ar.binaryOp(ctx, l, r, "add", true)
}

func (ar defaultArithmetic) Sub(ctx context.Context, l, r Val) Val {
	return ar.binaryOp(ctx, l, r, "sub", false)
}

func (ar defaultArithmetic) Mul(ctx context.Context, l, r Val) Val {
	return ar.binaryOp(ctx, l, r, "mul", false)
}

func (ar defaultArithmetic) Div(ctx context.Context, l, r Val) Val {
	return ar.binaryOp(ctx, l, r, "div", false)
}

func (ar defaultArithmetic) Mod(ctx context.Context, l, r Val) Val {
	return ar.binaryOp(ctx, l, r, "mod", false)
}

func (ar defaultArithmetic) Unm(ctx context.Context, l Val) Val {
	lt := Type(l)
	if lt == "number" {
		return Number(-l.Float(ctx))
	} else if lt == "object" {
		lo := l.(Object)
		if v, ok := lo.callMetaMethod(ctx, "__unm"); ok {
			return v
		}
	}
	panic(NewTypeError(lt, "", "unm"))
}

// Comparer defines the method required to compare two Values.
// Cmp() returns 1 if the first value is greater, 0 if
// it is equal, and -1 if it is lower.
type Comparer interface {
	Cmp(context.Context, Val, Val) int
}

var (
	// Unmutable, this would be a const if it was possible
	uneqMatrix = map[string]map[string]int{
		"nil": map[string]int{
			"number": -1,
			"string": -1,
			"bool":   -1,
			"object": -1,
			"func":   -1,
			"custom": -1,
		},
		"number": map[string]int{
			"nil":    1,
			"string": -1,
			"bool":   1,
			"object": -1,
			"func":   1,
			"custom": 1,
		},
		"string": map[string]int{
			"number": 1,
			"nil":    1,
			"bool":   1,
			"object": 1,
			"func":   1,
			"custom": 1,
		},
		"bool": map[string]int{
			"number": -1,
			"string": -1,
			"nil":    1,
			"object": -1,
			"func":   -1,
			"custom": -1,
		},
		"object": map[string]int{
			"number": 1,
			"string": -1,
			"bool":   1,
			"nil":    1,
			"func":   1,
			"custom": 1,
		},
		"func": map[string]int{
			"number": -1,
			"string": -1,
			"bool":   1,
			"object": -1,
			"nil":    1,
			"custom": 1,
		},
		"custom": map[string]int{
			"number": -1,
			"string": -1,
			"bool":   1,
			"object": -1,
			"func":   -1,
			"nil":    1,
		},
	}
)

// The default, standard agora comparer implementation.
type defaultComparer struct{}

func (dc defaultComparer) Cmp(ctx context.Context, l, r Val) int {
	lt, rt := Type(l), Type(r)
	if lt == rt {
		// Comparable types
		switch lt {
		case "nil":
			return 0
		case "number":
			lf, rf := l.Float(ctx), r.Float(ctx)
			if lf == rf {
				return 0
			} else if lf < rf {
				return -1
			} else {
				return 1
			}
		case "string":
			ls, rs := l.String(ctx), r.String(ctx)
			if ls == rs {
				return 0
			} else if ls < rs {
				return -1
			} else {
				return 1
			}
		case "bool":
			lb, rb := l.Bool(ctx), r.Bool(ctx)
			if lb == rb {
				return 0
			} else if lb {
				return 1 // true (1) is greater than false (0)
			} else {
				return -1
			}
		case "func":
			lf, rf := l.Native(ctx), r.Native(ctx)
			if lf == rf {
				return 0
			} else {
				// "greater" or "lower" has no sense for funcs, return -1
				return -1
			}
		case "object":
			// If left has meta method, use left, otherwise right, else compare
			lo, ro := l.(Object), r.(Object)
			if v, ok := lo.callMetaMethod(ctx, "__cmp", r, Bool(true)); ok {
				return int(v.Int(ctx))
			}
			if v, ok := ro.callMetaMethod(ctx, "__cmp", l, Bool(false)); ok {
				return int(v.Int(ctx))
			}
			if lo == ro {
				return 0
			} else {
				// "greater" or "lower" has no sense for objects, return -1
				return -1
			}
		case "custom":
			if l == r {
				return 0
			} else {
				// "greater" or "lower" has no sense for custom vals, return -1
				return -1
			}
		default:
			panic(NewTypeError(lt, "", "cmp"))
		}
	} else {
		// Uncomparable types, first check for meta-methods
		var (
			o      Object
			isLeft bool
			otherv Val
		)
		if lt == "object" {
			o = l.(Object)
			isLeft = true
			otherv = r
		} else if rt == "object" {
			o = r.(Object)
			isLeft = false
			otherv = l
		}
		if o != nil {
			if v, ok := o.callMetaMethod(ctx, "__cmp", otherv, Bool(isLeft)); ok {
				return int(v.Int(ctx))
			}
		}
		// Else, return arbitrary but constant result
		return uneqMatrix[lt][rt]
	}
}

// The Dumper interface defines the required behaviour to pretty-print
// the values in debug logs.
type Dumper interface {
	Dump() string
}

// Val is the representation of a value, any value, in the language.
// The supported value types are the following:
// * Number (float64)
// * String
// * Bool (bool)
// * Nil (null)
// * Object (interface)
// * Func (interface)
// * Custom (any other Val impl)
type Val interface {
	Converter
}

// Type returns the type of the value, which can be one of the following strings:
// * string
// * number
// * bool
// * func
// * object
// * nil
// * custom
func Type(v Val) string {
	switch v.(type) {
	case String:
		return "string"
	case Number:
		return "number"
	case Bool:
		return "bool"
	case Func:
		return "func"
	case Object:
		return "object"
	default:
		if v == Nil {
			return "nil"
		} else {
			return "custom"
		}
	}
}

// Helper function to pretty-print a value for debugging purpose.
func dumpVal(v Val) string {
	if dmp, ok := v.(Dumper); ok {
		return dmp.Dump()
	}
	return fmt.Sprintf("%v", v)
}
