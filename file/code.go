package file

// Code represents a line or block of code.
type Code struct {
	Lines []CodeLine
	Position
}

var _ ScopeItem = Code{}

func (Code) _typeScopeItem() {}

type CodeLine struct {
	// Code is the code in the line.
	Code string

	Position
}
