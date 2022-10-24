package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// ============================================================================
// Div
// ======================================================================================

// Div lexes a div shorthand.
//
// It assumes the next rune is either '.' or '#' followed by a non-whitespace
// character..
//
// It emits a [token.Div].
func Div(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.Emit(token.Div)
	return BehindElement
}

// BlockExpansionDiv lexes a div shorthand used as part of a block expansion.
//
// It assumes the next rune is either '.' or '#' followed by a non-whitespace
// character..
//
// It emits a [token.Div].
func BlockExpansionDiv(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.Emit(token.Div)
	return BehindBlockExpansionElement
}

// ============================================================================
// Comment
// ======================================================================================

// Comment lexes an HTML comment.
//
// It emits a [token.Comment] followed by either a single [token.Text], or
// followed by a [token.Indent], one or multiple [token.Text] items, and then a
// [token.Dedent].
func Comment(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("//")
	l.Emit(token.Comment)
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n': // either an empty comment or a block comment
		// handled after the switch
	default: // a one-line comment
		end := lexutil.EmitNextPredicate(l, token.Text, nil, lexer.IsNot('\n'))
		if end != nil {
			return end
		}

		return lexutil.AssertNewlineOrEOF(l, Next)
	}

	// we're possibly in a block comment, check if the next line is indented
	dIndent, _, err := l.ConsumeIndent(lexer.ConsumeSingleIncrease)
	if err != nil {
		return lexutil.ErrorState(err)
	}

	lexutil.EmitIndent(l, dIndent)

	if l.Peek() == lexer.EOF {
		return lexutil.EOFState()
	}

	// it's not, just an empty comment.
	if dIndent <= 0 {
		return Next
	}

	for {
		peek := l.NextWhile(lexer.IsNot('\n'))
		if peek == lexer.EOF {
			if !l.IsContentEmpty() {
				l.Emit(token.Text)
			}
			return lexutil.EOFState()
		}

		// even emit empty lines so that these are reflected in the HTML output
		l.Emit(token.Text)

		l.IgnoreNext()

		dIndent, skippedLines, err := l.ConsumeIndent(lexer.ConsumeNoIncrease)
		if err != nil {
			return lexutil.ErrorState(err)
		}

		lexutil.EmitIndent(l, dIndent)

		if l.Peek() == lexer.EOF {
			return lexutil.EOFState()
		}

		if dIndent >= 0 {
			for i := 0; i < skippedLines; i++ {
				l.Emit(token.Text)
			}
		} else {
			return Next
		}
	}
}

// ============================================================================
// Element
// ======================================================================================

// Element lexes the name of an element.
//
// It assumes the next rune is the first rune of the element name, although
// it needn't bee valid.
//
// It emits a [token.Element].
func Element(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitElement(l); end != nil {
		return end
	}

	return BehindElement
}

// BlockExpansionElement lexes the name of an element used as part of a block
// expansion.
//
// It assumes the next rune is the first rune of the element name, although
// it needn't bee valid.
//
// It emits a [token.Element].
func BlockExpansionElement(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitElement(l); end != nil {
		return end
	}

	return BehindBlockExpansionElement
}

// InterpolatedElement lexes an interpolated element.
//
// It assumes the next runes are the name of the element.
//
// It emits a [token.Element].
func InterpolatedElement(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := emitElement(l); end != nil {
		return end
	}

	return BehindInterpolatedElement
}

// emitElement consumes and emits the name of an element.
//
// It assumes the next rune is the first rune of the element name, although
// it needn't bee valid.
func emitElement(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	return lexutil.EmitNextPredicate(l, token.Element,
		&lexerr.UnknownItemError{Expected: "an element name"},
		lexutil.IsElementName)
}

// ============================================================================
// Behind Element
// ======================================================================================

// BehindElement lexes the directives after the name of an element.
func BehindElement(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '.':
		l.Next()

		// special case: this is not a class, but a dot-block
		if l.IsLineEmpty() {
			l.Backup()
			return DotBlock
		}

		l.Backup()

		return Class
	case '#':
		return ID
	case '(':
		return Attributes
	}

	return BehindAttributes
}

// BehindBlockExpansionElement lexes the directives after the name of an
// element used as part of a block expansion.
func BehindBlockExpansionElement(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '.':
		return BlockExpansionClass
	case '#':
		return BlockExpansionID
	case '(':
		return BlockExpansionAttributes
	}

	return BehindBlockExpansionAttributes
}

// BehindInterpolatedElement lexes the directives after the name of an
// interpolated element.
func BehindInterpolatedElement(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '.':
		return InterpolatedClass
	case '#':
		return InterpolatedID
	case '(':
		return InterpolatedAttributes
	}

	return BehindInterpolatedAttributes
}

// BehindAmpersand lexes the directives after an &-directive.
func BehindAmpersand(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '\n':
		l.Next()
		return Next
	case '.':
		return AmpersandClass
	case '#':
		return AmpersandID
	case '(':
		return AmpersandAttributes
	default:
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a class, id, attribute, or newline"})
	}
}

// ============================================================================
// BlockExpansion
// ======================================================================================

// BlockExpansion lexes a block expansion.
//
// It assumes the next rune is a ':'.
//
// It emits an BlockExpansion item followed by an element and optionally
// classes, ids and attributes.
func BlockExpansion(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString(":")
	l.Emit(token.BlockExpansion)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	peek := l.Peek()
	switch {
	case peek == lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case peek == '.' || peek == '#':
		return BlockExpansionDiv
	case l.PeekIsWord("block"):
		return BlockExpansionBlock
	default:
		return BlockExpansionElement
	}
}

// ============================================================================
// TagVoid
// ======================================================================================

// TagVoid lexes a void element indicator.
//
// It assumes the next rune is a '/'.
//
// It emits a [token.TagVoid].
func TagVoid(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("/")
	l.Emit(token.TagVoid)
	return Next
}
