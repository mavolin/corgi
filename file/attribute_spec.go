package file

import (
	"fmt"
	"sync"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

type AttributeSpec struct {
	// BUILD SYMBOLS
	//

	AST        *ast.AttributeSpec
	Definition *ast.AttributeDefinition
	File       *File

	// WildcardRule is the rule with the wildcard selector, if such a rule
	// exists.
	// It's presence indicates that the attribute is defined for all elements.
	WildcardRule *ast.AttributeRule

	explicitRules     map[*ElementSpec]*ast.AttributeRule
	explicitRulesOnce sync.Once

	// Prefix is the prefix of the attribute definition, if it has one.
	Prefix CanonicalAttributeName

	// Specificity is the specificity of the attribute selector.
	//
	// For basic attribute selectors the specificity is calculated as the length of
	// the prefix and name of the attribute, excluding the wildcard asterisk.
	// For example `foo` and `foo*` both have a specificity of 3.
	//
	// For regular expression selectors, the specificity is always 0.
	//
	// In a valid package, there are never two attribute specs with the same
	// specificity that match the same name.
	Specificity int
}

func (spec *AttributeSpec) MatchesHTMLName(name CanonicalAttributeName) bool {
	if spec.AST.Selector == nil {
		return false
	} else if len(name) < len(spec.Prefix) {
		return false
	}

	prefixPart := name[:len(spec.Prefix)]
	selectorPart := name[len(spec.Prefix):]
	return prefixPart == spec.Prefix && spec.AST.Selector.Matches(string(selectorPart))
}

func (spec *AttributeSpec) MatchesQualifiableName(name CanonicalQualifiableAttributeName) bool {
	return spec.AST.Selector.Matches(string(name))
}

// RuleFor returns the rule on the attribute definition for the given element.
//
// If there is no explicit match for the element, the wildcard rule is returned,
// if it exists.
//
// The file must have been linked.
func (spec *AttributeSpec) RuleFor(elemSpec *ElementSpec) *ast.AttributeRule {
	if !spec.File.Linked {
		panic("AttributeSpec.RuleFor called on unlinked file")
	}

	if spec.AST.Ruleset == nil {
		return nil
	} else if spec.WildcardRule != nil && len(spec.AST.Ruleset.List) == 1 { // fast path
		return spec.WildcardRule
	}

	spec.explicitRulesOnce.Do(func() {
		spec.explicitRules = make(map[*ElementSpec]*ast.AttributeRule)
		for _, rule := range spec.AST.Ruleset.List {
			if rule == nil {
				continue
			}

			switch sel := rule.Selector.(type) {
			case *ast.WildcardElementSelector:
				// used as fallback: spec.WildcardRule
			case *ast.ListElementSelector:
				for _, elemRefAST := range sel.List {
					elemRef := spec.File.ElementReferenceByNode(elemRefAST)
					if elemRef != nil && elemRef.Spec != nil {
						spec.explicitRules[elemRef.Spec] = rule
					}
				}
			default:
				panic(fmt.Sprintf("AttributeSpec.RuleFor: unknown selector type: %T", sel))
			}
		}
	})

	for elemSpec != nil {
		if rule := spec.explicitRules[elemSpec]; rule != nil {
			return rule
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

	return spec.WildcardRule
}

// TypeFor returns the type of the attribute for the given element.
func (spec *AttributeSpec) TypeFor(elemSpec *ElementSpec) attrtype.Type {
	r := spec.RuleFor(elemSpec)
	if r == nil || r.Type == nil {
		return nil
	}
	return r.Type.Type
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
