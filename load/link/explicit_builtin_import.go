package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckExplicitBuiltinImport() {
	logger := l.Logger.WithGroup("check.explicit_builtin_import")
	logger.Info("Checking for an illegal explicit import of the builtin package")

	for _, f := range l.Pkg.Files {
		logger := logger.With(slog.String("file", string(f.Name)))

		builtin := f.BuiltinImport()
		if builtin == nil {
			continue
		}

		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.CorgiPath != builtin.CorgiPath {
				continue
			}

			logger := logger.With(
				slog.String("pos", imp.AST.Start().String()),
				slog.String("import", string(imp.CorgiPath)))

			logger.Error("Explicit import of builtin package")
			l.Report(&diagnostic.Diagnostic{
				Message: "explicit import of builtin package",
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST, "explicit import of builtin package"),
				},
				Explanation: "The builtin package is implicitly imported in every file. " +
					"Explicitly importing it serves no purpose and is not allowed.",
			})
			imp.Loaded, imp.Package = true, nil
		}
	}
}
