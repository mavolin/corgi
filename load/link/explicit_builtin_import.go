package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckExplicitBuiltinImport() {
	logger := l.logger.WithGroup("check.explicit_builtin_import")
	logger.Info("Checking for an illegal explicit import of the builtin package")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Checking file")

		builtin := f.BuiltinImport()
		if builtin == nil {
			logger.Debug("File has no builtin import, skipping")
			continue
		}

		for _, imp := range f.Imports {
			if imp.AST == nil || imp.Path == "" {
				continue
			}

			logger := logger.With(
				slog.String("pos", imp.AST.Start().String()),
				slog.String("import", imp.Path))
			logger.Debug("Checking import")
			if imp.Path != builtin.Path {
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
			imp.LoadedWithErrors = true
		}
	}
}
