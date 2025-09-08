// Package file represents a high-level view around the File of a corgi file.
// This, most prominently, includes linked imports, component calls and additional
// metadata, as well as the result of [github.com/mavolin/corgi/load/analyze.Analyze].
package file

import (
	"cmp"
	"fmt"
	"path"
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/meta"
)

const (
	EscapeImport  = meta.Module + "/escape"
	SafeImport    = EscapeImport + "/safe"
	RuntimeImport = meta.Module + "/runtime"
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
	return path.Join(f.Package.GoImportPath(), f.Name)
}

func (f *File) PathInModule() string {
	if f.Package == nil {
		return "<unknown package>/" + f.Name
	}
	return path.Join(f.Package.PathInModule, f.Name)
}

// ============================================================================
// Symbols
// ======================================================================================

// Symbols contains the symbols of the file.
//
// Refer to [PackageSymbols] and [BuildSymbols] for more information.
type Symbols struct {
	// Imports are the imports of the file.
	//
	// It is recommended to use [Symbols.AddImport] and
	// [Symbols.AddBuiltinImport] to add imports to the file, as these
	// methods provide additional safeguards to prevent illegal states.
	Imports []*Import

	ComponentCalls       []*ComponentCall
	componentCallsByNode map[*ast.ComponentCall]*ComponentCall

	ElementReferences       []*ElementReference
	elementReferencesByNode map[*ast.ElementReference]*ElementReference

	Attributes       []*Attribute
	attributesByNode map[ast.Attribute]*Attribute

	AttributeReferences       []*AttributeReference
	attributeReferencesByNode map[*ast.AttributeReference]*AttributeReference

	//
	// LINKER

	// Linked indicates that the entire file has been linked, i.e. all symbols
	// have Linked set to true.
	Linked bool

	//
	// ANALYZER

	// Analyzed indicates that the entire file has been analyzed, i.e. all
	// symbols have Analyzed set to true.
	Analyzed bool
}

func buildSymbols(f *File) {
	f.Symbols = &Symbols{
		Imports:             make([]*Import, 0, 64),
		ComponentCalls:      make([]*ComponentCall, 0, 256),
		ElementReferences:   make([]*ElementReference, 0, 256),
		Attributes:          make([]*Attribute, 0, 512),
		AttributeReferences: make([]*AttributeReference, 0, 512),
	}
	defer func() {
		f.Imports = slices.Clip(f.Imports)
		f.ComponentCalls = slices.Clip(f.ComponentCalls)
		f.ElementReferences = slices.Clip(f.ElementReferences)
		f.Attributes = slices.Clip(f.Attributes)
		f.AttributeReferences = slices.Clip(f.AttributeReferences)
	}()

	for _, impStmt := range f.AST.Imports {
		for _, spec := range impStmt.Specs {
			imp := &Import{AST: spec}
			if spec.Alias != nil {
				imp.Alias = spec.Alias.Name
			}
			if spec.Path != nil {
				imp.CorgiPath = spec.Path.Unquote()
			}
			f.Imports = append(f.Imports, imp)
		}
	}

	var (
		cc          *ComponentCall
		attr        *Attribute
		comp        *Component
		parentBlock *BlockInstance
	)
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		switch n := n.(type) {
		case *ast.ComponentCall:
			ccw := &ComponentCall{AST: n, File: f, BlockSetters: make([]*BlockSetter, 0, 24)}
			f.ComponentCalls = append(f.ComponentCalls, ccw)
			f.componentCallsByNode[n] = ccw

			oldCC := cc
			cc = ccw
			n.Walk(walk)
			cc = oldCC

			ccw.BlockSetters = slices.Clip(ccw.BlockSetters)
			for _, with := range ccw.BlockSetters {
				with.Instances = slices.Clip(with.Instances)
			}
		case ast.BlockSetter:
			if cc == nil {
				n.Walk(walk)
				break
			}
			instance := &BlockSetterInstance{AST: n}
			group := cc.BlockSetterByName(n.Name())
			if group == nil {
				group = &BlockSetter{Name: n.Name(), Instances: make([]*BlockSetterInstance, 0, 16)}
				cc.BlockSetters = append(cc.BlockSetters, group)
			}
			instance.Group = group
			group.Instances = append(group.Instances, instance)

			n.Walk(walk)
		case *ast.Block:
			instance := &BlockInstance{AST: n, Parent: parentBlock}
			if n.Default != nil {
				instance.Default = &BlockInstanceDefault{AST: n.Default}
			}
			group := comp.BlockByName(n.Name())
			if group == nil {
				group = &Block{Name: n.Name(), Instances: make([]*BlockInstance, 0, 16)}
				comp.Blocks = append(comp.Blocks, group)
			}
			instance.Group = group
			group.Instances = append(group.Instances, instance)

			oldParent := parentBlock
			parentBlock = instance
			n.Walk(walk)
			parentBlock = oldParent
		case *ast.ElementReference:
			f.ElementReferences = append(f.ElementReferences, &ElementReference{AST: n})
			n.Walk(walk)
		case ast.Attribute:
			attr = &Attribute{AST: n}
			f.Attributes = append(f.Attributes, attr)
			n.Walk(walk)
		case *ast.AttributeReference:
			ref := &AttributeReference{AST: n}
			if attr != nil {
				attr.Reference = ref
				attr = nil
			}
			f.AttributeReferences = append(f.AttributeReferences, ref)
			n.Walk(walk)
		}
	}
	for _, n := range f.AST.TopLevel {
		astC, _ := n.(*ast.Component)
		if astC == nil {
			continue
		}
		c := f.Package.ComponentByNode(astC)
		c.Blocks = make([]*Block, 0, 24)

		ccsStart := len(f.ComponentCalls)
		n.Walk(walk)
		ccEnd := len(f.ComponentCalls)
		if ccEnd > ccsStart {
			c.ComponentCalls = f.ComponentCalls[ccsStart:ccEnd:ccEnd]
		}

		c.Blocks = slices.Clip(c.Blocks)
		for _, block := range c.Blocks {
			block.Instances = slices.Clip(block.Instances)
		}
	}

	f.RebuildLookupTables()
}

// AddBuiltinImport creates a new [Import] importing the given builtin package.
// The import is marked as not forwarded by default. It is automatically marked
// as loaded.
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
		CorgiPath: builtin.CorgiImportPath,
		GoPath:    builtin.GoImportPath(),
		Package:   builtin,
		Namespace: cmp.Or(alias, builtin.Name),
		Builtin:   true,
		Loaded:    true,
	}
	s.AddImport(imp)
}

// AddImport adds the given import to the file.
// Always use this method if adding implicit imports.
//
// AddImport panics if any of the following conditions are violated:
//   - If the import is implicit (except builtin), it must have a Go import path.
//   - If the import is explicit, it must have a corgi import path.
//   - If the import is a builtin import, the file must not already have a
//     builtin import, i.e. BuiltinImport() == nil.
//   - The import's namespace must match the alias, if set.
//   - If implicit, the import must not be a dot import.
//   - The file must not already have an import with the namespace.
//     You can ensure a unique namespace using [Import.EnsureUniqueNamespace].
//   - The import must be marked as forwarded, unless it is explicit.
func (s *Symbols) AddImport(imp *Import) {
	switch {
	case !imp.Builtin && imp.Implicit() && imp.GoPath == "":
		panic("cannot add implicit import with no Go import path")
	case imp.Explicit() && imp.CorgiPath == "":
		panic("cannot add explicit import with no corgi import path")
	case imp.Builtin && s.BuiltinImport() != nil:
		panic(fmt.Sprintf("symbols already contain builtin import for %q", s.BuiltinImport().CorgiPath))
	case imp.Implicit() && imp.Alias == ".":
		panic("cannot add implicit dot import")
	case !imp.Builtin && imp.Alias != "" && imp.Alias != imp.Namespace:
		panic(fmt.Sprintf("import alias %s does not match namespace %s", imp.Alias, imp.Namespace))
	case s.ImportByNamespace(imp.Namespace) != nil:
		panic(fmt.Sprintf("symbols already contain import with namespace %s: you need to chose a (different) alias", imp.Namespace))
	case !imp.Implicit() && !imp.Forward:
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
		if imp.CorgiPath == p {
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

func (s *Symbols) AttributeByNode(node ast.Attribute) *Attribute {
	return s.attributesByNode[node]
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

	s.attributesByNode = make(map[ast.Attribute]*Attribute, len(s.Attributes))
	for _, attr := range s.Attributes {
		s.attributesByNode[attr.AST] = attr
	}

	s.attributeReferencesByNode = make(map[*ast.AttributeReference]*AttributeReference, len(s.AttributeReferences))
	for _, ref := range s.AttributeReferences {
		s.attributeReferencesByNode[ref.AST] = ref
	}
}
