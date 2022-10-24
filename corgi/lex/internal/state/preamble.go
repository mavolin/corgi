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
		return lexutil.EOFState()
	case '\n':
		return lexutil.ErrorState(&lexerr.EOLError{After: "an string"})
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

// Import consumes an import directive.
//
// It assumes that the next string is 'import'.
//
// It emits an Import item.
// Then it either directly returns an import, or it emits an Indent indicating
// a list of imports is being read.
//
// Each import is a Literal containing the import path string.
// It is optionally preceded by an Ident declaring an import alias.
func Import(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("import")
	l.Emit(token.Import)

	spaceAfter := l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '\n': // a block import
		// handled below
	default: // a single import
		if !spaceAfter {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
		}

		if end := emitSingleImport(l); end != nil {
			return end
		}

		return lexutil.AssertNewlineOrEOF(l, Next)
	}

	l.IgnoreNext()

	dIndent, _, err := l.ConsumeIndent(lexer.ConsumeAllIndents)
	if err != nil {
		return lexutil.ErrorState(err)
	}

	lexutil.EmitIndent(l, dIndent)

	if dIndent <= 0 {
		return Next
	}

	for dIndent >= 0 {
		if l.PeekIsString("//-") {
			if end := lexutil.IgnoreCorgiComment(l); end != nil {
				return end
			}

			// consumeCorgiComment may consume indentation on the next
			// non-comment line and correctly emit it.
			// While this is normally not a problem, we need to check if the
			// import block ended.
			if l.Col() == 1 {
				return Next
			}
		} else {
			if end := emitSingleImport(l); end != nil {
				return end
			}
		}

		if l.Peek() == lexer.EOF {
			l.Next()
			return lexutil.EOFState()
		}

		dIndent, _, err = l.ConsumeIndent(lexer.ConsumeAllIndents)
		if err != nil {
			return lexutil.ErrorState(err)
		}

		lexutil.EmitIndent(l, dIndent)

		if dIndent > 0 { // can't increase indentation
			return lexutil.ErrorState(&lexerr.IllegalIndentationError{In: "an import block"})
		}
	}

	return Next
}

// emitSingleImport consumes and emits a single import directive.
//
// It emits an optional [token.Ident] (the alias) and then a [token.Literal]
// indicating the import path of the package.
func emitSingleImport(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return nil
	case '"', '`':
		// handled below
	default: // an alias
		if end := emitImportAlias(l); end != nil {
			return end
		}

		if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
		}

		switch l.Next() {
		case lexer.EOF:
			return nil
		case '\n':
			return lexutil.ErrorState(&lexerr.EOLError{After: "an import alias"})
		case '`', '"': // begin of the import path
			l.Backup()
			// handled below
		default: // invalid
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "an import path"})
		}
	}

	// we're at the beginning of the import path
	if end := lexutil.ConsumeString(l); end != nil {
		return end
	}

	l.Emit(token.Literal)

	return lexutil.AssertNewlineOrEOF(l, nil)
}

// emitImportAlias consumes and emits an import alias.
//
// It emits an Ident.
func emitImportAlias(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	end := lexutil.EmitIdent(l, nil)
	if end != nil {
		return end
	}

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n':
		return lexutil.ErrorState(&lexerr.EOLError{After: "an import alias"})
	default:
		l.Backup()
		return nil
	}
}

// ============================================================================
// Use
// ======================================================================================

// Use consumes a use directive.
//
// It assumes that the next string is 'use'.
//
// It emits an item of type [token.Use].
// Then it either directly emits a use directive, or it emits a
// [token.Indent] indicating a list of use directives is being read.
//
// Each use directive is a [token.Literal] containing the import path string.
// It is optionally preceded by a [token.Ident] declaring an import alias.
func Use(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("use")
	l.Emit(token.Use)

	spaceAfter := l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case '\n': // a block import
		l.IgnoreNext()
		// handled below
	default: // a single import
		if !spaceAfter {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
		}

		if end := emitSingleUse(l); end != nil {
			return end
		}

		return lexutil.AssertNewlineOrEOF(l, Next)
	}

	dIndent, _, err := l.ConsumeIndent(lexer.ConsumeAllIndents)
	if err != nil {
		return lexutil.ErrorState(err)
	}

	lexutil.EmitIndent(l, dIndent)

	if dIndent <= 0 {
		return Next
	}

	for dIndent >= 0 {
		if l.PeekIsString("//-") {
			if end := lexutil.IgnoreCorgiComment(l); end != nil {
				return end
			}

			// _corgiComment may consume indentation on the next non-comment
			// line and correctly emit it.
			// While this is normally not a problem, we need to check if the
			// import block ended.
			if l.Col() == 1 {
				return Next
			}
		} else {
			if end := emitSingleUse(l); end != nil {
				return end
			}
		}

		if l.Peek() == lexer.EOF {
			l.Next()
			return lexutil.EOFState()
		}

		dIndent, _, err = l.ConsumeIndent(lexer.ConsumeAllIndents)
		if err != nil {
			return lexutil.ErrorState(err)
		}

		lexutil.EmitIndent(l, dIndent)

		if dIndent > 0 { // can't increase indentation
			return lexutil.ErrorState(&lexerr.IllegalIndentationError{In: "a use block"})
		}
	}

	return Next
}

// emitSingleUse consumes and emits a single use directive.
//
// It emits an optional [token.Ident] (the alias) and then a [token.Literal]
// indicating the import path of the use directive.
func emitSingleUse(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	switch l.Peek() {
	case lexer.EOF:
		l.Next()
		return nil
	case '"', '`':
		// handled below
	default: // an alias
		if end := emitUseAlias(l); end != nil {
			return end
		}

		if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
		}

		switch l.Next() {
		case lexer.EOF:
			return nil
		case '\n':
			return lexutil.ErrorState(&lexerr.EOLError{After: "a use alias"})
		case '`', '"': // begin of the import path
			l.Backup()
			// handled below
		default: // invalid
			return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a use path"})
		}
	}

	// we're at the beginning of the import path
	if end := lexutil.ConsumeString(l); end != nil {
		return end
	}

	l.Emit(token.Literal)

	return lexutil.AssertNewlineOrEOF(l, nil)
}

// emitUseAlias consumes and emits a use alias.
//
// It emits a [token.Ident].
func emitUseAlias(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if end := lexutil.EmitIdent(l, nil); end != nil {
		return end
	}

	switch l.Next() {
	case lexer.EOF:
		return lexutil.EOFState()
	case '\n':
		return lexutil.ErrorState(&lexerr.EOLError{After: "a use alias"})
	default:
		l.Backup()
		return nil
	}
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
