package link

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/sourcegraph/conc"
)

func (l *linker) LoadImports(ctx context.Context) {
	(&importLoader{
		l:      l,
		logger: l.logger.WithGroup("imports"),
	}).load(ctx)
}

type importLoader struct {
	l         *linker
	logger    *slog.Logger
	reportMut sync.Mutex
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
		logger := loader.logger.With(slog.String("file", f.Name))

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
		logger := loader.logger.With(slog.String("file", f.Name))
		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.Alias == "" || !strings.HasPrefix(imp.Alias, "__corgi_") {
				continue
			}

			logger.Error("Import alias with reserved prefix",
				slog.String("alias", imp.Alias),
				slog.String("import_path", imp.Path))
			loader.l.report(&diagnostic.Diagnostic{
				Message: "import alias: cannot use `__corgi_` prefix",
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST.Alias, "illegal import alias"),
				},
				Explanation: "All import aliases starting with `__corgi_` are reserved for internal use by corgi.",
			})
		}
	}
}

func (loader *importLoader) loadImports(ctx context.Context) {
	loader.logger.Info("Concurrently loading imports")

	var wg conc.WaitGroup

	for _, f := range loader.l.p.Files {
		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.Path == "" || imp.LoadedWithErrors {
				continue
			}

			wg.Go(func() {
				loader.loadImport(ctx, f, imp)
			})
		}
	}

	builtin, builtinDiagnostics, builtinErr := loader.loadBuiltin(ctx)
	loader.logger.Debug("Waiting for all goroutines to finish")
	wg.Wait()
	if loader.l.builtinPath != "" {
		loader.setBuiltinImport(builtin, builtinDiagnostics, builtinErr)
	}
	loader.logger.Info("Finished loading imports")
}

func (loader *importLoader) loadImport(ctx context.Context, f *file.File, imp *file.Import) {
	logger := loader.logger.With(
		slog.String("file", f.Name),
		slog.String("import", imp.Path),
		slog.String("import_pos", imp.AST.Start().String()))
	logger.Debug("Loading import")

	var d diagnostic.List
	var err error
	imp.Package, d, err = loader.l.importer(ctx, imp.Path)
	if len(d) > 0 || err != nil {
		imp.LoadedWithErrors = true

		loader.reportMut.Lock()

		if err != nil {
			logger.Error("Failed to load import", slog.String("err", err.Error()))
			loader.l.report(&diagnostic.Diagnostic{
				Message: "import: failed to load package",
				Cause:   err,
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST, "failed to load import"),
				},
			})
		}
		if len(d) > 0 {
			loader.l.report(d...)
			logger.Error("Import contains errors", slog.String("err", d.Short()))
		}
		loader.reportMut.Unlock()
	}

	logger.Debug("Successfully loaded import")

	if imp.Package == nil {
		return
	}

	switch {
	case imp.Alias == ".":
		imp.Namespace = ""
	case imp.Alias != "":
		imp.Namespace = imp.Alias
	default:
		imp.Namespace = imp.Package.Name
		if strings.HasPrefix(imp.Namespace, "__corgi_") {
			logger.Error("Import has package name with reserved prefix",
				slog.String("namespace", imp.Namespace),
				slog.String("import_path", imp.Path))

			loader.reportMut.Lock()
			loader.l.report(&diagnostic.Diagnostic{
				Message: "import: import uses reserved `__corgi_` package name prefix",
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST, "illegal package name"),
				},
				Explanation: "All namespaces starting with `__corgi_` are reserved for internal use by corgi.",
				Hints: []diagnostic.Hint{
					{Hint: "Use an import alias."},
				},
			})
			loader.reportMut.Unlock()
		}
	}
}

func (loader *importLoader) loadBuiltin(ctx context.Context) (*file.Package, diagnostic.List, error) {
	if loader.l.builtinPath == "" {
		return nil, nil, nil // no builtin path set, nothing to load
	}
	logger := loader.logger.With(slog.String("import", loader.l.builtinPath))

	logger.Debug("Loading builtin import in current goroutine")

	p, d, err := loader.l.importer(ctx, loader.l.builtinPath)
	switch {
	case err != nil:
		logger.Error("Failed to load builtin import", slog.String("err", err.Error()))
	case len(d) > 0:
		logger.Error("Builtin import contains errors", slog.String("err", d.Short()))
	default:
		logger.Debug("Successfully loaded builtin import")
	}

	return p, d, err
}

func (loader *importLoader) setBuiltinImport(p *file.Package, d diagnostic.List, err error) {
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
			f.AddBuiltinImport(BuiltinAlias, p)
			if len(d) > 0 || err != nil {
				f.BuiltinImport().LoadedWithErrors = true
			}
		}
	}
}
