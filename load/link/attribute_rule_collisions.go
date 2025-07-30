package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

type duplicateAttributeRule struct {
	rule      *ast.AttributeRule
	highlight ast.Node
}

func (l *linker) CheckAttributeRuleCollisions() {
	logger := l.logger.WithGroup("checks.attribute_rule_collisions")
	logger.Debug("Checking for duplicate elements within the same attribute spec")

	for _, spec := range l.p.AttributeSpecs {
		if spec.AST == nil || spec.AST.Ruleset == nil {
			continue
		}

		logger := logger.With(
			slog.String("file", spec.File.Name),
			slog.String("spec_pos", spec.AST.Start().String()))

		if len(spec.AST.Ruleset.List) <= 1 {
			continue
		}

		wildcardDupls := duplicateWildcardElementSelectors(spec)
		if len(wildcardDupls) >= 2 {
			reportDuplicateElements(l, logger, spec.File, "*", wildcardDupls)
		}

		elemDupls := duplicateListElementSelectors(spec)
		for elemSpec, dupls := range elemDupls {
			if len(dupls) >= 2 {
				reportDuplicateElements(l, logger, spec.File, elemSpec.HTMLName(), dupls)
			}
		}
	}
}

func duplicateWildcardElementSelectors(spec *file.AttributeSpec) []*duplicateAttributeRule {
	sels := make([]*duplicateAttributeRule, 0, len(spec.AST.Ruleset.List))

	for _, rule := range spec.AST.Ruleset.List {
		if rule == nil || rule.Selector == nil {
			continue
		}

		sel, _ := rule.Selector.(*ast.WildcardElementSelector)
		if sel == nil {
			continue
		}

		sels = append(sels, &duplicateAttributeRule{
			rule:      rule,
			highlight: rule.Selector,
		})
	}

	return sels
}

func duplicateListElementSelectors(spec *file.AttributeSpec) map[*file.ElementSpec][]*duplicateAttributeRule {
	elemMap := make(map[*file.ElementSpec][]*duplicateAttributeRule)

	for _, rule := range spec.AST.Ruleset.List {
		if rule == nil || rule.Selector == nil {
			continue
		}

		sel, _ := rule.Selector.(*ast.ListElementSelector)
		if sel == nil {
			continue
		}

		for _, elemRefAST := range sel.List {
			elemRef := spec.File.ElementReferenceByNode(elemRefAST)
			if elemRef == nil {
				continue
			}

			elemMap[elemRef.Spec] = append(elemMap[elemRef.Spec], &duplicateAttributeRule{
				rule:      rule,
				highlight: elemRefAST,
			})
		}
	}

	return elemMap
}

func reportDuplicateElements(l *linker, logger *slog.Logger, f *file.File, name string, dupls []*duplicateAttributeRule) {
	logger.Error("Found duplicate attribute rules",
		slog.String("element_selector", name),
		slog.Int("count", len(dupls)))

	primaries := make([]diagnostic.Annotation, len(dupls))
	for i, dupl := range dupls {
		primaries[i] = anno.Anno(f, anno.Annotation{
			Context:    anno.ContextLines(dupl.rule.Start(), dupl.rule.End()),
			Highlight:  anno.HighlightNode(dupl.highlight),
			Annotation: "used here",
		})
	}

	l.report(&diagnostic.Diagnostic{
		Message: "attribute definition: element selector used multiple times",
		Primary: primaries,
	})
}
