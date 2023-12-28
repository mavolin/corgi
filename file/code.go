package file

// ============================================================================
// Code
// ======================================================================================

type Code struct {
	Statements []GoCode
	// Implicit indicates whether this code was implicitly detected as such,
	// i.e. it didn't use the '-' operator.
	//
	// This field has no relevance for global code and may be any value.
	Implicit bool
	Position
}

func (Code) _scopeItem() {}

// ============================================================================
// Return
// ======================================================================================

type Return struct {
	Err *GoCode
	Position
}

func (Return) _scopeItem() {}

// ============================================================================
// Break
// ======================================================================================

type Break struct {
	Label *Ident
	Position
}

func (Break) _scopeItem() {}

// ============================================================================
// Continue
// ======================================================================================

type Continue struct {
	Label *Ident
	Position
}

func (Continue) _scopeItem() {}
