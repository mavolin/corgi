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

func (l *linker) checkDuplicateAttributeDefinitionElementTypes(_ context.Context) {
	logger := l.logger.WithGroup("check.duplicate_attribute_definition_types")
	logger.Info("Checking for duplicate elements within attribute definitions within different selectors")

	duplElems := make([]*duplicateAttributeDefinitionElementType, 0, 8)

	for _, def := range l.p.AttributeDefinitions {
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

		(&duplicateAttributeDefinitionElementTypeChecker{
			reported:  l.takeStringSet(),
			duplElems: duplElems[:0],
		}).checkDefinition(l, logger, def)
	}
}

type (
	duplicateAttributeDefinitionElementTypeChecker struct { // attribute definition level
		reported  set.Set[elementName]
		duplElems []*duplicateAttributeDefinitionElementType
	}

	duplicateAttributeDefinitionElementType struct {
		rule *ast.AttributeRule
		elem *ast.ElementName
	}
)

func (c *duplicateAttributeDefinitionElementTypeChecker) checkDefinition(l *linker, logger *slog.Logger, def *file.AttributeDefinition) {
	for ai, a := range def.AST.Ruleset.Rules[:len(def.AST.Ruleset.Rules)-1] {
		if a == nil || a.Selector == nil {
			continue
		}

		logger := logger.With(slog.String("rule_pos", a.Start().String()))
		logger.Debug("Checking rule")

		aSel, _ := a.Selector.(*ast.ListElementSelector)
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

			for _, b := range def.AST.Ruleset.Rules[ai+1:] {
				if b == nil || b.Selector == nil {
					continue
				}

				bSel, _ := b.Selector.(*ast.ListElementSelector)
				if bSel == nil {
					continue
				}

				for _, bElem := range bSel.Elements {
					if bElem == nil || bElem.Name == "" {
						continue
					}

					if aLowerName == strings.ToLower(bElem.Name) {
						c.recordDuplicate(bElem, b)
						if len(c.duplElems) == 0 {
							c.reported.Add(aLowerName)
						}
					}
				}
			}

			if len(c.duplElems) > 0 {
				c.reportDuplicate(l, logger, def.File, duplicateAttributeDefinitionElementType{a, aElem}, c.duplElems)
			}
		}
	}
}

func (c *duplicateAttributeDefinitionElementTypeChecker) reportDuplicate(l *linker, logger *slog.Logger, f *file.File, first duplicateAttributeDefinitionElementType, dupls []*duplicateAttributeDefinitionElementType) {
	logger.Error("Found duplicate element")

	primaries := make([]diagnostic.Annotation, 1, len(dupls))
	primaries[0] = anno.Anno(f, anno.Annotation{
		Context:    anno.ContextLines(first.rule.Start(), first.rule.End()),
		Highlight:  anno.HighlightNode(first.elem),
		Annotation: "first defined here",
	})
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Anno(f, anno.Annotation{
			Context:    anno.ContextLines(dupl.rule.Start(), dupl.rule.End()),
			Highlight:  anno.HighlightNode(dupl.elem),
			Annotation: "then again here",
		}))
	}

	l.report(&diagnostic.Diagnostic{
		Message: "attribute definition: type for element defined multiple times",
		Primary: primaries,
	})
}

func (c *duplicateAttributeDefinitionElementTypeChecker) resetDuplicates() {
	c.duplElems = c.duplElems[:0]
}

func (c *duplicateAttributeDefinitionElementTypeChecker) recordDuplicate(elem *ast.ElementName, rule *ast.AttributeRule) {
	c.duplElems = append(c.duplElems, &duplicateAttributeDefinitionElementType{rule, elem})
}
