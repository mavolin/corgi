package file

// File represents a parsed corgi file.
type File struct {
	// Name is the name of the file.
	//
	// If this is the main file, it will be just the file's name.
	//
	// If this is an extended, included, or used file, this will be the path
	// specified to extend/include/use the file.
	// A file extension will be added, if the path did not include one.
	//
	// If this was parsed as part of a use directive on a directory, the Name
	// will be the path to the directory + "/" + the file's name.
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

	// GlobalCode is the code that is written outside the function body.
	//
	// Groupings are not preserved as they have no semantic influence.
	GlobalCode []Code

	// Func is the function header.
	// It is always present for main files, i.e. those files that are given
	// to the corgi command.
	Func Func

	// Scope is the global scope.
	Scope Scope
}

// Pos indicates the position where an element was encountered.
//
// It is not present for all elements, but only where needed to generated
// meaningful lexerr during linking.
type Pos struct {
	Line int
	Col  int
}

type (
	Extend struct {
		// Path is the unquoted path to the file.
		Path string
		File File

		Pos
	}

	// Import represents a single import.
	Import struct {
		// Alias is the alias of the import, if any.
		Alias GoIdent
		// Path is the unquoted path of the import.
		Path string

		Pos
		// Source is the source of the first file that made this import.
		Source string
		// File is the first file that made this import.
		File string
	}

	// Use represents a single use directive.
	Use struct {
		// Namespace is the namespace of the used files.
		Namespace Ident

		// Path is the path to the directory or file.
		Path string

		// Files are the files included by this use directive.
		// It is filled by the linker
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
	}
)
