package lex

// nextHTML consumes the next HTMl directive, excluding doctype.
func (l *Lexer) nextHTML() stateFn {
	if l.peek() == eof {
		return l.eof
	}

	l.ignoreRunes('\n') // empty lines

	if l.col == 1 {
		dIndent, _, err := l.consumeIndent(allIndents)
		if err != nil {
			return l.error(err)
		}

		l.emitIndent(dIndent)
	}

	// some keywords have spaces behind them to avoid confusion if they are
	// just the prefix of an element
	switch {
	case l.peek() == eof:
		return l.eof
	case l.peekIsWord("doctype"):
		return l.doctype
	case l.peekIsString("//"):
		return l.comment(l.nextHTML)
	case l.peekIsWord("-"):
		return l.code(l.nextHTML)
	case l.peekIsWord("include"):
		return l.include
	case l.peekIsWord("block"):
		return l.block
	case l.peekIsWord("append"):
		return l.blockAppend
	case l.peekIsWord("prepend"):
		return l.blockPrepend
	case l.peekIsWord("mixin"):
		return l.mixin
	case l.peekIsWord("if block"):
		return l.ifBlock
	case l.peekIsWord("if"):
		return l.if_
	case l.peekIsWord("else if"):
		return l.elseIf
	case l.peekIsWord("else"), l.peekIsWord("else:"): // block expansion
		return l.else_
	case l.peekIsWord("switch"):
		return l.switch_
	case l.peekIsWord("case"):
		return l.case_
	case l.peekIsWord("default"), l.peekIsWord("default:"): // block expansion
		return l.caseDefault
	case l.peekIsWord("for"):
		return l.for_
	case l.peekIsWord("while"):
		return l.while
	case l.peekIsString("&"):
		return l.and
	case l.peekIsString("."), l.peekIsString("#"):
		return l.divOrDotBlock
	case l.peekIsString("+"):
		if endState := l._mixinCall(true); endState != nil {
			return endState
		}

		return l.nextHTML
	case l.peekIsString("="), l.peekIsString("!="):
		return l.assign
	case l.peekIsString("|"):
		return l.pipe
	case l.peekIsString(":"):
		return l.filter
	default:
		return l.element
	}
}

// ============================================================================
// Doctype
// ======================================================================================

// doctype consumes a doctype directive.
//
// It assumes the next string is 'doctype'.
//
// It emits a doctype item followed by a single Literal containing ALL doctype
// text.
func (l *Lexer) doctype() stateFn {
	l.nextString("doctype")
	l.emit(Doctype)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l.emitUntil(Literal, &EOLError{In: "the doctype"}, '\n'); endState != nil {
		return endState
	}

	return l.newlineOrEOF(l.nextHTML)
}

// ============================================================================
// Comment
// ======================================================================================

// comment consumes a comment and then returns next.
//
// If the comments is an HTML comment, it emits a Comment item.
func (l *Lexer) comment(next stateFn) stateFn {
	return func() stateFn {
		if l.peekIsString("//-") {
			if endState := l._corgiComment(); endState != nil {
				return endState
			}

			return next
		}

		return l.htmlComment(next)
	}
}

// _corgiComment is the same as comment, but only for '//-' style comments.
//
// It emits nothing.
func (l *Lexer) _corgiComment() stateFn {
	l.nextString("//-")
	l.ignoreWhitespace()

	switch l.next() {
	case eof:
		return l.eof
	case '\n': // either an empty comment or a block comment
		// handled after the switch
	default: // a one-line comment
		// since this is a corgi comment and not an HTML comment, just ignore it
		peek := l.nextUntil('\n')
		if peek == eof {
			return l.eof
		}

		l.nextString("\n")
		l.ignore()
		return nil
	}

	// we're possibly in a block comment, check if the next line is indented
	dIndent, _, err := l.consumeIndent(singleIncrease)
	if err != nil {
		return l.error(err)
	}

	l.emitIndent(dIndent - 1)

	// it's not, just an empty comment.
	if dIndent <= 0 {
		return nil
	}

	for {
		peek := l.nextUntil('\n')
		if peek == eof {
			return l.eof
		}

		l.nextString("\n")
		l.ignore()

		dIndent, _, err := l.consumeIndent(noIncrease)
		if err != nil {
			return l.error(err)
		}

		// emit the change in indentation relative to when we encountered the '//-'
		l.emitIndent(dIndent + 1)

		if dIndent < 0 {
			return nil
		}
	}
}

// htmlComment lexes an HTML comment.
//
// It emits a Comment item followed by either one Text, or followed by an
// Indent, one or multiple Texts and then a Dedent.
func (l *Lexer) htmlComment(next stateFn) stateFn {
	return func() stateFn {
		l.nextString("//")
		l.emit(Comment)
		l.ignoreWhitespace()

		switch l.next() {
		case eof:
			return l.eof
		case '\n': // either an empty comment or a block comment
			// handled after the switch
		default: // a one-line comment
			if endState := l.emitUntil(Text, nil, '\n'); endState != nil {
				return endState
			}

			return l.newlineOrEOF(next)
		}

		// we're possibly in a block comment, check if the next line is indented
		dIndent, _, err := l.consumeIndent(singleIncrease)
		if err != nil {
			return l.error(err)
		}

		l.emitIndent(dIndent)

		if l.peek() == eof {
			return l.eof
		}

		// it's not, just an empty comment.
		if dIndent <= 0 {
			return next
		}

		for {
			peek := l.nextUntil('\n')
			if peek == eof {
				if !l.contentEmpty() {
					l.emit(Text)
				}
				return l.eof
			}

			// even emit empty lines so that these are reflected in the HTML output
			l.emit(Text)

			l.nextString("\n")
			l.ignore()

			dIndent, skippedLines, err := l.consumeIndent(noIncrease)
			if err != nil {
				return l.error(err)
			}

			l.emitIndent(dIndent)

			if l.peek() == eof {
				return l.eof
			}

			if dIndent >= 0 {
				for i := 0; i < skippedLines; i++ {
					l.emit(Text)
				}
			} else {
				return next
			}
		}
	}
}

// ============================================================================
// Include
// ======================================================================================

// include consumes a single include directive.
//
// It assumes the next string will be 'include'.
//
// It emits Include and then a Literal identifying the included file.
func (l *Lexer) include() stateFn {
	l.nextString("include")
	l.emit(Include)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	switch l.peek() {
	case eof:
		return l.eof
	case '\n':
		return l.error(&EOLError{After: "a string"})
	case '`', '"':
		// handled below
	default: // invalid
		return l.error(&UnknownItemError{Expected: "a string"})
	}

	if endState := l._string(); endState != nil {
		return endState
	}

	l.emit(Literal)

	return l.newlineOrEOF(l.nextHTML)
}

// ============================================================================
// Code
// ======================================================================================

// Code consumes a code line or block.
//
// It assumes the next rune is a '-'.
//
// It emits a CodeStart item and then either one Code item
// or an Indent and one, or multiple Code items, terminated by a Dedent.
func (l *Lexer) code(next stateFn) stateFn {
	return func() stateFn {
		l.nextString("-")
		l.emit(CodeStart)

		spaceAfter := l.ignoreWhitespace()

		switch l.next() {
		case eof:
			return l.eof
		case '\n': // this is a block of code
			l.ignore()
			// handled below
		default: // a single line of code
			// special case: empty line, for ✨visuals✨
			if l.isLineEmpty() {
				l.ignoreWhitespace()
				return l.newlineOrEOF(next)
			}

			if !spaceAfter {
				return l.error(&UnknownItemError{Expected: "a space"})
			}

			endState := l.emitUntil(Code, &EOLError{After: "a line of code"}, '\n')
			if endState != nil {
				return endState
			}

			return l.newlineOrEOF(next)
		}

		// we're at the beginning of a block of code

		dIndent, _, err := l.consumeIndent(singleIncrease)
		if err != nil {
			return l.error(err)
		}

		l.emitIndent(dIndent)

		if dIndent <= 0 {
			return next
		}

		for dIndent >= 0 {
			if endState := l.emitUntil(Code, nil, '\n'); endState != nil {
				return endState
			}

			if l.next() == eof {
				return l.eof
			}

			dIndent, _, err = l.consumeIndent(noIncrease)
			if err != nil {
				return l.error(err)
			}

			l.emitIndent(dIndent)
		}

		return next
	}
}

// ============================================================================
// Block
// ======================================================================================

// block consumes a block directive.
//
// It assumes the next string is 'block'.
//
// It emits a Block item.
// If the block is named, it emits an Ident with the name of the block.
//
// Lastly, it optionally emits a DotBlock, if the line ends with a '.'.
func (l *Lexer) block() stateFn {
	l.nextString("block")
	l.emit(Block)

	if l.peek() == '.' {
		return l.dotBlock
	}

	if l.isLineEmpty() {
		return l.newlineOrEOF(l.nextHTML)
	}

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	l.emitIdent(&UnknownItemError{Expected: "a block name"})

	if l.peek() == '.' {
		return l.dotBlock
	}

	return l.newlineOrEOF(l.nextHTML)
}

// block consumes an append directive.
//
// It assumes the next string is 'append'.
//
// It emits an BlockAppend item optionally followed by a BlockIfExists.
// Lastly, it emits an Ident with the name of the block to append.
func (l *Lexer) blockAppend() stateFn {
	l.nextString("append")
	l.emit(BlockAppend)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l.emitIdent(&EOLError{After: "'append'"}); endState != nil {
		return endState
	}

	if l.peek() == '.' {
		return l.dotBlock
	}

	return l.newlineOrEOF(l.nextHTML)
}

// block consumes a prepend directive.
//
// It assumes the next string is 'prepend'.
//
// It emits an BlockPrepend item optionally followed by a BlockIfExists.
// Lastly, it emits an Ident with the name of the block to append.
func (l *Lexer) blockPrepend() stateFn {
	l.nextString("prepend")
	l.emit(BlockPrepend)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l.emitIdent(&EOLError{After: "'prepend'"}); endState != nil {
		return endState
	}

	if l.peek() == '.' {
		return l.dotBlock
	}

	return l.newlineOrEOF(l.nextHTML)
}

// ============================================================================
// Mixin
// ======================================================================================

// mixin consumes a mixin directive.
//
// It assumes the next string is 'mixin'.
//
// It emits a Mixin item, then an Ident with the name of the mixin, and then a
// LParen followed by the list of parameters.
// mixin finished by emitting a RParen.
//
// Each parameter consists of an Ident (the name), optionally followed by an
// Assign and a Code, denoting the default value.
// After each but the last parameter, but optionally also after the last, a
// Comma is emitted.
func (l *Lexer) mixin() stateFn {
	l.nextString("mixin")
	l.emit(Mixin)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	endState := l.emitIdent(&UnknownItemError{Expected: "a mixin name"})
	if endState != nil {
		return endState
	}

	l.ignoreWhitespace()

	if l.next() != '(' {
		return l.error(&UnknownItemError{Expected: "a '('"})
	}

	l.emit(LParen)

	l.ignoreRunes(' ', '\t', '\n')

	// special case: no params
	if l.peek() == ')' {
		l.nextString(")")
		l.emit(RParen)
		return l.newlineOrEOF(l.nextHTML)
	}

Params:
	for {
		if endState := l._mixinParam(); endState != nil {
			return endState
		}

		switch l.next() {
		case eof:
			return l.eof
		case ')':
			break Params
		case ',':
			l.emit(Comma)

			l.ignoreRunes(' ', '\t', '\n')

			// special case: trailing comma
			switch l.next() {
			case eof:
				return l.eof
			case ')':
				break Params
			}

			l.backup()

			// no trailing comma, continue
		default:
			return l.error(&UnknownItemError{Expected: "a comma, a closing parenthesis, or a mixin parameter"})
		}

		l.ignoreRunes(' ', '\t', '\n')
	}

	l.emit(RParen)

	return l.newlineOrEOF(l.nextHTML)
}

// _mixinParam consumes a single mixin parameter.
// Each parameter consists of an Ident (the name), optionally followed by an
// Assign and an Expression, denoting the default value.
//
// It assumes the next string is the name of the parameter.
func (l *Lexer) _mixinParam() stateFn {
	endState := l.emitIdent(&UnknownItemError{Expected: "a mixin parameter name"})
	if endState != nil {
		return endState
	}

	l.ignoreWhitespace()

	switch l.peek() {
	case eof:
		l.next()
		return l.eof
	case '\n':
		l.next()
		return l.error(&EOLError{In: "mixin parameters"})
	case ',', ')':
		return l.error(&UnknownItemError{Expected: "type, default value, or both"})
	}

	if l.peek() != '=' {
		endState = l.emitUntil(Ident, nil, ' ', ',', ')', '=', '\t', '\n')
		if endState != nil {
			return endState
		}

		l.ignoreWhitespace()

		switch l.peek() {
		case eof:
			l.next()
			return l.eof
		case '\n':
			l.next()
			return l.error(&EOLError{In: "mixin parameters"})
		case ',', ')':
			return nil
		case '=':
			// handled below
		default:
			return l.error(&UnknownItemError{Expected: "',', ')', or an '='"})
		}
	}

	l.next()
	l.emit(Assign)

	l.ignoreRunes(' ', '\t', '\n')
	if l.peek() == eof {
		l.next()
		return l.eof
	}

	return l._expression(true, ',', ')')
}

// ============================================================================
// Conditionals
// ======================================================================================

// ifBlock consumes an 'if block' directive.
//
// It assumes the next string is 'if block'.
//
// It emits an IfBlock item optionally followed by an Ident with the name of
// the block.
func (l *Lexer) ifBlock() stateFn {
	l.nextString("if block")
	l.emit(IfBlock)

	if !l.ignoreWhitespace() {
		if l.isLineEmpty() {
			return l.newlineOrEOF(l.nextHTML)
		}

		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l.emitIdent(nil); endState != nil {
		return endState
	}

	return l.newlineOrEOF(l.nextHTML)
}

// if_ consumes an 'if' directive.
//
// It assumes the next string is 'if'.
//
// It emits an If item optionally followed by an Expression.
func (l *Lexer) if_() stateFn { //nolint:revive
	l.nextString("if")
	l.emit(If)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l._expression(true, '\n'); endState != nil {
		return endState
	}

	switch l.next() {
	case eof:
		return l.eof
	case '\n':
		l.ignore()
		return l.nextHTML
	case ':':
		l.backup()
		return l.blockExpansion
	default:
		return l.error(&UnknownItemError{Expected: "a newline or ':'"})
	}
}

// elseIf consumes an 'else if' directive.
//
// It assumes the next string is 'else if'.
//
// It emits an ElseIf item optionally followed by an Expression.
func (l *Lexer) elseIf() stateFn {
	l.nextString("else if")
	l.emit(ElseIf)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l._expression(true, ':', '\n'); endState != nil {
		return endState
	}

	switch l.next() {
	case eof:
		return l.eof
	case '\n':
		l.ignore()
		return l.nextHTML
	case ':':
		l.backup()
		return l.blockExpansion
	default:
		return l.error(&UnknownItemError{Expected: "a newline or ':'"})
	}
}

// else consumes an 'else' directive.
//
// It assumes the next string is 'else'.
//
// It emits an Else item.
func (l *Lexer) else_() stateFn { //nolint:revive
	l.nextString("else")
	l.emit(Else)

	l.ignoreWhitespace()

	switch l.next() {
	case eof:
		return l.eof
	case '\n':
		l.ignore()
		return l.nextHTML
	case ':':
		l.backup()
		return l.blockExpansion
	default:
		return l.error(&UnknownItemError{Expected: "a newline or ':'"})
	}
}

// ============================================================================
// Switch
// ======================================================================================

// switch_ consumes an 'switch' directive.
//
// It assumes the next string is 'switch'.
//
// It emits a Switch item optionally followed by an Expression.
func (l *Lexer) switch_() stateFn { //nolint:revive
	l.nextString("switch")
	l.emit(Switch)

	spaceAfter := l.ignoreWhitespace()

	switch l.next() {
	case eof:
		return l.eof
	case '\n': // no comparative value
		return l.nextHTML
	}

	l.backup()

	if !spaceAfter {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l._expression(false, '\n'); endState != nil {
		return endState
	}

	return l.nextHTML
}

// case_ consumes an 'case' directive.
//
// It assumes the next string is 'case'.
//
// It emits a Case item followed by an Expression.
func (l *Lexer) case_() stateFn { //nolint:revive
	l.nextString("case")
	l.emit(Case)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l._expression(true, ':', '\n'); endState != nil {
		return endState
	}

	switch l.peek() {
	case eof:
		l.next()
		return l.eof
	case ':':
		return l.blockExpansion
	default:
		return l.nextHTML
	}
}

// case_ consumes an 'default' directive.
//
// It assumes the next string is 'default'.
//
// It emits a DefaultCase.
func (l *Lexer) caseDefault() stateFn {
	l.nextString("default")
	l.emit(DefaultCase)

	l.ignoreWhitespace()

	switch l.next() {
	case eof:
		return l.eof
	case '\n':
		l.ignore()
		return l.nextHTML
	case ':':
		l.backup()
		return l.blockExpansion
	default:
		return l.error(&UnknownItemError{Expected: "a newline"})
	}
}

// ============================================================================
// For
// ======================================================================================

// for_ consumes an 'for' directive.
//
// It assumes the next string is 'for'.
//
// It emits a For item.
// Then it emits an Ident and optionally a Comma followed by another Ident.
// Then it emits a Range followed by another Expression.
func (l *Lexer) for_() stateFn { //nolint:revive
	l.nextString("for")
	l.emit(For)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	endState := l.emitIdent(&UnknownItemError{Expected: "an identifier"})
	if endState != nil {
		return endState
	}

	spaceAfter := l.ignoreWhitespace()

	if l.peek() == ',' {
		l.next()
		l.emit(Comma)

		l.ignoreWhitespace()

		endState = l.emitIdent(&UnknownItemError{Expected: "an identifier"})
		if endState != nil {
			return endState
		}

		spaceAfter = l.ignoreWhitespace()
	}

	if !spaceAfter {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if !l.peekIsString("range") {
		return l.error(&UnknownItemError{Expected: "'range'"})
	}

	l.nextString("range")
	l.emit(Range)

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState = l._expression(false, '\n'); endState != nil {
		return endState
	}

	return l.nextHTML
}

// ============================================================================
// While
// ======================================================================================

// while consumes a while directive.
//
// It emits While followed by an Expression.
func (l *Lexer) while() stateFn {
	l.nextString("while")
	l.emit(While)

	endState := l._expression(false, '\n')
	if endState != nil {
		return endState
	}

	return l.nextHTML
}

// ============================================================================
// Elements
// ======================================================================================

// divOrDotBlock lexes a div in short form or a dot block.
//
// It assumes the next rune is either a '.' or a '#'.
func (l *Lexer) divOrDotBlock() stateFn {
	if l.next() != '.' { // it's a '#'
		l.backup()
		l.emit(Div)
		return l.behindElement
	}

	if l.isLineEmpty() {
		l.backup()
		return l.dotBlock
	}

	l.backup()
	l.emit(Div)
	return l.behindElement
}

// element lexes the name of an element.
//
// It assumes the next rune is the first rune of the element name.
//
// It emits an Element item.
func (l *Lexer) element() stateFn {
	endState := l.emitUntil(Element, &UnknownItemError{Expected: "an element name"},
		'=', ':', '.', '#', '(', '/', ' ', '\t', '\n')
	if endState != nil {
		return endState
	}

	return l.behindElement
}

// behindElement lexes the directives after the name of an element.
func (l *Lexer) behindElement() stateFn {
Next:
	for {
		switch l.peek() {
		case '.':
			l.next()
			if l.isLineEmpty() {
				l.backup()
				return l.dotBlock
			}

			l.backup()

			if endState := l._class(); endState != nil {
				return endState
			}
		case '#':
			if endState := l._id(); endState != nil {
				return endState
			}
		case '(':
			if endState := l._attributes(); endState != nil {
				return endState
			}
		default:
			break Next
		}
	}

	if l.peek() == '/' {
		l.nextString("/")
		l.emit(TagVoid)
	}

	// end states

	switch l.peek() {
	case eof:
		l.next()
		return l.eof
	case '\n':
		l.next()
		l.ignore()
		return l.nextHTML
	case '!', '=':
		return l.assign
	case ':':
		return l.blockExpansion
	case '.':
		return l.dotBlock
	case ' ', '\t':
		if l.isLineEmpty() {
			l.ignoreWhitespace()
			return l.newlineOrEOF(l.nextHTML)
		}

		return l.elementInlineText
	default:
		l.next()
		return l.error(&UnknownItemError{Expected: "a class, id, attribute, '=', '!=', ':', a newline, or a space"})
	}
}

// _class consumes a class directive.
//
// It assumes the next string is '.'.
//
// It emits a Class item followed by a Literal.
func (l *Lexer) _class() stateFn {
	l.nextString(".")
	l.emit(Class)

	endState := l.emitUntil(Literal, &UnknownItemError{Expected: "a class name"},
		'.', '#', '(', '[', '{', '=', ':', ' ', '\t', '\n')
	if endState != nil {
		return endState
	}

	return nil
}

// _id consumes an id directive.
//
// It assumes the next string is '#'.
//
// It emits an ID item followed by a Literal.
func (l *Lexer) _id() stateFn {
	l.nextString("#")
	l.emit(ID)

	endState := l.emitUntil(Literal, &UnknownItemError{Expected: "an id"},
		'.', '#', '[', '{', '(', '=', ':', ' ', '\t', '\n')
	if endState != nil {
		return endState
	}

	return nil
}

// _attributes consumes a list of attributes directive.
//
// It assumes the next string is '(' and hence emits an LParen.
//
// For each attribute, it emits an Ident, followed by an Assign and then an
// Expression.
// Each but the last, but optionally also the last, must be followed by a
// comma, which will also be emitted.
//
// Upon returning, it emits an RParen.
func (l *Lexer) _attributes() stateFn {
	l.nextString("(")
	l.emit(LParen)

	l.ignoreRunes(' ', '\t', '\n')

	// special case, no attributes
	if l.peek() == ')' {
		l.nextString(")")
		l.emit(RParen)
		return nil
	}

Attributes:
	for {
		if endState := l._attribute(); endState != nil {
			return endState
		}

		switch l.next() {
		case eof:
			return l.eof
		case ')':
			break Attributes
		case ',':
			l.emit(Comma)

			l.ignoreRunes(' ', '\t', '\n')

			// special case: trailing comma
			switch l.peek() {
			case eof:
				l.next()
				return l.eof
			case ')':
				l.nextString(")")
				break Attributes
			}

			// no trailing comma, continue
		default:
			return l.error(&UnknownItemError{Expected: "a comma, or a closing parenthesis"})
		}

		l.ignoreRunes(' ', '\t', '\n')
	}

	l.emit(RParen)
	return nil
}

// attributes consumes a single attribute.
//
// It assumes the next string is the name of the attribute.
func (l *Lexer) _attribute() stateFn {
	var parenCount int

Name:
	for {
		switch l.next() {
		case eof:
			return l.eof
		case '=', '!', ',':
			fallthrough
		case ' ', '\t', '\n':
			l.backup()
			break Name

		// Support angular attributes, e.g. '(click)'.
		// This is kinda 🥴, but I don't know of any lib/framework/whatever
		// that uses unmatched parentheses in their attributes.
		//
		// To the person who is reading this because they actually use
		// attributes that include unmatched parentheses: OtherFile an issue, thx.
		case '(':
			parenCount++
		case ')':
			parenCount--

			if parenCount < 0 {
				l.backup()
				break Name
			}
		}
	}

	if l.contentEmpty() {
		return l.error(&UnknownItemError{Expected: "an attribute name"})
	}

	l.emit(Ident)
	l.ignoreWhitespace()

	switch l.next() {
	case eof:
		return l.eof
	case ',', ')': // boolean attribute
		l.backup()
		return nil
	case '=':
		l.emit(Assign)
	case '!':
		if l.next() != '=' {
			return l.error(&UnknownItemError{Expected: "'='"})
		}

		l.emit(AssignNoEscape)
	default:
		return l.error(&UnknownItemError{Expected: "'=' or '!='"})
	}

	l.ignoreRunes(' ', '\t', '\n')
	if err := l._expression(true, ',', ')'); err != nil {
		return err
	}

	return nil
}

// blockExpansion parses a block expansion.
//
// It assumes the next rune is a ':'.
//
// It emits a BlockExpansion item followed by an element and optionally
// classes, ids and attributes.
func (l *Lexer) blockExpansion() stateFn {
	l.nextString(":")
	l.emit(BlockExpansion)

	if l.peek() == eof {
		l.next()
		return l.eof
	}

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	switch l.peek() {
	case eof:
		l.next()
		return l.eof
	case '.', '#':
		return l.divOrDotBlock()
	default:
		if l.peekIsString("block") {
			return l.block
		}
		return l.element
	}
}

// elementInlineText lexes the text that is on the same line as the element.
//
// It assumes the next rune is ' '.
//
// It emit at least one Text or Hash item.
// Multiple may be emitted, if the hash is used.
func (l *Lexer) elementInlineText() stateFn {
	l.nextString(" ")
	l.ignore()

	if endState := l._text(); endState != nil {
		return endState
	}

	return l.nextHTML
}

// ============================================================================
// And (&)
// ======================================================================================

// and consumes an '&' directive.
//
// It assumes the next string is '&'.
//
// It emits an And item.
func (l *Lexer) and() stateFn {
	l.nextString("&")
	l.emit(And)

	if l.peek() == '\n' {
		l.next()
		return l.error(&EOLError{In: "an &"})
	}

	for {
		switch l.peek() {
		case eof:
			l.next()
			return l.eof
		case '\n':
			l.next()
			l.ignore()
			return l.nextHTML
		case '.':
			if endState := l._class(); endState != nil {
				return endState
			}
		case '#':
			if endState := l._id; endState() != nil {
				return endState
			}
		case '(':
			if endState := l._attributes(); endState != nil {
				return endState
			}
		default:
			return l.error(&UnknownItemError{Expected: "a class, id, or attribute"})
		}
	}
}

// ============================================================================
// Dot Block
// ======================================================================================

// dotBlock lexes a dot block.
//
// It assumes the next rune is '.'.
func (l *Lexer) dotBlock() stateFn {
	l.nextString(".")
	l.emit(DotBlock)

	if endState := l.newlineOrEOF(nil); endState != nil {
		return endState
	}

	dIndent, _, err := l.consumeIndent(singleIncrease)
	if err != nil {
		return l.error(err)
	}

	l.emitIndent(dIndent)

	if dIndent <= 0 {
		return l.nextHTML
	}

	for {
		l.emit(DotBlockLine)

		if endState := l._text(); endState != nil {
			return endState
		}

		if l.peek() == eof {
			l.next()
			return l.eof
		}

		dIndent, skippedLines, err := l.consumeIndent(noIncrease)
		if err != nil {
			return l.error(err)
		}

		l.emitIndent(dIndent)

		if dIndent < 0 {
			return l.nextHTML
		}

		for i := 0; i < skippedLines; i++ {
			l.emit(DotBlockLine)
		}
	}
}

// ============================================================================
// Pipe
// ======================================================================================

// pipe lexes a single pipe.
//
// It assumes the next rune is a '|'.
//
// It emits a Pipe item followed optionally by a Text or Hash.
// Multiple Text/Hash items may be emitted, if the hash is used.
func (l *Lexer) pipe() stateFn {
	l.nextString("|")
	l.emit(Pipe)

	if l.isLineEmpty() {
		return l.newlineOrEOF(l.nextHTML)
	}

	if l.next() != ' ' {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	l.ignore()

	if endState := l._text(); endState != nil {
		return endState
	}

	return l.nextHTML
}

// ============================================================================
// Assign
// ======================================================================================

// assign lexes an expression assignment to an element.
//
// It assumes the next rune is '!' or '='.
//
// It emits an Assign or AssignUnsescaped followed by an Expression.
func (l *Lexer) assign() stateFn {
	if l.next() == '!' {
		if l.next() != '=' {
			return l.error(&UnknownItemError{Expected: "'='"})
		}

		l.emit(AssignNoEscape)
	} else {
		l.emit(Assign)
	}

	if !l.ignoreWhitespace() {
		return l.error(&UnknownItemError{Expected: "a space"})
	}

	if endState := l._expression(true, '\n'); endState != nil {
		return endState
	}

	return l.nextHTML
}

// ============================================================================
// Filter
// ======================================================================================

// filter lexes a filter directive.
//
// It assumes the next rune is a ':'.
//
// It emits a Filter item followed by an Ident, the name of the filter.
// It then emits zero, one or multiple Literals representing the individual
// arguments.
// Each Literal is either a string, as denoted by its '"', or '`' prefix, or
// a regular text.
//
// Lastly, it emits zero, one, or multiple Text items.
func (l *Lexer) filter() stateFn {
	l.nextString(":")
	l.emit(Filter)

	endState := l.emitUntil(Ident, &UnknownItemError{Expected: "the name of the filter"}, ' ', '\t', '\n')
	if endState != nil {
		return endState
	}

Args:
	for {
		l.ignoreWhitespace()

		switch l.peek() {
		case eof:
			l.next()
			return l.eof
		case '\n':
			l.nextString("\n")
			l.ignore()
			break Args
		}

		if p := l.peek(); p == '"' || p == '`' {
			if endState = l._string(); endState != nil {
				return endState
			}

			l.emit(Literal)
		} else {
			if endState = l.emitUntil(Literal, nil, ' ', '\t', '\n'); endState != nil {
				return endState
			}
		}
	}

	if l.peek() == eof {
		l.next()
		return l.eof
	}

	dIndent, _, err := l.consumeIndent(singleIncrease)
	if err != nil {
		return l.error(err)
	}

	l.emitIndent(dIndent)

	if dIndent <= 0 {
		return l.nextHTML
	}

	for dIndent >= 0 {
		peek := l.nextUntil('\n')
		if peek == eof {
			if !l.contentEmpty() {
				l.emit(Text)
			}

			return l.eof
		}

		// empty lines are valid
		l.emit(Text)

		l.nextString("\n")
		l.ignore()

		dIndent, _, err = l.consumeIndent(noIncrease)
		if err != nil {
			return l.error(err)
		}

		l.emitIndent(dIndent)
	}

	return l.nextHTML()
}

// ============================================================================
// Mixin Call
// ======================================================================================

// _mixinCall lexes a mixin call.
//
// It assumes the next rune is '+'.
//
// It emits a MixinCall item, followed by an Ident.
//
// Optionally, it also emits the following:
// An LParen and then zero, one, or multiple parameters and finally a RParen.
//
// Regardless of whether the parameters were emitted, it may also emit a
// MixinBlockShortcut item.
//
// Each parameter consists of an Ident, an Assign, and an Expression.
func (l *Lexer) _mixinCall(allowNewlines bool) stateFn {
	l.nextString("+")
	l.emit(MixinCall)

	endState := l.emitIdent(&UnknownItemError{Expected: "a mixin name"})
	if endState != nil {
		return endState
	}

	if l.peek() == '.' { // this was just the namespace
		l.nextString(".")
		l.ignore()

		endState = l.emitIdent(&UnknownItemError{Expected: "a mixin name"})
		if endState != nil {
			return endState
		}

		switch l.next() {
		case eof:
			return l.eof
		case '\n':
			return nil
		default:
			l.backup()
		}
	}

	l.ignoreWhitespace()

	if l.peek() == '(' {
		l.nextString("(")
		l.emit(LParen)

		if allowNewlines {
			l.ignoreRunes(' ', '\t', '\n')
		} else {
			l.ignoreWhitespace()
		}

		// special case: no args
		if l.peek() == ')' {
			l.nextString(")")
		} else {
		Args:
			for {
				if endState = l._mixinArg(allowNewlines); endState != nil {
					return endState
				}

				switch l.next() {
				case eof:
					return l.eof
				case '\n':
					return l.error(&EOLError{In: "mixin arguments"})
				case ',':
					l.emit(Comma)

					if allowNewlines {
						l.ignoreRunes(' ', '\t', '\n')
					} else {
						l.ignoreWhitespace()
					}

					// special case: trailing comma before RParen
					switch l.next() {
					case eof:
						return l.eof
					case ')':
						break Args
					}

					l.backup()

					// no trailing comma, continue
				case ')':
					break Args
				}

				if allowNewlines {
					l.ignoreRunes(' ', '\t', '\n')
				} else {
					l.ignoreWhitespace()
				}
			}
		}

		l.emit(RParen)
	}

	if !allowNewlines {
		return nil
	}

	l.ignoreWhitespace()
	if l.peek() != '>' {
		return l.newlineOrEOF(nil)
	}

	l.nextString(">")
	l.emit(MixinBlockShortcut)

	switch l.peek() {
	case eof:
		l.next()
		return l.eof
	case '.':
		return l.dotBlock
	case ' ', '\t':
		return l.elementInlineText
	default:
		return l.newlineOrEOF(nil)
	}
}

func (l *Lexer) _mixinArg(allowNewlines bool) stateFn {
	endState := l.emitIdent(&UnknownItemError{Expected: "a mixin parameter name"})
	if endState != nil {
		return endState
	}

	l.ignoreWhitespace()

	switch l.next() {
	case eof:
		return l.eof
	case '\n':
		return l.error(&EOLError{In: "mixin arguments"})
	case '!':
		if l.next() != '=' {
			return l.error(&UnknownItemError{Expected: "'='"})
		}

		l.emit(AssignNoEscape)
	case '=':
		l.emit(Assign)
	default:
		return l.error(&UnknownItemError{Expected: "'='"})
	}

	if allowNewlines {
		l.ignoreRunes(' ', '\t', '\n')
	} else {
		l.ignoreWhitespace()
	}

	if l.peek() == eof {
		l.next()
		return l.eof
	}

	return l._expression(allowNewlines, ',', ')')
}

// ============================================================================
// Helpers
// ======================================================================================

// _text lexes a single line of text.
//
// It emits at least one Text item.
// Multiple may be emitted, if the current line makes use of the hash operator.
func (l *Lexer) _text() stateFn {
	for {
		if l.isLineEmpty() {
			l.ignoreWhitespace()
			return l.newlineOrEOF(nil)
		}

		peek := l.nextUntil('#', '\n')

		if l.peekIsString("##") { // hash escape
			l.nextString("##")
			continue
		}

		if !l.contentEmpty() {
			l.emit(Text)
		}

		if peek == eof || peek == '\n' {
			return l.newlineOrEOF(nil)
		}

		if endState := l._hash(); endState != nil {
			return endState
		}
	}
}

// _hash lexes a hash expression.
//
// It assumes the next rune is '#', but the rune following it is not '#'.
func (l *Lexer) _hash() stateFn {
	l.nextString("#")
	l.emit(Hash)

	peek := l.peek()
	switch peek {
	case '+':
		return l._mixinCall(false)
	case '!':
		l.nextString("!")
		l.emit(NoEscape)
		peek = l.peek()
	}

	if peek == '[' {
		l.nextString("[")
		l.emit(LBracket)

		endState := l.emitUntil(Text, &UnknownItemError{Expected: "text"}, ']', '\n')
		if endState != nil {
			return endState
		}

		switch l.next() {
		case eof:
			return l.eof
		case '\n':
			return l.error(&EOLError{In: "inline text"})
		}

		l.emit(RBracket)
		return nil
	}

	if peek != '{' {
		return l._inlineElement()
	}

	l.nextString("{")
	l.emit(LBrace)

	if endState := l._expression(false, '}'); endState != nil {
		return endState
	}

	if l.next() == eof {
		return l.eof
	}

	l.emit(RBrace)
	return nil
}

// _inlineElement lexes an inline element.
//
// It assumes the next runes are the name of the element.
func (l *Lexer) _inlineElement() stateFn {
	switch l.peek() {
	case '.', '#':
		l.emit(Div)
	default:
		endState := l.emitUntil(Element, &UnknownItemError{Expected: "an element name"},
			'{', '[', '#', '.', '(', ' ', '\t', '\n')
		if endState != nil {
			return endState
		}
	}

	for {
		switch l.peek() {
		case '[':
			l.nextString("[")
			l.emit(LBracket)

			endState := l.emitUntil(Text, nil, ']', '\n')
			if endState != nil {
				return endState
			}

			switch l.next() {
			case eof:
				return l.eof
			case '\n':
				return l.error(&EOLError{In: "inline element text"})
			default:
				l.emit(RBracket)
				return nil
			}
		case '{':
			l.nextString("{")
			l.emit(LBrace)

			endState := l._expression(false, '}')
			if endState != nil {
				return endState
			}

			if l.next() == eof {
				return l.eof
			}

			l.emit(RBrace)
			return nil
		case '.':
			if endState := l._class(); endState != nil {
				return endState
			}
		case '#':
			if endState := l._id(); endState != nil {
				return endState
			}
		case '(':
			if endState := l._attributes(); endState != nil {
				return endState
			}
		default:
			return l.error(&UnknownItemError{Expected: "a class, id, attribute, a '{', or a '['"})
		}
	}
}
