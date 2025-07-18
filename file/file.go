// Package file represents a high-level view around the File of a corgi file.
// This, most prominently, includes linked imports, component calls and additional
// metadata, as well as the result of [github.com/mavolin/corgi/load/analyze.Analyze].
package file

import (
	"path"
	"slices"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

type permanentImport struct {
	Alias string
	Path  string
}

var (
	EscapeImport  = &permanentImport{"__corgi_escape", "github.com/mavolin/corgi/v2/escape"}
	SafeImport    = &permanentImport{"__corgi_safe", "github.com/mavolin/corgi/v2/escape/safe"}
	RuntimeImport = &permanentImport{"__corgi_runtime", "github.com/mavolin/corgi/v2/runtime"}
)

// File represents a parsed corgi file.
type File struct {
	Package *Package

	// METADATA
	//

	// Name is the name of the file.
	Name string

	// Module is the path/name of the Go module providing this file.
	Module string
	// PathInModule is the path to the file in the Go module, relative to the
	// module root.
	//
	// It is always specified as a forward slash separated path.
	PathInModule string

	// NeedsSafeImport indicates whether this file needs to import the safe
	// package.
	NeedsSafeImport bool
	// NeedsEscapeImport indicates whether this file needs to import the
	// escape package.
	NeedsEscapeImport bool

	AST *ast.File

	*Symbols
}

type Symbols struct {
	Imports            []*Import
	importsByNamespace map[string]*Import

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
		Imports:            make([]*Import, 0, nImports),
		importsByNamespace: make(map[string]*Import, nImports),

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
			f.Imports = append(f.Imports, imp)
			if ns := imp.Namespace(); ns != "" {
				f.importsByNamespace[ns] = imp
			}
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
			ccsStart := len(f.Symbols.ComponentCalls)
			n.Walk(walk)
			ccEnd := len(f.Symbols.ComponentCalls)
			if ccEnd > ccsStart {
				f.Package.ComponentByNode(n).ComponentCalls = f.Symbols.ComponentCalls[ccsStart:ccEnd:ccEnd]
			}
		case *ast.Alias:
			ccsStart := len(f.Symbols.ComponentCalls)
			if n.Header != nil {
				n.Header.Walk(walk)
			}
			ccw := &ComponentCall{
				AST:      n.ComponentCall,
				AliasFor: f.Package.AliasByNode(n),
				File:     f,
			}
			f.Symbols.ComponentCalls = append(f.Symbols.ComponentCalls, ccw)
			f.Symbols.componentCallsByNode[n.ComponentCall] = ccw
			if n.ComponentCall.Header != nil {
				n.ComponentCall.Header.Walk(walk)
			}
			if n.ComponentCall.Body != nil {
				n.ComponentCall.Body.Walk(walk)
			}
			ccsEnd := len(f.Symbols.ComponentCalls)
			ccw.AliasFor.ComponentCalls = f.Symbols.ComponentCalls[ccsStart:ccsEnd:ccsEnd]
		}
	}

	f.ComponentCalls = slices.Clip(f.Symbols.ComponentCalls)
	f.ElementReferences = slices.Clip(f.Symbols.ElementReferences)
	f.AttributeReferences = slices.Clip(f.Symbols.AttributeReferences)
}

func (s *Symbols) ImportByNamespace(namespace string) *Import {
	return s.importsByNamespace[namespace]
}

func (s *Symbols) ImportByNode(node *ast.ImportSpec) *Import {
	for _, imp := range s.Imports {
		if imp.AST == node {
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

type Import struct {
	// BUILD SYMBOLS
	//

	AST *ast.ImportSpec

	// LINKER
	//

	// Package is the package this import resolves to.
	//
	// This may be nil, if no components are imported from the package.
	Package *Package
}

func (i *Import) ImportPath() string {
	if i.AST.Path == nil {
		return ""
	}
	return i.AST.Path.Unquote()
}

// Namespace returns the namespace of the import.
// For dot imports, it returns ".".
func (i *Import) Namespace() string {
	if i.AST.Alias != nil {
		return i.AST.Alias.Ident
	}
	return path.Base(i.ImportPath())
}

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
