package lexutil

import (
	"fmt"

	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// ============================================================================
// String
// ======================================================================================

// ConsumeString consumes a string.
//
// It assumes the next character is a '`' or a '"'.
//
// It stops after the end of the string.
func ConsumeString(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Next() == '"' {
		return consumeDoubleQuotedString(l)
	}

	return consumeBacktickString(l)
}

// consumeDoubleQuotedString consumes a double-quoted string, but doesn't emit it.
//
// It assumes the '"' has already been consumed.
//
// It returns at the end of the string.
func consumeDoubleQuotedString(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	for {
		l.NextWhile(lexer.IsNot('\\', '"', '\n'))
		switch l.Next() {
		case lexer.EOF:
			return ErrorState(&lexerr.EOFError{In: "a string"})
		case '\n':
			return ErrorState(&lexerr.EOLError{In: "a string"})
		case '\\': // an escape
			// skip the backslash and next character so that we don't possible
			// stop at an escaped quote in the next iteration
			l.Next()
		case '"':
			return nil
		}
	}
}

// consumeBacktickString consumes a `backtick` string.
//
// It assumes the '`' has already been consumed.
//
// It returns at the end of the string.
func consumeBacktickString(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	for {
		l.NextWhile(lexer.IsNot('`', '\n'))
		switch l.Next() {
		case lexer.EOF:
			return ErrorState(&lexerr.EOFError{In: "a string"})
		case '\n':
			return ErrorState(&lexerr.EOLError{In: "a string"})
		case '`':
			return nil
		}
	}
}

// ============================================================================
// Comment
// ======================================================================================

// IgnoreCorgiComment consumes and ignores a corgi comment ('//-').
//
// If the comment is a block comment, [IgnoreCorgiComment] will consume and
// emit changes in indentation at the end of the comment.
//
// Otherwise, it emits nothing.
func IgnoreCorgiComment(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("//-")
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOFState()
	case '\n': // either an empty comment or a block comment
		// handled after the switch
	default: // a one-line comment
		// since this is a corgi comment and not an HTML comment, just ignore it
		peek := l.NextWhile(lexer.IsNot('\n'))
		if peek == lexer.EOF {
			return EOFState()
		}

		l.IgnoreNext() // '\n'
		return nil
	}

	// we're possibly in a block comment, check if the next line is indented
	dIndent, _, err := l.ConsumeIndent(lexer.ConsumeSingleIncrease)
	if err != nil {
		return ErrorState(err)
	}

	if dIndent < 0 {
		EmitIndent(l, dIndent)
	} else if dIndent == 0 {
		return nil
	}

	for {
		peek := l.NextWhile(lexer.IsNot('\n'))
		if peek == lexer.EOF {
			return EOFState()
		}

		l.IgnoreNext()

		dIndent, _, err = l.ConsumeIndent(lexer.ConsumeNoIncrease)
		if err != nil {
			return ErrorState(err)
		}

		if dIndent < 0 {
			// emit the change in indentation relative to when we encountered the '//-'
			EmitIndent(l, dIndent+1)

			return nil
		}
	}
}

// ============================================================================
// EmitExpression
// ======================================================================================

// EmitExpression consumes an expression and emits the appropriate items.
//
// It assumes the expression starts at the next character.
func EmitExpression(
	l *lexer.Lexer[token.Token], allowNewlines bool, terminators ...rune,
) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		return EOFState()
	case '?':
		return ternary(l, allowNewlines)
	}

	var (
		// count of parens, brackets, braces
		parenCount  int
		inString    bool
		inRawString bool
	)

	// in case our terminator is any of those things
	switch l.Next() {
	case '"':
		inString = true
	case '`':
		inString = true
		inRawString = true
	case '(', '[', '{':
		parenCount++
	}

	for {
		next := l.Next()

		if parenCount <= 0 && !inString {
			for _, t := range terminators {
				if t == next {
					l.Backup()
					l.Emit(token.Expression)
					return nil
				}
			}
		}

		switch next {
		case lexer.EOF:
			if parenCount > 0 {
				return ErrorState(&lexerr.EOFError{
					In: fmt.Sprintf("an expression with unclosed parentheses, brackets, or braces (started at %d:%d)",
						l.StartLine(), l.StartCol()),
				})
			} else if inString {
				return ErrorState(&lexerr.EOFError{
					In: fmt.Sprintf("an expression with an unclosed string (started at %d:%d)",
						l.StartLine(), l.StartCol()),
				})
			}

			l.Emit(token.Expression)

			return EOFState()
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
				l.Next()
			}
		case '\'':
			if !inString {
				if l.PeekIsString(`\''`) {
					l.SkipString(`\''`)
				} else {
					peek := l.NextWhile(lexer.IsNot('\''))
					if peek == lexer.EOF {
						return ErrorState(&lexerr.EOFError{
							In: fmt.Sprintf("an expression with an unclosed rune literal (started at %d:%d)",
								l.StartLine(), l.StartCol()),
						})
					}

					l.Next()
				}
			}
		case '?':
			if !inString && parenCount <= 0 {
				l.Backup()
				if endState := nilCheck(l, allowNewlines); endState != nil {
					return endState
				}

				l.IgnoreWhile(lexer.IsHorizontalWhitespace)

				next = l.Next()
				if next == lexer.EOF {
					return EOFState()
				}

				for _, t := range terminators {
					if next == t {
						l.Backup()
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

				return ErrorState(&lexerr.UnknownItemError{Expected: "one of " + terminatorList})
			}
		case '\n':
			if !allowNewlines {
				return ErrorState(&lexerr.EOLError{In: "expression"})
			}
		}
	}
}

// ternary consumes a ternary expression.
//
// It assumes the next character is '?'.
func ternary(l *lexer.Lexer[token.Token], allowNewline bool) lexer.StateFn[token.Token] {
	l.SkipString("?")
	l.Emit(token.Ternary)

	if endState := EmitExpression(l, allowNewline, '('); endState != nil {
		return endState
	}

	if l.Next() == lexer.EOF {
		return EOFState()
	}

	l.Emit(token.LParen)

	if err := EmitExpression(l, allowNewline, ':'); err != nil {
		return err
	}

	if l.Next() == lexer.EOF {
		return EOFState()
	}

	l.Emit(token.TernaryElse)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)
	if l.Peek() == lexer.EOF {
		l.Next()
		return EOFState()
	}

	if endState := EmitExpression(l, allowNewline, ')'); endState != nil {
		return endState
	}

	if l.Next() == lexer.EOF {
		return EOFState()
	}

	l.Emit(token.RParen)
	return nil
}

// nilCheck consumes a nil check expression.
//
// It assumes the previous expression has not yet been emitted and that the
// next character is '?'.
//
// It emits the EmitExpression, then the NilCheck and then, if present, the default
// value contained in parens.
func nilCheck(l *lexer.Lexer[token.Token], allowNewline bool) lexer.StateFn[token.Token] {
	l.Emit(token.Expression)

	l.SkipString("?")
	l.Emit(token.NilCheck)

	if l.Peek() != '(' { // no default
		return nil
	}

	l.SkipString("(")
	l.Emit(token.LParen)

	if endState := EmitExpression(l, allowNewline, ')'); endState != nil {
		return endState
	}

	if l.Next() == lexer.EOF {
		return EOFState()
	}

	l.Emit(token.RParen)
	return nil
}
