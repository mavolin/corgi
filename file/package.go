package file

import (
	"path"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
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
	// ImportPath is the import path with which the package was imported.
	//
	// This needn't necessarily be a correct import, corresponding to
	// path.Join(Module, PathInModule), if the path is symbolic.
	// The most common case for that is a corgi stdlib import, that uses the
	// "corgi/" import path prefix, but is obviously located in this module.
	//
	// In that case ImportPath might be "corgi/fmt", while Module is
	// "github.com/mavolin/corgi/v2" and PathInModule is "std/fmt".
	ImportPath string // load

	Name string // analyze

	// Analyzed indicates whether the package has been analyzed, albeit with
	// errors.
	Analyzed bool // analyze

	*PackageSymbols

	Files []*File
}

func (p *Package) ModulePath() string {
	if p.Module != "" {
		return path.Join(p.Module, p.PathInModule)
	}
	return p.PathInModule
}

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
	State            []*State
	stateByName      map[string]*State
	stateByNode      map[*ast.StateSpec][]*State
	ElementSpecs     []*ElementSpec
	AttributeSpecs   []*AttributeSpec // ordered by specificity, descending
}

func (s *PackageSymbols) ComponentByNode(c *ast.Component) *Component {
	return s.componentByNode[c]
}

func (s *PackageSymbols) ComponentByName(name string) *Component {
	return s.componentsByName[name]
}

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

func (s *PackageSymbols) ElementSpecByFullName(name string) *ElementSpec {
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

// AttributeSpecByFullName returns the attribute spec that matches
// the given full name.
//
// It might return multiple definitions if there are multiple selectors with
// the same specificity that both match the name.
func (s *PackageSymbols) AttributeSpecByFullName(name string) []*AttributeSpec {
	var matches []*AttributeSpec
	for _, def := range s.AttributeSpecs {
		if def.MatchesFullName(name) {
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
	for _, c := range s.Components {
		s.componentByNode[c.AST] = c

		if c.AST.Header != nil && c.AST.Header.Name != nil {
			s.componentsByName[c.AST.Header.Name.Name] = c
		}
	}

	for _, state := range s.State {
		states := s.stateByNode[state.AST]
		if len(states) < len(state.AST.Names) {
			states = slices.Grow(states, len(state.AST.Names)-len(states))
		}
		states[state.Index] = state
		s.stateByNode[state.AST] = states

		if name := state.Name(); name != nil {
			s.stateByName[name.Name] = state
		}
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

type State struct {
	//
	// BUILD SYMBOLS

	AST  *ast.StateSpec
	File *File
	// Index of the variable in the Names and Values slices.
	Index int

	//
	// ANALYZER

	AnalyzedWithErrors bool

	// The InferredType of this value, if there is no explicit type.
	InferredType string
}

func (s *State) Name() *ast.Identifier {
	return s.AST.Names[s.Index]
}

func (s *State) Value() *ast.Expression {
	if len(s.AST.Values) == 1 {
		return s.AST.Values[0]
	}
	return s.AST.Values[s.Index]
}

func (s *State) ResolvedType() string {
	if s.AST.Type != nil {
		return s.AST.Type.Type
	}
	return s.InferredType
}

type ElementSpec struct {
	//
	// BUILD SYMBOLS

	AST         *ast.ElementSpec
	Definition  *ast.ElementDefinition
	File        *File
	lowerPrefix string
	lowerName   string

	//
	// ANALYZE

	AnalyzedWithErrors bool

	Type elemtype.Type
}

// QualifiedName is the name of the element, without the prefix.
func (d *ElementSpec) QualifiedName() string {
	if d.AST.Name != nil {
		return d.AST.Name.Name
	}
	return ""
}

func (d *ElementSpec) MatchesQualifiedName(name string) bool {
	if d.AST.Name == nil {
		return false
	}
	return strings.EqualFold(d.AST.Name.Name, name)
}

// FullName is the name of the element, including the prefix.
func (d *ElementSpec) FullName() string {
	if d.AST.Name == nil {
		return ""
	}
	if d.Definition != nil && d.Definition.Prefix != nil {
		return d.Definition.Prefix.Name + d.AST.Name.Name
	}
	return d.AST.Name.Name
}

func (d *ElementSpec) MatchesFullName(name string) bool {
	if d.AST.Name == nil {
		return false
	}

	name = strings.ToLower(name)
	if d.Definition != nil {
		if d.Definition.Prefix != nil {
			prefix := strings.ToLower(d.Definition.Prefix.Name)
			if !strings.HasPrefix(name, prefix) {
				return false
			}
			name = name[len(prefix):]
		}
	}
	return strings.ToLower(d.AST.Name.Name) == name
}

type AttributeSpec struct {
	// BUILD SYMBOLS
	//

	AST        *ast.AttributeSpec
	Definition *ast.AttributeDefinition
	File       *File

	// Specificity is the specificity of the attribute definition.
	//
	// For basic attribute selectors the specificity is calculated as the length of
	// the name of the attribute, excluding the wildcard asterisk.
	// For example `foo` and `foo*` both have a specificity of 3.
	//
	// For regular expression selectors, the specificity is always 0.
	//
	// In a valid package, there are never two attribute definitions with the same
	// specificity that match the same name.
	Specificity int
}

func (d *AttributeSpec) MatchesFullName(name string) bool {
	if d.AST.Selector == nil {
		return false
	}

	name = strings.ToLower(name)
	if d.Definition != nil {
		if d.Definition.Prefix != nil {
			prefix := strings.ToLower(d.Definition.Prefix.Name)
			if !strings.HasPrefix(name, prefix) {
				return false
			}
			name = name[len(prefix):]
		}
	}
	return d.AST.Selector.Matches(name)
}

func (d *AttributeSpec) MatchesQualifiedName(name string) bool {
	return d.AST.Selector.Matches(name)
}

// TypeFor returns the type of the attribute for the given element.
func (d *AttributeSpec) TypeFor(elemDef *ElementSpec) attrtype.Type {
	if d.AST.Ruleset == nil {
		return attrtype.Unknown
	}

	for elemDef != nil {
		elemName := elemDef.FullName()

		var wildcard attrtype.Type
		for _, rule := range d.AST.Ruleset.List {
			if rule == nil {
				continue
			}
			switch sel := rule.Selector.(type) {
			case *ast.WildcardElementSelector:
				wildcard = rule.Type.Type
			case *ast.ListElementSelector:
				if sel.Matches(elemName) {
					return rule.Type.Type
				}
			}
		}
		if wildcard != attrtype.Unknown {
			return wildcard
		}

		if elemDef.AST == nil || elemDef.AST.Type == nil {
			break
		}
		if alias, _ := elemDef.AST.Type.(*ast.AliasElementType); alias != nil {
			elemRef := d.File.ElementReferenceByNode(alias.Name)
			if elemRef != nil {
				elemDef = elemRef.Spec
			}
		}
	}

	return attrtype.Unknown
}

// GenericType returns the one type an attribute would have, regardless of the
// element it is used on.
//
// In other words, it only returns a type if the attribute definition contains
// a single wildcard selector rule.
func (d *AttributeSpec) GenericType() attrtype.Type {
	if d.AST.Ruleset == nil {
		return attrtype.Unknown
	}

	if len(d.AST.Ruleset.List) != 1 {
		return attrtype.Unknown
	}

	rule := d.AST.Ruleset.List[0]
	if rule.Selector == nil || rule.Type == nil {
		return attrtype.Unknown
	}

	wildcard, _ := rule.Selector.(*ast.WildcardElementSelector)
	if wildcard == nil {
		return attrtype.Unknown
	}
	return rule.Type.Type
}

func (d *AttributeSpec) specificity() int {
	switch sel := d.AST.Selector.(type) {
	case *ast.BasicAttributeSelector:
		if d.Definition != nil && d.Definition.Prefix != nil {
			return len(d.Definition.Prefix.Name) + len(sel.Name)
		}
		return len(sel.Name)
	case *ast.RegexpAttributeSelector:
		return 0
	default:
		return 0
	}
}
