package file

// ============================================================================
// If
// ======================================================================================

type If struct {
	Condition IfExpression
	Then      Body

	ElseIfs []ElseIf
	Else    *Else

	Position
}

func (If) _scopeItem() {}

type ElseIf struct {
	Condition IfExpression
	Then      Body

	Position
}

type Else struct {
	Then Body
	Position
}

// ============================================================================
// Switch
// ======================================================================================

type Switch struct {
	Comparator *GoCode // nil for case conditions
	Cases      []Case

	Position
}

func (Switch) _scopeItem() {}

// ======================================== Case ========================================

type Case struct {
	Expression Expression // nil for default case
	Colon      Position
	Then       Scope // has no [LR]Brace set

	Position
}

// ============================================================================
// For
// ======================================================================================

type For struct {
	Expression ForExpression // nil for infinite loop
	Body       Body

	Position
}

func (For) _scopeItem() {}
