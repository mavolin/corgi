package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// attrTerminators are the runes that can terminate a class literal, or an id
// literal.
//
// Instead of only allowing certain runes, we are purposefully allowing all
// except for this set to keep the options of classnames and ids as broad as
// possible, so that users won't need to use the more verbose attribute syntax
// every time they want to use rune that we potentially don't allow.
// The rationale here is that there are many applications for non-alphanumeric
// class names/ids, e.g. using the '@' for classes regarding media queries, or
// people using their native, non-latin, language to name things.
var attrTerminators = []rune{'.', '#', '(', '[', '{', '!', '=', ':', ' ', '\t', '\n'}

// ============================================================================
// Ampersand (&)
// ======================================================================================

// Ampersand consumes an '&' directive.
//
// It assumes the next string is '&'.
//
// It emits a [token.And] followed by either a [token.Class], a [token.ID], or
// a [token.LParen].
func Ampersand(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("&")
	l.Emit(token.And)

	return BehindAmpersand
}

// ============================================================================
// Class
// ======================================================================================

// Class lexes a class literal used behind a regular element.
//
// Refer to the documentation of [emitClass] for more information.
func Class(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitClass(l); end != nil {
		return end
	}

	return BehindElement
}

// BlockExpansionClass lexes a class literal used behind an element that is
// part of a block expansion.
//
// Refer to the documentation of [emitClass] for more information.
func BlockExpansionClass(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitClass(l); end != nil {
		return end
	}

	return BehindBlockExpansionElement
}

// InterpolatedClass lexes a class literal used behind an interpolated element.
//
// Refer to the documentation of [emitClass] for more information.
func InterpolatedClass(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitClass(l); end != nil {
		return end
	}

	return BehindInterpolatedElement
}

// AmpersandClass lexes a class literal used behind an &.
//
// Refer to the documentation of [emitClass] for more information.
func AmpersandClass(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitClass(l); end != nil {
		return end
	}

	return BehindAmpersand
}

// emitClass consumes and emits a class directive.
//
// It assumes the next string is '.'.
//
// It emits a [token.Class] followed by a [token.Literal].
func emitClass(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString(".")
	l.Emit(token.Class)

	end := lexutil.EmitNextPredicate(l, token.Literal,
		&lexerr.UnknownItemError{Expected: "a class name"}, lexer.IsNot(attrTerminators...))
	if end != nil {
		return end
	}

	return nil
}

// ============================================================================
// ID
// ======================================================================================

// ID lexes an id literal used behind a regular element.
//
// Refer to the documentation of [emitID] for more information.
func ID(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitID(l); end != nil {
		return end
	}

	return BehindElement
}

// BlockExpansionID lexes an id literal used behind an element that is part of
// a block expansion.
//
// Refer to the documentation of [emitID] for more information.
func BlockExpansionID(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitID(l); end != nil {
		return end
	}

	return BehindBlockExpansionElement
}

// InterpolatedID lexes an id literal used behind an interpolated element.
//
// Refer to the documentation of [emitID] for more information.
func InterpolatedID(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitID(l); end != nil {
		return end
	}

	return BehindInterpolatedElement
}

// AmpersandID lexes an id literal used behind an &.
//
// Refer to the documentation of [emitID] for more information.
func AmpersandID(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitID(l); end != nil {
		return end
	}

	return BehindAmpersand
}

// emitID consumes and emits an id directive.
//
// It assumes the next string is '#'.
//
// It emits a [token.ID] followed by a [token.Literal].
func emitID(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("#")
	l.Emit(token.ID)

	end := lexutil.EmitNextPredicate(l, token.Literal,
		&lexerr.UnknownItemError{Expected: "an id"}, lexer.IsNot(attrTerminators...))
	if end != nil {
		return end
	}

	return nil
}

// ============================================================================
// Attributes
// ======================================================================================

// Attributes lexes a list of attributes used behind a regular element.
//
// Refer to the documentation of [emitAttribute] for more information.
func Attributes(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("(")
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsWhitespace)

	return Attribute
}

// BlockExpansionAttributes lexes a list of attributes used behind an element
// that is part of a block expansion.
//
// It assumes the next string is '(' and hence emits a [token.LParen].
func BlockExpansionAttributes(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("(")
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	return BlockExpansionAttribute
}

// InterpolatedAttributes lexes a list of attributes used behind an
// interpolated element.
//
// It assumes the next string is '(' and hence emits a [token.LParen].
func InterpolatedAttributes(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("(")
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	return InterpolatedAttribute
}

// AmpersandAttributes lexes a list of attributes used behind an &.
//
// It assumes the next string is '(' and hence emits a [token.LParen].
func AmpersandAttributes(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("(")
	l.Emit(token.LParen)

	l.IgnoreWhile(lexer.IsWhitespace)

	return AmpersandAttribute
}

// ============================================================================
// Single Attribute
// ======================================================================================

// Attribute lexes a single attribute found behind a regular element.
//
// It assumes the next string is the name of the attribute.
func Attribute(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitAttributeName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case ',', ')': // boolean attribute
		return BehindAttribute
	case '=':
		l.Emit(token.Assign)
	case '!':
		if l.Next() != '=' {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'='"})
		}

		l.Emit(token.AssignNoEscape)
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, ')', '=', or '!='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)
	if err := lexutil.EmitExpression(l, true, ',', ')'); err != nil {
		return err
	}

	return BehindAttribute
}

// BlockExpansionAttribute lexes a single attribute found behind an element
// used in a block expansion.
//
// It assumes the next string is the name of the attribute.
func BlockExpansionAttribute(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitAttributeName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case ',', ')': // boolean attribute
		return BehindAttribute
	case '=':
		l.Emit(token.Assign)
	case '!':
		if l.Next() != '=' {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'='"})
		}

		l.Emit(token.AssignNoEscape)
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, ')', '=', or '!='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)
	if err := lexutil.EmitExpression(l, true, ',', ')'); err != nil {
		return err
	}

	return BehindBlockExpansionAttribute
}

// InterpolatedAttribute lexes a single attribute found behind an interpolated
// element.
//
// It assumes the next string is the name of the attribute.
func InterpolatedAttribute(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitAttributeName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case ',', ')': // boolean attribute
		return BehindAttribute
	case '=':
		l.Emit(token.Assign)
	case '!':
		if l.Next() != '=' {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'='"})
		}

		l.Emit(token.AssignNoEscape)
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, ')', '=', or '!='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)
	if err := lexutil.EmitExpression(l, true, ',', ')'); err != nil {
		return err
	}

	return BehindInterpolatedAttribute
}

// AmpersandAttribute lexes a single attribute found in a list of attributes
// on an &-directive.
//
// It assumes the next string is the name of the attribute.
func AmpersandAttribute(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitAttributeName(l); end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case ',', ')': // boolean attribute
		return BehindAttribute
	case '=':
		l.Emit(token.Assign)
	case '!':
		if l.Next() != '=' {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "'='"})
		}

		l.Emit(token.AssignNoEscape)
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, ')', '=', or '!='"})
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)
	if err := lexutil.EmitExpression(l, true, ',', ')'); err != nil {
		return err
	}

	return BehindAmpersandAttribute
}

func emitAttributeName(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	var parenCount int

Name:
	for {
		next := l.Next()
		switch {
		case next == lexer.EOF:
			return lexutil.EOFState()
		case !lexutil.IsAttributeName(next):
			fallthrough
		case next == '=' || next == '!' || next == ',':
			fallthrough
		case next == ' ' || next == '\t' || next == '\n':
			l.Backup()
			break Name

		// Support angular attributes, e.g. '(click)'.
		// This is kinda 🥴, but I don't know of any lib/framework/whatever
		// that uses unmatched parentheses in their attributes.
		//
		// To the person who is reading this because they actually use
		// attributes that include unmatched parentheses: File an issue, thx.
		case next == '(':
			parenCount++
		case next == ')':
			parenCount--

			// this was the closing paren of the list of attributes
			if parenCount < 0 {
				l.Backup()
				break Name
			}
		}
	}

	if l.IsContentEmpty() {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "an attribute name"})
	}

	l.Emit(token.Ident)
	return nil
}

// ============================================================================
// Behind Single Attributes
// ======================================================================================

// BehindAttribute lexes the tokens after an attribute.
//
// It assumes the next rune is either eof, ',', or ')'.
//
// It emits either a [token.Comma] optionally followed by a [token.RParen], or
// just a [token.RParen].
func BehindAttribute(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
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
			return BehindElement
		}

		return Attribute
	case ')':
		l.Emit(token.RParen)
		return BehindElement
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// BehindBlockExpansionAttribute lexes the tokens after an attribute that is
// part of an element used in a block expansion.
//
// It assumes the next rune is either eof, ',', or ')'.
//
// It emits either a [token.Comma] optionally followed by a [token.RParen], or
// just a [token.RParen].
func BehindBlockExpansionAttribute(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
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
			return BehindBlockExpansionElement
		}

		return BlockExpansionAttribute
	case ')':
		l.Emit(token.RParen)
		return BehindBlockExpansionElement
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// BehindInterpolatedAttribute lexes the tokens after an attribute that is
// part of an interpolated element.
//
// It assumes the next rune is either eof, ',', or ')'.
//
// It emits either a [token.Comma] optionally followed by a [token.RParen], or
// just a [token.RParen].
func BehindInterpolatedAttribute(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
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
			return BehindInterpolatedElement
		}

		return InterpolatedAttribute
	case ')':
		l.Emit(token.RParen)
		return BehindInterpolatedElement
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// BehindAmpersandAttribute lexes the tokens after an attribute that is part
// of an attribute list attached to an &-directive.
//
// It assumes the next rune is either eof, ',', or ')'.
//
// It emits either a [token.Comma] optionally followed by a [token.RParen], or
// just a [token.RParen].
func BehindAmpersandAttribute(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
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
			return BehindAmpersand
		}

		return AmpersandAttribute
	case ')':
		l.Emit(token.RParen)
		return BehindAmpersand
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a comma, or closing parenthesis"})
	}
}

// ============================================================================
// BehindAttributes
// ======================================================================================

func BehindAttributes(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '\n':
		l.IgnoreNext()
		return Next
	case '/':
		return TagVoid
	case '!', '=':
		return Assign
	case ':':
		return BlockExpansion
	case ' ', '\t':
		if l.IsLineEmpty() {
			return lexutil.AssertNewlineOrEOF(l, Next)
		}

		return Text
	case '.':
		return DotBlock
	default:
		l.Next()
		return lexutil.ErrorState(&lexerr.UnknownItemError{
			Expected: "a class, id, attribute, '/', '=', '!=', ':', a newline, or a space",
		})
	}
}

func BehindBlockExpansionAttributes(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '\n':
		l.IgnoreNext()
		return Next
	case '/':
		return TagVoid
	case '!', '=':
		return Assign
	case ':':
		return BlockExpansion
	case ' ', '\t':
		if l.IsLineEmpty() {
			return lexutil.AssertNewlineOrEOF(l, Next)
		}

		return Text
	default:
		l.Next()
		return lexutil.ErrorState(&lexerr.UnknownItemError{
			Expected: "a class, id, attribute, '/', '=', '!=', ':', a newline, or a space",
		})
	}
}

func BehindInterpolatedAttributes(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
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
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a class, id, attribute, a '{', or a '['"})
	}
}
