package link

import (
	"context"
	"log/slog"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

func (l *linker) CheckDuplicateAttributeDefinitionElementSelectors(_ context.Context) {
	logger := l.logger.WithGroup("check.duplicate_attribute_definition_element_selectors")
	logger.Info("Checking for duplicate elements within attribute definitions within the same selector")

	duplElems := make([]*ast.ElementName, 0, 8)

	for _, def := range l.p.AttributeDefinitions {
		if def.AST == nil || def.AST.Ruleset == nil {
			continue
		}

		logger := logger.With(
			slog.String("file", def.File.Name),
			slog.String("spec_pos", def.AST.Start().String()))
		logger.Debug("Checking attribute definition spec")

		for _, rule := range def.AST.Ruleset.Rules {
			if rule == nil || rule.Selector == nil {
				continue
			}

			logger := logger.With(slog.String("rule_pos", rule.Start().String()))
			logger.Debug("Checking rule")

			sel, _ := rule.Selector.(*ast.ListElementSelector)
			if sel == nil {
				logger.Debug("Not a list element selector, skipping")
				continue
			} else if len(sel.Elements) <= 1 {
				logger.Debug("One or no elements, skipping")
				continue
			}

			(&duplicateAttributeDefinitionElementSelectorChecker{
				duplElems: duplElems[:0],
				reported:  l.takeStringSet(),
			}).checkRule(l, logger, def, rule, sel)
		}
	}
}

type duplicateAttributeDefinitionElementSelectorChecker struct { // attribute definition rule level
	duplElems []*ast.ElementName
	reported  set.Set[elementName]
}

func (c *duplicateAttributeDefinitionElementSelectorChecker) checkRule(l *linker, logger *slog.Logger, def *file.AttributeDefinition, rule *ast.AttributeRule, sel *ast.ListElementSelector) {
	for ai, a := range sel.Elements[:len(sel.Elements)-1] {
		if a == nil || a.Name == "" {
			continue
		}

		aLowerName := strings.ToLower(a.Name)

		logger := logger.With(
			slog.String("element", a.Name),
			slog.String("element_pos", a.Start().String()))
		logger.Debug("Checking element")

		if c.reported.Contains(aLowerName) {
			logger.Debug("Already reported, skipping")
			continue
		}

		c.resetDuplicates()

		for _, b := range sel.Elements[ai+1:] {
			if b == nil || b.Name == "" {
				continue
			}

			if aLowerName == strings.ToLower(b.Name) {
				c.recordDuplicate(b)
				if len(c.duplElems) == 0 {
					c.reported.Add(aLowerName)
				}
			}
		}

		if len(c.duplElems) > 0 {
			c.reportDuplicate(l, logger, def, rule, a, c.duplElems)
		}
	}
}

func (c *duplicateAttributeDefinitionElementSelectorChecker) reportDuplicate(l *linker, logger *slog.Logger, def *file.AttributeDefinition, rule *ast.AttributeRule, first *ast.ElementName, duplElems []*ast.ElementName) {
	logger.Error("Found duplicate element")

	primaries := make([]diagnostic.Annotation, 1, len(duplElems))
	primaries[0] = anno.Anno(def.File, anno.Annotation{
		Context:    anno.ContextLines(rule.Start(), rule.End()),
		Highlight:  anno.HighlightNode(first),
		Annotation: "first specified here",
	})
	for _, b := range duplElems {
		primaries = append(primaries, anno.Node(def.File, b, "then again here"))
	}

	l.report(&diagnostic.Diagnostic{
		Message: "attribute definition: element specified multiple times in the same selector",
		Primary: primaries,
		Hints: []diagnostic.Hint{
			{Hint: "Remember that element names are case-insensitive."},
		},
	})
}

func (c *duplicateAttributeDefinitionElementSelectorChecker) resetDuplicates() {
	c.duplElems = c.duplElems[:0]
}

func (c *duplicateAttributeDefinitionElementSelectorChecker) recordDuplicate(b *ast.ElementName) {
	c.duplElems = append(c.duplElems, b)
}
