package file

import (
	"strings"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

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

	// Analyzed indicates whether the ElementSpec has been analyzed,
	// albeit with errors.
	Analyzed bool

	// Circular indicates that the element is defined by referencing itself.
	Circular bool

	Type Analysis[elemtype.Type]
}

// QualifiedName is the name of the element, without the prefix.
func (spec *ElementSpec) QualifiedName() string {
	if spec.AST.Name != nil {
		return spec.AST.Name.Name
	}
	return ""
}

func (spec *ElementSpec) MatchesQualifiedName(name string) bool {
	if spec.AST.Name == nil {
		return false
	}
	return strings.EqualFold(spec.AST.Name.Name, name)
}

// HTMLName is the name of the element, including the prefix.
func (spec *ElementSpec) HTMLName() string {
	if spec.AST.Name == nil {
		return ""
	}
	if spec.Definition != nil && spec.Definition.Prefix != nil {
		return spec.Definition.Prefix.Name + spec.AST.Name.Name
	}
	return spec.AST.Name.Name
}

func (spec *ElementSpec) MatchesHTMLName(name string) bool {
	if spec.AST.Name == nil {
		return false
	}

	name = strings.ToLower(name)
	if spec.Definition != nil {
		if spec.Definition.Prefix != nil {
			prefix := strings.ToLower(spec.Definition.Prefix.Name)
			if !strings.HasPrefix(name, prefix) {
				return false
			}
			name = name[len(prefix):]
		}
	}
	return strings.ToLower(spec.AST.Name.Name) == name
}
