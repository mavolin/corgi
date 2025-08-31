package file

import "github.com/mavolin/corgi/v2/file/ast"

type Import struct {
	//
	// BUILD SYMBOLS

	// AST is the AST node of the import, if this package was explicitly
	// imported.
	AST *ast.ImportSpec

	Alias string // may be empty
	// Path is the import path.
	//
	// Guaranteed to be non-empty for both explicit and implicit imports.
	Path string

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

	// Namespace is the namespace of the import.
	//
	// The responsibility of setting this field depends on whether the import
	// is explicit or implicit:
	//
	// For explicit imports, it is the linker's responsibility to set this
	// field, as it loads the package and reads the package name.
	// If the linker chooses not to load this import, the Namespace field
	// may remain empty.
	//
	// For implicit imports, it is the responsibility of the adder of the
	// import to set this field.
	//
	// All forwarded imports must have a valid namespace.
	//
	// For dot imports, this field is set to the empty sting.
	//
	// The corgi module reserves all namespaces prefixed with "__corgi_".
	Namespace string

	// Forward indicates whether this import should be forwarded, i.e. included,
	// in the output file's list of imports.
	//
	// Like with the Namespace field, this is set by the linker for explicit
	// imports, and by the adder of the import for implicit imports.
	//
	// All forwarded imports must have a valid, unique, namespace.
	// All forwarded explicit imports must have a valid Package.
	Forward bool

	// Builtin indicates that this is the single builtin import for the file.
	Builtin bool
}

func (imp *Import) Explicit() bool { return imp.AST != nil }
func (imp *Import) Implicit() bool { return !imp.Explicit() }

// EnsureUniqueNamespace ensures that the import's namespace is unique
// within the file's symbols.
//
// If the namespace is already taken, it appends underscores until it is
// unique and returns false.
// Otherwise, it returns true.
func (imp *Import) EnsureUniqueNamespace(s *Symbols) (ok bool) {
	if s.ImportByNamespace(imp.Namespace) == nil {
		return true
	}

	for s.ImportByNamespace(imp.Namespace) != nil {
		imp.Namespace += "_"
	}
	imp.Alias = imp.Namespace
	return false
}
