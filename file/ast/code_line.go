package ast

// ============================================================================
// Implicit Code Line
// ======================================================================================

type ImplicitCodeLine struct {
	Statement *Statement
}

var (
	_ ScopeNode    = (*ImplicitCodeLine)(nil)
	_ TopLevelNode = (*ImplicitCodeLine)(nil)
)

func (i *ImplicitCodeLine) Start() Position {
	if i.Statement != nil {
		if start := i.Statement.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (i *ImplicitCodeLine) End() Position {
	if i.Statement != nil {
		if end := i.Statement.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (i *ImplicitCodeLine) Walk(w func(Node)) {
	if i.Statement != nil {
		w(i.Statement)
	}
}

func (*ImplicitCodeLine) _node()         {}
func (*ImplicitCodeLine) _scopeNode()    {}
func (*ImplicitCodeLine) _topLevelNode() {}

// ============================================================================
// Explicit Code Line
// ======================================================================================

type ExplicitCodeLine struct {
	Minus     *Position
	Statement *Statement
}

var _ ScopeNode = (*ExplicitCodeLine)(nil)

func (e *ExplicitCodeLine) Start() Position {
	if e.Minus != nil {
		return *e.Minus
	}
	if e.Statement != nil {
		if start := e.Statement.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (e *ExplicitCodeLine) End() Position {
	if e.Statement != nil {
		if end := e.Statement.End(); end != NoPosition {
			return end
		}
	}
	if e.Minus != nil {
		return *e.Minus
	}
	return NoPosition
}

func (e *ExplicitCodeLine) Walk(w func(Node)) {
	if e.Statement != nil {
		w(e.Statement)
	}
}

func (*ExplicitCodeLine) _node()      {}
func (*ExplicitCodeLine) _scopeNode() {}
