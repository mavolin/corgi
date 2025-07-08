package ast

import "slices"

// ============================================================================
// Comment Group
// ======================================================================================

// CommentGroup is a group of adjacent comments.
//
// If a CommentGroup contains a /* General Comment */, it will be the only
// element in the group.
// Line comments which are on the same line as another node, are also
// grouped alone.
//
// This means
//
//	// This
//	// is a single group.
//	type Foo string
//
// while these are two groups:
//
//	if foo { // bar
//	  // baz
//
// and these are three groups:
//
//	/* g1 */ // g2
//	// g3
type CommentGroup struct {
	Comments []*Comment
	// Attached is the node attached to this comment.
	Attached Node // may be nil
}

func (g CommentGroup) Start() Position {
	for _, c := range g.Comments {
		if c != nil {
			return c.Start()
		}
	}
	return Position{}
}

func (g CommentGroup) End() Position {
	for _, c := range slices.Backward(g.Comments) {
		if c != nil {
			return c.End()
		}
	}
	return Position{}
}

func (g CommentGroup) Walk(w func(Node)) {
	for _, c := range g.Comments {
		if c != nil {
			w(c)
		}
	}
}

func (CommentGroup) _node() {}

// ============================================================================
// Comment
// ======================================================================================

// Comment is a comment not included in the output of the template.
type Comment struct {
	Open    *Position
	Comment string
	General bool
	Close   *Position // only set for general comments
	Until   Position  // excl. end position
}

func (c *Comment) Start() Position {
	if c.Open != nil {
		return *c.Open
	} else if c.Close != nil {
		return *c.Close
	}
	return c.Until
}

func (c *Comment) End() Position {
	return c.Until
}
func (c *Comment) Walk(func(Node)) {}

func (*Comment) _node() {}
