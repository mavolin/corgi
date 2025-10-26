package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckSelfImport() {
	logger := l.Logger.WithGroup("checks.self_import")
	logger.Debug("Checking if package imports itself")

	for _, f := range l.Pkg.Files {
		logger := logger.With(slog.String("file", string(f.Name)))

		var reported bool

		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.CorgiPath == "" {
				continue
			} else if imp.CorgiPath != l.Pkg.CorgiImportPath && imp.CorgiPath != file.CorgiImportPath(l.Pkg.GoImportPath()) {
				continue
			}

			imp.Loaded = true // prevent deadlock
			if reported {
				continue
			}

			logger.Error("Import to current package detected",
				slog.String("import", string(imp.CorgiPath)),
				slog.String("pos", imp.AST.Start().String()))
			l.Report(&diagnostic.Diagnostic{
				Message:     "package imports itself",
				Primary:     []diagnostic.Annotation{anno.Node(f, imp.AST.Path, "import to this package")},
				Explanation: "You cannot import the package you are currently in.",
			})
			reported = true // don't report duplicate imports, done elsewhere
		}
	}
}
