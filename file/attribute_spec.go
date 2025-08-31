package file

import (
	"fmt"
	"strings"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

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

func (spec *AttributeSpec) MatchesHTMLName(name string) bool {
	if spec.AST.Selector == nil {
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
	return spec.AST.Selector.Matches(name)
}

func (spec *AttributeSpec) MatchesQualifiedName(name string) bool {
	return spec.AST.Selector.Matches(name)
}

// RuleFor returns the rule on the attribute definition for the given element.
func (spec *AttributeSpec) RuleFor(elemSpec *ElementSpec) *ast.AttributeRule {
	if spec.AST.Ruleset == nil {
		return nil
	}

	for elemSpec != nil {
		if r := spec.ruleFor(elemSpec); r != nil {
			return r
		}

		// If the passed element is an alias of another element, check if
		// we match for that element.
		if elemSpec.AST == nil || elemSpec.AST.Type == nil {
			break
		}
		alias, _ := elemSpec.AST.Type.(*ast.AliasElementType)
		if alias == nil {
			break
		}
		elemRef := elemSpec.File.ElementReferenceByNode(alias.Name)
		if elemRef == nil {
			break
		}
		elemSpec = elemRef.Spec
	}

	return nil
}

func (spec *AttributeSpec) ruleFor(elemSpec *ElementSpec) *ast.AttributeRule {
	var fallback *ast.AttributeRule

	for _, rule := range spec.AST.Ruleset.List {
		if rule == nil {
			continue
		}

		switch sel := rule.Selector.(type) {
		case *ast.WildcardElementSelector:
			fallback = rule
		case *ast.ListElementSelector:
			for _, elemRefAST := range sel.List {
				elemRef := spec.File.ElementReferenceByNode(elemRefAST)
				if elemRef.Spec == elemSpec {
					return rule
				}
			}
		default:
			panic(fmt.Sprintf("AttributeSpec.RuleFor: unknown selector type: %T", sel))
		}
	}

	return fallback
}

// TypeFor returns the type of the attribute for the given element.
func (spec *AttributeSpec) TypeFor(elemSpec *ElementSpec) attrtype.Type {
	r := spec.RuleFor(elemSpec)
	if r == nil || r.Type == nil {
		return attrtype.Unknown
	}
	return r.Type.Type
}

// GenericType returns the one type an attribute would have, regardless of the
// element it is used on.
//
// In other words, it only returns a type if the attribute definition contains
// a single wildcard selector rule.
func (spec *AttributeSpec) GenericType() attrtype.Type {
	if spec.AST.Ruleset == nil {
		return attrtype.Unknown
	}

	if len(spec.AST.Ruleset.List) != 1 {
		return attrtype.Unknown
	}

	rule := spec.AST.Ruleset.List[0]
	if rule.Selector == nil || rule.Type == nil {
		return attrtype.Unknown
	}

	wildcard, _ := rule.Selector.(*ast.WildcardElementSelector)
	if wildcard == nil {
		return attrtype.Unknown
	}
	return rule.Type.Type
}

func (spec *AttributeSpec) specificity() int {
	switch sel := spec.AST.Selector.(type) {
	case *ast.BasicAttributeSelector:
		if spec.Definition != nil && spec.Definition.Prefix != nil {
			return len(spec.Definition.Prefix.Name) + len(sel.Name)
		}
		return len(sel.Name)
	case *ast.RegexpAttributeSelector:
		return 0
	default:
		panic(fmt.Sprintf("AttributeSpec.TypeFor: unknown selector type: %T", sel))
	}
}
