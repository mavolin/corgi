package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckSelfImport() {
	logger := l.logger.WithGroup("checks.self_import")
	logger.Debug("Checking if package imports itself")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))

		var reported bool

		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.Path == "" {
				continue
			} else if imp.Path != l.p.ImportPath && imp.Path != l.p.ModulePath() {
				continue
			}

			imp.Loaded = true // prevent deadlock
			if reported {
				continue
			}

			logger.Error("Import to current package detected",
				slog.String("import", imp.Path),
				slog.String("pos", imp.AST.Start().String()))
			l.report(&diagnostic.Diagnostic{
				Message:     "package imports itself",
				Primary:     []diagnostic.Annotation{anno.Node(f, imp.AST.Path, "import to this package")},
				Explanation: "You cannot import the package you are currently in.",
			})
			reported = true // don't report duplicate imports, done elsewhere
		}
	}
}
