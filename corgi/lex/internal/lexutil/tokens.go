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

// ConsumeString consumes a string that does not include any interpolation.
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
		l.NextWhile(lexer.MatchesNot('\\', '"', '\n'))
		switch l.Next() {
		case lexer.EOF:
			return errorState(&lexerr.EOFError{WhileParsing: "a string"})
		case '\n':
			return errorState(&lexerr.EOLError{In: "a string"})
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
		l.NextWhile(lexer.MatchesNot('`', '\n'))
		switch l.Next() {
		case lexer.EOF:
			return errorState(&lexerr.EOFError{WhileParsing: "a string"})
		case '\n':
			return errorState(&lexerr.EOLError{In: "a string"})
		case '`':
			return nil
		}
	}
}

// ============================================================================
// EmitExpression
// ======================================================================================

// EmitExpression consumes an expression.
//
// It emits a [token.Expression].
//
// It assumes the expression starts at the next character.
func EmitExpression(
	l *lexer.Lexer[token.Token], allowNewlines bool, terminators ...string,
) lexer.StateFn[token.Token] {
	if l.Peek() == lexer.EOF {
		l.Next()
		return eofState()
	}

	var (
		// count of parens, brackets, braces
		parenCount                int
		strRune                   rune
		strStartLine, strStartCol int
	)

	// in case our terminator is any of those things
	switch l.Next() {
	case '"':
		strRune = '"'
	case '`':
		strRune = '`'
	case '(', '[', '{':
		parenCount++
	}

	for {
		if parenCount == 0 && strRune == 0 {
			for _, t := range terminators {
				if l.PeekIsString(t) {
					l.Backup()
					l.Emit(token.Expression)
					return nil
				}
			}
		}

		next := l.Next()
		switch next {
		case lexer.EOF:
			if parenCount > 0 {
				return errorState(&lexerr.EOFError{
					WhileParsing: fmt.Sprintf("an expression with unclosed parentheses, brackets, or braces (started at %d:%d)",
						l.StartLine(), l.StartCol()),
				})
			} else if strRune != 0 {
				return errorState(&lexerr.EOFError{
					WhileParsing: fmt.Sprintf("an expression with an unclosed string (started at %d:%d)",
						strStartLine, strStartCol),
				})
			}

			l.Emit(token.Expression)

			return eofState()
		case '\n':
			if strRune != 0 {
				return errorState(&lexerr.EOLError{
					In: fmt.Sprintf("an expression with an unclosed string (started at %d:%d)",
						strStartLine, strStartCol),
				})
			}

			if !allowNewlines {
				return errorState(&lexerr.EOLError{In: "expression"})
			}
		case '(', '[', '{':
			if strRune == 0 {
				parenCount++
			}
		case ')', ']', '}':
			if strRune == 0 {
				parenCount--
			}
		case '"', '`':
			if strRune == 0 { // start of string
				strRune = next
				strStartLine, strStartCol = l.Line(), l.Col()
			} else if strRune == next { // end of string
				strRune = 0
			}
		case '\\':
			if strRune == '"' && l.Peek() == '"' {
				// skip the escaped quote, so it doesn't end the string next iteration
				l.Next()
			}
		case '\'':
			if strRune == 0 {
				startLine, startCol := l.Line(), l.Col()

				if l.PeekIsString(`\''`) {
					l.SkipString(`\''`)
				} else {
					peek := l.NextWhile(lexer.MatchesNot('\'', '\n'))
					if peek == lexer.EOF {
						return errorState(&lexerr.EOFError{
							WhileParsing: fmt.Sprintf("an expression with an unclosed rune literal (started at %d:%d)",
								startLine, startCol),
						})
					}

					l.Next()
				}
			}
		}
	}
}
