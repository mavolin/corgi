package file

type BadItem struct {
	// Line contains the entire bad line, minus indentation.
	Line string
	// Body contains the body of the item, if it has any.
	Body Scope
	Position
}

func (b BadItem) _typeScopeItem() {}

var _ ScopeItem = BadItem{}
