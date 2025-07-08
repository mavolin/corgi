package ast

// ============================================================================
// Implicit Code Line
// ======================================================================================

type ImplicitCodeLine struct {
	Statement *Statement
}

var _ ScopeNode = (*ImplicitCodeLine)(nil)

func (i *ImplicitCodeLine) Start() Position {
	if i.Statement != nil {
		return i.Statement.Start()
	}
	return Position{}
}

func (i *ImplicitCodeLine) End() Position {
	if i.Statement != nil {
		return i.Statement.End()
	}
	return Position{}
}

func (i *ImplicitCodeLine) Walk(w func(Node)) {
	if i.Statement != nil {
		w(i.Statement)
	}
}

func (*ImplicitCodeLine) _node()      {}
func (*ImplicitCodeLine) _scopeNode() {}

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
	} else if e.Statement != nil {
		return e.Statement.Start()
	}
	return Position{}
}

func (e *ExplicitCodeLine) End() Position {
	if e.Statement != nil {
		return e.Statement.End()
	} else if e.Minus != nil {
		return *e.Minus
	}
	return Position{}
}

func (e *ExplicitCodeLine) Walk(w func(Node)) {
	if e.Statement != nil {
		w(e.Statement)
	}
}

func (*ExplicitCodeLine) _node()      {}
func (*ExplicitCodeLine) _scopeNode() {}
