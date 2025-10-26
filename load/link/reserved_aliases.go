package link

import (
	"log/slog"
	"strings"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

type reservedAliasCheck struct{}

func (l *linker) CheckReservedAliases() {
	logger := l.Logger.WithGroup("checks.reserved_aliases")
	logger.Debug("Checking for reserved import aliases")
	defer l.Ran(l.Pkg, reservedAliasCheck{})

	for _, f := range l.Pkg.Files {
		logger := logger.With(slog.String("file", string(f.Name)))
		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.Alias == "" || !strings.HasPrefix(string(imp.Alias), "__corgi_") {
				continue
			}

			logger.Error("Import alias with reserved prefix",
				slog.String("alias", string(imp.Alias)),
				slog.String("import_path", string(imp.CorgiPath)))
			l.Report(&diagnostic.Diagnostic{
				Message: "import alias: cannot use `__corgi_` prefix",
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST.Alias, "illegal import alias"),
				},
				Explanation: "All import aliases starting with `__corgi_` are reserved for internal use.",
			})
		}
	}
}
