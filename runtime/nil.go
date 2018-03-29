package runtime

import "context"

const (
	// The string representation of the nil value
	NilString = "nil"
)

var (
	// The one and only Nil instance
	Nil = null{}
)

// Null is the representation of the null type. It is semantically equivalent
// to Go's nil value, but it is represented as an empty struct to implement
// the Val interface so that it is a valid agora value.
type null struct{}

// Dump pretty-prints the value for debugging purpose.
func (n null) Dump() string {
	return "[Nil]"
}

// Int is an invalid conversion.
func (n null) Int(context.Context) int64 {
	panic(NewTypeError(Type(n), "", "int"))
}

// Float is an invalid conversion.
func (n null) Float(context.Context) float64 {
	panic(NewTypeError(Type(n), "", "float"))
}

// String returns the string "nil".
func (n null) String(context.Context) string {
	return NilString
}

// Bool returns false.
func (n null) Bool(context.Context) bool {
	return false
}

// Native returns the Go native representation of the value.
func (n null) Native(context.Context) interface{} {
	return nil
}
