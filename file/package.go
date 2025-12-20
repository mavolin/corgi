package file

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
)

type Package struct {
	// Module is the path/name of the Go module providing this directory.
	//
	// Empty for Corgi stdlib.
	Module ModulePath // load
	// PathInModule is the path to the directory in the Go module, relative
	// to the module root.
	//
	// Always specified as a forward slash separated path.
	PathInModule PackagePath // load
	// CorgiImportPath is the import path with which the package should be
	// imported in Corgi code.
	//
	// This might be different from the [Package.GoImportPath], if the path is
	// symbolic.
	// The most common case for that is a corgi stdlib import, that uses the
	// "corgi/" import path prefix, but is obviously imported in the generated
	// Go code using another import path.
	CorgiImportPath CorgiImportPath // load

	Name Qualifier // analyze

	// Analyzed indicates whether the package has been analyzed, albeit with
	// errors.
	Analyzed bool // analyze

	*PackageSymbols

	Files []*File
}

// GoImportPath returns the import path with which the package should be
// imported in Go code.
//
// This might be different from the [Package.CorgiImportPath], if the path is
// symbolic.
// See [Package.CorgiImportPath] for more information.
func (p *Package) GoImportPath() GoImportPath {
	if p.Module != "" {
		return p.Module.ImportPathFor(p.PathInModule)
	}
	return GoImportPath(p.PathInModule)
}

// ============================================================================
// Symbols
// ======================================================================================

// PackageSymbols contains the symbols of the package.
//
// The usual way of populating these is by using the [BuildSymbols] function.
// Refer to its documentation for more information.
//
// Note that when you modify the slices in this struct and the corresponding
// file symbols, you need to at least call [BuildSymbols] with the
// [BuildSymbolsOptions.OnlyLookupTables] option set to true, to update the
// lookup tables used by the methods on the respective types.
type PackageSymbols struct {
	Components       []*Component
	componentsByName map[Identifier]*Component
	componentByNode  map[*ast.Component]*Component
	ElementSpecs     []*ElementSpec
	AttributeSpecs   []*AttributeSpec // ordered by specificity, descending

	//
	// LINKER

	// Linked indicates that the entire file has been linked, i.e. all symbols
	// have Linked/Loaded set to true.
	Linked bool

	//
	// ANALYZER

	// Analyzed indicates that the entire file has been analyzed, i.e. all
	// symbols have Analyzed set to true.
	Analyzed bool
}

// BuildSymbols builds the symbols of the package.
// That means it populates the PackageSymbols field of the package
// and the Symbols field on the files of the package.
//
// This is usually done prior to linking, after all files have been parsed and
// added to the package.
// BuildSymbols fills the all fields on the package's PackageSymbols struct,
// and all fields on the Symbols struct of each file.
func BuildSymbols(p *Package) {
	p.PackageSymbols = &PackageSymbols{
		Components:     make([]*Component, 0, 64),
		ElementSpecs:   make([]*ElementSpec, 0, 64),
		AttributeSpecs: make([]*AttributeSpec, 0, 256),
	}
	defer func() {
		p.Components = slices.Clip(p.Components)
		p.ElementSpecs = slices.Clip(p.ElementSpecs)
		p.AttributeSpecs = slices.Clip(p.AttributeSpecs)
	}()

	for _, f := range p.Files {
		for _, n := range f.AST.TopLevel {
			switch n := n.(type) {
			case *ast.Component:
				c := &Component{
					AST:  n,
					File: f,
				}
				if n.Header != nil {
					if n.Header.Name != nil {
						c.Name = Identifier(n.Header.Name.Name)
					}
					if n.Header.Parameters != nil && len(n.Header.Parameters.List) > 0 {
						c.Parameters = make([]*ComponentParameter, 0, len(n.Header.Parameters.List))
						for _, paramAST := range n.Header.Parameters.List {
							if paramAST != nil {
								param := ComponentParameter{AST: paramAST, Component: c}
								if paramAST.Name != nil {
									param.Name = Identifier(paramAST.Name.Name)
								}
								c.Parameters = append(c.Parameters, &param)
							}
						}
					}
				}
				p.Components = append(p.Components, c)
			case *ast.ElementDefinition:
				var stylizedPrefix, canonicalPrefix string
				if n.Prefix != nil {
					stylizedPrefix = n.Prefix.Name
					canonicalPrefix = n.Prefix.CanonicalName
				}

				for _, spec := range n.Specs {
					if spec != nil {
						var stylizedQualifiableName, canonicalQualifiableName string
						if spec.Name != nil {
							stylizedQualifiableName = spec.Name.Name
							canonicalQualifiableName = spec.Name.CanonicalName
						}

						p.ElementSpecs = append(p.ElementSpecs, &ElementSpec{
							Definition:       n,
							AST:              spec,
							File:             f,
							StylizedHTMLName: stylizedPrefix + stylizedQualifiableName,
							HTMLName:         CanonicalElementName(canonicalPrefix + canonicalQualifiableName),
							QualifiableName:  CanonicalQualifiableElementName(canonicalQualifiableName),
						})
					}
				}
			case *ast.AttributeDefinition:
				var canonicalPrefix CanonicalAttributeName
				if n.Prefix != nil {
					canonicalPrefix = CanonicalAttributeName(n.Prefix.CanonicalName)
				}

				for _, spec := range n.Specs {
					if spec == nil {
						continue
					}

					var wildcardRule *ast.AttributeRule
					if spec.Ruleset != nil {
						for _, rule := range spec.Ruleset.List {
							if _, ok := rule.Selector.(*ast.WildcardElementSelector); ok {
								wildcardRule = rule
								break
							}
						}
					}

					p.AttributeSpecs = append(p.AttributeSpecs, &AttributeSpec{
						Definition:   n,
						AST:          spec,
						File:         f,
						WildcardRule: wildcardRule,
						Prefix:       canonicalPrefix,
					})
				}
			}
		}
	}

	p.RebuildLookupTables()

	for _, f := range p.Files {
		buildSymbols(f)
	}
}

func (s *PackageSymbols) ComponentByNode(c *ast.Component) *Component {
	return s.componentByNode[c]
}

func (s *PackageSymbols) ComponentByName(name Identifier) *Component {
	return s.componentsByName[name]
}

func (s *PackageSymbols) ElementSpecByNode(spec *ast.ElementSpec) *ElementSpec {
	for _, def := range s.ElementSpecs {
		if def.AST == spec {
			return def
		}
	}
	return nil
}

func (s *PackageSymbols) ElementSpecByHTMLName(name CanonicalElementName) *ElementSpec {
	for _, def := range s.ElementSpecs {
		if name == def.HTMLName {
			return def
		}
	}
	return nil
}

func (s *PackageSymbols) ElementSpecByQualifiableName(name CanonicalQualifiableElementName) *ElementSpec {
	for _, def := range s.ElementSpecs {
		if name == def.QualifiableName {
			return def
		}
	}
	return nil
}

func (s *PackageSymbols) AttributeSpecByNode(spec *ast.AttributeSpec) *AttributeSpec {
	for _, def := range s.AttributeSpecs {
		if def.AST == spec {
			return def
		}
	}
	return nil
}

// AttributeSpecByHTMLName returns the attribute spec that matches
// the given full name.
//
// It might return multiple definitions if there are multiple selectors with
// the same specificity that both match the name.
func (s *PackageSymbols) AttributeSpecByHTMLName(name CanonicalAttributeName) []*AttributeSpec {
	var matches []*AttributeSpec
	for _, def := range s.AttributeSpecs {
		if def.MatchesHTMLName(name) {
			if def.Specificity > 0 {
				return []*AttributeSpec{def}
			}
			matches = append(matches, def)
		}
	}
	return matches
}

// AttributeSpecByQualifiableName returns the attribute definition that
// matches the given qualified name.
//
// It might return multiple definitions if there are multiple selectors with
// the same specificity that both match the name.
func (s *PackageSymbols) AttributeSpecByQualifiableName(name CanonicalQualifiableAttributeName) []*AttributeSpec {
	var matches []*AttributeSpec
	for _, def := range s.AttributeSpecs {
		if def.MatchesQualifiableName(name) {
			if def.Specificity > 0 {
				return []*AttributeSpec{def}
			}
			matches = append(matches, def)
		}
	}
	return matches
}

// RebuildLookupTables rebuilds the lookup tables used by the methods of this
// type.
// Note that this explicitly does not rebuild the lookup tables of the
// [File] symbols.
//
// Every time you modify the slices of this struct, you should call this
// method to ensure that the lookup tables are up-to-date.
func (s *PackageSymbols) RebuildLookupTables() {
	s.componentByNode = make(map[*ast.Component]*Component, len(s.Components))
	s.componentsByName = make(map[Identifier]*Component, len(s.Components))

	for _, c := range s.Components {
		s.componentByNode[c.AST] = c

		if c.AST.Header != nil && c.AST.Header.Name != nil {
			s.componentsByName[Identifier(c.AST.Header.Name.Name)] = c
		}
	}

	for _, spec := range s.AttributeSpecs {
		spec.Specificity = spec.specificity()
	}
	slices.SortFunc(s.AttributeSpecs, func(a, b *AttributeSpec) int {
		return a.Specificity - b.Specificity
	})
}
