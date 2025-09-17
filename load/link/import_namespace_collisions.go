package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckImportNamespaceCollisions() {
	logger := l.logger.WithGroup("checks.import_namespace_collisions")
	logger.Debug("Checking for import qualifier collisions")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", string(f.Name)))

		if len(f.Imports) <= 1 {
			continue
		}

		dupls := make(map[file.Qualifier][]*file.Import)

		for _, imp := range f.Imports {
			if imp.Explicit() && imp.Qualifier != "" {
				dupls[imp.Qualifier] = append(dupls[imp.Qualifier], imp)
			}
		}

		for qualifier, imports := range dupls {
			if len(imports) <= 1 {
				continue
			}

			logger.Error("Import qualifier collisions",
				slog.String("qualifier", string(qualifier)),
				slog.Int("count", len(imports)))

			primaries := make([]diagnostic.Annotation, len(imports))
			for i, imp := range imports {
				primaries[i] = anno.Node(f, imp.AST, "has qualifier `"+string(imp.Qualifier)+"`")
			}
			l.report(&diagnostic.Diagnostic{
				Message:     "import collision",
				Primary:     primaries,
				Explanation: "There can only be one import per qualifier.",
				Hints: []diagnostic.Hint{
					{Hint: "Use an import alias, so that each import has a unique qualifier."},
				},
			})
		}
	}
}
