package file

type Filter interface {
	_typeFilter()
	ScopeItem
}

type FilterLine struct {
	Line string
	Position
}

// ============================================================================
// Command Filter
// ======================================================================================

type CommandFilter struct {
	Name string
	Args []CommandFilterArg

	Body []FilterLine

	Position
}

var _ Filter = CommandFilter{}

func (CommandFilter) _typeFilter()    {}
func (CommandFilter) _typeScopeItem() {}

type CommandFilterArg interface {
	_typeCommandFilterArg()
	Poser
}

type RawCommandFilterArg struct {
	Value string
	Position
}

func (RawCommandFilterArg) _typeCommandFilterArg() {}

type StringCommandFilterArg String

func (StringCommandFilterArg) _typeCommandFilterArg() {}

// ============================================================================
// Raw Filter
// ======================================================================================

type RawFilter struct {
	Body []FilterLine
	Position
}

var _ Filter = RawFilter{}

func (RawFilter) _typeFilter()    {}
func (RawFilter) _typeScopeItem() {}
