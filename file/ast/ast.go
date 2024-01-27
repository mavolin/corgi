// Package ast defines the abstract syntax tree for a corgi file.
//
// Most fields of the individual nodes are pointers or interfaces.
// Unless they are marked as optional, they will only be nil if there was a
// parsing error that was recovered from.
// In other words, users can safely assume that all fields are non-nil, unless
// they are marked as optional.
package ast

import "fmt"

type AST struct {
	// Raw contains the raw input file, as it was parsed.
	Raw string
	// Lines are the lines of Raw, stripped of their CRLF/LF line endings.
	Lines []string

	PackageDoc       []*DevComment // optional
	PackageDirective *PackageDirective

	Scope *Scope // optional
}

type Poser interface {
	Pos() Position
}

// Position indicates the position where a token was encountered.
type Position struct {
	Line int
	Col  int
}

var InvalidPosition = Position{0, 0}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}
