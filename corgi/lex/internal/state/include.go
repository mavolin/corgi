package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// Include consumes a single include directive.
//
// It assumes the next string will be 'Include'.
//
// It emits [token.Include] and then a [token.Literal] identifying the included
// file.
func Include(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("include")
	l.Emit(token.Include)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return Error(&lexerr.UnknownItemError{Expected: "a space"})
	}

	switch l.Peek() {
	case lexer.EOF:
		return EOF
	case '\n':
		return Error(&lexerr.EOLError{After: "a string"})
	case '`', '"':
		// handled below
	default: // invalid
		return Error(&lexerr.UnknownItemError{Expected: "a string"})
	}

	if end := lexutil.ConsumeString(l); end != nil {
		return end
	}

	l.Emit(token.Literal)
	return lexutil.AssertNewlineOrEOF(l, Next)
}
