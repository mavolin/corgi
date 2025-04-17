package link

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

func (l *linker) CheckImportCycles(ctx context.Context) {
	(&importCycleChecker{
		importersGraph: importersGraph(ctx),
		reported:       l.takeStringSet(),
	}).check(ctx, l, l.logger)
}

type importCycleChecker struct { // package level
	importersGraph []*file.Package
	reported       set.Set[importPath]
}

func (c *importCycleChecker) check(ctx context.Context, l *linker, logger *slog.Logger) {
	logger = l.logger.WithGroup("check.import_cycles")

	if logger.Enabled(ctx, slog.LevelInfo) {
		imports := make([]string, len(c.importersGraph))
		for i, p := range c.importersGraph {
			imports[i] = p.ImportPath
		}
		logger.Info("Checking for import cycles", slog.Any("chain", imports))
	}
	if len(c.importersGraph) == 0 {
		logger.Info("This is the root package, skipping")
		return
	}

	for _, f := range l.p.Files {
		c.checkFile(l, logger, f)
	}
}

func (c *importCycleChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.With(slog.String("file", f.Name))
	logger.Debug("Checking file")

	for _, imp := range f.Symbols.Imports {
		c.checkImport(l, logger, f, imp)
	}
}

func (c *importCycleChecker) checkImport(l *linker, logger *slog.Logger, f *file.File, imp *file.Import) {
	impPath := imp.ImportPath()
	if impPath == "" {
		return
	}
	logger = logger.With(
		slog.String("pos", imp.AST.Start().String()),
		slog.String("import", impPath))
	logger.Debug("Checking import")
	if c.reported.Contains(impPath) {
		logger.Debug("Already reported, skipping")
		return
	}

	for i, parentPkg := range slices.Backward(c.importersGraph) {
		if parentPkg.ImportPath != impPath {
			continue
		}
		logger.Error("Circular import")

		var msg strings.Builder
		msg.Grow(512)
		msg.WriteString("This package imports:")
		for i, p := range c.importersGraph[i:] {
			if i > 0 {
				msg.WriteString(", which imports")
			}
			msg.WriteString("\n  ")
			msg.WriteString(p.ImportPath)
		}

		c.reported.Add(impPath)
		l.report(&diagnostic.Diagnostic{
			Message: "circular import",
			Primary: []diagnostic.Annotation{anno.Node(f, imp.AST.Path, msg.String())},
			Explanation: "A circular import occurs when two or more packages import each other, " +
				"directly or indirectly. " +
				"Break the cycle by removing one of the imports.",
		})
	}
}
