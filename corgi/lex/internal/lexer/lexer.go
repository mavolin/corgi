// Package lexer implements a basic lexer.
package lexer

import "strings"

// Lexer implements a non-corgi-specific state-operated lexer, whose [StateFn] implementations emit Items.
type Lexer[Token any] struct {
	// in is the text being read.
	in string
	// items is the channel emitting the [Item] values.
	items chan Item[Token]

	// start is the first [StateFn] that is called when lexing is commenced.
	start StateFn[Token]

	// startLine and startCol are the line and column at which the current item
	// starts.
	startLine, startCol int
	// line and col are the line and column of the next rune.
	line, col int

	// prevWith is the width of the previous rune.
	prevWidth int
	// prevLineLen is a helper used by [Lexer.backup] to set the correct col
	// when backing a line up.
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
	// This variable is only used for lexerr and has no practical use.
	indentRefLine int
	indentLvl     int

	Context map[any]any
}

// EOF is the rune returned by Next, peek and alike to indicate the end of the
// input was reached.
const EOF rune = -1

// StateFn represents a single lexing state.
type StateFn[Token any] func(*Lexer[Token]) StateFn[Token]

// Item represents a lexical item as emitted by the lexer.
type Item[Token any] struct {
	// Type is the type of the item.
	Type Token
	// Expression is the value of the item, if any.
	Val string
	// Err is the error.
	// It is only set when Type is Error.
	Err error
	// Line is the line where the item starts.
	Line int
	// Col is the column after which the item starts.
	Col int
}

// New creates a new [Lexer] that lexes the passed input.
//
// When calling [Lexer.Lex], start will be the first [StateFn] to be called.
func New[Token any](in string, start StateFn[Token]) *Lexer[Token] {
	return &Lexer[Token]{
		in:        strings.ReplaceAll(in, "\r\n", "\n"),
		items:     make(chan Item[Token], 1),
		start:     start,
		startLine: 1,
		startCol:  1,
		line:      1,
		col:       1,
	}
}

// NextItem returns the Next lexical item.
func (l *Lexer[Token]) NextItem() Item[Token] {
	return <-l.items
}

// Stop stops the lexer's goroutine by draining all input elements.
func (l *Lexer[Token]) Stop() {
	for range l.items {
		// drain
	}
}

// Lex starts a new goroutine that lexes the input.
// The lexical items can be retrieved by calling NextItem.
func (l *Lexer[Token]) Lex() {
	go func() {
		state := l.start

		for state != nil {
			state = state(l)
		}

		close(l.items)
	}()
}
