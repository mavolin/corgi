package link

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) LoadImports(ctx context.Context, g *importGraph) {
	(&importLoader{
		l:      l,
		logger: l.logger.WithGroup("imports"),
		graph:  g,
	}).load(ctx)
}

type importLoader struct {
	l      *linker
	logger *slog.Logger
	graph  *importGraph
}

func (loader *importLoader) load(ctx context.Context) {
	loader.logger.Debug("Loading imports")

	if loader.l.importer == nil {
		loader.localOnlyModeCheck()
		return
	}

	loader.checkIllegalAliases()
	loader.loadImports(ctx)
}

func (loader *importLoader) localOnlyModeCheck() {
	loader.logger.Info("Running in local-only mode, not allowed to use imports")

	for _, f := range loader.l.p.Files {
		logger := loader.logger.With(slog.String("file", string(f.Name)))

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
			loader.l.report(&diagnostic.Diagnostic{
				Message:     "local-only mode: file contains imports",
				Primary:     primaries,
				Explanation: "In local-only mode, files are not allowed to make any imports.",
			})
		}
	}
}

func (loader *importLoader) checkIllegalAliases() {
	loader.logger.Debug("Checking for import aliases using the reserved `__corgi_` prefix")

	for _, f := range loader.l.p.Files {
		logger := loader.logger.With(slog.String("file", string(f.Name)))
		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.Alias == "" || !strings.HasPrefix(string(imp.Alias), "__corgi_") {
				continue
			}

			logger.Error("Import alias with reserved prefix",
				slog.String("alias", string(imp.Alias)),
				slog.String("import_path", string(imp.CorgiPath)))
			loader.l.report(&diagnostic.Diagnostic{
				Message: "import alias: cannot use `__corgi_` prefix",
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST.Alias, "illegal import alias"),
				},
				Explanation: "All import aliases starting with `__corgi_` are reserved for internal use.",
			})
		}
	}
}

func (loader *importLoader) loadImports(ctx context.Context) {
	loader.logger.Info("Concurrently loading imports")

	for _, f := range loader.l.p.Files {
		for _, imp := range f.Imports {
			if imp.Explicit() && !imp.Loaded {
				loader.loadImport(ctx, f, imp)
			}
		}
	}

	loader.loadBuiltin(ctx)

	for _, f := range loader.l.p.Files {
		for _, imp := range f.Imports {
			if imp.Explicit() && !imp.Loaded {
				loader.processImport(f, imp)
			}
		}
	}

	loader.logger.Info("Finished loading imports")
}

func (loader *importLoader) loadImport(ctx context.Context, f *file.File, imp *file.Import) {
	logger := loader.logger.With(
		slog.String("file", string(f.Name)),
		slog.String("import", string(imp.CorgiPath)),
		slog.String("import_pos", imp.AST.Start().String()))
	logger.Debug("Loading import")

	cycle := loader.graph.AddImport(loader.l.p, imp.CorgiPath, func() (*file.Package, diagnostic.List, error) {
		return loader.l.importer(ctx, imp.CorgiPath)
	})
	if cycle != nil {
		loader.reportImportCycle(logger, f, imp, cycle)
	}
}

func (loader *importLoader) processImport(f *file.File, imp *file.Import) {
	logger := loader.logger.With(
		slog.String("file", string(f.Name)),
		slog.String("import", string(imp.CorgiPath)),
		slog.String("import_pos", imp.AST.Start().String()))

	var d diagnostic.List
	var err error
	imp.Package, d, err = loader.graph.AwaitImport(imp.CorgiPath)
	imp.Loaded = true

	if err != nil {
		logger.Error("Failed to load import", slog.String("err", err.Error()))
		loader.l.report(&diagnostic.Diagnostic{
			Message: "import: failed to load package",
			Cause:   err,
			Primary: []diagnostic.Annotation{
				anno.Node(f, imp.AST, "couldn't load this import"),
			},
		})
	}
	if len(d) > 0 {
		loader.l.report(d...)
		logger.Error("Import contains errors", slog.String("err", d.Short()))
	}

	logger.Debug("Successfully loaded import")

	if imp.Package == nil {
		return
	}

	imp.GoPath = imp.Package.GoImportPath()

	if imp.Package.CorgiImportPath != imp.CorgiPath {
		logger.Error("Import using non-symbolic path",
			slog.String("go_import_path", string(imp.GoPath)))
		loader.l.report(&diagnostic.Diagnostic{
			Message: "import: use of Go import path when symbolic path exists",
			Primary: []diagnostic.Annotation{
				anno.Node(f, imp.AST, "expected import path to be `"+string(imp.Package.CorgiImportPath)+"`"),
			},
			Explanation: "Special import paths, like the standard library imports starting with `corgi/` " +
				"are available under a special short symbolic path. " +
				"If such a symbolic path exists, you must use it, so the linker can properly resolve imports.",
		})
	}

	switch {
	case imp.Alias == ".":
		imp.Qualifier = ""
	case imp.Alias != "":
		imp.Qualifier = imp.Alias
	default:
		imp.Qualifier = imp.Package.Name
		if strings.HasPrefix(string(imp.Qualifier), "__corgi_") {
			logger.Error("Import has package name with reserved prefix",
				slog.String("qualifier", string(imp.Qualifier)),
				slog.String("import_path", string(imp.CorgiPath)))

			loader.l.report(&diagnostic.Diagnostic{
				Message: "import: import uses reserved `__corgi_` package name prefix",
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST, "illegal package name"),
				},
				Explanation: "All namespaces starting with `__corgi_` are reserved for internal use.",
				Hints: []diagnostic.Hint{
					{Hint: "Use an import alias."},
				},
			})
		}
	}
}

func (loader *importLoader) loadBuiltin(ctx context.Context) {
	if loader.l.builtinPath == "" {
		return // no builtin path set, nothing to load
	} else if loader.l.p.CorgiImportPath == loader.l.builtinPath {
		return // this is the builtin package, nothing to do
	}
	logger := loader.logger.With(slog.String("import", string(loader.l.builtinPath)))

	var needBuiltin bool
	for _, f := range loader.l.p.Files {
		if f.BuiltinImport() == nil {
			needBuiltin = true
			continue
		}

		logger.Error("File already has a builtin import",
			slog.String("file", string(f.Name)))
		loader.l.report(&diagnostic.Diagnostic{
			Type:    diagnostic.InternalError,
			Message: "file already has a builtin import",
			Primary: []diagnostic.Annotation{
				anno.Position(f, ast.Position{Line: 1, Col: 1}, "file's builtin import already set"),
			},
			Explanation: "This file already has a builtin package set, but the linker was given a non-empty builtin path " +
				"to load.",
		})
	}
	if !needBuiltin {
		logger.Debug("All files already have a builtin package, no need to load it again")
		return // no builtin package needed, nothing to load
	}

	logger.Debug("Loading builtin package")

	cycle := loader.graph.AddImport(loader.l.p, loader.l.builtinPath, func() (*file.Package, diagnostic.List, error) {
		return loader.l.importer(ctx, loader.l.builtinPath)
	})
	if cycle != nil {
		loader.reportImportCycle(logger, loader.l.p.Files[0], nil, cycle)
		return
	}

	p, d, err := loader.graph.AwaitImport(loader.l.builtinPath)
	switch {
	case err != nil:
		logger.Error("Failed to load builtin package", slog.String("err", err.Error()))
	case len(d) > 0:
		logger.Error("Builtin package contains errors", slog.String("err", d.Short()))
	default:
		logger.Debug("Successfully loaded builtin package")
	}

	if len(d) > 0 {
		loader.l.report(d...)
	}
	if err != nil {
		loader.l.report(&diagnostic.Diagnostic{
			Message:     "failed to load builtin package",
			Cause:       err,
			Explanation: fmt.Sprintf("Attempting to load %q.", loader.l.builtinPath),
		})
	}
	if p != nil {
		for _, f := range loader.l.p.Files {
			if f.BuiltinImport() != nil {
				continue
			}

			f.AddBuiltinImport(f.UniqueQualifier(BuiltinAlias), p)
		}
	}
}

func (loader *importLoader) reportImportCycle(logger *slog.Logger, f *file.File, imp *file.Import, cycle []file.CorgiImportPath) {
	logger.Error("Circular import detected")

	// Build the import cycle message
	var msg strings.Builder
	msg.Grow(512)
	msg.WriteString("This package imports:")
	for i, p := range cycle {
		if i > 0 {
			msg.WriteString(", which imports")
		}
		msg.WriteString("\n  ")
		msg.WriteString(string(p))
	}

	var primary []diagnostic.Annotation
	if imp != nil && imp.AST != nil {
		primary = []diagnostic.Annotation{anno.Node(f, imp.AST.Path, msg.String())}
	} else {
		primary = []diagnostic.Annotation{anno.Position(f, ast.Position{Line: 1, Col: 1}, msg.String())}
	}

	loader.l.report(&diagnostic.Diagnostic{
		Message: "circular import",
		Primary: primary,
		Explanation: "A circular import occurs when two or more packages import each other, " +
			"directly or indirectly. " +
			"Break the cycle by removing one of the imports.",
	})
	imp.Loaded = true
}
