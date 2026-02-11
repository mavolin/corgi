// Package ast defines the abstract syntax tree for a corgi file.
//
// Most fields of the individual nodes are pointers or interfaces.
// Unless they are marked as optional, they will only be nil if there was a
// parsing error that was recovered from.
// In other words, users can safely assume that all fields are non-nil, unless
// they are marked as optional.
//
// # A note on compatibility
//
// In order to allow for syntax changes in the future, a small reminder that
// the sets of the sum types this package defines may expand in the future.
// It should also be said, that fields with comments narrowing the set of
// possible types may, therefore, eventually be broadened in future versions.
// They exist to facilitate understanding of the AST, not to set a contract.
// This goes against the usual precedent that a comment's contract is not
// to be broken between versions.
package ast

import (
	"fmt"
	"slices"
)

// A File holds the abstract syntax tree for a corgi file.
type File struct {
	Package *PackageDirective
	Imports []*Import

	TopLevel TopLevel
	Comments []*CommentGroup
}

var _ Node = (*File)(nil)

func (f *File) Start() Position {
	if f.Package != nil {
		return f.Package.Start()
	}
	for _, imp := range f.Imports {
		if imp != nil {
			return imp.Start()
		}
	}
	return NoPosition
}

func (f *File) End() Position {
	var cEnd Position
	if len(f.Comments) > 0 {
		cEnd = f.Comments[len(f.Comments)-1].End()
	}

	var end Position
	if f.TopLevel != nil {
		end = f.TopLevel.End()
	} else {
		for _, imp := range slices.Backward(f.Imports) {
			if imp != nil {
				end = imp.End()
				break
			}
		}
		if end.Line == 0 {
			end = f.Package.End()
		}
	}

	if cEnd.Line > end.Line || (cEnd.Line == end.Line && cEnd.Col > end.Col) {
		return cEnd
	}
	return end
}

func (f *File) Walk(w func(Node)) {
	if f.Package != nil {
		w(f.Package)
	}
	for _, imp := range f.Imports {
		if imp != nil {
			w(imp)
		}
	}
	if f.TopLevel != nil {
		w(f.TopLevel)
	}
}
func (f *File) _node() {}

type Node interface {
	_node()
	// Start returns the inclusive start position of the node.
	Start() Position
	// End returns the exclusive end position of the node.
	End() Position

	// Walk calls the passed function for each direct child node of the
	// receiver in the order they appear in the source code.
	Walk(func(Node))
}

// ============================================================================
// Position
// ======================================================================================

// Position is a position in a file.
type Position struct {
	Line int
	Col  int
}

var NoPosition = Position{}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

func deltaPos(p Position, delta int) Position {
	p.Col += delta
	return p
}

// ============================================================================
// Bad Node
// ======================================================================================

type BadNode struct {
	From, Until Position
}

var (
	_ ScopeNode    = (*BadNode)(nil)
	_ TopLevelNode = (*BadNode)(nil)
)

func (b *BadNode) Start() Position   { return b.From }
func (b *BadNode) End() Position     { return b.Until }
func (b *BadNode) Walk(_ func(Node)) {}

func (*BadNode) _node()         {}
func (*BadNode) _scopeNode()    {}
func (*BadNode) _topLevelNode() {}
