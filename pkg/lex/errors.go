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
	// ErrMixedIndentation is the error returned if both tabs and spaces are used
	// in a single indent.
	ErrMixedIndentation = fmt.Errorf("you cannot mix tabs and spaces when indenting")

	ErrIndentationIncrease = errors.New("you can only singleIncrease one indentation level at a time")
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
	default:
		return "unsupported indentation"
	}
}
