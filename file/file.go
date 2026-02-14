// Package file represents a high-level view around the File of a corgi file.
// This, most prominently, includes linked imports, component calls and additional
// metadata, as well as the result of [github.com/mavolin/corgi/load/analyze.Analyze].
package file

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
)

// File represents a parsed corgi file.
type File struct {
	Package *Package

	//
	// METADATA

	// Name is the name of the file.
	Name Name

	// Raw contains the raw input file, as it was parsed.
	Raw string
	// Lines are the lines of Raw, stripped of their CRLF/LF line endings.
	Lines []string

	AST *ast.File

	*Symbols
}

func (f *File) PathInModule() Path {
	if f.Package == nil {
		return Path("<unknown package>/" + f.Name)
	}
	return f.Package.PathInModule.FilePath(f.Name)
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
				imp.Alias = Qualifier(spec.Alias.Name)
			}
			if spec.Path != nil {
				if unq, ok := spec.Path.ConstantValue(); ok {
					imp.CorgiPath = CorgiImportPath(unq)
				}
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
			ccw := &ComponentCall{
				AST:          n,
				File:         f,
				BlockSetters: make([]*BlockSetter, 0, 24),
			}
			f.ComponentCalls = append(f.ComponentCalls, ccw)

			if n.Header != nil {
				switch name := n.Header.Name.(type) {
				case *ast.Identifier:
					ccw.Name = Identifier(name.Name)
				case *ast.QualifiedIdentifier:
					if name.Package != nil {
						ccw.Qualifier = Qualifier(name.Package.Name)
					}
					if name.Name != nil {
						ccw.Name = Identifier(name.Name.Name)
					}
				default:
					panic(fmt.Sprintf("unhandled component call name type %T", name))
				}

				if n.Header.Arguments != nil && len(n.Header.Arguments.List) > 0 {
					ccw.ComponentArguments = make([]*ComponentArgument, 0, len(n.Header.Arguments.List))
					for _, argAST := range n.Header.Arguments.List {
						argAST, _ := argAST.(*ast.ComponentArgument)
						if argAST != nil {
							arg := ComponentArgument{AST: argAST, ComponentCall: ccw}
							if arg.AST.Name != nil {
								arg.Name = Identifier(arg.AST.Name.Name)
							}
							ccw.ComponentArguments = append(ccw.ComponentArguments, &arg)
						}
					}
					ccw.ComponentArguments = slices.Clip(ccw.ComponentArguments)
				}
			}

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
			group := cc.BlockSetterByName(Identifier(n.Name()))
			if group == nil {
				group = &BlockSetter{
					ComponentCall: cc,
					Name:          Identifier(n.Name()),
					Instances:     make([]*BlockSetterInstance, 0, 16),
				}
				cc.BlockSetters = append(cc.BlockSetters, group)
			}
			instance.Group = group
			group.Instances = append(group.Instances, instance)

			n.Walk(walk)
		case *ast.Block:
			instance := &BlockInstance{AST: n, ContainingInstance: parentBlock}
			if n.Default != nil {
				instance.Default = &BlockInstanceDefault{AST: n.Default}
			}
			group := comp.BlockByName(Identifier(n.Name()))
			if group == nil {
				group = &Block{
					Component: comp,
					Name:      Identifier(n.Name()),
					Instances: make([]*BlockInstance, 0, 16),
				}
				comp.Blocks = append(comp.Blocks, group)
			}
			instance.Group = group
			group.Instances = append(group.Instances, instance)

			oldParent := parentBlock
			parentBlock = instance
			n.Walk(walk)
			parentBlock = oldParent
		case *ast.ElementReference:
			var qual Qualifier
			var qualifiableName CanonicalQualifiableElementName
			var unqualifiedName CanonicalElementName
			if n.Package != nil {
				qual = Qualifier(n.Package.Name)
			}
			if n.Name != nil {
				if n.Package != nil || n.Dot != nil {
					qualifiableName = CanonicalQualifiableElementName(n.Name.CanonicalName)
				} else {
					unqualifiedName = CanonicalElementName(n.Name.CanonicalName)
				}
			}

			f.ElementReferences = append(f.ElementReferences, &ElementReference{
				AST:             n,
				Qualifier:       qual,
				QualifiableName: qualifiableName,
				UnqualifiedName: unqualifiedName,
			})
			n.Walk(walk)
		case ast.Attribute:
			attr = &Attribute{AST: n}
			f.Attributes = append(f.Attributes, attr)
			n.Walk(walk)
		case *ast.AttributeReference:
			var qual Qualifier
			var qualifiableName CanonicalQualifiableAttributeName
			var unqualifiedName CanonicalAttributeName
			if n.Package != nil {
				qual = Qualifier(n.Package.Name)
			}
			if n.Name != nil {
				if n.Package != nil || n.Dot != nil {
					qualifiableName = CanonicalQualifiableAttributeName(n.Name.CanonicalName)
				} else {
					unqualifiedName = CanonicalAttributeName(n.Name.CanonicalName)
				}
			}

			ref := &AttributeReference{
				AST:             n,
				Qualifier:       qual,
				QualifiableName: qualifiableName,
				UnqualifiedName: unqualifiedName,
			}
			if attr != nil {
				attr.Reference = ref
				attr = nil
			}
			f.AttributeReferences = append(f.AttributeReferences, ref)
			n.Walk(walk)
		default:
			n.Walk(walk)
		}
	}
	for _, n := range f.AST.TopLevel {
		astC, _ := n.(*ast.Component)
		if astC == nil {
			continue
		}
		comp = f.Package.ComponentByNode(astC)
		comp.Blocks = make([]*Block, 0, 24)

		n.Walk(walk)

		comp.Blocks = slices.Clip(comp.Blocks)
		for _, block := range comp.Blocks {
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
// You can generate a unique alias using [Symbols.UniqueQualifier].
//
// The package must not contain any exported symbols.
//
// You must add a builtin import using this method, not by adding it to the
// [Symbols.Imports] slice directly.
func (s *Symbols) AddBuiltinImport(alias Qualifier, builtin *Package) {
	s.AddImport(&Import{
		Alias:     alias,
		CorgiPath: builtin.CorgiImportPath,
		GoPath:    builtin.GoImportPath(),
		Package:   builtin,
		Qualifier: cmp.Or(alias, builtin.Name),
		Builtin:   true,
		Loaded:    true,
	})
}

// AddImport adds the given import to the file.
// Always use this method if adding implicit imports.
//
// AddImport panics if any of the following conditions are violated:
//   - The Go or corgi import paths must be syntactically valid.
//   - If the import is implicit, it must have a Go import path.
//   - If the import is explicit, it must have a corgi import path and no Go
//     import path.
//   - If the import is a builtin import, the file must not already have a
//     builtin import, i.e. BuiltinImport() == nil.
//   - If the import is a builtin import, it must have a package.
//   - The import's qualifier must match the alias, if set.
//   - If implicit, the import must not be a dot import.
//   - The file must not already have an import with the qualifier.
//     You can ensure a unique qualifier using [Import.EnsureUniqueQualifier].
//   - The import must be marked as forwarded, unless it is explicit or the
//     builtin import.
func (s *Symbols) AddImport(imp *Import) {
	switch {
	case imp.Implicit() && imp.GoPath == "":
		panic("implicit import with no Go import path")
	case imp.Explicit() && imp.CorgiPath == "":
		panic("explicit import with no corgi import path")
	case imp.Explicit() && imp.GoPath != "":
		panic("explicit import with Go import path")
	case imp.Builtin && s.BuiltinImport() != nil:
		panic(fmt.Sprintf("symbols already contain builtin import for %q", s.BuiltinImport().GoPath))
	case imp.Builtin && imp.Package == nil:
		panic("builtin import with no package")
	case imp.Implicit() && imp.Alias == ".":
		panic("implicit dot import")
	case !imp.Builtin && imp.Alias != "" && imp.Alias != imp.Qualifier:
		panic(fmt.Sprintf("import alias %s does not match qualifier %s", imp.Alias, imp.Qualifier))
	case s.ImportByQualifier(imp.Qualifier) != nil:
		panic(fmt.Sprintf("symbols already contain import with qualifier %s: you need to chose a (different) alias", imp.Qualifier))
	case imp.Implicit() && !imp.Builtin && !imp.Forward:
		panic("implicit import that is not forwarded")
	}

	if imp.CorgiPath != "" {
		if err := imp.CorgiPath.CheckValid(); err != nil {
			panic(fmt.Sprintf("invalid corgi import path %q: %v", imp.CorgiPath, err))
		}
	}
	if imp.GoPath != "" {
		if err := imp.GoPath.CheckValid(); err != nil {
			panic(fmt.Sprintf("invalid Go import path %q: %v", imp.GoPath, err))
		}
	}

	s.Imports = append(s.Imports, imp)
}

// ImportByQualifier returns the first import with the given qualifier.
//
// Does not work for the "." qualifier.
//
// Only available after linking.
func (s *Symbols) ImportByQualifier(qual Qualifier) *Import {
	for _, imp := range s.Imports {
		if imp.Qualifier == qual {
			return imp
		}
	}
	return nil
}

// ImportByCorgiPath returns the first import with the given corgi import path.
func (s *Symbols) ImportByCorgiPath(p CorgiImportPath) *Import {
	for _, imp := range s.Imports {
		if imp.CorgiPath == p {
			return imp
		}
	}
	return nil
}

// ImportByGoPath returns the first import with the given Go import path.
//
// For explicit imports to be included in the search, the file must have been
// linked.
// Otherwise, only implicit imports and others with an already set/resolved Go
// import path are included.
func (s *Symbols) ImportByGoPath(p GoImportPath) *Import {
	for _, imp := range s.Imports {
		if imp.GoPath == p {
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

// UniqueQualifier returns a unique qualifier in the file, using the given
// qualifier as base, by appending underscores ("_") until the qualifier is
// unique.
func (s *Symbols) UniqueQualifier(base Qualifier) Qualifier {
	for s.ImportByQualifier(base) != nil {
		base += "_"
	}
	return base
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
// Every time you modify the slices of this struct directly, you must call this
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
