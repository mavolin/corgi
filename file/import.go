package file

import "github.com/mavolin/corgi/v2/file/ast"

type Import struct {
	//
	// BUILD SYMBOLS

	// AST is the AST node of the import, if this package was explicitly
	// imported.
	AST *ast.ImportSpec

	Alias Qualifier // may be empty
	// CorgiPath is the import path as found in the corgi file it was sourced
	// from.
	//
	// Only set for explicit imports.
	CorgiPath CorgiImportPath
	// GoPath is the Go import path of the import.
	//
	// For explicit imports, the Go import path is determined by the linker.
	//
	// See the documentation of [Package].GoImportPath for more information.
	//
	// Always set for implicit imports, and always set for explicit imports
	// that were successfully loaded.
	GoPath GoImportPath

	//
	// LINKER

	// Loaded indicates the linker determined that this import is
	// relevant, and it attempted to load the package.
	//
	// If true, but Package is nil, the linker encountered an error while
	// loading the package.
	Loaded bool

	// Package is the package this import resolves to.
	//
	// The linker will not load the packages of implicitly imported packages,
	// the only exception being a builtin package, if provided.
	Package *Package

	// Qualifier is the qualifier of the import.
	//
	// The responsibility of setting this field depends on whether the import
	// is explicit or implicit:
	//
	// For explicit imports, it is the linker's responsibility to set this
	// field, as it loads the package and reads the package name.
	// If the linker chooses not to load this import, the Qualifier field
	// may remain empty.
	//
	// For implicit imports, it is the responsibility of the adder of the
	// import to set this field.
	//
	// All forwarded imports must have a valid qualifier.
	//
	// For dot imports, this field is set to the empty sting.
	//
	// The corgi module reserves all namespaces prefixed with "__corgi_".
	Qualifier Qualifier

	// Forward indicates whether this import should be forwarded, i.e. included,
	// in the output file's list of imports.
	//
	// Like with the Qualifier field, this is set by the linker for explicit
	// imports, and by the adder of the import for implicit imports.
	//
	// All forwarded imports must have a valid, unique, qualifier.
	// All forwarded explicit imports must have a valid Package.
	Forward bool

	// Builtin indicates that this is the single builtin import for the file.
	Builtin bool
}

func (imp *Import) Explicit() bool { return imp.AST != nil }
func (imp *Import) Implicit() bool { return !imp.Explicit() }

// EnsureUniqueQualifier ensures that the import's qualifier is unique
// within the file's symbols.
//
// If the qualifier is already taken, it appends underscores until it is
// unique and sets the import's Alias to the new qualifier.
func (imp *Import) EnsureUniqueQualifier(f *File) {
	if f.ImportByQualifier(imp.Qualifier) == nil {
		return
	}

	imp.Qualifier = f.UniqueQualifier(imp.Qualifier)
	imp.Alias = imp.Qualifier
}
