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

	AST *ast.File
	*Symbols
}

type (
	Symbols struct {
		Imports            []*Import
		importsByNamespace map[string]*Import

		ComponentCalls       []*ComponentCall
		componentCallsByNode map[*ast.ComponentCall]*ComponentCall

		ElementReferences       []*ElementReference
		elementReferencesByNode map[*ast.ElementReference]*ElementReference

		AttributeReferences       []*AttributeReference
		attributeReferencesByNode map[*ast.AttributeReference]*AttributeReference
	}

	Import struct {
		AST *ast.ImportSpec // buildSymbols

		// Package is the package this import resolves to.
		//
		// This may be nil, if no components are imported from the package.
		Package *Package // linker
	}

	ElementReference struct {
		AST *ast.ElementReference // buildSymbols

		// Definition is the definition providing the type of the Element.
		Definition *ElementDefinition // linker
	}

	AttributeReference struct {
		AST *ast.AttributeReference // buildSymbols

		// Definition is the definition declaring the attribute.
		//
		// Since attributes can also be explicitly typed, this field may be
		// nil.
		// Hence, linker implementations should not report errors if they
		// cannot resolve the definition that belongs to the reference.
		Definition *AttributeDefinition // linker, may be nil

		Element *ElementDefinition // analyze, nil if not attached to an element
		// Rule is the rule that is relevant for the element/attribute pair.
		Rule *ast.AttributeRule // analyze, nil if not attached to an element
		Type attrtype.Type      // analyze
	}
)

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

func (r *ElementReference) Type() elemtype.Type {
	if r.Definition != nil {
		return r.Definition.Type
	}
	return elemtype.Unknown
}

func buildSymbols(f *File) {
	var nImports int
	for _, imp := range f.AST.Imports {
		nImports += len(imp.Specs)
	}

	f.Symbols = &Symbols{
		Imports:            make([]*Import, 0, nImports),
		importsByNamespace: make(map[string]*Import, nImports),

		ComponentCalls:       make([]*ComponentCall, 0, 128),
		componentCallsByNode: make(map[*ast.ComponentCall]*ComponentCall, 128),

		ElementReferences:       make([]*ElementReference, 0, 256),
		elementReferencesByNode: make(map[*ast.ElementReference]*ElementReference, 256),

		AttributeReferences:       make([]*AttributeReference, 0, 512),
		attributeReferencesByNode: make(map[*ast.AttributeReference]*AttributeReference, 512),
	}

	for _, impStmt := range f.AST.Imports {
		for _, spec := range impStmt.Specs {
			imp := &Import{AST: spec}
			f.Symbols.Imports = append(f.Symbols.Imports, imp)
			if ns := imp.Namespace(); ns != "" {
				f.Symbols.importsByNamespace[ns] = imp
			}
		}
	}

	f.Symbols.componentCallsByNode = make(map[*ast.ComponentCall]*ComponentCall)
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		if cc, _ := n.(*ast.ComponentCall); cc != nil {
			ccw := &ComponentCall{AST: cc}
			f.Symbols.ComponentCalls = append(f.Symbols.ComponentCalls, ccw)
			f.Symbols.componentCallsByNode[cc] = ccw
		} else if r, _ := n.(*ast.ElementReference); r != nil {
			rw := &ElementReference{AST: r}
			f.Symbols.ElementReferences = append(f.Symbols.ElementReferences, rw)
			f.Symbols.elementReferencesByNode[r] = rw
		} else if r, _ := n.(*ast.AttributeReference); r != nil {
			rw := &AttributeReference{AST: r}
			f.Symbols.AttributeReferences = append(f.Symbols.AttributeReferences, rw)
			f.Symbols.attributeReferencesByNode[r] = rw
		}
		n.Walk(walk)
	}
	for _, n := range f.AST.TopLevel {
		walk(n)
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

func IsExported(s string) bool {
	if len(s) == 0 {
		return false
	}
	return 'A' <= s[0] && s[0] <= 'Z'
}
