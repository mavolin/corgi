// Package assert provides simple assertion functions.
package assert

import (
	"testing"
)

var DebugEnabled = testing.Testing()

// Always panics with the given message if the condition is false.
//
// The assertion is run always.
func Always(cond bool, msg string) {
	if !cond {
		panic("assertion failed: " + msg)
	}
}

// NoError panics with the given message if the given error is not nil.
func NoError(err error, msg string) {
	if err != nil {
		panic("assertion failed: " + msg + ": " + err.Error())
	}
}

// Debug panics with the given message if debug mode is enabled and the given
// condition function returns false.
// The function is only evaluated if debug mode is enabled, preventing
// expensive computations in non-debug settings.
//
// The debug mode is controlled by the DebugEnabled variable.
func Debug(cond func() bool, msg string) {
	if DebugEnabled && !cond() {
		panic("debug assertion failed: " + msg)
	}
}
