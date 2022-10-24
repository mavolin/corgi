package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// ============================================================================
// IfBlock
// ======================================================================================

// IfBlock consumes an 'if block' directive.
//
// It assumes the next string is 'if block'.
//
// It emits a [token.IfBlock] optionally followed by a [token.Ident] with the
// name of the block.
func IfBlock(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("if block")
	l.Emit(token.IfBlock)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		if l.IsLineEmpty() {
			return lexutil.AssertNewlineOrEOF(l, Next)
		}

		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	if end := lexutil.EmitIdent(l, nil); end != nil {
		return end
	}

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// ============================================================================
// If
// ======================================================================================

// If consumes an 'if' directive.
//
// It assumes the next string is 'if'.
//
// It emits a [token.If] optionally followed by an expression.
func If(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("if")
	l.Emit(token.If)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	if end := lexutil.EmitExpression(l, true, '\n'); end != nil {
		return end
	}

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n':
		l.Ignore()
		return Next
	case ':':
		l.Backup()
		return BlockExpansion
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a newline or ':'"})
	}
}

// ElseIf consumes an 'else if' directive.
//
// It assumes the next string is 'else if'.
//
// It emits a [token.ElseIf] optionally followed by an expression.
func ElseIf(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("else if")
	l.Emit(token.ElseIf)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	if end := lexutil.EmitExpression(l, true, ':', '\n'); end != nil {
		return end
	}

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n':
		l.Ignore()
		return Next
	case ':':
		l.Backup()
		return BlockExpansion
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a newline or ':'"})
	}
}

// Else consumes an 'else' directive.
//
// It assumes the next string is 'else'.
//
// It emits a [token.Else] item.
func Else(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("else")
	l.Emit(token.Else)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n':
		l.Ignore()
		return Next
	case ':':
		l.Backup()
		return BlockExpansion
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a newline or ':'"})
	}
}

// ============================================================================
// Switch
// ======================================================================================

// Switch consumes an 'switch' directive.
//
// It assumes the next string is 'switch'.
//
// It emits a [token.Switch] optionally followed by an expression.
func Switch(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("switch")
	l.Emit(token.Switch)

	spaceAfter := l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n': // no comparative value
		l.Ignore()
		return Next
	}

	l.Backup()

	if !spaceAfter {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	if end := lexutil.EmitExpression(l, false, '\n'); end != nil {
		return end
	}

	return Next
}

// Case consumes an 'case' directive.
//
// It assumes the next string is 'case'.
//
// It emits a [token.Case] followed by an expression.
func Case(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("case")
	l.Emit(token.Case)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	if end := lexutil.EmitExpression(l, true, ':', '\n'); end != nil {
		return end
	}

	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case ':':
		return BlockExpansion
	default:
		return Next
	}
}

// Default consumes a 'default' directive.
//
// It assumes the next string is 'default'.
//
// It emits a [token.Default].
func Default(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("default")
	l.Emit(token.Default)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n':
		l.Ignore()
		return Next
	case ':':
		l.Backup()
		return BlockExpansion
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a newline"})
	}
}

// ============================================================================
// For
// ======================================================================================

// For consumes an 'for' directive.
//
// It assumes the next string is 'for'.
//
// It emits a [token.For].
// Then it emits an optional [token.Ident] and optionally a [token.Comma]
// followed by another [token.Ident].
// Then it emits a [token.Range] followed by an expression.
func For(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("for")
	l.Emit(token.For)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	if !l.PeekIsWord("range") {
		end := lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "an identifier"})
		if end != nil {
			return end
		}

		spaceAfter := l.IgnoreWhile(lexer.IsHorizontalWhitespace)

		if l.Peek() == ',' {
			l.Next()
			l.Emit(token.Comma)

			l.IgnoreWhile(lexer.IsHorizontalWhitespace)

			end = lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "an identifier"})
			if end != nil {
				return end
			}

			spaceAfter = l.IgnoreWhile(lexer.IsHorizontalWhitespace)
		}

		if !spaceAfter {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
		}
	}

	if !l.PeekIsWord("range") {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'range'"})
	}

	l.SkipString("range")
	l.Emit(token.Range)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	if end := lexutil.EmitExpression(l, false, '\n'); end != nil {
		return end
	}

	return Next
}

// ============================================================================
// While
// ======================================================================================

// While consumes a while directive.
//
// It emits a [token.While] followed by an Expression.
func While(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("while")
	l.Emit(token.While)

	end := lexutil.EmitExpression(l, false, '\n')
	if end != nil {
		return end
	}

	return Next
}
