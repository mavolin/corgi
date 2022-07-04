package lex

// This file contains all state functions of the lexer as well as some helpers
// denoted by their leading '_'.

// start lexes directives that can be found in the preamble of a file.
// Upon the first non-preamble directive, it switches to next.
func (l *Lexer) start() stateFn {
	l.ignoreRunes('\n')
	if l.peek() == eof {
		return l.eof()
	}

	switch {
	case l.peekIsString("//"):
		return l.comment(l.start)
	case l.peekIsWord("extend"):
		return l.extend
	case l.peekIsWord("import"):
		return l.import_
	case l.peekIsWord("use"):
		return l.use
	case l.peekIsString("-"):
		return l.code(l.start)
	case l.peekIsWord("func"):
		return l.func_
	default:
		return l.nextHTML
	}
}

// ============================================================================
// Extend
// ======================================================================================

// extend consumes an extend directive.
//
// It emits Extend and then the Literal identifying the extended file.
func (l *Lexer) extend() stateFn {
	l.nextString("extend")
	l.emit(Extend)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	switch l.peek() {
	case eof:
		return l.eof
	case '\n':
		return l.error(&EOLError{After: "an string"})
	case '`', '"':
		// handled below
	default: // invalid
		return l.error(&UnknownItemError{Expected: "a string"})
	}

	if endState := l._string(); endState != nil {
		return endState
	}

	l.emit(Literal)

	return l.newlineOrEOF(l.start)
}

// ============================================================================
// Import
// ======================================================================================

// import consumes an import directive
//
// It emits an Import item.
// Then it either directly returns an import, or it emits an Indent indicating
// a list of imports is being read.
//
// Each import is a Literal containing the import path string.
// It is optionally preceded by an Ident declaring an import alias.
func (l *Lexer) import_() stateFn {
	l.nextString("import")
	l.emit(Import)

	spaceAfter := l.ignoreWhitespace()

	switch l.peek() {
	case eof:
		l.next()
		return l.eof()
	case '\n': // a block import
		l.next()
		l.ignore()
		// handled below
	default: // a single import
		if !spaceAfter {
			return l.error(&UnknownItemError{Expected: "a space"})
		}

		if endState := l._singleImport(); endState != nil {
			return endState
		}

		return l.newlineOrEOF(l.start)
	}

	dIndent, _, err := l.consumeIndent(allIndents)
	if err != nil {
		return l.error(err)
	}

	l.emitIndent(dIndent)

	if dIndent <= 0 {
		return l.start
	}

	for dIndent >= 0 {
		if l.peekIsString("//-") {
			if endState := l._corgiComment(); endState != nil {
				return endState
			}

			// _corgiComment may consume indentation on the next non-comment
			// line and correctly emit it.
			// While this is normally not a problem, we need to check if the
			// import block ended.
			if l.indentLen == 0 {
				return l.start
			}
		} else {
			if endState := l._singleImport(); endState != nil {
				return endState
			}
		}

		if l.peek() == eof {
			l.next()
			return l.eof()
		}

		dIndent, _, err = l.consumeIndent(allIndents)
		if err != nil {
			return l.error(err)
		}

		l.emitIndent(dIndent)

		if dIndent > 0 { // can't increase indentation
			return l.error(&IllegalIndentationError{In: "an import block"})
		}
	}

	return l.start
}

// _singleImport consumes a single import directive.
//
// It emits an optional Ident (the alias) and then a Literal indicating the
// import path of the package.
func (l *Lexer) _singleImport() stateFn {
	switch l.peek() {
	case eof:
		l.next()
		return nil
	case '"', '`':
		// handled below
	default: // an alias
		if endState := l._importAlias(); endState != nil {
			return endState
		}

		if !l.ignoreWhitespace() {
			return l.error(&UnknownItemError{Expected: "a space"})
		}

		switch l.next() {
		case eof:
			return nil
		case '\n':
			return l.error(&EOLError{After: "an import alias"})
		case '`', '"': // begin of the import path
			l.backup()
			// handled below
		default: // invalid
			return l.error(&UnknownItemError{Expected: "an import path"})
		}
	}

	// we're at the beginning of the import path
	if endState := l._string(); endState != nil {
		return endState
	}

	l.emit(Literal)

	return l.newlineOrEOF(nil)
}

// _importAlias lexes an import alias.
//
// It emits an Ident.
func (l *Lexer) _importAlias() stateFn {
	if endState := l.emitUntil(Ident, nil, ' ', '\t', '\n'); endState != nil {
		return endState
	}

	switch l.next() {
	case eof:
		return l.eof
	case '\n':
		return l.error(&EOLError{After: "an import alias"})
	default:
		l.backup()
		return nil
	}
}

// ============================================================================
// Use
// ======================================================================================

// use consumes a use directive
//
// It emits a Use item.
// Then it either directly emits a use directive, or it emits an Indent
// indicating a list of use directives is being read.
//
// Each use directive is a Literal containing the import path string.
// It is optionally preceded by an Ident declaring an import alias.
func (l *Lexer) use() stateFn {
	l.nextString("use")
	l.emit(Use)

	spaceAfter := l.ignoreWhitespace()

	switch l.peek() {
	case eof:
		l.next()
		return l.eof()
	case '\n': // a block import
		l.next()
		l.ignore()
		// handled below
	default: // a single import
		if !spaceAfter {
			return l.error(&UnknownItemError{Expected: "a space"})
		}

		if endState := l._singleUse(); endState != nil {
			return endState
		}

		return l.newlineOrEOF(l.start)
	}

	dIndent, _, err := l.consumeIndent(allIndents)
	if err != nil {
		return l.error(err)
	}

	l.emitIndent(dIndent)

	if dIndent <= 0 {
		return l.start
	}

	for dIndent >= 0 {
		if l.peekIsString("//-") {
			if endState := l._corgiComment(); endState != nil {
				return endState
			}

			// _corgiComment may consume indentation on the next non-comment
			// line and correctly emit it.
			// While this is normally not a problem, we need to check if the
			// import block ended.
			if l.indentLen == 0 {
				return l.start
			}
		} else {
			if endState := l._singleUse(); endState != nil {
				return endState
			}
		}

		if l.peek() == eof {
			l.next()
			return l.eof()
		}

		dIndent, _, err = l.consumeIndent(allIndents)
		if err != nil {
			return l.error(err)
		}

		l.emitIndent(dIndent)

		if dIndent > 0 { // can't increase indentation
			return l.error(&IllegalIndentationError{In: "a use block"})
		}
	}

	return l.start
}

// _singleUse consumes a single use directive.
//
// It emits an optional Ident (the alias) and then a Literal indicating the
// import path of the use.
func (l *Lexer) _singleUse() stateFn {
	switch l.peek() {
	case eof:
		l.next()
		return nil
	case '"', '`':
		// handled below
	default: // an alias
		if endState := l._useAlias(); endState != nil {
			return endState
		}

		if !l.ignoreWhitespace() {
			return l.error(&UnknownItemError{Expected: "a space"})
		}

		switch l.next() {
		case eof:
			return nil
		case '\n':
			return l.error(&EOLError{After: "a use alias"})
		case '`', '"': // begin of the import path
			l.backup()
			// handled below
		default: // invalid
			return l.error(&UnknownItemError{Expected: "a use path"})
		}
	}

	// we're at the beginning of the import path
	if endState := l._string(); endState != nil {
		return endState
	}

	l.emit(Literal)

	return l.newlineOrEOF(nil)
}

// _useAlias lexes a use alias.
//
// It emits an Ident.
func (l *Lexer) _useAlias() stateFn {
	if endState := l.emitIdent(nil); endState != nil {
		return endState
	}

	switch l.next() {
	case eof:
		return l.eof
	case '\n':
		return l.error(&EOLError{After: "a use alias"})
	default:
		l.backup()
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
// It emits a Func item followed by an Ident (the functions name) and then a
// Literal containing the parentheses and all function parameters.
func (l *Lexer) func_() stateFn { //nolint:revive
	l.nextString("func")
	l.emit(Func)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	endState := l.emitUntil(Ident, &EOLError{In: "the function name"}, '(', ' ', '\t', '\n')
	if endState != nil {
		return endState
	}

	l.ignoreWhitespace()

	l.nextString("(")
	peek := l.nextUntil(')')
	if peek == eof {
		l.next()
		return l.eof()
	}

	l.nextString(")")
	l.emit(Literal)

	l.ignoreWhitespace()
	return l.newlineOrEOF(l.start)
}
