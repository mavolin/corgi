package file

// ============================================================================
// If
// ======================================================================================

// ===================================== Regular If =====================================

// If represents an 'if' statement.
type If struct {
	// Condition is the condition of the if statement.
	Condition Expression

	// Then is scope of the code that is executed if the condition evaluates
	// to true.
	Then Scope

	// ElseIfs are the else if statements, if this If has any.
	ElseIfs []ElseIf
	// Else is the scope of the Else statement, if this If has one.
	Else *Else

	Position
}

var _ ScopeItem = If{}

func (If) _typeScopeItem() {}

// ElseIf represents an 'else if' statement.
type ElseIf struct {
	// Condition is the condition of the else if statement.
	Condition Expression

	// Then is scope of the code that is executed if the condition evaluates
	// to true.
	Then Scope

	Position
}

// ====================================== If Block ======================================

// IfBlock represents an 'if block' directive.
type IfBlock struct {
	// Name is the name of the block, whose existence is checked.
	Name Ident

	// Then is the scope of the code that is executed if the block exists.
	Then Scope
	// ElseIfs are the else if statements, if this IfBlock has any.
	ElseIfs []ElseIfBlock
	// Else is the scope of the code that is executed if the block does not
	// exist.
	Else *Else

	Position
}

var _ ScopeItem = IfBlock{}

func (IfBlock) _typeScopeItem() {}

type ElseIfBlock struct {
	// Name is the name of the block, whose existence is checked.
	Name Ident
	// Then is the scope of the code that is executed if the block exists.
	Then Scope

	Position
}

// ======================================== Else ========================================

type Else struct {
	Then Scope
	Position
}

// ============================================================================
// Switch
// ======================================================================================

// Switch represents a 'switch' statement.
type Switch struct {
	// Comparator is the expression that is compared against.
	//
	// It may be nil, in which case the cases will contain boolean
	// expressions.
	Comparator *Expression

	// Cases are the cases of the Switch.
	Cases []Case
	// Default is the default case, if there is one.
	Default *Case

	Position
}

var _ ScopeItem = Switch{}

func (Switch) _typeScopeItem() {}

// ======================================== Case ========================================

type Case struct {
	// Expression is the expression written behind 'case'.
	//
	// Nil for the default case.
	Expression *Expression
	// Then is the scope of the code that is executed if the condition
	// evaluates to true.
	Then Scope

	Position
}

// ============================================================================
// For
// ======================================================================================

// For represents a for loop.
type For struct {
	// Expression is the expression written in the head of the for, or nil if
	// this is an infinite loop.
	Expression *Expression
	Body       Scope

	Position
}

func (For) _typeScopeItem() {}
