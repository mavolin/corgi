package link

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckImportCycles(ctx context.Context) {
	logger := l.logger.WithGroup("checks.import_cycles")

	importersGraph := importersGraph(ctx)

	if logger.Enabled(ctx, slog.LevelInfo) {
		imports := make([]string, len(importersGraph))
		for i, p := range importersGraph {
			imports[i] = p.ImportPath
		}
		logger.Debug("Checking for import cycles", slog.Any("chain", imports))
	}

	if len(importersGraph) == 0 {
		return
	}

	// don't report the same import multiple times
	reported := make(map[string]bool)

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))

		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.Path == "" {
				continue
			}

			if reported[imp.Path] {
				continue
			}

			for i, parentPackage := range slices.Backward(importersGraph) {
				if parentPackage.ImportPath != imp.Path {
					continue
				}

				logger.Error("Circular import detected",
					slog.String("import", imp.Path),
					slog.String("pos", imp.AST.Start().String()))

				// Build the import cycle message
				var msg strings.Builder
				msg.Grow(512)
				msg.WriteString("This package imports:")
				for i, p := range importersGraph[i:] {
					if i > 0 {
						msg.WriteString(", which imports")
					}
					msg.WriteString("\n  ")
					msg.WriteString(p.ImportPath)
				}
				l.report(&diagnostic.Diagnostic{
					Message: "circular import",
					Primary: []diagnostic.Annotation{anno.Node(f, imp.AST.Path, msg.String())},
					Explanation: "A circular import occurs when two or more packages import each other, " +
						"directly or indirectly. " +
						"Break the cycle by removing one of the imports.",
				})

				imp.LoadedWithErrors = true
				reported[imp.Path] = true

				// once we've found a cycle for this import, no need to check more parent packages
				break
			}
		}
	}
}
