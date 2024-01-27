package ast

// ============================================================================
// If
// ======================================================================================

type If struct {
	Condition *IfExpression
	Then      Body

	ElseIfs []*ElseIf
	Else    *Else // may be nil

	Position Position
}

var _ ScopeItem = (*If)(nil)

func (i *If) Pos() Position { return i.Position }
func (*If) _scopeItem()     {}

type ElseIf struct {
	Condition *IfExpression
	Then      Body
	Position  Position
}

type Else struct {
	Then     Body
	Position Position
}

// ============================================================================
// Switch
// ======================================================================================

type Switch struct {
	Comparator *GoCode // nil for case conditions
	Cases      []*Case
	Position   Position
}

var _ ScopeItem = (*Switch)(nil)

func (s *Switch) Pos() Position { return s.Position }
func (*Switch) _scopeItem()     {}

// ======================================== Case ========================================

type Case struct {
	Expression Expression // nil for default case
	Colon      *Position
	Then       *Scope // has no L-/RBrace set

	Position Position
}

// ============================================================================
// For
// ======================================================================================

type For struct {
	Expression ForExpression // nil for infinite loop
	Body       Body
	Position   Position
}

var _ ScopeItem = (*For)(nil)

func (f *For) Pos() Position { return f.Position }
func (*For) _scopeItem()     {}
