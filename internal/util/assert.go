package util

import "fmt"

// Assert checks a condition and panics if it's false. If the condition is true,
// this function does nothing.
func Assert(condition bool, detail string, args ...any) {
	if !condition {
		panic(fmt.Sprintf(detail, args...))
	}
}
