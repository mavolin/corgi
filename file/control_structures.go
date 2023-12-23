package file

// ============================================================================
// If
// ======================================================================================

// If represents an 'if' statement.
type If struct {
	// Condition is the condition of the if statement.
	Condition IfCondition

	// Then is scope of the code that is executed if the condition evaluates
	// to true.
	Then Body

	// ElseIfs are the else if statements, if this If has any.
	ElseIfs []ElseIf
	// Else is the scope of the Else statement, if this If has one.
	Else *Else

	Position
}

func (If) _scopeItem() {}

// ElseIf represents an 'else if' statement.
type ElseIf struct {
	// Condition is the condition of the else if statement.
	Condition Expression

	// Then is scope of the code that is executed if the condition evaluates
	// to true.
	Then Body

	Position
}

type Else struct {
	Then Scope

	Position
}

// ============================================================================
// IfCondition
// ======================================================================================

// IfCondition is either an [Expression] or a [BlockCondition].
type IfCondition interface {
	_ifCondition()
	Poser
}

// BlockCondition represents the special block existence check function.
type BlockCondition struct {
	// Block is the name of the block.
	Block Ident

	Position
}

func (BlockCondition) _ifCondition() {}

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

func (Switch) _scopeItem() {}

// ======================================== Case ========================================

type Case struct {
	// Expression is the expression written behind 'case'.
	//
	// Nil for the default case.
	Expression IfCondition
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
	Body       Body

	Position
}

func (For) _scopeItem() {}

// ============================================================================
// Return
// ======================================================================================

type Return struct {
	Err *Expression
	Position
}

func (Return) _scopeItem() {}

// ============================================================================
// Break
// ======================================================================================

type Break struct {
	Position
}

func (Break) _scopeItem() {}

// ============================================================================
// Continue
// ======================================================================================

type Continue struct {
	Position
}

func (Continue) _scopeItem() {}
