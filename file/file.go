// Package file represents a high-level view around the File of a corgi file.
// This, most prominently, includes linked imports, component calls and additional
// metadata, as well as the result of [github.com/mavolin/corgi/load/analyze.Analyze].
package file

import (
	"cmp"
	"fmt"
	"path"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

const (
	EscapeImport  = "github.com/mavolin/corgi/v2/escape"
	SafeImport    = "github.com/mavolin/corgi/v2/escape/safe"
	RuntimeImport = "github.com/mavolin/corgi/v2/runtime"
)

// File represents a parsed corgi file.
type File struct {
	Package *Package

	//
	// METADATA

	// Name is the name of the file.
	Name string

	AST *ast.File

	*Symbols
}

func (f *File) ModulePath() string {
	return path.Join(f.Package.ModulePath(), f.Name)
}

func (f *File) PathInModule() string {
	return path.Join(f.Package.PathInModule, f.Name)
}

// Symbols contains the symbols of the file.
//
// Refer to [PackageSymbols] and [BuildSymbols] for more information.
type Symbols struct {
	Imports []*Import

	ComponentCalls       []*ComponentCall
	componentCallsByNode map[*ast.ComponentCall]*ComponentCall

	ElementReferences       []*ElementReference
	elementReferencesByNode map[*ast.ElementReference]*ElementReference

	AttributeReferences       []*AttributeReference
	attributeReferencesByNode map[*ast.AttributeReference]*AttributeReference
}

// AddBuiltinImport creates a new [Import] importing the given builtin package.
// The import is marked as not forwarded by default.
//
// You needn't specify an alias, however, the alias must not be ".".
//
// The file must not already have a builtin import or use the given alias.
//
// The package must not contain any exported symbols.
//
// You must add a builtin import using this method, not by adding it to the
// [Symbols.Imports] slice directly.
func (s *Symbols) AddBuiltinImport(alias string, builtin *Package) {
	imp := &Import{
		Alias:     alias,
		Path:      builtin.ImportPath,
		Package:   builtin,
		Namespace: cmp.Or(alias, builtin.Name),
		Builtin:   true,
	}
	s.AddImport(imp)
}

// AddImport adds the given import to the file.
// Always use this method over manipulating the [Symbols.Imports] slice
// directly.
//
// AddImport panics if any of the following conditions are violated:
//   - If the import is a builtin import, the file must not already have a
//     builtin import.
//   - The import's namespace must match the alias, if set.
//   - If implicit, the import must not be a dot import.
//   - The file must not already have an import with the namespace.
//   - The import must be marked as forwarded, unless it is implicit or the
//     builtin import.
func (s *Symbols) AddImport(imp *Import) {
	switch {
	case imp.Builtin && s.BuiltinImport() != nil:
		panic(fmt.Sprintf("symbols already contain builtin import for %q", s.BuiltinImport().Path))
	case imp.Implicit() && imp.Alias == ".":
		panic("cannot add implicit dot import")
	case !imp.Builtin && imp.Alias != "" && imp.Alias != imp.Namespace:
		panic(fmt.Sprintf("import alias %s does not match namespace %s", imp.Alias, imp.Namespace))
	case s.ImportByNamespace(imp.Namespace) != nil:
		panic(fmt.Sprintf("symbols already contain import with namespace %s: you need to chose a (different) alias", imp.Namespace))
	case !imp.Implicit() && !imp.Builtin && !imp.Forward:
		panic("cannot add implicit import that is not forwarded")
	}

	s.Imports = append(s.Imports, imp)
}

// ImportByNamespace returns the first import with the given namespace.
//
// Does not work for the "." namespace.
//
// Only available after linking.
func (s *Symbols) ImportByNamespace(namespace string) *Import {
	for _, imp := range s.Imports {
		if imp.Namespace == namespace {
			return imp
		}
	}
	return nil
}

func (s *Symbols) ImportByPath(p string) *Import {
	for _, imp := range s.Imports {
		if imp.Path == p {
			return imp
		}
	}
	return nil
}

func (s *Symbols) ImportByNode(node *ast.ImportSpec) *Import {
	for _, imp := range s.Imports {
		if imp.AST == node {
			return imp
		}
	}
	return nil
}

func (s *Symbols) BuiltinImport() *Import {
	for _, imp := range s.Imports {
		if imp.Builtin {
			return imp
		}
	}
	return nil
}

func (s *Symbols) ComponentCallByNode(node *ast.ComponentCall) *ComponentCall {
	return s.componentCallsByNode[node]
}

func (s *Symbols) ElementReferenceByNode(node *ast.ElementReference) *ElementReference {
	return s.elementReferencesByNode[node]
}

func (s *Symbols) AttributeReferenceByNode(node *ast.AttributeReference) *AttributeReference {
	return s.attributeReferencesByNode[node]
}

// RebuildLookupTables rebuilds the lookup tables used by the methods of this
// type.
//
// Every you modify the slices of this struct directly, you must call this
// method to ensure that the lookup tables are up-to-date.
func (s *Symbols) RebuildLookupTables() {
	s.componentCallsByNode = make(map[*ast.ComponentCall]*ComponentCall, len(s.ComponentCalls))
	for _, cc := range s.ComponentCalls {
		s.componentCallsByNode[cc.AST] = cc
	}

	s.elementReferencesByNode = make(map[*ast.ElementReference]*ElementReference, len(s.ElementReferences))
	for _, ref := range s.ElementReferences {
		s.elementReferencesByNode[ref.AST] = ref
	}

	s.attributeReferencesByNode = make(map[*ast.AttributeReference]*AttributeReference, len(s.AttributeReferences))
	for _, ref := range s.AttributeReferences {
		s.attributeReferencesByNode[ref.AST] = ref
	}
}

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

	// Package is the package this import resolves to.
	//
	// The linker will not load the packages of implicitly imported packages,
	// the only exception being a builtin package, if provided.
	Package *Package

	// LoadedWithErrors indicates that the linker wanted to load the package,
	// but encountered an error while doing so.
	//
	// Details will be available in the diagnostics returned by the linker.
	LoadedWithErrors bool

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

type ElementReference struct {
	//
	// BUILD SYMBOLS

	AST *ast.ElementReference

	//
	// LINKER

	// Spec is the spec providing the type of the Element.
	Spec *ElementSpec
}

// HTMLName returns the name of the element.
//
// Can only be called after successful linking.
func (r *ElementReference) HTMLName() string {
	return r.Spec.HTMLName()
}

type AttributeReference struct {
	//
	// BUILD SYMBOLS

	AST *ast.AttributeReference

	//
	// LINKER

	// Spec is the spec declaring the attribute.
	//
	// Since attributes can also be explicitly typed, this field may be
	// nil.
	// Hence, linker implementations should not report errors if they
	// cannot resolve the spec that belongs to the reference.
	Spec *AttributeSpec // may be nil

	//
	// ANALYZER

	AnalyzedWithErrors bool

	Element *ElementSpec // nil if not attached to an element
	// Rule is the rule that is relevant for the element/attribute pair.
	Rule *ast.AttributeRule // nil if not attached to an element
	Type attrtype.Type
}

// HTMLName returns the name of the attribute.
//
// Can only be called after successful linking.
func (r *AttributeReference) HTMLName() string {
	// possibly has a prefix
	if r.AST.Package != nil {
		if r.Spec == nil { // externally defined attribute, but no spec?
			// Every attribute reference with a package name set, will have its
			// spec set by the linker, because the package name alone means
			// that the attribute is defined in the package, making it a linker
			// issue.
			// Ergo, someone called this method before linking, or there are
			// linker errors.
			panic("AttributeReference.HTMLName called before linking or with linker errors")
		}

		// prepend the prefix
		if r.Spec.Definition.Prefix != nil {
			return r.Spec.Definition.Prefix.Name + r.AST.Name.Name
		}

		// fallthrough, no prefix
	}

	return r.AST.Name.Name
}
