package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// EOF emits a [token.EOF] and then returns nil.
func EOF(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.Emit(token.EOF)
	return nil
}

// Error returns a [lexer.StateFn] that emits the passed error err and then
// returns nil.
func Error(err error) lexer.StateFn[token.Token] {
	return func(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
		l.EmitError(token.Error, err)
		return nil
	}
}
