package link

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

func (l *linker) CheckDuplicateComponents(_ context.Context) {
	(&duplicateComponentChecker{
		duplComps: make([]*file.Component, 0, 8),
		checked:   l.takeStringSet(),
	}).check(l, l.logger)
}

type duplicateComponentChecker struct { // package level
	duplComps []*file.Component
	checked   set.Set[componentName]
}

func (c *duplicateComponentChecker) check(l *linker, logger *slog.Logger) {
	logger = logger.WithGroup("check.duplicate_components")
	logger.Info("Checking for duplicate components")

	if len(l.p.Components) <= 1 {
		logger.Info("One or no components, skipping")
		return
	}

	for ai, a := range l.p.Components[:len(l.p.Components)-1] {
		if c.shouldCheck(a) {
			continue
		}

		aName := a.Header().Name.Ident
		logger := logger.With(
			slog.String("file", a.File.Name),
			slog.String("component_pos", a.Start().String()),
			slog.String("component", aName))
		logger.Debug("Checking component")

		if c.checked.Contains(aName) {
			logger.Debug("Already checked for components with that name")
			continue
		}
		c.checked.Add(aName)
		c.resetDuplicates()

		for _, b := range l.p.Components[ai+1:] {
			if c.shouldCheck(b) {
				continue
			}
			bName := b.Header().Name.Ident

			if aName == bName {
				c.recordDuplicate(b)
			}
		}

		if len(c.duplComps) > 0 {
			c.reportDuplicate(l, logger, a, c.duplComps)
		}
	}
}

func (c *duplicateComponentChecker) reportDuplicate(l *linker, logger *slog.Logger, first *file.Component, dupls []*file.Component) {
	logger.Error("Found duplicate components")

	primaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
	primaries[0] = anno.Node(first.File, first.Header().Name, "first defined here")
	for _, b := range dupls {
		primaries = append(primaries, anno.Node(b.File, b.Header().Name, "then again here"))
	}

	l.report(&diagnostic.Diagnostic{
		Message: "component defined multiple times",
		Primary: primaries,
	})
}

func (c duplicateComponentChecker) shouldCheck(comp *file.Component) bool {
	return comp != nil && comp.Header() != nil && comp.Header().Name != nil
}

func (c *duplicateComponentChecker) resetDuplicates() {
	c.duplComps = c.duplComps[:0]
}

func (c *duplicateComponentChecker) recordDuplicate(comp *file.Component) {
	c.duplComps = append(c.duplComps, comp)
}
