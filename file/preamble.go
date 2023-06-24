package file

// ============================================================================
// Extend
// ======================================================================================

type Extend struct {
	// Path to the file.
	Path String
	File *File

	Position
}

// ============================================================================
// Import
// ======================================================================================

// Import represents a single import.
type Import struct {
	Imports []ImportSpec

	// Position is the position of the import keyword.
	// Hence, multiple imports may share the same position, if they are
	// grouped in a block.
	Position
}

type ImportSpec struct {
	// Alias is the alias of the import, if any.
	Alias *GoIdent
	Path  String

	Position

	// Source points to the file that made this import.
	//
	// This field will only be set after linking, and will only be set if
	// the import was made in a file different from the one containing this
	// import.
	//
	// Source will be the first file encountered that made this import.
	// However, there may be other files that also imported the same
	// package.
	Source *File
}

// ============================================================================
// Use
// ======================================================================================

type Use struct {
	Uses []UseSpec

	// Position is the position of the use keyword.
	// Hence, multiple uses may share the same position, if they are
	// grouped in a block.
	Position
}

type UseSpec struct {
	// Alias is the alias of the used files, if any.
	Alias *Ident
	// Path is the path to the directory or file.
	Path String

	Position

	// Library is the used library.
	//
	// It is filled by the linker.
	Library *Library
}

// ============================================================================
// Func
// ======================================================================================

// Func is the function header for the generated function.
type Func struct {
	// Name is the name of the function.
	Name GoIdent

	LParenPos Position
	Params    []FuncParam
	RParenPos Position

	Position
}

type FuncParam struct {
	Names    []GoIdent
	Variadic bool
	Type     GoType
}

func (p FuncParam) Pos() Position {
	if len(p.Names) == 0 {
		return InvalidPosition
	}

	return p.Names[0].Pos()
}
