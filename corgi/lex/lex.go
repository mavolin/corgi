// Package lex provides the lexer for corgi.
//
// This package is split into five additional smaller package besides this
// one to keep the code organized and more readable.
//
// [lexer] provides a basic lexer, responsible for reading the input
// with helpers and keeping track of indentation.
// It does not depend on any other lex package except lexerr.
//
// [lexutil] provides utilities that require package token, but that
// do not resemble actual states as found in this package.
//
// [lexerr] provides the error types used across all packages.
// Arguably, keeping errors in a separate package as opposed to lex is less
// than elegant, but it is necessary to avoid circular dependencies, and I
// believe the benefits in terms of readability and maintainability we gain by
// splitting lex into these subpackages outweigh this small inelegance.
//
// [token] provides the token types used by the lexer, again inelegantly kept
// in a separate package to avoid circular dependencies (remember the
// benefits! :) ).
//
// [state] provides the individual states for the lexer.
//
// Lastly, this package glues everything together.
package lex

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/state"
	"github.com/mavolin/corgi/corgi/lex/token"
)

type Lexer struct {
	lex *lexer.Lexer[token.Token]
}

type Item = lexer.Item[token.Token]

// New creates a new lexer.
func New(in string) *Lexer {
	return &Lexer{lex: lexer.New(in, state.Next)}
}

// NextItem returns the Next lexical item.
func (l *Lexer) NextItem() Item {
	return l.lex.NextItem()
}

// Stop stops the lexer's goroutine by draining all input elements.
func (l *Lexer) Stop() {
	l.lex.Stop()
}

// Lex starts a new goroutine that lexes the input.
// The lexical items can be retrieved by calling NextItem.
func (l *Lexer) Lex() {
	l.lex.Lex()
}
