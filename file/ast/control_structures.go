package ast

import "slices"

// ============================================================================
// Conditional
// ======================================================================================

// Conditional is a wrapper around an if statement, for more straightforward
// walking.
type Conditional struct {
	If      *If
	ElseIfs []*ElseIf
	Else    *Else
}

var (
	_ ScopeNode   = (*Conditional)(nil)
	_ Highlighter = (*Conditional)(nil)
)

func (c *Conditional) Start() Position {
	if c.If != nil {
		if start := c.If.Start(); start != NoPosition {
			return start
		}
	}
	for _, e := range c.ElseIfs {
		if e != nil {
			if start := e.Start(); start != NoPosition {
				return start
			}
		}
	}
	if c.Else != nil {
		if start := c.Else.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (c *Conditional) End() Position {
	if c.Else != nil {
		if end := c.Else.End(); end != NoPosition {
			return end
		}
	}
	for _, e := range slices.Backward(c.ElseIfs) {
		if e != nil {
			if end := e.End(); end != NoPosition {
				return end
			}
		}
	}
	if c.If != nil {
		if end := c.If.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (c *Conditional) Walk(w func(Node)) {
	if c.If != nil {
		w(c.If)
	}
	for _, e := range c.ElseIfs {
		if e != nil {
			w(e)
		}
	}
	if c.Else != nil {
		w(c.Else)
	}
}

func (c *Conditional) Highlight() (start, end Position) {
	if c.If != nil {
		return c.If.Highlight()
	}
	return c.Start(), c.End()
}

func (*Conditional) _node()      {}
func (*Conditional) _scopeNode() {}

// ============================================================================
// If
// ======================================================================================

type If struct {
	If     *Position
	Header *IfHeader
	Then   Body
}

var (
	_ Node        = (*If)(nil)
	_ Highlighter = (*If)(nil)
)

func (i *If) Start() Position {
	if i.If != nil {
		return *i.If
	}
	if i.Header != nil {
		if start := i.Header.Start(); start != NoPosition {
			return start
		}
	}
	if i.Then != nil {
		if start := i.Then.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (i *If) End() Position {
	if i.Then != nil {
		if end := i.Then.End(); end != NoPosition {
			return end
		}
	}
	if i.Header != nil {
		if start := i.Header.Start(); start != NoPosition {
			return start
		}
	}
	if i.If != nil {
		return deltaPos(*i.If, len("if"))
	}
	return NoPosition
}

func (i *If) Walk(w func(Node)) {
	if i.Header != nil {
		w(i.Header)
	}
	if i.Then != nil {
		w(i.Then)
	}
}

func (i *If) Highlight() (start, end Position) {
	if i.If != nil {
		return *i.If, deltaPos(*i.If, len("if"))
	}
	return i.Start(), i.End()
}

func (*If) _node() {}

// ============================================================================
// Else If
// ======================================================================================

type ElseIf struct {
	Else   *Position
	If     *Position
	Header *IfHeader
	Then   Body
}

var (
	_ Node        = (*ElseIf)(nil)
	_ Highlighter = (*ElseIf)(nil)
)

func (e *ElseIf) Start() Position {
	if e.Else != nil {
		return *e.Else
	}
	if e.If != nil {
		return *e.If
	}
	if e.Header != nil {
		if start := e.Header.Start(); start != NoPosition {
			return start
		}
	}
	if e.Then != nil {
		if start := e.Then.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (e *ElseIf) End() Position {
	if e.Then != nil {
		if end := e.Then.End(); end != NoPosition {
			return end
		}
	}
	if e.Header != nil {
		if start := e.Header.Start(); start != NoPosition {
			return start
		}
	}
	if e.If != nil {
		return deltaPos(*e.If, len("if"))
	}
	if e.Else != nil {
		return deltaPos(*e.Else, len("else"))
	}
	return NoPosition
}

func (e *ElseIf) Walk(w func(Node)) {
	if e.Header != nil {
		w(e.Header)
	}
	if e.Then != nil {
		w(e.Then)
	}
}

func (e *ElseIf) Highlight() (start, end Position) {
	if e.Else != nil {
		if e.If != nil {
			return *e.Else, deltaPos(*e.If, len("if"))
		}
		return *e.Else, deltaPos(*e.Else, len("else"))
	}
	return e.Start(), e.End()
}

func (*ElseIf) _node() {}

// ============================================================================
// Else
// ======================================================================================

type Else struct {
	Else *Position
	Then Body
}

var (
	_ Node        = (*Else)(nil)
	_ Highlighter = (*Else)(nil)
)

func (e *Else) Start() Position {
	if e.Else != nil {
		return *e.Else
	}
	if e.Then != nil {
		if start := e.Then.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (e *Else) End() Position {
	if e.Then != nil {
		if end := e.Then.End(); end != NoPosition {
			return end
		}
	}
	if e.Else != nil {
		return deltaPos(*e.Else, len("else"))
	}
	return NoPosition
}

func (e *Else) Walk(w func(Node)) {
	if e.Then != nil {
		w(e.Then)
	}
}

func (e *Else) Highlight() (start, end Position) {
	if e.Else != nil {
		return *e.Else, deltaPos(*e.Else, len("else"))
	}
	return e.Start(), e.End()
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
	return NoPosition
}

func (e *IfHeader) End() Position {
	if e.Condition != nil {
		return e.Condition.End()
	} else if e.Statement != nil {
		return e.Statement.End()
	}
	return NoPosition
}

func (e *IfHeader) Walk(w func(Node)) {
	if e.Statement != nil {
		w(e.Statement)
	}
	if e.Condition != nil {
		w(e.Condition)
	}
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

var (
	_ ScopeNode   = (*Switch)(nil)
	_ Highlighter = (*Switch)(nil)
)

func (s *Switch) Start() Position {
	if s.Switch != nil {
		return *s.Switch
	}
	if s.Comparator != nil {
		if start := s.Comparator.Start(); start != NoPosition {
			return start
		}
	}
	if s.LBrace != nil {
		return *s.LBrace
	}
	for _, c := range s.Cases {
		if c != nil {
			if start := c.Start(); start != NoPosition {
				return start
			}
		}
	}
	if s.RBrace != nil {
		return *s.RBrace
	}
	return NoPosition
}

func (s *Switch) End() Position {
	if s.RBrace != nil {
		return deltaPos(*s.RBrace, len("}"))
	}
	for _, c := range slices.Backward(s.Cases) {
		if c != nil {
			if end := c.End(); end != NoPosition {
				return end
			}
		}
	}
	if s.LBrace != nil {
		return deltaPos(*s.LBrace, len("{"))
	}
	if s.Comparator != nil {
		if end := s.Comparator.End(); end != NoPosition {
			return end
		}
		return deltaPos(*s.Switch, len("switch"))
	}
	if s.Switch != nil {
		return deltaPos(*s.Switch, len("switch"))
	}
	return NoPosition
}

func (s *Switch) Walk(w func(Node)) {
	if s.Comparator != nil {
		w(s.Comparator)
	}
	for _, c := range s.Cases {
		if c != nil {
			w(c)
		}
	}
}

func (s *Switch) Highlight() (start, end Position) {
	if s.Switch != nil {
		return *s.Switch, deltaPos(*s.Switch, len("switch"))
	}
	return s.Start(), s.End()
}

func (*Switch) _node()      {}
func (*Switch) _scopeNode() {}

// ======================================== Case ========================================

type Case struct {
	Case       *Position   // either this or Default is set
	Default    *Position   // nil, if not default
	Expression *Expression // nil for the default case
	Colon      *Position
	Then       []ScopeNode
}

var (
	_ Node        = (*Case)(nil)
	_ Highlighter = (*Case)(nil)
)

func (c *Case) Start() Position {
	if c.Case != nil {
		return *c.Case
	}
	if c.Default != nil {
		return *c.Default
	}
	if c.Expression != nil {
		if start := c.Expression.Start(); start != NoPosition {
			return start
		}
	}
	if c.Colon != nil {
		return *c.Colon
	}
	for _, n := range c.Then {
		if n != nil {
			if start := n.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (c *Case) End() Position {
	for _, n := range slices.Backward(c.Then) {
		if n != nil {
			if end := n.End(); end != NoPosition {
				return end
			}
		}
	}
	if c.Colon != nil {
		return deltaPos(*c.Colon, len(":"))
	}
	if c.Expression != nil {
		if end := c.Expression.End(); end != NoPosition {
			return end
		}
	}
	if c.Default != nil {
		return deltaPos(*c.Default, len("default"))
	}
	if c.Case != nil {
		return deltaPos(*c.Case, len("case"))
	}
	return NoPosition
}

func (c *Case) Walk(w func(Node)) {
	if c.Expression != nil {
		w(c.Expression)
	}
	for _, n := range c.Then {
		if n != nil {
			w(n)
		}
	}
}

func (c *Case) Highlight() (start, end Position) {
	if c.Case != nil {
		return *c.Case, deltaPos(*c.Case, len("case"))
	} else if c.Default != nil {
		return *c.Default, deltaPos(*c.Default, len("default"))
	}
	return c.Start(), c.End()
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

var (
	_ ScopeNode   = (*For)(nil)
	_ Highlighter = (*For)(nil)
)

func (f *For) Start() Position {
	if f.For != nil {
		return *f.For
	}
	if f.Header != nil {
		if start := f.Header.Start(); start != NoPosition {
			return start
		}
	}
	if f.Body != nil {
		if start := f.Body.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (f *For) End() Position {
	if f.Body != nil {
		if end := f.Body.End(); end != NoPosition {
			return end
		}
	}
	if f.Header != nil {
		if end := f.Header.Start(); end != NoPosition {
			return end
		}
	}
	if f.For != nil {
		return deltaPos(*f.For, len("for"))
	}
	return NoPosition
}

func (f *For) Walk(w func(Node)) {
	if f.Header != nil {
		w(f.Header)
	}
	if f.Body != nil {
		w(f.Body)
	}
}

func (f *For) Highlight() (start, end Position) {
	if f.For != nil {
		return *f.For, deltaPos(*f.For, len("for"))
	}
	return f.Start(), f.End()
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
		if start := h.Condition.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (h *ForConditionHeader) End() Position {
	if h.Condition != nil {
		if end := h.Condition.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (h *ForConditionHeader) Walk(w func(Node)) {
	if h.Condition != nil {
		w(h.Condition)
	}
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
		if start := h.Init.Start(); start != NoPosition {
			return start
		}
	}
	if h.Condition != nil {
		if start := h.Condition.Start(); start != NoPosition {
			return start
		}
	}
	if h.Post != nil {
		if start := h.Post.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (h *ForClauseHeader) End() Position {
	if h.Post != nil {
		if end := h.Post.End(); end != NoPosition {
			return end
		}
	}
	if h.Condition != nil {
		if end := h.Condition.End(); end != NoPosition {
			return end
		}
	}
	if h.Init != nil {
		if end := h.Init.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (h *ForClauseHeader) Walk(w func(Node)) {
	if h.Init != nil {
		w(h.Init)
	}
	if h.Condition != nil {
		w(h.Condition)
	}
	if h.Post != nil {
		w(h.Post)
	}
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
	Range     *Position // Position of the "range" keyword

	// Expression is the expression that is being iterated over.
	Expression *Expression
}

var _ ForHeader = (*ForRangeHeader)(nil)

func (h *ForRangeHeader) Start() Position {
	if h.Var1 != nil {
		if start := h.Var1.Start(); start != NoPosition {
			return start
		}
	}
	if h.Var2 != nil {
		if start := h.Var2.Start(); start != NoPosition {
			return start
		}
	}
	if h.Colon != nil {
		return *h.Colon
	}
	if h.EqualSign != nil {
		return *h.EqualSign
	}
	if h.Range != nil {
		return *h.Range
	}
	if h.Expression != nil {
		if start := h.Expression.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (h *ForRangeHeader) End() Position {
	if h.Expression != nil {
		if end := h.Expression.End(); end != NoPosition {
			return end
		}
	}
	if h.Range != nil {
		return deltaPos(*h.Range, len("range"))
	}
	if h.Colon != nil {
		return deltaPos(*h.Colon, len(":"))
	}
	if h.EqualSign != nil {
		return deltaPos(*h.EqualSign, len("="))
	}
	if h.Var2 != nil {
		if end := h.Var2.End(); end != NoPosition {
			return end
		}
	}
	if h.Var1 != nil {
		if end := h.Var1.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (h *ForRangeHeader) Walk(w func(Node)) {
	if h.Var1 != nil {
		w(h.Var1)
	}
	if h.Var2 != nil {
		w(h.Var2)
	}
	if h.Expression != nil {
		w(h.Expression)
	}
}

func (*ForRangeHeader) _node()      {}
func (*ForRangeHeader) _forHeader() {}
