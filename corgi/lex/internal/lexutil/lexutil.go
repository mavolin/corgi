// Package lexutil provides utilities to be used by package state.
package lexutil

import (
	"unicode"

	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// EmitIdent consumes a [token.Ident] and then does one of the following:
//
// If the ident is not empty, it emits a [lexer.Item] of type [token.Ident] and
// returns nil.
//
// If it is empty and the end of file is reached, it returns a [lexer.StateFn]
// that emits [token.EOF].
//
// If it is empty and ifEmptyErr is set and the end of file is not reached,
// it returns a [lexer.StateFn] that emits ifEmptyErr.
//
// In any other case, it returns nil.
func EmitIdent(l *lexer.Lexer[token.Token], ifEmptyErr error) lexer.StateFn[token.Token] {
	next := l.Next()
	if next == lexer.EOF {
		return EOFState()
	}

	if next != '_' && !unicode.IsLetter(next) {
		if ifEmptyErr != nil {
			return ErrorState(ifEmptyErr)
		}

		return nil
	}

	l.NextWhile(func(r rune) bool {
		switch {
		case r == '_':
		case r >= '0' && r <= '9':
		case unicode.IsLetter(r):
		default:
			return false
		}

		return true
	})

	l.Emit(token.Ident)
	return nil
}

// EmitIndent emits the change in indentation level.
//
// If delta is 0, emitIndent will do nothing.
func EmitIndent(l *lexer.Lexer[token.Token], delta int) {
	if delta != 0 {
		typ := token.Indent
		if delta < 0 {
			delta = -delta
			typ = token.Dedent
		}

		for i := 0; i < delta; i++ {
			l.Emit(typ)
		}
	}
}

// EmitNextPredicate calls [lexer.Lexer.NextWhile] with the passed
// predicate.
//
// If at least one rune was consumed, it emits a [lexer.Item] of type t.
//
// If [lexer.Lexer.NextWhile] returns [lexer.EOF], a [lexer.StateFn]
// emitting [token.EOF] is returned.
//
// If no rune was consumed and ifEmptyErr, a [lexer.StateFn] emitting
// ifEmptyErr is returned.
//
// In any other case, nil is returned.
func EmitNextPredicate(
	l *lexer.Lexer[token.Token], t token.Token, ifEmptyErr error, predicate func(rune) bool,
) lexer.StateFn[token.Token] {
	peek := l.NextWhile(predicate)

	empty := l.IsContentEmpty()
	if !empty {
		l.Emit(t)
	}

	if peek == lexer.EOF {
		return EOFState()
	}

	if ifEmptyErr != nil && empty {
		return ErrorState(ifEmptyErr)
	}

	return nil
}

// AssertNewlineOrEOF skips all horizontal whitespaces and then asserts that
// the next rune is either a newline or the end of file.
//
// If it encounters the end of file, it returns a [lexer.StateFn] that emits
// [token.EOF].
//
// If it encounters a newline, it consumes and ignores it and returns next.
//
// If it encounters another character, it returns a [lexer.StateFn] that
// emits an [*UnknownItemError].
func AssertNewlineOrEOF(l *lexer.Lexer[token.Token], next lexer.StateFn[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOFState()
	case '\n':
		l.Ignore()
		return next
	default:
		return ErrorState(&lexerr.UnknownItemError{Expected: "a newline"})
	}
}

// EOFState returns a [lexer.StateFn] that emits a [token.EOF] and then returns
// nil.
func EOFState() lexer.StateFn[token.Token] {
	return func(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
		l.Emit(token.EOF)
		return nil
	}
}

// ErrorState returns a [lexer.StateFn] that emits the passed error err and
// then returns nil.
func ErrorState(err error) lexer.StateFn[token.Token] {
	return func(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
		l.EmitError(token.Error, err)
		return nil
	}
}
