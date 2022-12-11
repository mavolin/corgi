package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// ============================================================================
// Text
// ======================================================================================

// Text lexes a single line of text.
//
// It emits at least one [token.Text] item, however, multiple may be emitted,
// if the text makes use of the hash operator.
func Text(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.IsLineEmpty() {
		return lexutil.AssertNewlineOrEOF(l, Next)
	}

	for {
		l.NextWhile(lexer.MatchesNot('#', '\n'))
		if !l.PeekIsString("##") { // hash escape
			break
		}

		l.SkipString("##")
	}

	l.Emit(token.Text)

	peek := l.Peek()
	if peek == '#' {
		return Hash
	}

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// ============================================================================
// Interpolation
// ======================================================================================

// Hash lexes a hash expression.
//
// It assumes the next rune is '#', but the rune following it is not '#'.
func Hash(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("#")
	l.Emit(token.Interpolation)

	switch l.Peek() {
	case '+':
		return InterpolatedMixinCall
	case '!':
		l.SkipString("!")
		l.Emit(token.UnescapedInterpolation)
	}

	switch l.Peek() {
	case '[':
		return InterpolatedText
	case '{':
		return InterpolatedExpression
	default:
		return InterpolatedElement
	}
}

// InterpolatedText lexes an interpolated text.
//
// It assumes the next rune is a '['.
//
// It emits a [token.LBracket], followed by a [token.Text], followed by a
// [token.RBracket].
func InterpolatedText(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("[")
	l.Emit(token.LBracket)

	end := lexutil.EmitNextPredicate(l, token.Text,
		&lexerr.UnknownItemError{Expected: "text"}, lexer.MatchesNot(']', '\n'))
	if end != nil {
		return end
	}

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case '\n':
		return Error(&lexerr.EOLError{In: "interpolated text"})
	}

	l.Emit(token.RBracket)
	return Text
}

// InterpolatedExpression lexes an interpolated expression.
//
// It assumes the next rune is a '{'.
//
// It emits a [token.LBracket], followed by a [token.Expression], followed by a
// [token.RBracket].
func InterpolatedExpression(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("{")
	l.Emit(token.LBrace)

	if end := lexutil.EmitExpression(l, false, "}"); end != nil {
		return end
	}

	if l.Next() == lexer.EOF {
		return EOF
	}

	l.Emit(token.RBrace)
	return Text
}

// ============================================================================
// Dot Block
// ======================================================================================

type inDotBlockKey struct{}

var InDotBlockKey = inDotBlockKey{}

// DotBlock lexes a dot block.
//
// It assumes the next rune is '.'.
//
// It emits a [token.DotBlock].
func DotBlock(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString(".")
	l.Emit(token.DotBlock)

	if end := lexutil.AssertNewlineOrEOF(l, nil); end != nil {
		return end
	}

	dIndent, skippedLines, err := l.ConsumeIndent(lexer.ConsumeSingleIncrease)
	if err != nil {
		return Error(err)
	}

	lexutil.EmitIndent(l, dIndent)

	if dIndent <= 0 {
		return Next
	}

	for i := 0; i < skippedLines; i++ {
		l.Emit(token.DotBlockLine)
	}

	l.Context[InDotBlockKey] = true

	return DotBlockLine
}

// DotBlockLine lexes a single non-empty line inside a dot block.
//
// It emits a [token.DotBlockLine], followed by at least one [token.Text], or
// [token.Interpolation].
func DotBlockLine(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.Emit(token.DotBlockLine)
	return Text
}

// ============================================================================
// Pipe
// ======================================================================================

// Pipe lexes a single pipe line.
//
// It assumes the next rune is a '|'.
//
// It emits a [token.Pipe] followed by at least one [token.Text] item, however,
// multiple may be emitted, if the text makes use of the hash operator.
func Pipe(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("|")
	l.Emit(token.Pipe)

	if l.IsLineEmpty() {
		return lexutil.AssertNewlineOrEOF(l, Next)
	}

	if l.Next() != ' ' {
		return Error(&lexerr.UnknownItemError{Expected: "a space"})
	}

	l.Ignore()

	return Text
}

// ============================================================================
// Assign
// ======================================================================================

// Assign lexes a [token.Expression] assignment to an element.
//
// It assumes the next rune is '!' or '='.
//
// It emits a [token.Assign] or a [token.AssignUnescaped] followed by a
// [token.Expression].
func Assign(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Next() == '!' {
		if l.Next() != '=' {
			return Error(&lexerr.UnknownItemError{Expected: "'='"})
		}

		l.Emit(token.AssignNoEscape)
	} else {
		l.Emit(token.Assign)
	}

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return Error(&lexerr.UnknownItemError{Expected: "a space"})
	}

	if end := lexutil.EmitExpression(l, true, "\n"); end != nil {
		return end
	}

	return Next
}

// ============================================================================
// Filter
// ======================================================================================

// Filter lexes a filter directive.
//
// It assumes the next rune is a ':'.
//
// It emits a [token.Filter] followed optionally by a [token.Ident], the name
// of the filter.
// If a [token.Ident] was emitted, it then emits zero, one, or multiple
// [token.Literal] items representing the individual arguments.
// Each [token.Literal] is either a string, as denoted by its '"', or '`'
// prefix, or regular text.
//
// Lastly, it emits zero, one, or multiple [token.Text] constituting the body
// of the filter.
func Filter(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString(":")
	l.Emit(token.Filter)

	if l.IsLineEmpty() {
		return lexutil.AssertNewlineOrEOF(l, Next)
	}

	end := lexutil.EmitNextPredicate(l, token.Ident,
		&lexerr.UnknownItemError{Expected: "the name of the filter"},
		lexer.MatchesNot(' ', '\t', '\n'))
	if end != nil {
		return end
	}

Args:
	for {
		l.IgnoreWhile(lexer.IsHorizontalWhitespace)

		switch l.Peek() {
		case lexer.EOF:
			l.Next()
			return EOF
		case '\n':
			l.IgnoreNext()
			break Args
		}

		if p := l.Peek(); p == '"' || p == '`' {
			if end = lexutil.ConsumeString(l); end != nil {
				return end
			}

			l.Emit(token.Literal)
		} else {
			end = lexutil.EmitNextPredicate(l, token.Literal, nil, lexer.MatchesNot(' ', '\t', '\n'))
			if end != nil {
				return end
			}
		}
	}

	if l.Peek() == lexer.EOF {
		l.Next()
		return EOF
	}

	dIndent, skippedLines, err := l.ConsumeIndent(lexer.ConsumeSingleIncrease)
	if err != nil {
		return Error(err)
	}

	lexutil.EmitIndent(l, dIndent)

	if dIndent <= 0 {
		return Next
	}

	for i := 0; i < skippedLines; i++ {
		l.Emit(token.Text)
	}

	for dIndent >= 0 {
		peek := l.NextWhile(lexer.MatchesNot('\n'))
		if peek == lexer.EOF {
			if !l.IsContentEmpty() {
				l.Emit(token.Text)
			}

			return EOF
		}

		// empty lines are valid
		l.Emit(token.Text)

		l.IgnoreNext()

		dIndent, _, err = l.ConsumeIndent(lexer.ConsumeNoIncrease)
		if err != nil {
			return Error(err)
		}

		lexutil.EmitIndent(l, dIndent)
	}

	return Next
}
