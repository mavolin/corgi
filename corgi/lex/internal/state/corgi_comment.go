package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// CorgiComment lexes a corgi comment, ignoring it completely.
//
// It assumes the next string is '//-'.
//
// It emits nothing.
func CorgiComment(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	lexutil.IgnoreCorgiComment(l)
	return Next
}
