package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// ============================================================================
// Mixin
// ======================================================================================

// Mixin consumes a mixin directive.
//
// It assumes the next string is 'mixin'.
//
// It emits a [token.Mixin], then a [token.Ident] with the name of the mixin,
// and then a [token.LParen] followed by the list of parameters.
// The mixin header is marked finished by emitting a [token.RParen].
//
// Each parameter consists of a [token.Ident] (the name), optionally followed
// by a [token.Assign] and an expression, denoting the default value.
// After each but the last parameter, but optionally also after the last, a
// [token.Comma] is emitted.
func Mixin(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("mixin")
	l.Emit(token.Mixin)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	end := lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a mixin name"})
	if end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	if l.Next() != '(' {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a '('"})
	}

	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsWhitespace)

	// special case: no params
	if l.Peek() == ')' {
		l.SkipString(")")
		l.Emit(token.RParen)
		return lexutil.AssertNewlineOrEOF(l, Next)
	}

Params:
	for {
		if end = emitMixinParam(l); end != nil {
			return end
		}

		switch l.Next() {
		case lexer.EOF:
			return lexutil.EOFState()
		case ')':
			break Params
		case ',':
			l.Emit(token.Comma)

			l.IgnoreWhile(lexer.IsWhitespace)

			// special case: trailing comma
			switch l.Next() {
			case lexer.EOF:
				return lexutil.EOFState()
			case ')':
				break Params
			}

			l.Backup()

			// no trailing comma, continue
		default:
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, a closing parenthesis, or a mixin parameter"})
		}

		l.IgnoreWhile(lexer.IsWhitespace)
	}

	l.Emit(token.RParen)

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// emitMixinParam consumes and emits a single mixin parameter.
// Each parameter consists of a [token.Ident] (the name), optionally followed
// by a [token.Assign] and an expression, denoting the default value.
//
// It assumes the next string is the name of the parameter.
func emitMixinParam(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	end := lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a mixin parameter name"})
	if end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '\n':
		l.Next()
		return lexutil.ErrorState(&lexerr.EOLError{In: "mixin parameters"})
	case ',', ')':
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "type, default value, or both"})
	}

	if l.Peek() != '=' {
		end = lexutil.EmitNextPredicate(l, token.Ident, nil, lexer.IsNot(' ', ',', ')', '=', '\t', '\n'))
		if end != nil {
			return end
		}

		l.IgnoreWhile(lexer.IsHorizontalWhitespace)

		switch l.Peek() {
		case lexer.EOF:
			l.Next()
			return lexutil.EOFState()
		case '\n':
			l.Next()
			return lexutil.ErrorState(&lexerr.EOLError{In: "mixin parameters"})
		case ',', ')':
			return nil
		case '=':
			// handled below
		default:
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "',', ')', or an '='"})
		}
	}

	l.Next()
	l.Emit(token.Assign)

	l.IgnoreWhile(lexer.IsWhitespace)
	if l.Peek() == lexer.EOF {
		l.Next()
		return lexutil.EOFState()
	}

	return lexutil.EmitExpression(l, true, ',', ')')
}

// ============================================================================
// Mixin Call
// ======================================================================================

func MixinCall(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallName(l); end != nil {
		return end
	}

	return BehindMixinCallName
}

func InterpolatedMixinCall(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallName(l); end != nil {
		return end
	}

	return BehindInterpolatedMixinCallName
}

// emitMixinCallName lexes a mixin call name.
//
// It assumes the next rune is '+'.
//
// It emits a [token.MixinCall], followed by a [token.Ident].
// Optionally, it emits another [token.Ident] if the mixin call is namespaced.
func emitMixinCallName(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("+")
	l.Emit(token.MixinCall)

	end := lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a mixin name"})
	if end != nil {
		return end
	}

	if l.Peek() == '.' { // this was just the namespace
		l.IgnoreNext()

		end = lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a mixin name"})
		if end != nil {
			return end
		}

	}

	return nil
}

// ============================================================================
// Behind Mixin Call Name
// ======================================================================================

func BehindMixinCallName(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Peek() == '(' {
		return MixinCallArgs
	}

	return BehindMixinCallArgs
}

func BehindInterpolatedMixinCallName(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Peek() == '(' {
		return InterpolatedMixinCallArgs
	}

	return BehindInterpolatedMixinCallArgs
}

// ============================================================================
// Mixin Call Args
// ======================================================================================

// MixinCallArgs lexes a mixin call's arguments.
//
// It emits a [token.LParen], followed by zero, one, or multiple parameters,
// and finally a [token.RParen].
//
// Each parameter consists of an [token.Ident], a [token.Assign], and an
// expression.
func MixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("(")
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsWhitespace)

	// special case: no args
	if l.Peek() == ')' {
		l.SkipString(")")
		l.Emit(token.RParen)
		return BehindMixinCallArgs
	}

	return MixinCallArg
}

// InterpolatedMixinCallArgs lexes an interpolated mixin call's arguments.
//
// It emits a [token.LParen], followed by zero, one, or multiple parameters,
// and finally a [token.RParen].
//
// Each parameter consists of an [token.Ident], a [token.Assign], and an
// expression.
func InterpolatedMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("(")
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	// special case: no args
	if l.Peek() == ')' {
		l.SkipString(")")
		l.Emit(token.RParen)
		return BehindInterpolatedMixinCallArgs
	}

	return InterpolatedMixinCallArg
}

// ============================================================================
// Single Mixin Call Arg
// ======================================================================================

func MixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallParamName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '!':
		if l.Next() != '=' {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'='"})
		}

		l.Emit(token.AssignNoEscape)
	case '=':
		l.Emit(token.Assign)
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	end := lexutil.EmitExpression(l, true, ',', ')')
	if end != nil {
		return end
	}

	return BehindMixinCallArg
}

func InterpolatedMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallParamName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '!':
		if l.Next() != '=' {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'='"})
		}

		l.Emit(token.AssignNoEscape)
	case '=':
		l.Emit(token.Assign)
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	end := lexutil.EmitExpression(l, true, ',', ')')
	if end != nil {
		return end
	}

	return BehindInterpolatedMixinCallArg
}

func emitMixinCallParamName(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	return lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a mixin parameter name"})
}

// ============================================================================
// Behind Mixin Call Arg
// ======================================================================================

// BehindMixinCallArg lexes the tokens after a mixin call argument.
//
// It emits a [token.Comma] if the next token is a comma, or a [token.RParen]
// if the next token is a right parenthesis.
// The [token.Comma] can optionally be followed by a [token.RParen].
func BehindMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case ',':
		l.Emit(token.Comma)
		l.IgnoreWhile(lexer.IsWhitespace)

		if l.Peek() == ')' {
			l.Next()
			l.Emit(token.RParen)
			return BehindMixinCallArgs
		}

		return MixinCallArg
	case ')':
		l.Emit(token.RParen)
		return BehindMixinCallArgs
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// BehindInterpolatedMixinCallArg lexes the tokens after a mixin call argument
// used in an interpolated mixin call.
//
// It emits a [token.Comma] if the next token is a comma, or a [token.RParen]
// if the next token is a right parenthesis.
// The [token.Comma] can optionally be followed by a [token.RParen].
func BehindInterpolatedMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case ',':
		l.Emit(token.Comma)
		l.IgnoreWhile(lexer.IsHorizontalWhitespace)

		if l.Peek() == ')' {
			l.Next()
			l.Emit(token.RParen)
			return BehindInterpolatedMixinCallArgs
		}

		return MixinCallArg
	case ')':
		l.Emit(token.RParen)
		return BehindInterpolatedMixinCallArgs
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// ============================================================================
// Behind Mixin Call Args
// ======================================================================================

func BehindMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n':
		l.Next()
		return Next
	case '>':
		return MixinBlockShorthand
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space, a '{', or a '['"})
	}
}

func BehindInterpolatedMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '[':
		return InterpolatedText
	case '{':
		return InterpolatedExpression
	case ' ', '\n':
		l.Backup()
		return Text
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space, a '{', or a '['"})
	}
}

// ============================================================================
// Mixin Block Shorthand
// ======================================================================================

// MixinBlockShorthand lexes a mixin block shorthand.
//
// It assumes that the next rune is a '>'.
//
// It emits [token.Shorthand]
func MixinBlockShorthand(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString(">")
	l.Emit(token.MixinBlockShorthand)

	switch l.Peek() {
	case '.':
		return DotBlock
	case ' ', '\t':
		return Text
	default:
		return lexutil.AssertNewlineOrEOF(l, Next)
	}
}
