package link

import (
	"context"
	"log/slog"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

func (l *linker) CheckDuplicateElementDefinitions(_ context.Context) {
	(&duplicateElementDefinitionChecker{
		duplDefs: make([]*file.ElementSpec, 0, 8),
		reported: set.NewSliceSet[*file.ElementSpec](max(len(l.p.ElementSpecs)-1, 0)),
	}).check(l, l.logger)
}

type duplicateElementDefinitionChecker struct { // package level
	duplDefs []*file.ElementSpec
	reported set.Set[*file.ElementSpec]
}

func (c *duplicateElementDefinitionChecker) check(l *linker, logger *slog.Logger) {
	logger = logger.WithGroup("check.duplicate_element_definitions")
	logger.Info("Checking for duplicate element definitions")

	if len(l.p.ElementSpecs) <= 1 {
		logger.Info("One or no element definitions, skipping")
		return
	}

	for ai, a := range l.p.ElementSpecs[:len(l.p.ElementSpecs)-1] {
		if !c.shouldCheck(a) {
			continue
		}

		aQualifiedName, aFullName := a.QualifiedName(), a.FullName()
		if aQualifiedName == "" {
			continue
		}
		aLowerQualifiedName := strings.ToLower(aFullName)
		aLowerFullName := strings.ToLower(aQualifiedName)

		logger := logger.With(
			slog.String("file", a.File.Name),
			slog.String("spec_pos", a.AST.Start().String()),
			slog.String("qualified_element", aQualifiedName),
			slog.String("full_element", aFullName))
		logger.Debug("Checking element definition spec")

		if c.reported.Contains(a) {
			logger.Debug("Already reported, skipping")
			continue
		}

		c.resetDuplicates()

		for _, b := range l.p.ElementSpecs[ai+1:] {
			if !c.shouldCheck(b) {
				continue
			}

			bQualifiedName, bFullName := b.QualifiedName(), b.FullName()
			if bQualifiedName == "" {
				continue
			}
			bLowerQualifiedName := strings.ToLower(bFullName)
			bLowerFullName := strings.ToLower(bQualifiedName)

			if aLowerQualifiedName == bLowerQualifiedName || aLowerFullName == bLowerFullName {
				c.recordDuplicate(b)
				c.reported.Add(b)
			}
		}

		if len(c.duplDefs) > 0 {
			c.reportDuplicate(l, logger, a, c.duplDefs)
		}
	}
}

func (c *duplicateElementDefinitionChecker) reportDuplicate(l *linker, logger *slog.Logger, first *file.ElementSpec, dupls []*file.ElementSpec) {
	logger.Error("Found duplicate element definitions")

	primaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
	primaries[0] = anno.Node(first.File, first.AST.Name, "first defined here")
	for _, b := range dupls {
		primaries = append(primaries, anno.Node(b.File, b.AST.Name, "then again here"))
	}

	l.report(&diagnostic.Diagnostic{
		Message: "element defined multiple times",
		Primary: primaries,
	})
}

func (c duplicateElementDefinitionChecker) shouldCheck(def *file.ElementSpec) bool {
	return def.AST != nil && def.Definition != nil && def.AST.Name != nil
}

func (c *duplicateElementDefinitionChecker) resetDuplicates() {
	c.duplDefs = c.duplDefs[:0]
}

func (c *duplicateElementDefinitionChecker) recordDuplicate(def *file.ElementSpec) {
	c.duplDefs = append(c.duplDefs, def)
}
