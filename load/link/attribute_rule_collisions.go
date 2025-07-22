package link

import (
	"log/slog"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

type duplicateAttributeRule struct {
	rule *ast.AttributeRule
	name string
	pos  ast.Position
}

func (l *linker) CheckAttributeRuleCollisions() {
	logger := l.logger.WithGroup("checks.attribute_rule_collisions")
	logger.Debug("Checking for duplicate elements within the same attribute spec but in different rules")

	for _, spec := range l.p.AttributeSpecs {
		if spec.AST == nil || spec.AST.Ruleset == nil {
			continue
		}

		logger := logger.With(
			slog.String("file", spec.File.Name),
			slog.String("spec_pos", spec.AST.Start().String()))

		if len(spec.AST.Ruleset.Rules) <= 1 {
			logger.Debug("One or no rules, skipping")
			continue
		}

		wildcardDupls := duplicateWildcardElementSelectors(spec)
		if len(wildcardDupls) >= 2 {
			reportDuplicateElements(l, logger, spec.File, wildcardDupls)
		}

		elemDupls := duplicateListElementSelectors(spec)
		for _, dupls := range elemDupls {
			if len(dupls) >= 2 {
				reportDuplicateElements(l, logger, spec.File, dupls)
			}
		}
	}
}

func duplicateWildcardElementSelectors(spec *file.AttributeSpec) []*duplicateAttributeRule {
	sels := make([]*duplicateAttributeRule, 0, len(spec.AST.Ruleset.Rules))

	for _, rule := range spec.AST.Ruleset.Rules {
		if rule == nil || rule.Selector == nil {
			continue
		}

		sel, _ := rule.Selector.(*ast.WildcardElementSelector)
		if sel == nil {
			continue
		}

		sels = append(sels, &duplicateAttributeRule{
			rule: rule,
			name: "*",
			pos:  sel.Start(),
		})
	}

	return sels
}

func duplicateListElementSelectors(spec *file.AttributeSpec) map[elementName][]*duplicateAttributeRule {
	elemMap := make(map[elementName][]*duplicateAttributeRule)

	for _, rule := range spec.AST.Ruleset.Rules {
		if rule == nil || rule.Selector == nil {
			continue
		}

		sel, _ := rule.Selector.(*ast.ListElementSelector)
		if sel == nil {
			continue
		}

		for _, elem := range sel.Elements {
			if elem == nil || elem.Name == "" {
				continue
			}

			name := strings.ToLower(elem.Name)
			elemMap[name] = append(elemMap[name], &duplicateAttributeRule{
				rule: rule,
				name: elem.Name,
				pos:  elem.Start(),
			})
		}
	}

	return elemMap
}

func reportDuplicateElements(l *linker, logger *slog.Logger, f *file.File, dupls []*duplicateAttributeRule) {
	logger.Error("Found duplicate attribute rules",
		slog.String("element_selector", dupls[0].name),
		slog.Int("count", len(dupls)))

	primaries := make([]diagnostic.Annotation, len(dupls))
	for i, dupl := range dupls {
		primaries[i] = anno.Anno(f, anno.Annotation{
			Context:    anno.ContextLines(dupl.rule.Start(), dupl.rule.End()),
			Highlight:  anno.HighlightNRunes(dupl.pos, len(dupl.name)),
			Annotation: "used here",
		})
	}

	l.report(&diagnostic.Diagnostic{
		Message: "attribute definition: element selector used multiple times",
		Primary: primaries,
	})
}
