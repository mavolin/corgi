package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// ============================================================================
// Extend
// ======================================================================================

// Extend consumes an Extend directive.
//
// It assumes that the next string is 'extend'.
//
// It emits [token.Extend] and then the [token.Literal] identifying the
// extended file.
func Extend(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("extend")
	l.Emit(token.Extend)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '\n':
		return lexutil.ErrorState(&lexerr.EOLError{After: "a string"})
	case '`', '"':
		// handled below
	default: // invalid
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a string"})
	}

	if end := lexutil.ConsumeString(l); end != nil {
		return end
	}

	l.Emit(token.Literal)

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// ============================================================================
// Import
// ======================================================================================

type inImportBlockKey struct{}

var InImportBlockKey = inImportBlockKey{}

// Import consumes an import directive.
//
// It assumes that the next string is 'import'.
//
// It emits a [token.Import] item.
func Import(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("import")
	l.Emit(token.Import)

	if l.Peek() == lexer.EOF {
		l.Next()
		return lexutil.EOFState()
	}

	if l.Peek() == '\n' { // a block of imports
		l.IgnoreNext()
		dlvl, _, err := l.ConsumeIndent(lexer.ConsumeAllIndents)
		if err != nil {
			return lexutil.ErrorState(err)
		}

		if dlvl != 1 {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a non-empty block of imports"})
		}

		lexutil.EmitIndent(l, dlvl)

		l.Context[InImportBlockKey] = true
	} else { // a single import
		if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space or newline"})
		}
	}

	if peek := l.Peek(); peek == '"' || peek == '`' {
		return ImportPath
	}

	return ImportAlias
}

// ImportAlias lexes an import alias.
//
// It emits a [token.Ident].
func ImportAlias(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	end := lexutil.EmitIdent(l, nil)
	if end != nil {
		return end
	}

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	return ImportPath
}

// ImportPath lexes an import path.
//
// It assumes the next rune is either '`' or '"'.
//
// It emits a [token.Literal].
func ImportPath(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := lexutil.ConsumeString(l); end != nil {
		return end
	}

	l.Emit(token.Literal)

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// ============================================================================
// Use
// ======================================================================================

type inUseBlockKey struct{}

var InUseBlockKey = inUseBlockKey{}

// Use consumes a use directive.
//
// It assumes that the next string is 'use'.
//
// It emits a [token.Use].
func Use(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("use")
	l.Emit(token.Use)

	if l.Peek() == lexer.EOF {
		l.Next()
		return lexutil.EOFState()
	}

	if l.Peek() == '\n' { // a block of uses
		l.IgnoreNext()
		dlvl, _, err := l.ConsumeIndent(lexer.ConsumeAllIndents)
		if err != nil {
			return lexutil.ErrorState(err)
		}

		if dlvl != 1 {
			return lexutil.ErrorState(&lexerr.UnknownItemError{
				Expected: "a non-empty block of use directives",
			})
		}

		lexutil.EmitIndent(l, dlvl)

		l.Context[InUseBlockKey] = true
	} else { // a single use directive
		if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space or newline"})
		}
	}

	if peek := l.Peek(); peek == '"' || peek == '`' {
		return UsePath
	}

	return UseAlias
}

// UseAlias lexes a use directive alias.
//
// It emits a [token.Ident].
func UseAlias(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	end := lexutil.EmitIdent(l, nil)
	if end != nil {
		return end
	}

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	return UsePath
}

// UsePath lexes a use directive path.
//
// It assumes the next rune is either '`' or '"'.
//
// It emits a [token.Literal].
func UsePath(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := lexutil.ConsumeString(l); end != nil {
		return end
	}

	l.Emit(token.Literal)

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// ============================================================================
// Func
// ======================================================================================

// Func consumes the function definition.
//
// It assumes that the next string is 'func'.
//
// It emits a [token.Func] followed by a [token.Ident] (the functions name) and
// then a [token.Literal] containing the parentheses and all function
// parameters.
func Func(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] { //nolint:revive
	l.SkipString("func")
	l.Emit(token.Func)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	end := lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "the function name"})
	if end != nil {
		return end
	}

	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	if l.Next() != '(' {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a '('"})
	}

	peek := l.NextWhile(lexer.IsNot(')'))
	if peek == lexer.EOF {
		l.Next()
		return lexutil.EOFState()
	}

	l.SkipString(")")
	l.Emit(token.Literal)

	return lexutil.AssertNewlineOrEOF(l, Next)
}
