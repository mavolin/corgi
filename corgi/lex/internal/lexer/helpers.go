package lexer

import (
	"unicode/utf8"
)

// ============================================================================
// Pos
// ======================================================================================

// Line returns the line number of the rune last returned by Next.
func (l *Lexer[Token]) Line() int {
	if l.col-1 <= 0 {
		return l.line - 1
	}

	return l.line
}

// Col returns the column number of the last returned by Next.
func (l *Lexer[Token]) Col() int {
	col := l.col - 1

	if col <= 0 {
		col = l.prevLineLen
	}

	return col
}

// NextLine returns the line number of the rune next rune.
func (l *Lexer[Token]) NextLine() int {
	return l.line
}

// NextCol returns the column number of the next rune.
func (l *Lexer[Token]) NextCol() int {
	return l.col
}

// StartLine returns the line number of the first rune of the current item.
func (l *Lexer[Token]) StartLine() int {
	return l.startLine
}

// StartCol returns the column number of the first rune of the current item.
func (l *Lexer[Token]) StartCol() int {
	return l.startCol
}

// ============================================================================
// Next
// ======================================================================================

// Next returns the next rune from the input.
// It increments the current position and updates line and col,
// so that the next call to [Lexer.Next] will yield the rune after the
// current.
//
// If there are no more characters, EOF is returned.
func (l *Lexer[Token]) Next() rune {
	if l.pos >= len(l.in) {
		return EOF
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

// NextWhile calls Next so long as the predicate is fulfilled or the
// end of file is reached.
//
// If it finds a rune r that does not match the predicate, it will backup
// and return r, making the next call [Lexer.Next] return the same rune
//
// If it, however, encounters the end of file before that, it will return EOF
// and won't backup.
func (l *Lexer[Token]) NextWhile(predicate func(r rune) bool) rune {
	for next := l.Next(); ; next = l.Next() {
		if next == EOF {
			return EOF
		}

		if !predicate(next) {
			l.Backup()
			return next
		}
	}
}

// SkipString skips the upcoming occurrence of s, i.e. it advances the position
// so that the next rune will be the rune after s.
//
// SkipString won't check if the upcoming actually matches s.
// Instead, it is callers responsibility to check themselves using
// [Lexer.PeekIsString].
//
// s must not contain any newlines.
// The only exception is s being just a newline, i.e. s = "\n".
func (l *Lexer[Token]) SkipString(s string) {
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

// Backup goes back one rune as if the previous [Lexer.Next] call never
// happened.
//
// Backup can only be called once for every call to [Lexer.Next].
func (l *Lexer[Token]) Backup() {
	l.pos -= l.prevWidth
	l.col--

	if l.col <= 0 {
		l.col = l.prevLineLen
		l.line--
	}
}

// ============================================================================
// Peek
// ======================================================================================

// Peek is the same as [Lexer.Next] but without increasing position.
func (l *Lexer[Token]) Peek() rune {
	if l.pos >= len(l.in) {
		return EOF
	}

	defer l.Backup()
	return l.Next()
}

// PeekIsString peeks ahead to see if the next runes match s in its full length.
//
// If so, it returns true.
// If not, false is returned.
func (l *Lexer[Token]) PeekIsString(s string) bool {
	endPos := l.pos + len(s)
	if endPos > len(l.in) {
		return false
	}

	return l.in[l.pos:endPos] == s
}

// PeekIsWord is the same as [Lexer.PeekIsString], but it additionally asserts
// that after the occurrence of s, a space, tab, newline, or EOF follows.
func (l *Lexer[Token]) PeekIsWord(s string) bool {
	if !l.PeekIsString(s) {
		return false
	}

	afterIndex := l.pos + len(s)

	if len(l.in) <= afterIndex { // EOF
		return true
	}

	after := l.in[afterIndex]
	return after == ' ' || after == '\t' || after == '\n'
}

// ============================================================================
// Ignore
// ======================================================================================

// Ignore ignores all input up to this point.
// It sets the start of the current item to the current position.
func (l *Lexer[Token]) Ignore() {
	l.startPos = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// IgnoreNext is shorthand for:
//
//	l.Next()
//	l.Ignore()
func (l *Lexer[Token]) IgnoreNext() {
	l.Next()
	l.Ignore()
}

// IgnoreWhile ignores all input up to this point.
// Additionally, it ignores all upcoming runes that fulfill the passed
// predicate.
//
// UponReturning, the next call to Next will yield the first rune that did
// not fulfill the predicate.
//
// If the end of file is reached before the predicate fails,
// [Lexer.IgnoreWhile] will stop at the end of file and return.
// The current rune will be EOF.
//
// It returns true if at least one rune was ignored.
func (l *Lexer[Token]) IgnoreWhile(predicate func(rune) bool) bool {
	l.Ignore()
	l.NextWhile(predicate)

	ignored := l.pos > l.startPos
	l.Ignore()

	return ignored
}

// ============================================================================
// Misc
// ======================================================================================

// IsLineEmpty checks if the rest of the current line is empty.
func (l *Lexer[Token]) IsLineEmpty() bool {
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

// IsContentEmpty reports whether at least one rune would be emitted, if
// [Lexer.Emit] were to be called now.
func (l *Lexer[Token]) IsContentEmpty() bool {
	return l.pos <= l.startPos
}

// ============================================================================
// Emit
// ======================================================================================

// Emit makes the lexed contents available through [Lexer.NextItem].
// After emitting, the next item will start at the next rune.
func (l *Lexer[Token]) Emit(t Token) {
	itm := Item[Token]{
		Type: t,
		Val:  l.in[l.startPos:l.pos],
		Line: l.startLine,
		Col:  l.startCol,
	}

	// If we're emitting the eof, we need different lines and cols.
	if l.pos > len(l.in) {
		itm.Col = l.col - 1
		itm.Line = l.line
		itm.Val = ""
	}

	l.items <- itm

	l.startPos = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// EmitError emits an [Item] of type t containing err for the current position.
func (l *Lexer[Token]) EmitError(t Token, err error) {
	if err == nil {
		panic("error called with nil error")
	}

	itm := Item[Token]{
		Type: t,
		Err:  err,
		Line: l.line,
		Col:  l.col - 1,
	}

	if itm.Col == 0 {
		itm.Line--
		itm.Col = l.prevLineLen
	}

	l.items <- itm
}
