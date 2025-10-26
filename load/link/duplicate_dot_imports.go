package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

type duplicateDotImportsCheck struct{}

func (l *linker) CheckDuplicateDotImports() {
	logger := l.Logger.WithGroup("checks.duplicate_dot_imports")
	logger.Debug("Checking for duplicate dot imports")
	defer l.Ran(l.Pkg, duplicateDotImportsCheck{})

	for _, f := range l.Pkg.Files {
		logger := logger.With(slog.String("file", string(f.Name)))

		if len(f.Imports) <= 1 {
			continue
		}

		// l.dotImports is deduplicated
		dotImports := make(map[file.CorgiImportPath][]*file.Import)

		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.AST.Alias == nil || imp.AST.Alias.Name != "." {
				continue
			}

			dotImports[imp.CorgiPath] = append(dotImports[imp.CorgiPath], imp)
		}

		for path, imports := range dotImports {
			if len(imports) <= 1 {
				continue
			}

			logger.Error("Duplicate dot imports",
				slog.String("import_path", string(path)),
				slog.Int("count", len(imports)))

			primaries := make([]diagnostic.Annotation, 0, len(imports))
			for i, imp := range imports {
				var msg string
				if i == 0 {
					msg = "package dot-imported for the first time here"
				} else {
					msg = "then again here"
				}
				primaries = append(primaries, anno.Node(f, imp.AST, msg))
			}

			l.Report(&diagnostic.Diagnostic{
				Message:     "duplicated dot imports",
				Primary:     primaries,
				Explanation: "You can only dot-import a package once per file.",
				Hints: []diagnostic.Hint{
					{Hint: "Remove the duplicate dot imports."},
				},
			})
		}
	}
}
