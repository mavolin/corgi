package element

import "fmt"

// ============================================================================
// NoAndElementError
// ======================================================================================

// NoAndElementError is the error returned by AndChecker.Check if an & is used
// outside an element.
type NoAndElementError struct {
	// Source is the source of the file.
	Source string
	// File is the name of the file.
	File string
	// Line is the line of the &.
	Line int
	// Col is the column of the &.
	Col int
}

var _ error = (*NoAndElementError)(nil)

func (e *NoAndElementError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: cannot use & outside of an element", e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// AndPlacementError
// ======================================================================================

type AndPlacementError struct {
	// Source is the source of the file.
	Source string
	// File is the name of the file.
	File string
	// Line is the line of the &.
	Line int
	// Col is the column of the &.
	Col int
}

var _ error = (*AndPlacementError)(nil)

func (e *AndPlacementError) Error() string {
	return fmt.Sprintf(
		"%s/%s:%d:%d: cannot use `&` at the start of extend blocks or after writing to an element's body",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// LoopAndError
// ======================================================================================

type LoopAndError struct {
	// Source is the source of the file.
	Source string
	// File is the name of the file.
	File string
	// Line is the line of the for.
	Line int
	// Col is the column of the for.
	Col int
}

var _ error = (*LoopAndError)(nil)

func (e *LoopAndError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: loops cannot use `&` while also writing to an element's body",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// SelfClosingBodyError
// ======================================================================================

type SelfClosingBodyError struct {
	// Source is the source of the file.
	Source string
	// File is the name of the file.
	File string
	// Line is the line of the element.
	Line int
	// Col is the column of the element.
	Col int
}

var _ error = (*SelfClosingBodyError)(nil)

func (e *SelfClosingBodyError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: self-closing elements and void elements must not write to their body",
		e.Source, e.File, e.Line, e.Col)
}
