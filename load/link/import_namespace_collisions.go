package link

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

func (l *linker) CheckImportNamespaceCollisions(_ context.Context) {
	logger := l.logger.WithGroup("check.import_namespace_collisions")
	logger.Info("Checking for import collisions")

	duplNamespace := make([]*file.Import, 0, 8)

	for _, f := range l.p.Files {
		(&importNamespaceCollisionChecker{
			reported:      l.takeStringSet(),
			duplNamespace: duplNamespace[:0],
		}).checkFile(l, logger, f)
	}
}

type importNamespaceCollisionChecker struct { // file level
	reported      set.Set[importPath]
	duplNamespace []*file.Import
}

func (c *importNamespaceCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.With(slog.String("file", f.Name))
	logger.Debug("Checking file")
	if len(f.Symbols.Imports) <= 1 {
		logger.Debug("One or no imports, skipping")
		return
	}

	for ai, a := range f.Symbols.Imports[:len(f.Symbols.Imports)-1] {
		aNamespace := a.Namespace()
		if aNamespace == "" || aNamespace == "." {
			continue
		}
		aImpPath := a.ImportPath()
		logger := logger.With(
			slog.String("pos", a.AST.Start().String()),
			slog.String("import", aImpPath),
			slog.String("namespace", aNamespace))
		logger.Debug("Checking import")

		if c.reported.Contains(aImpPath) {
			logger.Debug("Already reported, skipping")
			continue
		}
		c.resetDuplicates()

		for _, b := range f.Symbols.Imports[ai+1:] {
			bNamespace := b.Namespace()
			if bNamespace == "" {
				continue
			} else if aNamespace == bNamespace {
				logger.Debug("Found duplicate", slog.String("duplicate_import", b.ImportPath()))
				c.recordDuplicate(b)
			}
		}

		if len(c.duplNamespace) > 0 {
			c.reported.Add(aImpPath)
			c.reportCollision(l, logger, f, a, c.duplNamespace)
		}
	}
}

func (c *importNamespaceCollisionChecker) reportCollision(l *linker, logger *slog.Logger, f *file.File, first *file.Import, dupls []*file.Import) {
	ns := first.Namespace()
	logger.Error("Import collisions", slog.Int("n_collisions", len(ns)))

	primaries := make([]diagnostic.Annotation, 1, 1+len(dupls))
	primaries[0] = anno.Node(f, first.AST, "namespace `"+ns+"` used for the first time here")
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.AST, "then again here"))
	}

	l.report(&diagnostic.Diagnostic{
		Message:     "import collision",
		Primary:     primaries,
		Explanation: "There can only be one import per namespace.",
		Hints: []diagnostic.Hint{
			{Hint: "Use an import alias, so that each import has a unique namespace."},
		},
	})
}

func (c *importNamespaceCollisionChecker) resetDuplicates() {
	c.duplNamespace = c.duplNamespace[:0]
}

func (c *importNamespaceCollisionChecker) recordDuplicate(imp *file.Import) {
	c.duplNamespace = append(c.duplNamespace, imp)
}
