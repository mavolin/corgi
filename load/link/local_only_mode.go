package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

type localOnlyModeCheck struct{}

func (l *linker) CheckLocalOnlyMode() {
	defer l.Ran(l.Pkg, localOnlyModeCheck{})

	if !l.isLocalOnlyMode() {
		return
	}

	l.Logger.Debug("Running in local-only mode: Checking for no imports")

	for _, f := range l.Pkg.Files {
		logger := l.Logger.With(slog.String("file", string(f.Name)))

		if len(f.Imports) == 0 {
			continue
		}

		primaries := make([]diagnostic.Annotation, 0, len(f.Imports))
		for _, imp := range f.Imports {
			if imp.Explicit() && imp.AST != nil {
				primaries = append(primaries, anno.Node(f, imp.AST, "illegal import"))
			}
		}
		if len(primaries) > 0 {
			logger.Error("Local-only mode: File contains imports")
			l.Report(&diagnostic.Diagnostic{
				Message:     "local-only mode: file contains imports",
				Primary:     primaries,
				Explanation: "In local-only mode, files are not allowed to make any imports.",
			})
		}
	}
}

func (l *linker) isLocalOnlyMode() bool {
	return l.importer == nil
}
