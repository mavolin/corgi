// Package file represents a high-level view around the File of a corgi file.
// This, most prominently, includes linked imports, component calls and additional
// metadata, as well as the result of [github.com/mavolin/corgi/load/analyze.Analyze].
package file

import (
	"cmp"
	"fmt"
	"path"
	"slices"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
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

	// METADATA
	//

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

type Symbols struct {
	Imports []*Import

	ComponentCalls       []*ComponentCall
	componentCallsByNode map[*ast.ComponentCall]*ComponentCall

	ElementReferences       []*ElementReference
	elementReferencesByNode map[*ast.ElementReference]*ElementReference

	AttributeReferences       []*AttributeReference
	attributeReferencesByNode map[*ast.AttributeReference]*AttributeReference
}

func buildSymbols(f *File) {
	var nImports int
	for _, imp := range f.AST.Imports {
		nImports += len(imp.Specs)
	}

	f.Symbols = &Symbols{
		Imports: make([]*Import, 0, nImports),

		ComponentCalls:       make([]*ComponentCall, 0, 256),
		componentCallsByNode: make(map[*ast.ComponentCall]*ComponentCall, 256),

		ElementReferences:       make([]*ElementReference, 0, 256),
		elementReferencesByNode: make(map[*ast.ElementReference]*ElementReference, 256),

		AttributeReferences:       make([]*AttributeReference, 0, 512),
		attributeReferencesByNode: make(map[*ast.AttributeReference]*AttributeReference, 512),
	}

	for _, impStmt := range f.AST.Imports {
		for _, spec := range impStmt.Specs {
			imp := &Import{AST: spec}
			if spec.Alias != nil {
				imp.Alias = spec.Alias.Name
			}
			if spec.Path != nil {
				imp.Path = spec.Path.Unquote()
			}
			f.Imports = append(f.Imports, imp)
		}
	}

	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		switch n := n.(type) {
		case *ast.ComponentCall:
			ccw := &ComponentCall{AST: n, File: f}
			f.ComponentCalls = append(f.ComponentCalls, ccw)
			f.componentCallsByNode[n] = ccw
		case *ast.ElementReference:
			rw := &ElementReference{AST: n}
			f.ElementReferences = append(f.ElementReferences, rw)
			f.elementReferencesByNode[n] = rw
		case *ast.AttributeReference:
			rw := &AttributeReference{AST: n}
			f.AttributeReferences = append(f.AttributeReferences, rw)
			f.attributeReferencesByNode[n] = rw
		}
		n.Walk(walk)
	}
	for _, n := range f.AST.TopLevel {
		switch n := n.(type) {
		case *ast.Component:
			ccsStart := len(f.ComponentCalls)
			n.Walk(walk)
			ccEnd := len(f.ComponentCalls)
			if ccEnd > ccsStart {
				f.Package.ComponentByNode(n).ComponentCalls = f.ComponentCalls[ccsStart:ccEnd:ccEnd]
			}
		case *ast.Alias:
			ccsStart := len(f.ComponentCalls)
			if n.Header != nil {
				n.Header.Walk(walk)
			}
			ccw := &ComponentCall{
				AST:      n.ComponentCall,
				AliasFor: f.Package.AliasByNode(n),
				File:     f,
			}
			f.ComponentCalls = append(f.ComponentCalls, ccw)
			f.componentCallsByNode[n.ComponentCall] = ccw
			if n.ComponentCall.Header != nil {
				n.ComponentCall.Header.Walk(walk)
			}
			if n.ComponentCall.Body != nil {
				n.ComponentCall.Body.Walk(walk)
			}
			ccsEnd := len(f.ComponentCalls)
			ccw.AliasFor.ComponentCalls = f.ComponentCalls[ccsStart:ccsEnd:ccsEnd]
		}
	}

	f.ComponentCalls = slices.Clip(f.ComponentCalls)
	f.ElementReferences = slices.Clip(f.ElementReferences)
	f.AttributeReferences = slices.Clip(f.AttributeReferences)
}

// AddBuiltinImport creates a new [Import] importing the given builtin package.
// The import is marked as not forwarded by default.
// You may choose to not use any alias at all.
//
// The file must not already have a builtin import or use the given alias.
// The function returns a pointer to the created import, which may also be
// retrieved by calling [Symbols.BuiltinImport].
//
// The package must not contain any exported symbols.
func (s *Symbols) AddBuiltinImport(alias string, builtin *Package) {
	if builtinImp := s.BuiltinImport(); builtinImp != nil {
		panic(fmt.Sprintf("symbols already contain builtin import for %q", builtinImp.Path))
	}

	namespace := cmp.Or(alias, builtin.Name)
	if imp := s.ImportByNamespace(namespace); imp != nil {
		panic(fmt.Sprintf("symbols already contain import with namespace %s: you need to chose a (different) alias", namespace))
	}

	imp := &Import{
		Alias:     alias,
		Path:      builtin.ImportPath,
		Package:   builtin,
		Namespace: "",
	}
	s.Imports = append(s.Imports, imp)
}

// AddImport adds the given import to the file.
// Always use this method over manipulating the [Symbols.Imports] slice
// directly.
//
// AddImport panics if any of the following conditions are violated:
//   - The import's namespace must match the alias, if set.
//   - The import must not be a dot import
//   - The import must not have the AST field set.
//   - The file must not already have an import with the namespace.
//   - The import must be marked as forwarded: There would be no point in
//     adding an implicit import that is not forwarded, unless you are
//     doing sketchy AST manipulation (that should've happened before building
//     symbols instead).
func (s *Symbols) AddImport(imp *Import) {
	if imp.Alias != "" && imp.Alias != imp.Namespace {
		panic(fmt.Sprintf("import alias %s does not match namespace %s", imp.Alias, imp.Namespace))
	} else if imp.Alias == "." {
		panic("cannot add implicit dot import")
	} else if imp := s.ImportByNamespace(imp.Namespace); imp != nil {
		panic(fmt.Sprintf("symbols already contain import with namespace %s: you need to chose a (different) alias", imp.Namespace))
	} else if imp.AST != nil {
		panic("cannot add implicit import with AST set")
	} else if !imp.Forward {
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

func (s *Symbols) ImportByPackage(p *Package) *Import {
	for _, imp := range s.Imports {
		if imp.Package == p {
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
	return s.ImportByNamespace("")
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

type Import struct {
	// BUILD SYMBOLS
	//

	// AST is the AST node of the import, if this package was explicitly
	// imported.
	AST *ast.ImportSpec

	Alias string // may be empty
	// Path is the import path.
	//
	// Guaranteed to be non-empty for both explicit and implicit imports.
	Path string

	// LINKER
	//

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
	//
	// For implicit imports, it is the responsibility of the adder of the
	// import to set this field.
	//
	// All forwarded imports must have a valid namespace.
	//
	// For dot imports, this field is set to "."
	//
	// For the builtin package, this field is set to the empty string.
	// The builtin package is the only package where Namespace differs from the
	// computed namespace, i.e. where Namespace is neither the Alias, if set,
	// nor the package name as specified in the source files.
	//
	// Therefore, when outputting the file, rely on Alias to produce correct
	// import statement aliases.
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
}

func (imp *Import) Explicit() bool { return imp.AST != nil }
func (imp *Import) Implicit() bool { return !imp.Explicit() }

type ElementReference struct {
	// BUILD SYMBOLS
	//

	AST *ast.ElementReference

	// LINKER
	//

	// Spec is the spec providing the type of the Element.
	Spec *ElementSpec
}

// Type returns the type of the element.
//
// Only returns a valid type after linking.
// Returns [elemtype.Unknown] if there was a linker error, or if called before
// linking.
func (r *ElementReference) Type() elemtype.Type {
	if r.Spec != nil {
		return r.Spec.Type
	}
	return elemtype.Unknown
}

type AttributeReference struct {
	// BUILD SYMBOLS
	//

	AST *ast.AttributeReference

	// LINKER
	//

	// Spec is the spec declaring the attribute.
	//
	// Since attributes can also be explicitly typed, this field may be
	// nil.
	// Hence, linker implementations should not report errors if they
	// cannot resolve the spec that belongs to the reference.
	Spec *AttributeSpec // may be nil

	// ANALYZER
	//

	AnalyzedWithErrors bool

	Element *ElementSpec // nil if not attached to an element
	// Rule is the rule that is relevant for the element/attribute pair.
	Rule *ast.AttributeRule // nil if not attached to an element
	Type attrtype.Type
}

// Name returns the name of the attribute.
//
// Can only be called after successful linking.
func (r *AttributeReference) Name() string {
	// possibly has a prefix
	if r.AST.Package != nil {
		if r.Spec == nil { // externally defined attribute, but no spec?
			// Every attribute reference with a package name set, will have its
			// spec set by the linker, because the package name alone means
			// that the attribute is defined in the package, making it a linker
			// issue.
			// Ergo, someone called this method before linking, or there are
			// linker errors.
			panic("AttributeReference.Name called before linking or with linker errors")
		}

		// prepend the prefix
		if r.Spec.Definition.Prefix != nil {
			return r.Spec.Definition.Prefix.Name + r.AST.Name.Name
		}

		// fallthrough, no prefix
	}

	return r.AST.Name.Name
}
