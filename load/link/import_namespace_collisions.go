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
	if len(f.Imports) <= 1 {
		logger.Debug("One or no imports, skipping")
		return
	}

	for ai, a := range f.Imports[:len(f.Imports)-1] {
		if a.Namespace == "" || a.Namespace == "." {
			continue
		}
		logger := logger.With(
			slog.String("pos", a.AST.Start().String()),
			slog.String("import", a.Path),
			slog.String("namespace", a.Namespace))
		logger.Debug("Checking import")

		if c.reported.Contains(a.Path) {
			logger.Debug("Already reported, skipping")
			continue
		}
		c.resetDuplicates()

		for _, b := range f.Imports[ai+1:] {
			if b.Namespace == "" || b.Namespace == "." {
				continue
			} else if a.Namespace == b.Namespace {
				logger.Debug("Found duplicate", slog.String("duplicate_import", b.Path))
				c.recordDuplicate(b)
			}
		}

		if len(c.duplNamespace) > 0 {
			c.reported.Add(a.Path)
			c.reportCollision(l, logger, f, a, c.duplNamespace)
		}
	}
}

func (c *importNamespaceCollisionChecker) reportCollision(l *linker, logger *slog.Logger, f *file.File, first *file.Import, dupls []*file.Import) {
	logger.Error("Import collisions",
		slog.String("namespace", first.Namespace),
		slog.Int("n_collisions", len(first.Namespace)))

	primaries := make([]diagnostic.Annotation, 1, 1+len(dupls))
	primaries[0] = anno.Node(f, first.AST, "has namespace `"+first.Namespace+"`")
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.AST, "also has namespace `"+dupl.Namespace+"`"))
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
