package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckImportNamespaceCollisions() {
	logger := l.logger.WithGroup("checks.import_namespace_collisions")
	logger.Debug("Checking for import namespace collisions")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))

		if len(f.Imports) <= 1 {
			continue
		}

		dupls := make(map[namespace][]*file.Import)

		for _, imp := range f.Imports {
			if imp.Explicit() && imp.Namespace != "" {
				dupls[imp.Namespace] = append(dupls[imp.Namespace], imp)
			}
		}

		for namespace, imports := range dupls {
			if len(imports) <= 1 {
				continue
			}

			logger.Error("Import namespace collisions",
				slog.String("namespace", namespace),
				slog.Int("count", len(imports)))

			primaries := make([]diagnostic.Annotation, len(imports))
			for i, imp := range imports {
				primaries[i] = anno.Node(f, imp.AST, "has namespace `"+imp.Namespace+"`")
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
	}
}
