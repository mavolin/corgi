package file

import (
	"path"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file/ast"
)

type Package struct {
	// Module is the path/name of the Go module providing this directory.
	//
	// Empty for Corgi stdlib.
	Module string // load
	// PathInModule is the path to the directory in the Go module, relative
	// to the module root.
	//
	// Always specified as a forward slash separated path.
	PathInModule string // load
	// CorgiImportPath is the import path with which the package was
	// imported in the corgi source file.
	//
	// This might be different from the [Package.GoImportPath], if the path is
	// symbolic.
	// The most common case for that is a corgi stdlib import, that uses the
	// "corgi/" import path prefix, but is obviously imported in the generated
	// Go code using another import path.
	CorgiImportPath string // load

	Name string // analyze

	// Analyzed indicates whether the package has been analyzed, albeit with
	// errors.
	Analyzed bool // analyze

	*PackageSymbols

	Files []*File
}

func (p *Package) GoImportPath() string {
	if p.Module != "" {
		return path.Join(p.Module, p.PathInModule)
	}
	return p.PathInModule
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
	componentsByName map[string]*Component
	componentByNode  map[*ast.Component]*Component
	// State are individual state variables in the package.
	// States belonging to the same spec must be grouped together.
	State          []*State
	stateByName    map[string]*State
	stateByNode    map[*ast.StateSpec][]*State
	ElementSpecs   []*ElementSpec
	AttributeSpecs []*AttributeSpec // ordered by specificity, descending

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
		State:          make([]*State, 0, 64),
		ElementSpecs:   make([]*ElementSpec, 0, 64),
		AttributeSpecs: make([]*AttributeSpec, 0, 256),
	}
	defer func() {
		p.Components = slices.Clip(p.Components)
		p.State = slices.Clip(p.State)
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
				if n.Header != nil && n.Header.Parameters != nil && len(n.Header.Parameters.List) > 0 {
					c.Parameters = make([]*ComponentParameter, 0, len(n.Header.Parameters.List))
					for _, param := range n.Header.Parameters.List {
						if param != nil {
							c.Parameters = append(c.Parameters, &ComponentParameter{AST: param})
						}
					}
				}
				p.Components = append(p.Components, c)
			case *ast.StateDeclaration:
				for _, spec := range n.Specs {
					if spec == nil {
						continue
					}
					for i, name := range spec.Names {
						if name != nil {
							p.State = append(p.State, &State{AST: spec, File: f, Index: i})
						}
					}
				}
			case *ast.ElementDefinition:
				for _, spec := range n.Specs {
					if spec != nil {
						p.ElementSpecs = append(p.ElementSpecs, &ElementSpec{Definition: n, AST: spec, File: f})
					}
				}
			case *ast.AttributeDefinition:
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

func (s *PackageSymbols) ComponentByName(name string) *Component {
	return s.componentsByName[name]
}

// StateByNode returns the state symbol for the given state spec node
// and index in the names list.
// An index might not exist if there were parse errors and the name is nil.
// Excess values generally are not part of the package's symbols.
func (s *PackageSymbols) StateByNode(spec *ast.StateSpec, index int) *State {
	if states := s.stateByNode[spec]; states != nil && index < len(states) {
		return states[index]
	}
	return nil
}

func (s *PackageSymbols) StateByName(name string) *State {
	return s.stateByName[name]
}

func (s *PackageSymbols) ElementSpecByNode(spec *ast.ElementSpec) *ElementSpec {
	for _, def := range s.ElementSpecs {
		if def.AST == spec {
			return def
		}
	}
	return nil
}

func (s *PackageSymbols) ElementSpecByHTMLName(name string) *ElementSpec {
	name = strings.ToLower(name)
	for _, def := range s.ElementSpecs {
		if len(name) <= len(def.lowerPrefix) {
			continue
		}
		if name[:len(def.lowerPrefix)] == def.lowerPrefix && name[len(def.lowerPrefix):] == def.lowerName {
			return def
		}
	}
	return nil
}

func (s *PackageSymbols) ElementSpecByQualifiedName(name string) *ElementSpec {
	name = strings.ToLower(name)
	for _, def := range s.ElementSpecs {
		if def.lowerName == name {
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
func (s *PackageSymbols) AttributeSpecByHTMLName(name string) []*AttributeSpec {
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

// AttributeSpecByQualifiedName returns the attribute definition that
// matches the given qualified name.
//
// It might return multiple definitions if there are multiple selectors with
// the same specificity that both match the name.
func (s *PackageSymbols) AttributeSpecByQualifiedName(name string) []*AttributeSpec {
	var matches []*AttributeSpec
	for _, def := range s.AttributeSpecs {
		if def.MatchesQualifiedName(name) {
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
	s.componentsByName = make(map[string]*Component, len(s.Components))

	for _, c := range s.Components {
		s.componentByNode[c.AST] = c

		if c.AST.Header != nil && c.AST.Header.Name != nil {
			s.componentsByName[c.AST.Header.Name.Name] = c
		}
	}

	s.stateByNode = make(map[*ast.StateSpec][]*State, len(s.State))
	s.stateByName = make(map[string]*State, len(s.State))

	var specStart int
	var lastSpec *ast.StateSpec
	for i, state := range s.State {
		if name := state.Name(); name != nil {
			s.stateByName[name.Name] = state
		}

		// We require that states belonging to the same spec are grouped, so
		// we can save memory by creating only views into the State slice,
		// instead of allocating a new slice for every spec.
		if state.AST != lastSpec {
			if lastSpec != nil {
				s.stateByNode[lastSpec] = s.State[specStart:i]
			}
			lastSpec = state.AST
			specStart = i
		}
	}
	if lastSpec != nil { // last group
		s.stateByNode[lastSpec] = s.State[specStart:]
	}

	for _, spec := range s.ElementSpecs {
		if spec.Definition.Prefix != nil {
			spec.lowerPrefix = strings.ToLower(spec.Definition.Prefix.Name)
		}
		if spec.AST.Name != nil {
			spec.lowerName = strings.ToLower(spec.AST.Name.Name)
		}
	}

	for _, spec := range s.AttributeSpecs {
		spec.Specificity = spec.specificity()
	}
	slices.SortFunc(s.AttributeSpecs, func(a, b *AttributeSpec) int {
		return a.Specificity - b.Specificity
	})
}
