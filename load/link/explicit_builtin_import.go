package link

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckExplicitBuiltinImport(_ context.Context) {
	logger := l.logger.WithGroup("check.explicit_builtin_import")
	logger.Info("Checking for an illegal explicit import of the builtin package")

	if l.builtin == nil {
		logger.Info("No builtin package set, skipping")
		return
	}

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Checking file")

		for _, imp := range f.Imports {
			impPath := imp.ImportPath()
			logger := logger.With(
				slog.String("pos", imp.AST.Start().String()),
				slog.String("import", impPath))
			logger.Debug("Checking import")
			if impPath == "" || impPath != l.builtin.ImportPath {
				continue
			}

			logger.Error("Explicit import of builtin package")
			l.report(&diagnostic.Diagnostic{
				Message: "import of builtin package not allowed",
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST, "explicit import of builtin package"),
				},
				Explanation: "The builtin package is implicitly imported in every file. " +
					"Explicitly importing it serves no purpose and is not allowed.",
			})
		}
	}
}
