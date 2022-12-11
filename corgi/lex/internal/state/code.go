package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// Code consumes a code line or block.
//
// It assumes the next rune is a '-'.
//
// It emits a CodeStart item and then either one Code item
// or an Indent and one, or multiple Code items, terminated by a Dedent.
func Code(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("-")
	l.Emit(token.CodeStart)

	spaceAfter := l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case '\n': // this is a block of code
		l.Ignore()
		// handled below
	default: // a single line of code
		l.Backup()
		// special case: empty line, for ✨visuals✨
		if l.IsLineEmpty() {
			return lexutil.AssertNewlineOrEOF(l, Next)
		}

		if !spaceAfter {
			return Error(&lexerr.UnknownItemError{Expected: "a space"})
		}

		end := lexutil.EmitNextPredicate(l, token.Code, nil, lexer.MatchesNot('\n'))
		if end != nil {
			return end
		}

		// indents are allowed after single line code blocks

		dIndent, _, err := l.ConsumeIndent(lexer.ConsumeAllIndents)
		if err != nil {
			return Error(err)
		}

		lexutil.EmitIndent(l, dIndent)

		return Next
	}

	// we're at the beginning of a block of code

	dIndent, _, err := l.ConsumeIndent(lexer.ConsumeSingleIncrease)
	if err != nil {
		return Error(err)
	}

	lexutil.EmitIndent(l, dIndent)

	if dIndent <= 0 {
		return Next
	}

	for dIndent >= 0 {
		end := lexutil.EmitNextPredicate(l, token.Code, nil, lexer.MatchesNot('\n'))
		if end != nil {
			return end
		}

		if l.Next() == lexer.EOF {
			return EOF
		}

		dIndent, _, err = l.ConsumeIndent(lexer.ConsumeNoIncrease)
		if err != nil {
			return Error(err)
		}

		lexutil.EmitIndent(l, dIndent)
	}

	return Next
}
