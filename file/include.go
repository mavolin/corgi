package file

type Include struct {
	// Path is the path to the file to include.
	Path String

	// Include is the included file.
	// It is populated by the linker.
	Include IncludeFile

	Position
}

var _ ScopeItem = Include{}

func (Include) _typeScopeItem() {}

// IncludeFile is the type used to represent an included file.
//
// Its concrete type is either a CorgiInclude or a OtherInclude.
type IncludeFile interface {
	_typeIncludeFile()
}

// ============================================================================
// Corgi Include
// ======================================================================================

type CorgiInclude struct {
	File *File
}

func (CorgiInclude) _typeIncludeFile() {}

// ============================================================================
// Other Include
// ======================================================================================

// OtherInclude represents an included file other than a Corgi file.
type OtherInclude struct {
	Contents string
}

func (OtherInclude) _typeIncludeFile() {}
