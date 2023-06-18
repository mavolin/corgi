package link

import "fmt"

// Error is the error returned if an extend, use, or include directive
// could not be linked.
type Error struct {
	// Source is the source of the directive.
	Source string
	// File is the name of the directive.
	File string
	// Line is the line of the directive.
	Line int
	// Col is the column of the directive.
	Col int

	// Cause is the error that prevented linking.
	Cause error
}

var _ error = (*Error)(nil)

func (e *Error) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: %s", e.Source, e.File, e.Line, e.Col, e.Cause)
}
