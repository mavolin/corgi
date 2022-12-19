package file

// File represents a parsed corgi file.
type File struct {
	// Name is the name of the file.
	//
	// If this is the main file, it will be just the file's name.
	//
	// If this is an extended, included, or used file, this will be the path
	// of the specified file in its source.
	Name string
	// Source is the name of the source of this file.
	Source string

	// Extend is the file that this file extends, if any.
	Extend *Extend

	// Imports is a list of imports made by this file.
	//
	// After linking, it will also include all imports made by used, extended
	// and included files.
	Imports []Import

	// Uses is a list of used library files.
	Uses []Use

	// GlobalCode is the code that is written above the output function.
	GlobalCode []Code

	// Func is the function header.
	// It is always present for main files, i.e. those files that are given
	// to the corgi command.
	Func Func

	// Scope is the global scope.
	Scope Scope
}

// Pos indicates the position where a token was encountered.
type Pos struct {
	Line int
	Col  int
}

func (p Pos) Position() (line, col int) {
	return p.Line, p.Col
}

type (
	Extend struct {
		// Path to the file.
		Path String
		File File

		Pos
	}

	// Import represents a single import.
	Import struct {
		// Alias is the alias of the import, if any.
		Alias GoIdent
		Path  String

		// Source is the source of the first file that made this import.
		Source string
		// File is the name of the first file that made this import.
		File string

		Pos
	}

	// Use represents a single use directive.
	Use struct {
		// Namespace is the namespace of the used files.
		Namespace Ident

		// Path is the path to the directory or file.
		Path string

		// Files are the files included by this use directive.
		// It is filled by the linker.
		Files []File

		Pos
	}

	// Func is the function header for the generated function.
	Func struct {
		// Name is the name of the function.
		Name GoIdent
		// Params are the parameters of the function.
		// They are enclosed in parentheses.
		Params GoExpression

		Pos
	}
)
