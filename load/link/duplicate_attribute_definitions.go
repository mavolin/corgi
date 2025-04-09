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

func (l *linker) checkDuplicateAttributeDefinitions(_ context.Context) {
	(&duplicateAttributeDefinitionChecker{
		reported:          set.NewSliceSet[*file.AttributeDefinition](max(len(l.p.AttributeDefinitions)-1, 0)),
		duplDefs:          make([]*file.AttributeDefinition, 0, 8),
		duplQualifiedDefs: make([]*file.AttributeDefinition, 0, 8),
	}).check(l, l.logger)
}

type (
	duplicateAttributeDefinitionChecker struct { // package-level
		reported          set.Set[*file.AttributeDefinition]
		duplDefs          []*file.AttributeDefinition
		duplQualifiedDefs []*file.AttributeDefinition
	}
)

func (c *duplicateAttributeDefinitionChecker) check(l *linker, logger *slog.Logger) {
	logger = logger.WithGroup("check.duplicate_attribute_definitions")
	logger.Info("Checking for duplicate attribute definitions")

	if len(l.p.AttributeDefinitions) <= 1 {
		logger.Info("One or no attribute definitions, skipping")
		return
	}

	for ai, a := range l.p.AttributeDefinitions[:len(l.p.AttributeDefinitions)-1] {
		aInfo := attrDefInfo(a)
		if aInfo == nil {
			continue
		}
		aLowerSelector, aLowerFullSelector := strings.ToLower(aInfo.selector()), strings.ToLower(aInfo.fullSelector())

		logger := logger.With(
			slog.String("file", a.File.Name),
			slog.String("pos", a.AST.Selector.Start().String()),
			slog.String("qualified_selector", aInfo.selector()),
			slog.String("full_selector", aInfo.fullSelector()))
		logger.Debug("Checking attribute definition")

		if c.reported.Contains(a) {
			logger.Debug("Already reported, skipping")
			continue
		}

		c.resetDuplicates()

		for _, b := range l.p.AttributeDefinitions[ai+1:] {
			bInfo := attrDefInfo(b)
			if bInfo == nil {
				continue
			}
			bLowerSelector, bLowerFullSelector := strings.ToLower(bInfo.selector()), strings.ToLower(bInfo.fullSelector())

			if aLowerSelector == bLowerSelector || aLowerFullSelector == bLowerFullSelector {
				c.recordDuplicate(b)
				c.reported.Add(b)
			}
		}

		if len(c.duplDefs) > 0 {
			c.reportDuplicate(l, logger, a, c.duplDefs)
		}
	}
}

func (c *duplicateAttributeDefinitionChecker) reportDuplicate(l *linker, logger *slog.Logger, first *file.AttributeDefinition, duplDefs []*file.AttributeDefinition) {
	logger.Error("Found duplicate attribute definitions")

	primaries := make([]diagnostic.Annotation, 0, 2*(len(duplDefs)+1))
	reportedPrefixes := set.NewSliceSet[*ast.AttributeDefinition](len(duplDefs) + 1)
	primaries = c.appendDuplicateAnnotation(primaries, first, reportedPrefixes)
	for _, b := range duplDefs {
		primaries = c.appendDuplicateAnnotation(primaries, b, reportedPrefixes)
	}

	l.report(&diagnostic.Diagnostic{
		Message: "attribute defined multiple times",
		Primary: primaries,
		Hints: []diagnostic.Hint{
			{Hint: "Remember that attribute names are case-insensitive."},
		},
	})
}

func (c *duplicateAttributeDefinitionChecker) appendDuplicateAnnotation(as []diagnostic.Annotation, def *file.AttributeDefinition, reportedPrefixes *set.SliceSet[*ast.AttributeDefinition]) []diagnostic.Annotation {
	txt := "first defined here"
	if len(as) > 0 {
		txt = "then again here"
	}

	if def.Definition.LParen == nil && def.Definition.Prefix != nil {
		return append(as,
			anno.Range(def.File, def.Definition.Prefix.Start(), def.AST.Selector.End(), txt))
	}

	if def.Definition.Prefix != nil && reportedPrefixes.Add(def.Definition) {
		as = append(as,
			anno.Node(def.File, def.Definition.Prefix, "with this prefix"))
	}
	return append(as, anno.Node(def.File, def.AST.Selector, txt))
}

func (c *duplicateAttributeDefinitionChecker) resetDuplicates() {
	c.duplDefs = c.duplDefs[:0]
}

func (c *duplicateAttributeDefinitionChecker) recordDuplicate(def *file.AttributeDefinition) {
	c.duplDefs = append(c.duplDefs, def)
}
