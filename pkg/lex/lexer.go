package lex

import (
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

// NewLexer creates a new lexer.
func NewLexer(in string) *Lexer {
	return &Lexer{
		in:    in,
		items: make(chan Item),
		line:  1,
		col:   1,
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

// Run starts a new goroutine that lexes the input.
// The lexical items can be retrieved by calling Next.
func (l *Lexer) Run() {
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
		l.col += w
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
func (l *Lexer) nextString(s string) {
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
				break Next
			}
		}
	}

	l.backup()
	return
}

// ======================================= backup =======================================

// backup goes back one rune as if the previous next call never happened.
// backup can only be called once per next.
func (l *Lexer) backup() {
	l.pos -= l.prevWidth
	l.col -= l.prevWidth

	if l.col <= 0 {
		l.col = l.prevLineLen
		l.line--
	}
}

// ======================================== peek ========================================

// peek is the same as next but without increasing position.
func (l *Lexer) peek() rune {
	defer l.backup()
	return l.next()
}

// peekIsString peeks ahead to see if the next runes match s.
// If so it returns true.
// If not or if the remaining input is less than s's length, false is returned.
func (l *Lexer) peekIsString(s string) bool {
	endPos := l.pos + len(s)
	if endPos >= len(l.in) {
		return false
	}

	return l.in[l.pos:endPos] == s
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
func (l *Lexer) ignoreRunes(rs ...rune) {
	l.nextRunes(rs...)
	l.ignore()
}

// ignoreWhitespace is short for l.ignoreRunes(' ', '\t').
func (l *Lexer) ignoreWhitespace() {
	l.ignoreRunes(' ', '\t')
}

// ignoreUntil ignores all upcoming runes until one matching r is found or the
// end of file is encountered.
// Upon returning, the next() rune will be one of rs or eof.
func (l *Lexer) ignoreUntil(rs ...rune) {
	for {
		next := l.next()
		if next == eof {
			break
		}

		for _, r := range rs {
			if next == r {
				break
			}
		}
	}

	l.backup()
	l.ignore()
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
	// all consumes all indentation changes.
	// If it encounters an increase by more than one level, it returns
	// ErrIndentationIncrease.
	all
)

// consumeIndent consumes the upcoming indentation.
// For that it first skips all empty lines.
// Then, when it encounters a line with content, it consumes all indentation
// before it and determines the delta in indentation, which it returns.
// This will either be a number negative number indicating by how many numbers
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
func (l *Lexer) consumeIndent(mode indentConsumptionMode) (dlvl int, err error) {
	if l.col != 1 {
		panic("consumeIndent called with non-start-of-line position")
	}

	for {
		if mode == all {
			dlvl, err, stop := l.consumeAllLineIndent()
			if stop {
				return dlvl, err
			}
		} else {
			dlvl, err, stop := l.consumeOtherLineIndent(mode)
			if stop {
				return dlvl, err
			}
		}
	}
}

func (l *Lexer) consumeAllLineIndent() (dlvl int, err error, stop bool) {
	indent := l.next()
	switch indent {
	case eof:
		return 0, nil, true
	case '\n': // empty line
		return 0, nil, false
	case ' ', '\t': // indent
		// handled below
	default: // no indent
		dlvl = -l.indentLvl
		l.indentLvl = 0

		l.backup()
		return dlvl, nil, true
	}

	count := 1

	for n := l.next(); n == ' ' || n == '\t'; n = l.next() {
		if n != indent {
			return 0, ErrMixedIndentation, true
		}

		count++
	}

	l.backup()

	switch l.peek() {
	case eof:
		return 0, nil, true
	case '\n': // empty line
		return 0, nil, false
	}

	// this is our first indent, use it for reference
	if l.indent == 0 {
		l.indent = indent
		l.indentLen = count
		l.indentRefLine = l.line

		l.indentLvl = 1
		return 1, nil, true
	}

	// illegal switching between tabs and spaces or inconsistent indentation
	if indent != l.indent || count%l.indentLen != 0 {
		return 0, &IndentationError{
			Expect:    l.indent,
			ExpectLen: l.indentLen,
			RefLine:   l.indentRefLine,
			Actual:    indent,
			ActualLen: count,
		}, true
	}

	newLvl := count / l.indentLen
	dlvl = newLvl - l.indentLvl
	if dlvl > 1 {
		return 0, ErrIndentationIncrease, true
	}

	l.indentLvl = newLvl
	return dlvl, nil, true
}

func (l *Lexer) consumeOtherLineIndent(mode indentConsumptionMode) (dlvl int, err error, stop bool) {
	l.ignore()

	indent := l.next()
	switch indent {
	case eof:
		return 0, nil, true
	case '\n': // empty line
		return 0, nil, false
	case ' ', '\t': // indent
		// handled below
	default: // no indent
		dlvl = -l.indentLvl
		l.indentLvl = 0

		l.backup()
		return dlvl, nil, true
	}

	count := 1

	limit := l.indentLvl*l.indentLen - 1
	if mode == singleIncrease {
		limit += l.indentLen
	}

	for i := 0; i < limit; i++ {
		switch n := l.next(); n {
		case eof:
			return 0, nil, true
		case ' ', '\t':
			if n != indent {
				return 0, ErrMixedIndentation, true
			}

			count++
		default:
			break
		}
	}

	switch l.peek() {
	case eof:
		return 0, nil, true
	case '\n': // empty line
		return 0, nil, false
	case ' ', '\t': // possibly empty line
	Peek:
		for i := l.pos + 1; i < len(l.in); i++ {
			switch l.in[i] {
			case '\n':
				return 0, nil, false
			case ' ', '\t':
			default: // not an empty line
				break Peek
			}
		}

	}

	// this is our first indent, use it for reference
	if l.indent == 0 {
		l.indent = indent
		l.indentLen = count
		l.indentRefLine = l.line

		l.indentLvl = 1
		return 1, nil, true
	}

	// illegal switching between tabs and spaces or inconsistent indentation
	if indent != l.indent || count%l.indentLen != 0 {
		return 0, &IndentationError{
			Expect:    l.indent,
			ExpectLen: l.indentLen,
			RefLine:   l.indentRefLine,
			Actual:    indent,
			ActualLen: count,
		}, true
	}

	newLvl := count / l.indentLen
	dlvl = newLvl - l.indentLvl
	if dlvl > 1 {
		return 0, ErrIndentationIncrease, true
	}

	l.indentLvl = newLvl
	return dlvl, nil, true
}

// ======================================== emit ========================================

// emit sends the lexed contents in the items channel, starting at startPos.
// After emitting, startPos is updated to the current position.
func (l *Lexer) emit(t ItemType) {
	l.items <- Item{
		Type: t,
		Val:  string(l.in[l.startPos:l.pos]),
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
	if delta == 0 {
		return
	}

	typ := Indent
	if delta < 0 {
		delta = -delta
		typ = Dedent
	}

	for i := 0; i < delta; i++ {
		l.items <- Item{
			Type: typ,
			Val:  string(l.indent),
			Line: l.startLine,
			Col:  l.startCol,
		}
	}

	l.startPos = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// emitError sends an Error item in the items channel using the result of
// calling fmt.Sprintf with the passed arguments as value.
//
// It then returns a nil stateFn, indicating the lexer should stop.
func (l *Lexer) error(err error) stateFn {
	l.items <- Item{
		Type: Error,
		Err:  err,
		Line: l.startLine,
		Col:  l.startLine,
	}

	return nil
}

// eof emits an EOF item for the current position, not the starting postion.
func (l *Lexer) eof() stateFn {
	l.items <- Item{
		Type: EOF,
		Line: l.line,
		Col:  l.col - 1,
	}

	return nil
}
