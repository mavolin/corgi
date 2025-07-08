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

var _ ScopeNode = (*Conditional)(nil)

func (c *Conditional) Start() Position {
	if c.If != nil {
		return c.If.Start()
	}
	for _, e := range c.ElseIfs {
		if e != nil {
			return e.Start()
		}
	}
	if c.Else != nil {
		return c.Else.Start()
	}
	return Position{}
}

func (c *Conditional) End() Position {
	if c.Else != nil {
		return c.Else.End()
	}
	for _, e := range slices.Backward(c.ElseIfs) {
		if e != nil {
			return e.End()
		}
	}
	if c.If != nil {
		return c.If.End()
	}
	return Position{}
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

var _ Node = (*If)(nil)

func (i *If) Start() Position {
	switch {
	case i.If != nil:
		return *i.If
	case i.Header != nil:
		return i.Header.Start()
	case i.Then != nil:
		return i.Then.Start()
	}

	return Position{}
}

func (i *If) End() Position {
	switch {
	case i.Then != nil:
		return i.Then.End()
	case i.Header != nil:
		return i.Header.Start()
	case i.If != nil:
		return deltaPos(*i.If, len("if"))
	}
	return Position{}
}

func (i *If) Walk(w func(Node)) {
	if i.Header != nil {
		w(i.Header)
	}
	if i.Then != nil {
		w(i.Then)
	}
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

var _ Node = (*ElseIf)(nil)

func (e *ElseIf) Start() Position {
	switch {
	case e.If != nil:
		return *e.If
	case e.Else != nil:
		return *e.Else
	case e.Header != nil:
		return e.Header.Start()
	case e.Then != nil:
		return e.Then.Start()
	}
	return Position{}
}

func (e *ElseIf) End() Position {
	switch {
	case e.Then != nil:
		return e.Then.End()
	case e.Header != nil:
		return e.Header.Start()
	case e.Else != nil:
		return deltaPos(*e.Else, len("else"))
	case e.If != nil:
		return deltaPos(*e.If, len("if"))
	}
	return Position{}
}

func (e *ElseIf) Walk(w func(Node)) {
	if e.Header != nil {
		w(e.Header)
	}
	if e.Then != nil {
		w(e.Then)
	}
}

func (*ElseIf) _node() {}

// ============================================================================
// Else
// ======================================================================================

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

func (e *Else) Walk(w func(Node)) {
	if e.Then != nil {
		w(e.Then)
	}
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

var _ ScopeNode = (*Switch)(nil)

func (s *Switch) Start() Position {
	switch {
	case s.Switch != nil:
		return *s.Switch
	case s.Comparator != nil:
		return s.Comparator.Start()
	case s.LBrace != nil:
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
	switch {
	case s.LBrace != nil:
		return deltaPos(*s.LBrace, len("{"))
	case s.Comparator != nil:
		return s.Comparator.End()
	case s.Switch != nil:
		return deltaPos(*s.Switch, len("switch"))
	}
	return Position{}
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

var _ Node = (*Case)(nil)

func (c *Case) Start() Position {
	switch {
	case c.Case != nil:
		return *c.Case
	case c.Default != nil:
		return *c.Default
	case c.Expression != nil:
		return c.Expression.Start()
	case c.Colon != nil:
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
	switch {
	case c.Colon != nil:
		return deltaPos(*c.Colon, len(":"))
	case c.Expression != nil:
		return c.Expression.End()
	case c.Default != nil:
		return deltaPos(*c.Default, len("default"))
	case c.Case != nil:
		return deltaPos(*c.Case, len("case"))
	}
	return Position{}
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
	switch {
	case f.For != nil:
		return *f.For
	case f.Header != nil:
		return f.Header.Start()
	case f.Body != nil:
		return f.Body.Start()
	}
	return Position{}
}

func (f *For) End() Position {
	switch {
	case f.Body != nil:
		return f.Body.End()
	case f.Header != nil:
		return f.Header.Start()
	case f.For != nil:
		return deltaPos(*f.For, len("for"))
	}
	return Position{}
}

func (f *For) Walk(w func(Node)) {
	if f.Header != nil {
		w(f.Header)
	}
	if f.Body != nil {
		w(f.Body)
	}
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
	switch {
	case h.Init != nil:
		return h.Init.Start()
	case h.Condition != nil:
		return h.Condition.Start()
	case h.Post != nil:
		return h.Post.Start()
	}
	return Position{}
}

func (h *ForClauseHeader) End() Position {
	switch {
	case h.Post != nil:
		return h.Post.End()
	case h.Condition != nil:
		return h.Condition.End()
	case h.Init != nil:
		return h.Init.End()
	}
	return Position{}
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
	Ordered   *Position // Position of the "ordered", if range is ordered
	Range     *Position // Position of the "range" keyword

	// Expression is the expression that is being iterated over.
	Expression *Expression
}

var _ ForHeader = (*ForRangeHeader)(nil)

func (h *ForRangeHeader) Start() Position {
	switch {
	case h.Var1 != nil:
		return h.Var1.Start()
	case h.Var2 != nil:
		return h.Var2.Start()
	case h.Colon != nil:
		return *h.Colon
	case h.EqualSign != nil:
		return *h.EqualSign
	case h.Ordered != nil:
		return *h.Ordered
	case h.Range != nil:
		return *h.Range
	case h.Expression != nil:
		return h.Expression.Start()
	}
	return Position{}
}

func (h *ForRangeHeader) End() Position {
	switch {
	case h.Expression != nil:
		return h.Expression.End()
	case h.Range != nil:
		return deltaPos(*h.Range, len("range"))
	case h.Ordered != nil:
		return deltaPos(*h.Ordered, len("ordered"))
	case h.Colon != nil:
		return deltaPos(*h.Colon, len(":"))
	case h.EqualSign != nil:
		return deltaPos(*h.EqualSign, len("="))
	case h.Var2 != nil:
		return h.Var2.End()
	case h.Var1 != nil:
		return h.Var1.End()
	}
	return Position{}
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
