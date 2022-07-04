package lex

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	// in is the text being read.
	in string
	// items is the channel emitting the items.
	items chan Item

	// startLine and startCol are the line and column at which the current item
	// starts.
	startLine, startCol int
	// line and col are the line and column of the next rune.
	line, col int

	// prevWith is the width of the previous rune.
	prevWidth int
	// prevLineLen is a helper used by backup to set the correct col when
	// backing a line up.
	prevLineLen int

	// startPos is the position at which the current item starts.
	startPos int
	// pos is the index of the next rune.
	pos int

	// indent represents the character that is used for indentation.
	// It is either a space or a tab.
	indent rune
	// indentLen is the number of times indent is used to form a single indent
	// level.
	indentLen int
	// indentRefLine is the line used for reference for indentation.
	// It is the first line that uses indentation.
	//
	// This variable is only used for errors and has no practical use.
	indentRefLine int
	indentLvl     int
}

// eof is the rune returned by next, peek and alike to indicate the end of the
// input was reached.
const eof rune = -1

type stateFn func() stateFn

// New creates a new lexer.
func New(in string) *Lexer {
	return &Lexer{
		in:        in,
		items:     make(chan Item),
		startLine: 1,
		line:      1,
		startCol:  1,
		col:       1,
	}
}

// ============================================================================
// Exported
// ======================================================================================

// Next returns the next lexical item.
func (l *Lexer) Next() Item {
	return <-l.items
}

// Stop stops the lexer's goroutine by draining all input elements.
func (l *Lexer) Stop() {
	for range l.items {
		// drain
	}
}

// Lex starts a new goroutine that lexes the input.
// The lexical items can be retrieved by calling Next.
func (l *Lexer) Lex() {
	go func() {
		state := l.start

		for state != nil {
			state = state()
		}

		close(l.items)
	}()
}

// ============================================================================
// Helpers
// ======================================================================================

// ======================================== next ========================================

// next returns the next rune from the input.
// It increments pos and updates line and col, so that the next call to
// next will yield the rune after the current.
//
// If there are no more characters, eof is returned.
func (l *Lexer) next() rune {
	if l.pos >= len(l.in) {
		return eof
	}

	r, w := utf8.DecodeRuneInString(l.in[l.pos:])
	if r == '\n' {
		l.line++
		l.prevLineLen = l.col
		l.col = 1
	} else {
		l.col++
	}

	l.prevWidth = w
	l.pos += w

	return r
}

// nextUntil calls next until it encounters a rune that matches r, or it
// reaches the end of file.
//
// If the end of file is reached, nextUntil will return eof and the current
// rune will be eof.
//
// Otherwise, nextUntil will return the rune it encountered.
// Upon returning the next call to next will return one of rs.
func (l *Lexer) nextUntil(rs ...rune) rune {
	var next rune

Next:
	for {
		next = l.next()
		if next == eof {
			return eof
		}

		for _, r := range rs {
			if r == next {
				break Next
			}
		}
	}

	l.backup()
	return next
}

// nextString assumes that the peekIsString(s) evaluates to true and skips
// the upcoming occurrence of s, i.e. it advances the position so that the next
// rune will be the rune after s.
//
// s must not contain any newlines.
// The only exception is s being just a newline, i.e. s = "\n".
func (l *Lexer) nextString(s string) {
	if s == "\n" {
		l.pos++
		l.line++
		l.prevLineLen = l.col
		l.col = 1

		return
	}

	l.pos += len(s)
	l.col += len(s)
}

// nextRunes calls next so long as the next rune matches rs or the end of file
// is reached.
// Upon returning the next rune will be the one after the last occurrence of
// one of rs's runes.
// If the end of file was reached, the current rune will be eof.
func (l *Lexer) nextRunes(rs ...rune) {
Next:
	for next := l.next(); ; next = l.next() {
		if next == eof {
			return
		}

		for _, r := range rs {
			if r == next {
				continue Next
			}
		}

		break
	}

	l.backup()
}

// consumeIdent consumes an upcoming Ident.
//
// An Ident starts with a unicode letter or underscore.
// It is followed by any number of unicode letters, decimal digits, or
// underscores.
// This is the same pattern as Go uses for its identifiers.
func (l *Lexer) consumeIdent() {
	next := l.next()
	if next != '_' && !unicode.IsLetter(next) {
		l.backup()
		return
	}

	for {
		next = l.next()
		switch {
		case next == eof:
			return
		case next == '_':
		case next >= '0' && next <= '9':
		case unicode.IsLetter(next):
		default:
			l.backup()
			return
		}
	}
}

// ======================================= backup =======================================

// backup goes back one rune as if the previous next call never happened.
// backup can only be called once per next.
func (l *Lexer) backup() {
	l.pos -= l.prevWidth
	l.col--

	if l.col <= 0 {
		l.col = l.prevLineLen
		l.line--
	}
}

// ======================================== peek ========================================

// peek is the same as next but without increasing position.
func (l *Lexer) peek() rune {
	if l.pos >= len(l.in) {
		return eof
	}

	defer l.backup()
	return l.next()
}

// peekIsString peeks ahead to see if the next runes match s.
// If so it returns true.
// If not or if the remaining input is less than s's length, false is returned.
func (l *Lexer) peekIsString(s string) bool {
	endPos := l.pos + len(s)
	if endPos > len(l.in) {
		return false
	}

	return l.in[l.pos:endPos] == s
}

// peekIsWord peeks ahead to see if the next runes match s.
// Additionally, it asserts that after the occurrence of s, a space, tab,
// newline, or EOF must follow.
//
// If either of these conditions is not met, false is returned.
func (l *Lexer) peekIsWord(s string) bool {
	if !l.peekIsString(s) {
		return false
	}

	afterIndex := l.pos + len(s)

	if len(l.in) <= afterIndex { // eof
		return true
	}

	after := l.in[afterIndex]
	return after == ' ' || after == '\t' || after == '\n'
}

// ======================================= ignore =======================================

// ignore ignores the input up to this point.
// It sets the start to the current position.
func (l *Lexer) ignore() {
	l.startPos = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// ignoreRunes ignores all upcoming runes that match rs.
// Upon returning, the next() rune will be the rune next to the last of the
// sequence of rs.
//
// If the end of file is reached, ignoreRunes will stop at the end of file and
// return.
// The current rune will be eof.
//
// It returns true if at least one rune was ignored.
func (l *Lexer) ignoreRunes(rs ...rune) bool {
	l.ignore()
	l.nextRunes(rs...)

	ignored := l.pos > l.startPos
	l.ignore()

	return ignored
}

// ignoreWhitespace is short for l.ignoreRunes(' ', '\t').
func (l *Lexer) ignoreWhitespace() bool {
	return l.ignoreRunes(' ', '\t')
}

// ignoreUntil ignores all upcoming runes until one matching r is found or the
// end of file is encountered.
//
// If the end of file is reached, ignoreUntil will stop with the end of file
// as current rune return.
//
// Otherwise, upon returning, the next() rune will be one of rs .
func (l *Lexer) ignoreUntil(rs ...rune) {
Next:
	for {
		next := l.next()
		if next == eof {
			l.ignore()
			return
		}

		for _, r := range rs {
			if next == r {
				break Next
			}
		}
	}

	l.backup()
	l.ignore()
}

// ==================================== isLineEmpty =====================================

// isLineEmpty checks if the rest of the current line is empty.
func (l *Lexer) isLineEmpty() bool {
	for i := l.pos; i < len(l.in); i++ {
		switch l.in[i] {
		case '\n':
			return true
		case ' ', '\t':
		default: // not an empty line
			return false
		}
	}

	return true
}

// ====================================== content =======================================

// contentEmpty reports whether any content has been lexed.
func (l *Lexer) contentEmpty() bool {
	return l.pos <= l.startPos
}

// ======================================= Indent =======================================

type indentConsumptionMode uint8

const (
	// noIncrease consumes indentation with a negative or no change in
	// indentation.
	noIncrease indentConsumptionMode = iota + 1
	// singleIncrease consumes indentation with a negative change, no change, or an
	// increase by one.
	// It does not return an error when it encounters more than one increase
	// in indentation levels.
	singleIncrease
	// allIndents consumes all indentation changes.
	// If it encounters an increase by more than one level, it returns
	// ErrIndentationIncrease.
	allIndents
)

// consumeIndent consumes the upcoming indentation.
// For that it first skips all empty lines.
// Then, when it encounters a line with content, it consumes all indentation
// before it and determines the delta in indentation, which it returns.
// This will either be a negative number indicating by how many levels
// the indention was removed, 0 to indicate no change in indentation, or 1 to
// indicate an increase in indention.
//
// Callers can use mode to indicate how much indentation should be consumed.
// Refer to the documentation of the mode constants for more information.
//
// When calling, the current position must be at the beginning of a line.
// Otherwise, consumeIndent will panic.
//
// If consumeIndent encounters the end of file, it will return 0 and nil.
func (l *Lexer) consumeIndent(mode indentConsumptionMode) (dlvl int, skippedLines int, err error) {
	if l.col != 1 {
		panic(fmt.Sprintf("consumeIndent called with non-start-of-line position (%d:%d)", l.line, l.col))
	}

	l.ignore()

	switch mode {
	case noIncrease:
		for {
			dlvl, stop, err := l.consumeNoIncreaseLineIndent()
			if stop {
				return dlvl, skippedLines, err
			}

			skippedLines++
		}
	case singleIncrease:
		for {
			dlvl, stop, err := l.consumeSingleIncreaseLineIndent()
			if stop {
				return dlvl, skippedLines, err
			}

			skippedLines++
		}
	case allIndents:
		for {
			dlvl, stop, err := l.consumeAllLineIndent()
			if stop {
				return dlvl, skippedLines, err
			}

			skippedLines++
		}
	default:
		panic(fmt.Sprintf("unknown indent consumption mode: %d", mode))
	}
}

func (l *Lexer) consumeNoIncreaseLineIndent() (dlvl int, stop bool, err error) {
	indent := l.next()
	switch indent {
	case eof:
		return 0, true, nil
	case '\n': // empty line
		return 0, false, nil
	case ' ', '\t': // indent
		// handled below
	default: // no indent
		dlvl = -l.indentLvl
		l.indentLvl = 0

		l.backup()
		return dlvl, true, nil
	}

	if l.indentLvl == 0 {
		l.backup()

		if l.isLineEmpty() {
			l.nextUntil('\n')
			l.next()
			return 0, false, nil
		}

		return 0, true, nil
	}

	count := 1

	limit := l.indentLvl*l.indentLen - 1

Next:
	for i := 0; i < limit; i++ {
		switch next := l.next(); next {
		case eof:
			return 0, true, nil
		case ' ', '\t':
			if next != indent {
				return 0, true, ErrMixedIndentation
			}

			count++
		default:
			l.backup()
			break Next
		}
	}

	switch l.peek() {
	case eof:
		return 0, true, nil
	case '\n': // empty line
		l.next()
		return 0, false, nil
	case ' ', '\t': // possibly empty line
		if l.isLineEmpty() {
			l.nextUntil('\n')
			l.next()
			return 0, false, nil
		}
	}

	// illegal switching between tabs and spaces or inconsistent indentation
	if indent != l.indent || count%l.indentLen != 0 {
		return 0, true, &IndentationError{
			Expect:    l.indent,
			ExpectLen: l.indentLen,
			RefLine:   l.indentRefLine,
			Actual:    indent,
			ActualLen: count,
		}
	}

	newLvl := count / l.indentLen
	dlvl = newLvl - l.indentLvl

	l.indentLvl = newLvl
	return dlvl, true, nil
}

func (l *Lexer) consumeSingleIncreaseLineIndent() (dlvl int, stop bool, err error) {
	indent := l.next()
	switch indent {
	case eof:
		return 0, true, nil
	case '\n': // empty line
		return 0, false, nil
	case ' ', '\t': // indent
		// handled below
	default: // no indent
		dlvl = -l.indentLvl
		l.indentLvl = 0

		l.backup()
		return dlvl, true, nil
	}

	// this is our first indent, use it for reference
	if l.indent == 0 {
		count := 1

		for n := l.next(); n == ' ' || n == '\t'; n = l.next() {
			if n != indent {
				return 0, true, ErrMixedIndentation
			}

			count++
		}

		if l.peek() == eof {
			return 0, true, nil
		}

		l.backup()
		if l.peek() == '\n' { // empty line
			return 0, false, nil
		}

		l.indent = indent
		l.indentLen = count
		l.indentRefLine = l.line
		l.indentLvl = 1

		return 1, true, nil
	}

	count := 1

	limit := (l.indentLvl+1)*l.indentLen - 1

Next:
	for i := 0; i < limit; i++ {
		switch next := l.next(); next {
		case eof:
			return 0, true, nil
		case ' ', '\t':
			if next != indent {
				return 0, true, ErrMixedIndentation
			}

			count++
		default:
			l.backup()
			break Next
		}
	}

	switch l.peek() {
	case eof:
		return 0, true, nil
	case '\n': // empty line
		l.next()
		return 0, false, nil
	case ' ', '\t': // possibly empty line
		if l.isLineEmpty() {
			l.nextUntil('\n')
			l.next()
			return 0, false, nil
		}
	}

	// illegal switching between tabs and spaces or inconsistent indentation
	if indent != l.indent || count%l.indentLen != 0 {
		return 0, true, &IndentationError{
			Expect:    l.indent,
			ExpectLen: l.indentLen,
			RefLine:   l.indentRefLine,
			Actual:    indent,
			ActualLen: count,
		}
	}

	newLvl := count / l.indentLen
	dlvl = newLvl - l.indentLvl

	l.indentLvl = newLvl
	return dlvl, true, nil
}

func (l *Lexer) consumeAllLineIndent() (dlvl int, stop bool, err error) {
	indent := l.next()
	switch indent {
	case eof:
		return 0, true, nil
	case '\n': // empty line
		return 0, false, nil
	case ' ', '\t': // indent
		// handled below
	default: // no indent
		dlvl = -l.indentLvl
		l.indentLvl = 0

		l.backup()
		return dlvl, true, nil
	}

	count := 1

	for next := l.next(); next == ' ' || next == '\t'; next = l.next() {
		if next != indent {
			return 0, true, ErrMixedIndentation
		}

		count++
	}

	if l.peek() == eof {
		return 0, true, nil
	}

	l.backup()
	if l.peek() == '\n' { // empty line
		return 0, false, nil
	}

	// this is our first indent, use it for reference
	if l.indent == 0 {
		l.indent = indent
		l.indentLen = count
		l.indentRefLine = l.line

		l.indentLvl = 1
		return 1, true, nil
	}

	// illegal switching between tabs and spaces or inconsistent indentation
	if indent != l.indent || count%l.indentLen != 0 {
		return 0, true, &IndentationError{
			Expect:    l.indent,
			ExpectLen: l.indentLen,
			RefLine:   l.indentRefLine,
			Actual:    indent,
			ActualLen: count,
		}
	}

	newLvl := count / l.indentLen
	dlvl = newLvl - l.indentLvl
	if dlvl > 1 {
		return 0, true, ErrIndentationIncrease
	}

	l.indentLvl = newLvl
	return dlvl, true, nil
}

// ==================================== newlineOrEOF ====================================

// newlineOrEOF skips all whitespaces.
// It then asserts that the next rune is either a newline or the end of file.
//
// If it encounters the end of file, it returns l.eof.
//
// If it encounters a newline, it consumes and ignores it and returns next.
//
// If it encounters another character, it returns with l.error.
func (l *Lexer) newlineOrEOF(next stateFn) stateFn {
	l.ignoreWhitespace()

	switch l.next() {
	case eof:
		return l.eof
	case '\n':
		l.ignore()
		return next
	default:
		return l.error(&UnknownItemError{Expected: "a newline"})
	}
}

// ======================================== emit ========================================

// emit sends the lexed contents in the items channel, starting at startPos.
// After emitting, startPos is updated to the current position.
func (l *Lexer) emit(t ItemType) {
	l.items <- Item{
		Type: t,
		Val:  l.in[l.startPos:l.pos],
		Line: l.startLine,
		Col:  l.startCol,
	}

	l.startPos = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// emitIndent emits the change in indentation level.
//
// If delta is 0, emitIndent will do nothing.
func (l *Lexer) emitIndent(delta int) {
	if delta != 0 {
		typ := Indent
		if delta < 0 {
			delta = -delta
			typ = Dedent
		}

		for i := 0; i < delta; i++ {
			l.items <- Item{
				Type: typ,
				Line: l.startLine,
				Col:  l.startCol,
			}
		}
	}

	l.startPos = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// emitIdent calls consumeIdent and then does one of the following:
//
// If the content is not empty, it emits an item of type t.
// If it is empty and ifEmptyErr is set and the end of file is not reached,
// ifEmptyErr is emitted.
func (l *Lexer) emitIdent(ifEmptyErr error) stateFn {
	l.consumeIdent()

	empty := l.contentEmpty()
	if !empty {
		l.emit(Ident)
	}

	if l.next() == eof {
		return l.eof
	}

	l.backup()

	if ifEmptyErr != nil && empty {
		return l.error(ifEmptyErr)
	}

	return nil
}

// emitUntil consumes all runes until it reaches one of rs or the end of file.
//
// If the content is not empty, it emits an item of type t.
// If it is empty and ifEmptyErr is set and the end of file is not reached,
// ifEmptyErr is emitted.
//
// If the next rune is one of rs, it returns nil.
// The next call rule will yield that rune.
//
// If the next rune is the end of file, it returns l.eof.
func (l *Lexer) emitUntil(t ItemType, ifEmptyErr error, rs ...rune) stateFn {
	peek := l.nextUntil(rs...)

	empty := l.contentEmpty()
	if !empty {
		l.emit(t)
	}

	if peek == eof {
		return l.eof
	}

	if ifEmptyErr != nil && empty {
		return l.error(ifEmptyErr)
	}

	return nil
}

// emitError returns a function that sends an Error item in the items channel
// for the current position.
//
// It then returns a nil stateFn, indicating the lexer should stop.
func (l *Lexer) error(err error) stateFn {
	if err == nil {
		panic("error called with nil error")
	}

	return func() stateFn {
		i := Item{
			Type: Error,
			Err:  err,
			Line: l.line,
			Col:  l.col - 1,
		}

		if i.Col == 0 {
			i.Line--
			i.Col = l.prevLineLen
		}

		l.items <- i

		return nil
	}
}

// eof emits an EOF item for the current position, not the starting position.
func (l *Lexer) eof() stateFn {
	l.items <- Item{
		Type: EOF,
		Line: l.line,
		Col:  l.col - 1,
	}

	return nil
}
