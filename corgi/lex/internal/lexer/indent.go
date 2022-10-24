package lexer

import (
	"fmt"

	"github.com/mavolin/corgi/corgi/lex/lexerr"
)

type IndentConsumptionMode uint8

const (
	// ConsumeNoIncrease consumes indentation with a negative or no change in
	// indentation.
	//
	// It stops after it has consumed the indent on the first non-empty line,
	// never consuming more than the current indentation level.
	// There may be remaining whitespace ahead, which will not lead to an error
	// return like when consuming [ConsumeAllIndents].
	ConsumeNoIncrease IndentConsumptionMode = iota + 1
	// ConsumeSingleIncrease consumes indentation with a negative change, no change,
	// or an increase by one.
	//
	// It stops after it has consumed the indent on the first non-empty line,
	// never consuming more than one additional indentation level.
	// There may be remaining whitespace ahead, which will not lead to an error
	// return like when consuming [ConsumeAllIndents].
	ConsumeSingleIncrease
	// ConsumeAllIndents consumes all indentation changes.
	//
	// If it encounters an increase by more than one level, it returns
	// ErrIndentationIncrease.
	ConsumeAllIndents
)

// ConsumeIndent consumes the upcoming indentation.
// For that it first skips all empty lines.
// Then, when it encounters a line with content, it consumes all indentation
// before it and determines the delta in indentation, which it returns.
// This will either be a negative number indicating by how many levels
// the indention was removed, 0 to indicate no change in indentation, or 1 to
// indicate an increase in indention.
//
// Callers can use mode to indicate how much indentation should be consumed.
// Refer to the documentation of the mode constants for more information.
//
// When calling, the current position must be at the beginning of a line.
// Otherwise, ConsumeIndent will panic.
//
// If ConsumeIndent encounters the end of file, it will return 0 and nil.
func (l *Lexer[Token]) ConsumeIndent(mode IndentConsumptionMode) (dlvl int, skippedLines int, err error) {
	if l.col != 1 {
		panic(fmt.Sprintf("ConsumeIndent called with non-start-of-line position (%d:%d)", l.line, l.col))
	}

	l.Ignore()

	switch mode {
	case ConsumeNoIncrease:
		for {
			dlvl, stop, err := l.consumeNoIncreaseLineIndent()
			if stop {
				return dlvl, skippedLines, err
			}

			skippedLines++
		}
	case ConsumeSingleIncrease:
		for {
			dlvl, stop, err := l.consumeSingleIncreaseLineIndent()
			if stop {
				return dlvl, skippedLines, err
			}

			skippedLines++
		}
	case ConsumeAllIndents:
		for {
			dlvl, stop, err := l.consumeAllLineIndent()
			if stop {
				return dlvl, skippedLines, err
			}

			skippedLines++
		}
	default:
		panic(fmt.Sprintf("unknown indent consumption mode: %d", mode))
	}
}

func (l *Lexer[Token]) consumeNoIncreaseLineIndent() (dlvl int, stop bool, err error) {
	indent := l.Next()
	switch indent {
	case EOF:
		return 0, true, nil
	case '\n': // empty line
		return 0, false, nil
	case ' ', '\t': // indent
		// handled below
	default: // no indent
		dlvl = -l.indentLvl
		l.indentLvl = 0

		l.Backup()
		return dlvl, true, nil
	}

	if l.indentLvl == 0 {
		l.Backup()

		if l.IsLineEmpty() {
			l.NextWhile(IsNot('\n'))
			l.Next()
			return 0, false, nil
		}

		return 0, true, nil
	}

	count := 1

	limit := l.indentLvl*l.indentLen - 1

Next:
	for i := 0; i < limit; i++ {
		switch next := l.Next(); next {
		case EOF:
			return 0, true, nil
		case ' ', '\t':
			if next != indent {
				return 0, true, lexerr.ErrMixedIndentation
			}

			count++
		default:
			l.Backup()
			break Next
		}
	}

	switch l.Peek() {
	case EOF:
		return 0, true, nil
	case '\n': // empty line
		l.Next()
		return 0, false, nil
	case ' ', '\t': // possibly empty line
		if l.IsLineEmpty() {
			l.NextWhile(IsNot('\n'))
			l.Next()
			return 0, false, nil
		}
	}

	// illegal switching between tabs and spaces or inconsistent indentation
	if indent != l.indent || count%l.indentLen != 0 {
		return 0, true, &lexerr.IndentationError{
			Expect:    l.indent,
			ExpectLen: l.indentLen,
			RefLine:   l.indentRefLine,
			Actual:    indent,
			ActualLen: count,
		}
	}

	newLvl := count / l.indentLen
	dlvl = newLvl - l.indentLvl

	l.indentLvl = newLvl
	return dlvl, true, nil
}

func (l *Lexer[Token]) consumeSingleIncreaseLineIndent() (dlvl int, stop bool, err error) {
	indent := l.Next()
	switch indent {
	case EOF:
		return 0, true, nil
	case '\n': // empty line
		return 0, false, nil
	case ' ', '\t': // indent
		// handled below
	default: // no indent
		dlvl = -l.indentLvl
		l.indentLvl = 0

		l.Backup()
		return dlvl, true, nil
	}

	// this is our first indent, use it for reference
	if l.indent == 0 {
		count := 1

		for n := l.Next(); n == ' ' || n == '\t'; n = l.Next() {
			if n != indent {
				return 0, true, lexerr.ErrMixedIndentation
			}

			count++
		}

		if l.Peek() == EOF {
			return 0, true, nil
		}

		l.Backup()
		if l.Peek() == '\n' { // empty line
			return 0, false, nil
		}

		l.indent = indent
		l.indentLen = count
		l.indentRefLine = l.line
		l.indentLvl = 1

		return 1, true, nil
	}

	count := 1

	limit := (l.indentLvl+1)*l.indentLen - 1

Next:
	for i := 0; i < limit; i++ {
		switch next := l.Next(); next {
		case EOF:
			return 0, true, nil
		case ' ', '\t':
			if next != indent {
				return 0, true, lexerr.ErrMixedIndentation
			}

			count++
		default:
			l.Backup()
			break Next
		}
	}

	switch l.Peek() {
	case EOF:
		return 0, true, nil
	case '\n': // empty line
		l.Next()
		return 0, false, nil
	case ' ', '\t': // possibly empty line
		if l.IsLineEmpty() {
			l.NextWhile(IsNot('\n'))
			l.Next()
			return 0, false, nil
		}
	}

	// illegal switching between tabs and spaces or inconsistent indentation
	if indent != l.indent || count%l.indentLen != 0 {
		return 0, true, &lexerr.IndentationError{
			Expect:    l.indent,
			ExpectLen: l.indentLen,
			RefLine:   l.indentRefLine,
			Actual:    indent,
			ActualLen: count,
		}
	}

	newLvl := count / l.indentLen
	dlvl = newLvl - l.indentLvl

	l.indentLvl = newLvl
	return dlvl, true, nil
}

func (l *Lexer[Token]) consumeAllLineIndent() (dlvl int, stop bool, err error) {
	indent := l.Next()
	switch indent {
	case EOF:
		return 0, true, nil
	case '\n': // empty line
		return 0, false, nil
	case ' ', '\t': // indent
		// handled below
	default: // no indent
		dlvl = -l.indentLvl
		l.indentLvl = 0

		l.Backup()
		return dlvl, true, nil
	}

	count := 1

	for next := l.Next(); next == ' ' || next == '\t'; next = l.Next() {
		if next != indent {
			return 0, true, lexerr.ErrMixedIndentation
		}

		count++
	}

	if l.Peek() == EOF {
		return 0, true, nil
	}

	l.Backup()
	if l.Peek() == '\n' { // empty line
		return 0, false, nil
	}

	// this is our first indent, use it for reference
	if l.indent == 0 {
		l.indent = indent
		l.indentLen = count
		l.indentRefLine = l.line

		l.indentLvl = 1
		return 1, true, nil
	}

	// illegal switching between tabs and spaces or inconsistent indentation
	if indent != l.indent || count%l.indentLen != 0 {
		return 0, true, &lexerr.IndentationError{
			Expect:    l.indent,
			ExpectLen: l.indentLen,
			RefLine:   l.indentRefLine,
			Actual:    indent,
			ActualLen: count,
		}
	}

	newLvl := count / l.indentLen
	dlvl = newLvl - l.indentLvl
	if dlvl > 1 {
		return 0, true, lexerr.ErrIndentationIncrease
	}

	l.indentLvl = newLvl
	return dlvl, true, nil
}
