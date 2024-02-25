package ast

// ============================================================================
// If
// ======================================================================================

type If struct {
	Header *IfHeader
	Then   Body

	ElseIfs []*ElseIf
	Else    *Else // may be nil

	Position Position
}

var _ ScopeNode = (*If)(nil)

func (i *If) Pos() Position { return i.Position }
func (i *If) End() Position {
	if i.Else != nil {
		return i.Else.End()
	} else if len(i.ElseIfs) > 0 {
		return i.ElseIfs[len(i.ElseIfs)-1].End()
	} else if i.Then != nil {
		return i.Then.End()
	} else if i.Header != nil {
		return i.Header.Pos()
	}
	return deltaPos(i.Position, len("if"))
}

func (*If) _node()      {}
func (*If) _scopeNode() {}

// ====================================== Else If =======================================

type ElseIf struct {
	Header   *IfHeader
	Then     Body
	Position Position
}

var _ Node = (*ElseIf)(nil)

func (e *ElseIf) Pos() Position { return e.Position }
func (e *ElseIf) End() Position {
	if e.Then != nil {
		return e.Then.End()
	} else if e.Header != nil {
		return e.Header.Pos()
	}
	return deltaPos(e.Position, len("else if"))
}

func (*ElseIf) _node() {}

// ======================================== Else ========================================

type Else struct {
	Then     Body
	Position Position
}

var _ Node = (*Else)(nil)

func (e *Else) Pos() Position { return e.Position }
func (e *Else) End() Position {
	if e.Then != nil {
		return e.Then.End()
	}
	return deltaPos(e.Position, len("else"))
}

func (*Else) _node() {}

// ===================================== If Header ======================================

type IfHeader struct {
	Statement *GoCode
	Condition Expression
}

var _ Node = (*IfHeader)(nil)

func (e *IfHeader) Pos() Position {
	if e.Statement != nil {
		return e.Statement.Pos()
	}

	return e.Condition.Pos()
}
func (e *IfHeader) End() Position {
	if e.Condition != nil {
		return e.Condition.End()
	} else if e.Statement != nil {
		return e.Statement.End()
	}
	return InvalidPosition
}

func (*IfHeader) _node() {}

// ============================================================================
// Switch
// ======================================================================================

type Switch struct {
	Comparator *GoCode // nil for case conditions
	LBrace     *Position
	Cases      []*Case
	RBrace     *Position
	Position   Position
}

var _ ScopeNode = (*Switch)(nil)

func (s *Switch) Pos() Position { return s.Position }
func (s *Switch) End() Position {
	if s.RBrace != nil {
		return deltaPos(*s.RBrace, 1)
	} else if len(s.Cases) > 0 {
		return s.Cases[len(s.Cases)-1].End()
	} else if s.LBrace != nil {
		return *s.LBrace
	} else if s.Comparator != nil {
		return s.Comparator.End()
	}
	return deltaPos(s.Position, len("switch"))
}

func (*Switch) _node()      {}
func (*Switch) _scopeNode() {}

// ======================================== Case ========================================

type Case struct {
	Expression Expression // nil for default case
	Colon      *Position
	Then       *Scope // has no L-/RBrace set

	Position Position
}

var _ Node = (*Case)(nil)

func (c *Case) Pos() Position { return c.Position }
func (c *Case) End() Position {
	if c.Then != nil {
		return c.Then.End()
	} else if c.Colon != nil {
		return deltaPos(*c.Colon, 1)
	} else if c.Expression != nil {
		return c.Expression.End()
	}
	return c.Position
}

func (*Case) _node() {}

// ============================================================================
// For
// ======================================================================================

type For struct {
	Header   ForHeader // nil for infinite loop
	Body     Body
	Position Position
}

var _ ScopeNode = (*For)(nil)

func (f *For) Pos() Position { return f.Position }
func (f *For) End() Position {
	if f.Body != nil {
		return f.Body.End()
	} else if f.Header != nil {
		return f.Header.Pos()
	}
	return deltaPos(f.Position, len("for"))
}

func (*For) _node()      {}
func (*For) _scopeNode() {}

// ===================================== For Header =====================================

// ForHeader is either a [ForRangeHeader], [GoCode], or a [ChainExpression].
type ForHeader interface {
	Node
	_forHeader()
}

// if this is changed, change the comment above
var (
	_ ForHeader = (*ForRangeHeader)(nil)
	_ ForHeader = (*GoCode)(nil)
	_ ForHeader = (*ChainExpression)(nil)
)

// ================================== For Range Header ==================================

// ForRangeHeader represents a range for loop header.
type ForRangeHeader struct {
	Var1  *Ident
	Comma *Position // nil if no var2
	Var2  *Ident    // nil if not present

	Declares  bool      // true if the range expression declares new variables (':=')
	EqualSign *Position // Position of the '=' or ':='
	Ordered   *Position // Position of the "ordered", if range is ordered
	Range     *Position // Position of the "range" keyword

	// Expression is the expression that is being iterated over.
	Expression Expression

	Position Position
}

var _ ForHeader = (*ForRangeHeader)(nil)

func (h *ForRangeHeader) Pos() Position { return h.Position }
func (h *ForRangeHeader) End() Position {
	if h.Expression != nil {
		return h.Expression.End()
	} else if h.Range != nil {
		return deltaPos(*h.Range, len("range"))
	} else if h.Ordered != nil {
		return deltaPos(*h.Ordered, len("ordered"))
	} else if h.EqualSign != nil {
		if h.Declares {
			return deltaPos(*h.EqualSign, len(":="))
		}
		return deltaPos(*h.EqualSign, len("="))
	} else if h.Var2 != nil {
		return h.Var2.End()
	} else if h.Comma != nil {
		return deltaPos(*h.Comma, 1)
	} else if h.Var1 != nil {
		return h.Var1.End()
	}
	return InvalidPosition
}

func (*ForRangeHeader) _node()      {}
func (*ForRangeHeader) _forHeader() {}
