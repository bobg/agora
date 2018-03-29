package runtime

import (
	"context"
	"fmt"
	"strconv"
)

// String is the representation of the String type. It is equivalent
// to Go's string type.
type String string

// Pretty-prints the string value.
func (s String) Dump() string {
	return fmt.Sprintf("\"%s\" (String)", string(s))
}

// Int converts the string representation of an integer to an integer value.
// If the string doesn't hold a valid integer representation,
// it panics.
func (s String) Int(context.Context) int64 {
	i, err := strconv.ParseInt(string(s), 10, 0)
	if err != nil {
		panic(err)
	}
	return int64(i)
}

// Float converts the string representation of a float to a float value.
// If the string doesn't hold a valid float representation,
// it panics.
func (s String) Float(context.Context) float64 {
	f, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		panic(err)
	}
	return f
}

// String returns itself.
func (s String) String(context.Context) string {
	return string(s)
}

// Bool returns true if the string value is not empty, false otherwise.
func (s String) Bool(context.Context) bool {
	return len(string(s)) > 0
}

// Native returns the Go native representation of the value.
func (s String) Native(context.Context) interface{} {
	return string(s)
}
