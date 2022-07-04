package voidwriter

import "io"

// Writer is an io.Writer that never returns an error.
var Writer = writer{}

type writer struct{}

var _ io.Writer = writer{}

func (w writer) Write(data []byte) (int, error) {
	return len(data), nil
}
