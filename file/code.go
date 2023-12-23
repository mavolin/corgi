package file

// Code represents a line or block of code.
type Code struct {
	Lines []CodeLine
	// Implicit indicates whether this code was implicitly detected as such,
	// i.e. it didn't use the '-' operator.
	Implicit bool
	Position
}

func (Code) _scopeItem() {}

type CodeLine struct {
	Code string
	Position
}
