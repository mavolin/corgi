package lex

import (
	"errors"
	"fmt"
	"strings"
)

// ============================================================================
// IndentationError
// ======================================================================================

var (
	// ErrMixedIndentation is the error returned if both tabs and spaces are
	// used in a single indent.
	ErrMixedIndentation = errors.New("you cannot mix tabs and spaces when indenting")

	ErrIndentationIncrease = errors.New("you can only increase one indentation level at a time")
)

// IndentationError is the error returned when the indentation of a line is
// inconsistent with the first indented line, which is used as reference.
// This occurs either when using both tabs and spaces for indentation or when
// the amount of indentation is not consistent.
type IndentationError struct {
	Expect    rune
	ExpectLen int
	RefLine   int

	Actual    rune
	ActualLen int
}

var _ error = (*IndentationError)(nil)

func (e *IndentationError) Error() string {
	return fmt.Sprintf("use of inconsistent indentation: expected %s or a multiple a thereof as in line %d, but got %s",
		verbalizeIndent(e.Expect, e.ExpectLen), e.RefLine, verbalizeIndent(e.Actual, e.ActualLen))
}

func verbalizeIndent(indent rune, indentLen int) string {
	switch {
	case indent == ' ':
		if indentLen == 1 {
			return `a single space (" ")`
		}

		return fmt.Sprintf(`%d spaces ("%s")`, indentLen, strings.Repeat(" ", indentLen))
	case indent == '\t':
		if indentLen == 1 {
			return "a single tab (\"\t\")"
		}

		return fmt.Sprintf(`%d tabs (\"%s\")`, indentLen, strings.Repeat("\t", indentLen))
	default: // not a valid indentation
		return `"` + strings.Repeat(string(indent), indentLen) + `"`
	}
}

// ============================================================================
// IllegalIndentationError
// ======================================================================================

// IllegalIndentationError is the error returned when using indentation in a
// place where it can't be used, such as an import statement.
type IllegalIndentationError struct {
	In string
}

var _ error = (*IndentationError)(nil)

func (e *IllegalIndentationError) Error() string {
	if e.In == "" {
		return "illegal indentation"
	}

	return fmt.Sprintf("you cannot indent while in %s", e.In)
}

// ============================================================================
// EOLError
// ======================================================================================

// EOLError is the error returned when a newline was encountered unexpectedly.
type EOLError struct {
	After string
	In    string
}

var _ error = (*EOLError)(nil)

func (e *EOLError) Error() string {
	switch {
	case e.After != "":
		return fmt.Sprintf("unexpected end of line after %s", e.After)
	case e.In != "":
		return fmt.Sprintf("unexpected end of line while parsing %s", e.In)
	default:
		return "unexpected end of line"
	}
}

// ============================================================================
// EOFError
// ======================================================================================

// EOFError is the error returned when an end of file was encountered
// unexpectedly.
type EOFError struct {
	After string
	In    string
}

var _ error = (*EOFError)(nil)

func (e *EOFError) Error() string {
	switch {
	case e.After != "":
		return fmt.Sprintf("unexpected end of file after %s", e.After)
	case e.In != "":
		return fmt.Sprintf("unexpected end of file while parsing %s", e.In)
	default:
		return "unexpected end of file"
	}
}

// ============================================================================
// UnknownItemError
// ======================================================================================

// UnknownItemError is the error returned when an unknown item is encountered.
type UnknownItemError struct {
	Expected string
}

var _ error = (*UnknownItemError)(nil)

func (e *UnknownItemError) Error() string {
	return fmt.Sprintf("unknown item: expected %s", e.Expected)
}
