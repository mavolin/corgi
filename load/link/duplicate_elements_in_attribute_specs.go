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

func (l *linker) CheckDuplicateElementsInAttributeSpecs(_ context.Context) {
	logger := l.logger.WithGroup("check.duplicate_elements_in_attribute_specs")
	logger.Info("Checking for duplicate elements within the same attribute spec but in different rules")

	duplElems := make([]*duplicateElementInAttributeSpec, 0, 8)

	for _, def := range l.p.AttributeSpecs {
		if def.AST == nil || def.AST.Ruleset == nil {
			continue
		}

		logger := logger.With(
			slog.String("file", def.File.Name),
			slog.String("spec_pos", def.AST.Start().String()))
		logger.Debug("Checking attribute definition spec")

		if len(def.AST.Ruleset.Rules) <= 1 {
			logger.Debug("One or no rules, skipping")
			continue
		}

		(&duplicateElementsInAttributeSpecsChecker{
			reported:  l.takeStringSet(),
			duplElems: duplElems[:0],
		}).checkSpec(l, logger, def)
	}
}

type (
	duplicateElementsInAttributeSpecsChecker struct { // attribute definition level
		reported  set.Set[elementName]
		duplElems []*duplicateElementInAttributeSpec
	}

	duplicateElementInAttributeSpec struct {
		rule *ast.AttributeRule
		name string
		pos  ast.Position
	}
)

func (c *duplicateElementsInAttributeSpecsChecker) checkSpec(l *linker, logger *slog.Logger, spec *file.AttributeSpec) {
	c.checkWildcardElementSelector(l, logger, spec)
	c.checkListElementSelectors(l, logger, spec)
}

func (c *duplicateElementsInAttributeSpecsChecker) checkListElementSelectors(
	l *linker, logger *slog.Logger, spec *file.AttributeSpec,
) {
	logger = logger.WithGroup("list_element_selectors")
	logger.Debug("Checking for duplicate list element selectors")

	for ai, aRule := range spec.AST.Ruleset.Rules[:len(spec.AST.Ruleset.Rules)-1] {
		if aRule == nil || aRule.Selector == nil {
			continue
		}

		logger := logger.With(slog.String("rule_pos", aRule.Start().String()))
		logger.Debug("Checking rule")

		aSel, _ := aRule.Selector.(*ast.ListElementSelector)
		if aSel == nil {
			logger.Debug("Not a list element selector, skipping")
			continue
		}

		for _, aElem := range aSel.Elements {
			if aElem == nil || aElem.Name == "" {
				continue
			}

			aLowerName := strings.ToLower(aElem.Name)

			logger := logger.With(
				slog.String("element", aElem.Name),
				slog.String("element_pos", aElem.Start().String()))
			logger.Debug("Checking element")

			if c.reported.Contains(aLowerName) {
				logger.Debug("Already reported, skipping")
				continue
			}

			c.resetDuplicates()

			for _, bRule := range spec.AST.Ruleset.Rules[ai+1:] {
				if bRule == nil || bRule.Selector == nil {
					continue
				}

				bSel, _ := bRule.Selector.(*ast.ListElementSelector)
				if bSel == nil {
					continue
				}

				for _, bElem := range bSel.Elements {
					if bElem == nil || bElem.Name == "" {
						continue
					}

					if aLowerName == strings.ToLower(bElem.Name) {
						logger.Debug("Found duplicate element in another rule",
							slog.String("other_rule_pos", bRule.Start().String()),
							slog.String("other_element_pos", bElem.Start().String()))
						c.recordDuplicate(bRule, bElem.Name, bElem.Start())
						if len(c.duplElems) == 0 {
							c.reported.Add(aLowerName)
						}
					}
				}
			}

			if len(c.duplElems) > 0 {
				first := &duplicateElementInAttributeSpec{aRule, aElem.Name, aElem.Start()}
				c.reportDuplicate(l, logger, spec.File, first, c.duplElems)
			}
		}
	}
}

func (c *duplicateElementsInAttributeSpecsChecker) checkWildcardElementSelector(
	l *linker, logger *slog.Logger, spec *file.AttributeSpec,
) {
	logger = logger.WithGroup("wildcard_element_selectors")
	logger.Debug("Checking for duplicate wildcard selectors")

	c.resetDuplicates()

	for _, rule := range spec.AST.Ruleset.Rules {
		if rule == nil || rule.Selector == nil {
			continue
		}

		logger := logger.With(slog.String("rule_pos", rule.Start().String()))
		logger.Debug("Checking rule")

		sel, _ := rule.Selector.(*ast.WildcardElementSelector)
		if sel == nil {
			logger.Debug("Not a wildcard element selector, skipping")
			continue
		}

		c.recordDuplicate(rule, "*", sel.Start())
	}

	if len(c.duplElems) >= 2 {
		c.reportDuplicate(l, logger, spec.File, c.duplElems[0], c.duplElems[1:])
	}
}

func (c *duplicateElementsInAttributeSpecsChecker) reportDuplicate(
	l *linker, logger *slog.Logger, f *file.File,
	first *duplicateElementInAttributeSpec, dupls []*duplicateElementInAttributeSpec,
) {
	logger.Error("Found duplicate element")

	primaries := make([]diagnostic.Annotation, 1, len(dupls))
	primaries[0] = anno.Anno(f, anno.Annotation{
		Context:    anno.ContextLines(first.rule.Start(), first.rule.End()),
		Highlight:  anno.HighlightNRunes(first.pos, len(first.name)),
		Annotation: "first defined here",
	})
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Anno(f, anno.Annotation{
			Context:    anno.ContextLines(dupl.rule.Start(), dupl.rule.End()),
			Highlight:  anno.HighlightNRunes(dupl.pos, len(dupl.name)),
			Annotation: "then again here",
		}))
	}

	l.report(&diagnostic.Diagnostic{
		Message: "attribute definition: type for element defined multiple times",
		Primary: primaries,
	})
}

func (c *duplicateElementsInAttributeSpecsChecker) resetDuplicates() {
	c.duplElems = c.duplElems[:0]
}

func (c *duplicateElementsInAttributeSpecsChecker) recordDuplicate(rule *ast.AttributeRule, name string, pos ast.Position) {
	c.duplElems = append(c.duplElems, &duplicateElementInAttributeSpec{rule, name, pos})
}
