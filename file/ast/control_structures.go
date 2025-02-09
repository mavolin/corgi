package ast

import "slices"

// ============================================================================
// If
// ======================================================================================

type If struct {
	If      *Position
	Header  *IfHeader
	Then    Body
	ElseIfs []*ElseIf
	Else    *Else // may be nil
}

var _ ScopeNode = (*If)(nil)

func (i *If) Start() Position {
	if i.If != nil {
		return *i.If
	} else if i.Header != nil {
		return i.Header.Start()
	} else if i.Then != nil {
		return i.Then.Start()
	}
	for _, e := range i.ElseIfs {
		if e != nil {
			return e.Start()
		}
	}
	if i.Else != nil {
		return i.Else.Start()
	}
	return Position{}
}
func (i *If) End() Position {
	if i.Else != nil {
		return i.Else.End()
	}
	for _, e := range slices.Backward(i.ElseIfs) {
		if e != nil {
			return e.End()
		}
	}
	if i.Then != nil {
		return i.Then.End()
	} else if i.Header != nil {
		return i.Header.Start()
	} else if i.If != nil {
		return deltaPos(*i.If, len("if"))
	}
	return Position{}
}

func (*If) _node()      {}
func (*If) _scopeNode() {}

// ====================================== Else If =======================================

type ElseIf struct {
	Else   *Position
	If     *Position
	Header *IfHeader
	Then   Body
}

var _ Node = (*ElseIf)(nil)

func (e *ElseIf) Start() Position {
	if e.If != nil {
		return *e.If
	} else if e.Else != nil {
		return *e.Else
	} else if e.Header != nil {
		return e.Header.Start()
	} else if e.Then != nil {
		return e.Then.Start()
	}
	return Position{}
}
func (e *ElseIf) End() Position {
	if e.Then != nil {
		return e.Then.End()
	} else if e.Header != nil {
		return e.Header.Start()
	} else if e.Else != nil {
		return deltaPos(*e.Else, len("else"))
	} else if e.If != nil {
		return deltaPos(*e.If, len("if"))
	}
	return Position{}
}

func (*ElseIf) _node() {}

// ======================================== Else ========================================

type Else struct {
	Else *Position
	Then Body
}

var _ Node = (*Else)(nil)

func (e *Else) Start() Position {
	if e.Else != nil {
		return *e.Else
	} else if e.Then != nil {
		return e.Then.Start()
	}
	return Position{}
}
func (e *Else) End() Position {
	if e.Then != nil {
		return e.Then.End()
	} else if e.Else != nil {
		return deltaPos(*e.Else, len("else"))
	}
	return Position{}
}

func (*Else) _node() {}

// ============================================================================
// If Header
// ======================================================================================

type IfHeader struct {
	Statement *SimpleStatement // optional
	Condition *Expression
}

var _ Node = (*IfHeader)(nil)

func (e *IfHeader) Start() Position {
	if e.Statement != nil {
		return e.Statement.Start()
	} else if e.Condition != nil {
		return e.Condition.Start()
	}
	return Position{}
}
func (e *IfHeader) End() Position {
	if e.Condition != nil {
		return e.Condition.End()
	} else if e.Statement != nil {
		return e.Statement.End()
	}
	return Position{}
}

func (*IfHeader) _node() {}

// ============================================================================
// Switch
// ======================================================================================

type Switch struct {
	Switch     *Position
	Comparator *SimpleStatement // nil for case conditions
	LBrace     *Position
	Cases      []*Case
	RBrace     *Position
}

var _ ScopeNode = (*Switch)(nil)

func (s *Switch) Start() Position {
	if s.Switch != nil {
		return *s.Switch
	} else if s.Comparator != nil {
		return s.Comparator.Start()
	} else if s.LBrace != nil {
		return *s.LBrace
	}
	for _, c := range s.Cases {
		if c != nil {
			return c.Start()
		}
	}
	if s.RBrace != nil {
		return *s.RBrace
	}
	return Position{}
}
func (s *Switch) End() Position {
	if s.RBrace != nil {
		return deltaPos(*s.RBrace, len("}"))
	}
	for _, c := range slices.Backward(s.Cases) {
		if c != nil {
			return c.End()
		}
	}
	if s.LBrace != nil {
		return deltaPos(*s.LBrace, len("{"))
	} else if s.Comparator != nil {
		return s.Comparator.End()
	} else if s.Switch != nil {
		return deltaPos(*s.Switch, len("switch"))
	}
	return Position{}
}

func (*Switch) _node()      {}
func (*Switch) _scopeNode() {}

// ======================================== Case ========================================

type Case struct {
	Case       *Position   // either this or Default is set
	Default    *Position   // nil if not default
	Expression *Expression // nil for default case
	Colon      *Position
	Then       []ScopeNode
}

var _ Node = (*Case)(nil)

func (c *Case) Start() Position {
	if c.Case != nil {
		return *c.Case
	} else if c.Default != nil {
		return *c.Default
	} else if c.Expression != nil {
		return c.Expression.Start()
	} else if c.Colon != nil {
		return *c.Colon
	}
	for _, n := range c.Then {
		if n != nil {
			return n.Start()
		}
	}
	return Position{}
}
func (c *Case) End() Position {
	for _, n := range slices.Backward(c.Then) {
		if n != nil {
			return n.End()
		}
	}
	if c.Colon != nil {
		return deltaPos(*c.Colon, len(":"))
	} else if c.Expression != nil {
		return c.Expression.End()
	} else if c.Default != nil {
		return deltaPos(*c.Default, len("default"))
	} else if c.Case != nil {
		return deltaPos(*c.Case, len("case"))
	}
	return Position{}
}

func (*Case) _node() {}

// ============================================================================
// For
// ======================================================================================

type For struct {
	For    *Position
	Header ForHeader // nil for infinite loop
	Body   Body
}

var _ ScopeNode = (*For)(nil)

func (f *For) Start() Position {
	if f.For != nil {
		return *f.For
	} else if f.Header != nil {
		return f.Header.Start()
	} else if f.Body != nil {
		return f.Body.Start()
	}
	return Position{}
}
func (f *For) End() Position {
	if f.Body != nil {
		return f.Body.End()
	} else if f.Header != nil {
		return f.Header.Start()
	} else if f.For != nil {
		return deltaPos(*f.For, len("for"))
	}
	return Position{}
}

func (*For) _node()      {}
func (*For) _scopeNode() {}

// ============================================================================
// For Header
// ======================================================================================

// ForHeader is either a [ForRangeHeader] or an [Expression].
type ForHeader interface {
	Node
	_forHeader()
}

// if this is changed, change the comment above
var (
	_ ForHeader = (*ForConditionHeader)(nil)
	_ ForHeader = (*ForClauseHeader)(nil)
	_ ForHeader = (*ForRangeHeader)(nil)
)

// ================================ For Condition Header ================================

type ForConditionHeader struct {
	Condition *Expression
}

var _ ForHeader = (*ForConditionHeader)(nil)

func (h *ForConditionHeader) Start() Position {
	if h.Condition != nil {
		return h.Condition.Start()
	}
	return Position{}
}
func (h *ForConditionHeader) End() Position {
	if h.Condition != nil {
		return h.Condition.End()
	}
	return Position{}
}
func (*ForConditionHeader) _node()      {}
func (*ForConditionHeader) _forHeader() {}

// ================================= For Clause Header ==================================

type ForClauseHeader struct {
	Init      *SimpleStatement // may be nil
	Condition *Expression      // may be nil
	Post      *SimpleStatement // may be nil
}

var _ ForHeader = (*ForClauseHeader)(nil)

func (h *ForClauseHeader) Start() Position {
	if h.Init != nil {
		return h.Init.Start()
	} else if h.Condition != nil {
		return h.Condition.Start()
	} else if h.Post != nil {
		return h.Post.Start()
	}
	return Position{}
}
func (h *ForClauseHeader) End() Position {
	if h.Post != nil {
		return h.Post.End()
	} else if h.Condition != nil {
		return h.Condition.End()
	} else if h.Init != nil {
		return h.Init.End()
	}
	return Position{}
}
func (*ForClauseHeader) _node()      {}
func (*ForClauseHeader) _forHeader() {}

// ================================== For Range Header ==================================

// ForRangeHeader is a range for loop header.
type ForRangeHeader struct {
	Var1 *Expression // optional
	Var2 *Expression // nil if not present

	Colon     *Position // optional
	EqualSign *Position // optional, Position of the '='
	Ordered   *Position // Position of the "ordered", if range is ordered
	Range     *Position // Position of the "range" keyword

	// Expression is the expression that is being iterated over.
	Expression *Expression
}

var _ ForHeader = (*ForRangeHeader)(nil)

func (h *ForRangeHeader) Start() Position {
	if h.Var1 != nil {
		return h.Var1.Start()
	} else if h.Var2 != nil {
		return h.Var2.Start()
	} else if h.Colon != nil {
		return *h.Colon
	} else if h.EqualSign != nil {
		return *h.EqualSign
	} else if h.Ordered != nil {
		return *h.Ordered
	} else if h.Range != nil {
		return *h.Range
	} else if h.Expression != nil {
		return h.Expression.Start()
	}
	return Position{}
}
func (h *ForRangeHeader) End() Position {
	if h.Expression != nil {
		return h.Expression.End()
	} else if h.Range != nil {
		return deltaPos(*h.Range, len("range"))
	} else if h.Ordered != nil {
		return deltaPos(*h.Ordered, len("ordered"))
	} else if h.Colon != nil {
		return deltaPos(*h.Colon, len(":"))
	} else if h.EqualSign != nil {
		return deltaPos(*h.EqualSign, len("="))
	} else if h.Var2 != nil {
		return h.Var2.End()
	} else if h.Var1 != nil {
		return h.Var1.End()
	}
	return Position{}
}

func (*ForRangeHeader) _node()      {}
func (*ForRangeHeader) _forHeader() {}
