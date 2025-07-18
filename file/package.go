package file

import (
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

type Package struct {
	// Module is the path/name of the Go module providing this directory.
	//
	// Empty for Go stdlib.
	Module string // load
	// PathInModule is the path to the directory in the Go module, relative
	// to the module root.
	//
	// Always specified as a forward slash separated path.
	PathInModule string // load
	// ImportPath is the actual import path of the package.
	//
	// Only differs from Module/PathInModule if the package is a stdlib package.
	ImportPath string // load

	Name string // analyze

	*PackageSymbols

	Files []*File
}

type PackageSymbols struct {
	Components           []*Component
	State                []*State
	ElementDefinitions   []*ElementSpec
	AttributeDefinitions []*AttributeSpec // ordered by specificity, descending
}

func BuildSymbols(p *Package) {
	var nComponents, nState, nElementDefinitions, nAttributeDefinitions int
	for _, f := range p.Files {
		for _, n := range f.AST.TopLevel {
			switch n := n.(type) {
			case *ast.Alias:
				nComponents++
			case *ast.Component:
				nComponents++
			case *ast.StateDeclaration:
				for _, spec := range n.Specs {
					nState += len(spec.Names)
				}
			case *ast.ElementDefinition:
				nElementDefinitions += len(n.Specs)
			case *ast.AttributeDefinition:
				nAttributeDefinitions += len(n.Specs)
			}
		}
	}
	p.PackageSymbols = &PackageSymbols{
		Components:           make([]*Component, 0, nComponents),
		State:                make([]*State, 0, nState),
		ElementDefinitions:   make([]*ElementSpec, 0, nElementDefinitions),
		AttributeDefinitions: make([]*AttributeSpec, 0, nAttributeDefinitions),
	}

	for _, f := range p.Files {
		for _, n := range f.AST.TopLevel {
			switch n := n.(type) {
			case *ast.Alias:
				c := &Component{AliasAST: n, File: f}
				p.Components = append(p.Components, c)
			case *ast.Component:
				c := &Component{DefinedAST: n, File: f}
				p.Components = append(p.Components, c)
			case *ast.StateDeclaration:
				for _, spec := range n.Specs {
					if spec == nil {
						continue
					}
					for i := range spec.Names {
						s := &State{AST: spec, File: f, Index: i}
						p.State = append(p.State, s)
					}
				}
			case *ast.ElementDefinition:
				for _, spec := range n.Specs {
					if spec == nil {
						continue
					}
					e := &ElementSpec{Definition: n, AST: spec, File: f}
					if n.Prefix != nil {
						e.lowerPrefix = strings.ToLower(n.Prefix.Name)
					}
					if spec.Name != nil {
						e.lowerName = strings.ToLower(spec.Name.Name)
					}
					p.ElementDefinitions = append(p.ElementDefinitions, e)
				}
			case *ast.AttributeDefinition:
				for _, spec := range n.Specs {
					if spec == nil {
						continue
					}
					a := &AttributeSpec{Definition: n, AST: spec, File: f}
					a.Specificity = a.specificity()
					p.AttributeDefinitions = append(p.AttributeDefinitions, a)
				}
			}
		}
	}

	for _, f := range p.Files {
		buildSymbols(f)
	}

	slices.SortFunc(p.AttributeDefinitions, func(a, b *AttributeSpec) int {
		return a.Specificity - b.Specificity
	})
}

func (s *PackageSymbols) ComponentByNode(c *ast.Component) *Component {
	for _, comp := range s.Components {
		if comp.DefinedAST == c {
			return comp
		}
	}
	return nil
}

func (s *PackageSymbols) AliasByNode(a *ast.Alias) *Component {
	for _, comp := range s.Components {
		if comp.AliasAST == a {
			return comp
		}
	}
	return nil
}

func (s *PackageSymbols) ComponentByName(name string) *Component {
	for _, comp := range s.Components {
		h := comp.Header()
		if h != nil && h.Name != nil && h.Name.Ident == name {
			return comp
		}
	}
	return nil
}

func (s *PackageSymbols) StateByNode(spec *ast.StateSpec, index int) *State {
	for i := 0; i < len(s.State); {
		state := s.State[i]
		if state.AST == spec {
			return s.State[i+index]
		}
		i += len(state.AST.Names)
	}
	return nil
}

func (s *PackageSymbols) StateByName(name string) *State {
	for _, state := range s.State {
		if state.AST.Names[state.Index] != nil && state.AST.Names[state.Index].Ident == name {
			return state
		}
	}
	return nil
}

func (s *PackageSymbols) ElementDefinitionByNode(spec *ast.ElementSpec) *ElementSpec {
	for _, def := range s.ElementDefinitions {
		if def.AST == spec {
			return def
		}
	}
	return nil
}

func (s *PackageSymbols) ElementDefinitionByFullName(name string) *ElementSpec {
	name = strings.ToLower(name)
	for _, def := range s.ElementDefinitions {
		if len(name) <= len(def.lowerPrefix) {
			continue
		}
		if name[:len(def.lowerPrefix)] == def.lowerPrefix && name[len(def.lowerPrefix):] == def.lowerName {
			return def
		}
	}
	return nil
}

func (s *PackageSymbols) ElementDefinitionByQualifiedName(name string) *ElementSpec {
	name = strings.ToLower(name)
	for _, def := range s.ElementDefinitions {
		if def.lowerName == name {
			return def
		}
	}
	return nil
}

func (s *PackageSymbols) AttributeDefinitionByNode(spec *ast.AttributeSpec) *AttributeSpec {
	for _, def := range s.AttributeDefinitions {
		if def.AST == spec {
			return def
		}
	}
	return nil
}

// AttributeDefinitionByFullName returns the attribute definition that matches
// the given full name.
//
// It might return multiple definitions if there are multiple selectors with
// the same specificity that both match the name.
// This, however, is only the case for invalid packages.
// If the linker passes with no errors, it is guaranteed that this function
// returns at most one definition.
func (s *PackageSymbols) AttributeDefinitionByFullName(name string) []*AttributeSpec {
	var matches []*AttributeSpec
	for _, def := range s.AttributeDefinitions {
		if def.MatchesFullName(name) {
			if def.Specificity > 0 {
				return []*AttributeSpec{def}
			}
			matches = append(matches, def)
		}
	}
	return matches
}

// AttributeDefinitionByQualifiedName returns the attribute definition that
// matches the given qualified name.
//
// It might return multiple definitions if there are multiple selectors with
// the same specificity that both match the name.
// This, however, is only the case for invalid packages.
// If the linker passes with no errors, it is guaranteed that this function
// returns at most one definition.
func (s *PackageSymbols) AttributeDefinitionByQualifiedName(name string) []*AttributeSpec {
	var matches []*AttributeSpec
	for _, def := range s.AttributeDefinitions {
		if def.MatchesQualifiedName(name) {
			if def.Specificity > 0 {
				return []*AttributeSpec{def}
			}
			matches = append(matches, def)
		}
	}
	return matches
}

type State struct {
	// BUILD SYMBOLS
	//

	AST  *ast.StateSpec
	File *File
	// Index of the variable in the Names and Values slices.
	Index int

	// ANALYZER
	//

	AnalyzedWithErrors bool

	// The InferredType of this value, if there is no explicit type.
	InferredType string
}

func (s *State) Name() *ast.Ident {
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
	// BUILD SYMBOLS
	//

	AST         *ast.ElementSpec
	Definition  *ast.ElementDefinition
	File        *File
	lowerPrefix string
	lowerName   string

	// ANALYZE
	//

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
		for _, rule := range d.AST.Ruleset.Rules {
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

	if len(d.AST.Ruleset.Rules) != 1 {
		return attrtype.Unknown
	}

	rule := d.AST.Ruleset.Rules[0]
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
