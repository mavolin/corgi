package lex

import "fmt"

// ============================================================================
// String
// ======================================================================================

// _string consumes a string.
//
// It assumes the next character is a '`' or a '"'.
//
// It stops after the end of the string.
func (l *Lexer) _string() stateFn {
	if l.next() == '"' {
		return l._doubleQuotedString()
	}

	return l._backtickString()
}

// _doubleQuotedString consumes a double-quoted string, but doesn't emit it.
//
// It assumes the '"' has already been consumed.
//
// It returns at the end of the string.
func (l *Lexer) _doubleQuotedString() stateFn {
	for {
		l.nextUntil('\\', '"', '\n')
		switch l.next() {
		case eof:
			return l.error(&EOFError{In: "a string"})
		case '\n':
			return l.error(&EOLError{In: "a string"})
		case '\\': // an escape
			// skip the backslash and next character so that we don't possible
			// stop at an escaped quote in the next iteration
			l.next()
		case '"':
			return nil
		}
	}
}

// _backtickString consumes a `backtick` string.
//
// It assumes the '`' has already been consumed.
//
// It returns at the end of the string.
func (l *Lexer) _backtickString() stateFn {
	for {
		l.nextUntil('\\', '`', '\n')
		switch l.next() {
		case eof:
			return l.error(&EOFError{In: "a string"})
		case '\n':
			return l.error(&EOLError{In: "a string"})
		case '`':
			return nil
		}
	}
}

// ============================================================================
// Expression
// ======================================================================================

// _expression consumes an expression and emits the appropriate items.
//
// It assumes the expression starts at the next character.
func (l *Lexer) _expression(allowNewlines bool, terminators ...rune) stateFn {
	switch l.peek() {
	case eof:
		return l.eof
	case '?':
		return l._ternary(allowNewlines)
	}

	var (
		// count of parens, brackets, braces
		parenCount  int
		inString    bool
		inRawString bool
	)

	// in case our terminator is any of those things
	switch l.next() {
	case '"':
		inString = true
	case '`':
		inString = true
		inRawString = true
	case '(', '[', '{':
		parenCount++
	}

	for {
		next := l.next()

		if parenCount <= 0 && !inString {
			for _, t := range terminators {
				if t == next {
					l.backup()
					l.emit(Expression)
					return nil
				}
			}
		}

		switch next {
		case eof:
			if parenCount > 0 {
				return l.error(&EOFError{
					In: fmt.Sprintf("an expression with unclosed parentheses, brackets, or braces (started at %d:%d)",
						l.startLine, l.startCol),
				})
			} else if inString {
				return l.error(&EOFError{
					In: fmt.Sprintf("an expression with an unclosed string (started at %d:%d)",
						l.startLine, l.startCol),
				})
			}

			l.emit(Expression)

			return l.eof
		case '(', '[', '{':
			if !inString {
				parenCount++
			}
		case ')', ']', '}':
			if !inString {
				parenCount--
			}
		case '"':
			inString = !inString
			inRawString = false
		case '`':
			inString = !inString
			inRawString = inString
		case '\\':
			if inString && !inRawString {
				// skip the escaped rune, in case it's a quote
				l.next()
			}
		case '\'':
			if !inString {
				if l.peekIsString(`\'''`) {
					l.nextString(`\'''`)
				} else {
					peek := l.nextUntil('\'')
					if peek == eof {
						return l.error(&EOFError{In: "an expression with an unclosed rune literal"})
					}

					l.next()
				}
			}
		case '?':
			if !inString && parenCount <= 0 {
				l.backup()
				if endState := l._nilCheck(allowNewlines); endState != nil {
					return endState
				}

				l.ignoreWhitespace()

				next = l.next()
				if next == eof {
					return l.eof
				}

				for _, t := range terminators {
					if next == t {
						l.backup()
						return nil
					}
				}

				var terminatorList string
				for i, t := range terminators {
					if i > 0 {
						terminatorList += ", "
					}

					terminatorList += "'" + string(t) + "'"
				}

				return l.error(&UnknownItemError{
					Expected: "one of " + terminatorList,
				})
			}
		case '\n':
			if !allowNewlines {
				return l.error(&EOLError{In: "expression"})
			}
		}
	}
}

// _ternary consumes a ternary expression.
//
// It assumes the next character is '?'.
func (l *Lexer) _ternary(allowNewline bool) stateFn {
	l.nextString("?")
	l.emit(Ternary)

	if endState := l._expression(allowNewline, '('); endState != nil {
		return endState
	}

	if l.next() == eof {
		return l.eof
	}

	l.emit(LParen)

	if err := l._expression(allowNewline, ':'); err != nil {
		return err
	}

	if l.next() == eof {
		return l.eof
	}

	l.emit(TernaryElse)

	l.ignoreWhitespace()
	if l.peek() == eof {
		l.next()
		return l.eof
	}

	if endState := l._expression(allowNewline, ')'); endState != nil {
		return endState
	}

	if l.next() == eof {
		return l.eof
	}

	l.emit(RParen)
	return nil
}

// _nilCheck consumes a nil check expression.
//
// It assumes the previous expression has not yet been emitted and that the
// next character is '?'.
//
// It emits the Expression, then the NilCheck and then, if present, the default
// value contained in parens.
func (l *Lexer) _nilCheck(allowNewline bool) stateFn {
	l.emit(Expression)

	l.nextString("?")
	l.emit(NilCheck)

	if l.peek() != '(' { // no default
		return nil
	}

	l.nextString("(")
	l.emit(LParen)

	if endState := l._expression(allowNewline, ')'); endState != nil {
		return endState
	}

	if l.next() == eof {
		return l.eof
	}

	l.emit(RParen)
	return nil
}
