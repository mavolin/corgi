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
// by a [token.Assign] and a [token.Expression], denoting the default value.
// After each but the last parameter, but optionally also after the last, a
// [token.Comma] is emitted.
func Mixin(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("mixin")
	l.Emit(token.Mixin)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return Error(&lexerr.UnknownItemError{Expected: "a space"})
	}

	end := lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a mixin name"})
	if end != nil {
		return end
	}

	if l.Next() != '(' {
		return Error(&lexerr.UnknownItemError{Expected: "a '('"})
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
			return EOF
		case ')':
			break Params
		case ',':
			l.Emit(token.Comma)

			l.IgnoreWhile(lexer.IsWhitespace)

			// special case: trailing comma
			switch l.Next() {
			case lexer.EOF:
				return EOF
			case ')':
				break Params
			}

			l.Backup()

			// no trailing comma, continue
		default:
			return Error(&lexerr.UnknownItemError{Expected: "a comma, a closing parenthesis, or a mixin parameter"})
		}

		l.IgnoreWhile(lexer.IsWhitespace)
	}

	l.Emit(token.RParen)

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// emitMixinParam consumes and emits a single mixin parameter.
// Each parameter consists of a [token.Ident] (the name), followed by one of
// the following:
//
//  1. a [token.Expression] denoting the default type
//  2. a [token.Assign] followed by a [token.Expression] denoting the default
//     value
//  3. a [token.Expression] denoting the default type, followed by a
//     [token.Assign], followed by a [token.Expression] denoting the default
//     value.
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
		return EOF
	case '\n':
		l.Next()
		return Error(&lexerr.EOLError{In: "mixin parameters"})
	case ',', ')':
		return Error(&lexerr.UnknownItemError{Expected: "type, default value, or both"})
	}

	// type
	if l.Peek() != '=' {
		end = lexutil.EmitNextPredicate(l, token.Expression, nil, lexer.Matches(' ', ',', ')', '=', '\t', '\n'))
		if end != nil {
			return end
		}

		l.IgnoreWhile(lexer.IsHorizontalWhitespace)

		switch l.Peek() {
		case lexer.EOF:
			l.Next()
			return EOF
		case '\n':
			l.Next()
			return Error(&lexerr.EOLError{In: "mixin parameters"})
		case ',', ')':
			return nil
		case '=':
			// handled below
		default:
			return Error(&lexerr.UnknownItemError{Expected: "',', ')', or an '='"})
		}
	}

	l.Next()
	l.Emit(token.Assign)

	l.IgnoreWhile(lexer.IsWhitespace)
	if l.Peek() == lexer.EOF {
		l.Next()
		return EOF
	}

	return lexutil.EmitExpression(l, true, ",", ")")
}

func Return(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("return")
	l.Emit(token.Return)
	return lexutil.AssertNewlineOrEOF(l, Next)
}

// ============================================================================
// Mixin Call
// ======================================================================================

// MixinCall consumes a regular mixin call directive.
func MixinCall(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallName(l); end != nil {
		return end
	}

	return MixinCallArgs
}

// BlockExpansionMixinCall consumes a mixin call directive that is used as part
// of a block expansion.
func BlockExpansionMixinCall(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallName(l); end != nil {
		return end
	}

	return BlockExpansionMixinCallArgs
}

// AttributeValueMixinCall consumes a mixin call directive that is used as the
// value for an attribute.
//
// Note that after the mixin call is lexed, the last state in the chain will
// return a nil state.
// Although inelegant, I made this decision because I didn't want to add four
// more mixin call states (and their partial states).
// Hence, callers are expected to call the returned state function themselves,
// and the state functions returned thereafter until they encounter a nil state.
func AttributeValueMixinCall(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallName(l); end != nil {
		return end
	}

	return AttributeValueMixinCallArgs
}

// MixinCallArgValueMixinCall consumes a mixin call directive that is used as
// the value for  a mixin argument.
//
// Note that after the mixin call is lexed, the last state in the chain will
// return a nil state.
// This is necessary because mixin calls used as args can be nested, and
// otherwise we wouldn't be able to keep track of the nesting.
// Hence, callers are expected to call the returned state function themselves,
// and the state functions returned thereafter until they encounter a nil state.
func MixinCallArgValueMixinCall(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallName(l); end != nil {
		return end
	}

	return MixinCallArgValueMixinCallArgs
}

// InterpolatedMixinCall consumes a mixin call directive that is used as
// part of a text interpolation.
func InterpolatedMixinCall(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallName(l); end != nil {
		return end
	}

	return InterpolatedMixinCallArgs
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
// Mixin Call Args
// ======================================================================================

// MixinCallArgs lexes a mixin call's arguments.
//
// It emits a [token.LParen], followed by zero, one, or multiple parameters,
// and finally a [token.RParen].
//
// Each parameter consists of an [token.Ident], a [token.Assign], and a
// [token.Expression].
func MixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Next() != '(' {
		return Error(&lexerr.UnknownItemError{Expected: "a '('"})
	}
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

// BlockExpansionMixinCallArgs lexes a mixin call's arguments, used as part of
// a block expansion.
//
// It emits a [token.LParen], followed by zero, one, or multiple parameters,
// and finally a [token.RParen].
//
// Each parameter consists of an [token.Ident], a [token.Assign], and a
// [token.Expression].
func BlockExpansionMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Next() != '(' {
		return Error(&lexerr.UnknownItemError{Expected: "a '('"})
	}
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	// special case: no args
	if l.Peek() == ')' {
		l.SkipString(")")
		l.Emit(token.RParen)
		return BehindBlockExpansionMixinCallArgs
	}

	return BlockExpansionMixinCallArg
}

// AttributeValueMixinCallArgs lexes a mixin call's arguments used as a value
// for an attribute.
//
// It emits a [token.LParen], followed by zero, one, or multiple parameters,
// and finally a [token.RParen].
//
// Each parameter consists of an [token.Ident], a [token.Assign], and a
// [token.Expression].
func AttributeValueMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Next() != '(' {
		return Error(&lexerr.UnknownItemError{Expected: "a '('"})
	}
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	// special case: no args
	if l.Peek() == ')' {
		l.SkipString(")")
		l.Emit(token.RParen)
		return BehindAttributeValueMixinCallArgs
	}

	return AttributeValueMixinCallArg
}

// MixinCallArgValueMixinCallArgs lexes a mixin call's arguments used as a
// value for a mixin argument.
//
// It emits a [token.LParen], followed by zero, one, or multiple parameters,
// and finally a [token.RParen].
//
// Each parameter consists of an [token.Ident], a [token.Assign], and a
// [token.Expression].
func MixinCallArgValueMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Next() != '(' {
		return Error(&lexerr.UnknownItemError{Expected: "a '('"})
	}
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	// special case: no args
	if l.Peek() == ')' {
		l.SkipString(")")
		l.Emit(token.RParen)
		return BehindMixinCallArgValueMixinCallArgs
	}

	return MixinCallArgValueMixinCallArg
}

// InterpolatedMixinCallArgs lexes an interpolated mixin call's arguments.
//
// It emits a [token.LParen], followed by zero, one, or multiple parameters,
// and finally a [token.RParen].
//
// Each parameter consists of an [token.Ident], a [token.Assign], and a
// [token.Expression].
func InterpolatedMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Next() != '(' {
		return Error(&lexerr.UnknownItemError{Expected: "a '('"})
	}
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

	var unescaped bool

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case '=':
		l.Emit(token.Assign)
	default:
		return Error(&lexerr.UnknownItemError{Expected: "'=' or '!='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	if l.Peek() == '+' {
		if unescaped {
			return Error(&lexerr.UnknownItemError{Expected: "expression or '=' instead of '!='"})
		}

		return MixinCallArgValueMixinCall
	}

	end := lexutil.EmitExpression(l, true, ",", ")")
	if end != nil {
		return end
	}

	return BehindMixinCallArg
}

func BlockExpansionMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallParamName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	var unescaped bool

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case '=':
		l.Emit(token.Assign)
	default:
		return Error(&lexerr.UnknownItemError{Expected: "'='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	if l.Peek() == '+' {
		if unescaped {
			return Error(&lexerr.UnknownItemError{Expected: "expression or '=' instead of '!='"})
		}

		return MixinCallArgValueMixinCall
	}

	end := lexutil.EmitExpression(l, true, ",", ")")
	if end != nil {
		return end
	}

	return BehindBlockExpansionMixinCallArg
}

func AttributeValueMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallParamName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case '=':
		l.Emit(token.Assign)
	default:
		return Error(&lexerr.UnknownItemError{Expected: "'='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	end := lexutil.EmitExpression(l, true, ",", ")")
	if end != nil {
		return end
	}

	return BehindAttributeValueMixinCallArg
}

func MixinCallArgValueMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallParamName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case '=':
		l.Emit(token.Assign)
	default:
		return Error(&lexerr.UnknownItemError{Expected: "'='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	end := lexutil.EmitExpression(l, true, ",", ")")
	if end != nil {
		return end
	}

	return BehindMixinCallArgValueMixinCallArg
}

func InterpolatedMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitMixinCallParamName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case '=':
		l.Emit(token.Assign)
	default:
		return Error(&lexerr.UnknownItemError{Expected: "'='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	if l.Peek() == '+' {
		if end := emitMixinCallArgValueMixinCall(l); end != nil {
			return end
		}

		return BehindInterpolatedMixinCallArg
	}

	end := lexutil.EmitExpression(l, true, ",", ")")
	if end != nil {
		return end
	}

	return BehindInterpolatedMixinCallArg
}

func emitMixinCallParamName(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	return lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a mixin parameter name"})
}

func emitMixinCallArgValueMixinCall(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	state := MixinCallArgValueMixinCall(l)
	for state != nil {
		state = state(l)
	}

	return nil
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
		return EOF
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
		return Error(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// BehindBlockExpansionMixinCallArg lexes the tokens after a mixin call argument
// used in a mixin call used as part of a block expansion.
//
// It emits a [token.Comma] if the next token is a comma, or a [token.RParen]
// if the next token is a right parenthesis.
// The [token.Comma] can optionally be followed by a [token.RParen].
func BehindBlockExpansionMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case ',':
		l.Emit(token.Comma)
		l.IgnoreWhile(lexer.IsHorizontalWhitespace)

		if l.Peek() == ')' {
			l.Next()
			l.Emit(token.RParen)
			return BehindBlockExpansionMixinCallArgs
		}

		return MixinCallArg
	case ')':
		l.Emit(token.RParen)
		return BehindBlockExpansionMixinCallArgs
	default:
		return Error(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// BehindAttributeValueMixinCallArg lexes the tokens after a mixin call
// argument used in a mixin call used as an attribute.
//
// It emits a [token.Comma] if the next token is a comma, or a [token.RParen]
// if the next token is a right parenthesis.
// The [token.Comma] can optionally be followed by a [token.RParen].
func BehindAttributeValueMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case ',':
		l.Emit(token.Comma)
		l.IgnoreWhile(lexer.IsHorizontalWhitespace)

		if l.Peek() == ')' {
			l.Next()
			l.Emit(token.RParen)
			return BehindAttributeValueMixinCallArgs
		}

		return AttributeValueMixinCallArg
	case ')':
		l.Emit(token.RParen)
		return BehindAttributeValueMixinCallArgs
	default:
		return Error(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// BehindMixinCallArgValueMixinCallArg lexes the tokens after a mixin call
// argument used as a mixin argument.
//
// It emits a [token.Comma] if the next token is a comma, or a [token.RParen]
// if the next token is a right parenthesis.
// The [token.Comma] can optionally be followed by a [token.RParen].
func BehindMixinCallArgValueMixinCallArg(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case ',':
		l.Emit(token.Comma)
		l.IgnoreWhile(lexer.IsHorizontalWhitespace)

		if l.Peek() == ')' {
			l.Next()
			l.Emit(token.RParen)
			return BehindMixinCallArgValueMixinCallArgs
		}

		return MixinCallArgValueMixinCallArg
	case ')':
		l.Emit(token.RParen)
		return BehindMixinCallArgValueMixinCallArgs
	default:
		return Error(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
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
		return EOF
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
		return Error(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// ============================================================================
// Behind Mixin Call Args
// ======================================================================================

func BehindMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return EOF
	case '\n':
		l.IgnoreNext()
		return Next
	case ':':
		return BlockExpansion
	case '=', '!':
		return Assign
	case ' ':
		if l.IsLineEmpty() {
			return lexutil.AssertNewlineOrEOF(l, Next)
		}

		l.IgnoreNext()
		return Text
	case '.':
		return DotBlock
	case '>':
		return MixinMainBlockShorthand
	default:
		return Error(&lexerr.UnknownItemError{Expected: "a space, a '{', or a '['"})
	}
}

func BehindBlockExpansionMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return EOF
	case '\n':
		l.IgnoreNext()
		return Next
	case ':':
		return BlockExpansion
	case '=', '!':
		return Assign
	case ' ':
		if l.IsLineEmpty() {
			return lexutil.AssertNewlineOrEOF(l, Next)
		}

		l.IgnoreNext()
		return Text
	default:
		return Text
	}
}

func BehindAttributeValueMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return EOF
	case '[':
		return AttributeValueMixinCallMainBlockText
	case '{':
		return AttributeValueMixinCallMainBlockExpression
	default:
		return nil
	}
}

func BehindMixinCallArgValueMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return EOF
	case '[':
		return MixinCallArgValueMixinCallMainBlockText
	case '{':
		return MixinCallArgValueMixinCallMainBlockExpression
	default:
		return nil
	}
}

func BehindInterpolatedMixinCallArgs(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return EOF
	case '[':
		return InterpolatedText
	case '{':
		return InterpolatedExpression
	default:
		return Text
	}
}

// ============================================================================
// Mixin Call Main Block Text
// ======================================================================================

func AttributeValueMixinCallMainBlockText(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("[")
	l.Emit(token.LBracket)

	end := lexutil.EmitNextPredicate(l, token.Text, nil, lexer.MatchesNot(']', '\n'))
	if end != nil {
		return end
	}

	if l.Peek() == '\n' {
		return Error(&lexerr.EOLError{In: "a single-line main block"})
	}

	l.SkipString("]")
	l.Emit(token.RBracket)

	return nil
}

func MixinCallArgValueMixinCallMainBlockText(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("[")
	l.Emit(token.LBracket)

	end := lexutil.EmitNextPredicate(l, token.Text, nil, lexer.MatchesNot(']', '\n'))
	if end != nil {
		return end
	}

	if l.Peek() == '\n' {
		return Error(&lexerr.EOLError{In: "a single-line main block"})
	}

	l.SkipString("]")
	l.Emit(token.RBracket)

	return nil
}

// ============================================================================
// Mixin Call Main Block Expression
// ======================================================================================

func AttributeValueMixinCallMainBlockExpression(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("{")
	l.Emit(token.LBrace)

	end := lexutil.EmitExpression(l, false, "}")
	if end != nil {
		return end
	}

	if l.Peek() == '\n' {
		return Error(&lexerr.EOLError{In: "a single-line main block expression"})
	}

	l.SkipString("}")
	l.Emit(token.RBrace)

	return nil
}

func MixinCallArgValueMixinCallMainBlockExpression(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("{")
	l.Emit(token.LBrace)

	end := lexutil.EmitExpression(l, false, "}")
	if end != nil {
		return end
	}

	if l.Peek() == '\n' {
		return Error(&lexerr.EOLError{In: "a single-line main block expression"})
	}

	l.SkipString("}")
	l.Emit(token.RBrace)

	return nil
}

// ============================================================================
// Mixin Main Block Shorthand
// ======================================================================================

// MixinMainBlockShorthand lexes a mixin block shorthand.
//
// It assumes that the next rune is a '>'.
//
// It emits [token.MixinMainBlockShorthand]
func MixinMainBlockShorthand(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString(">")
	l.Emit(token.MixinMainBlockShorthand)

	return lexutil.AssertNewlineOrEOF(l, Next)
}
