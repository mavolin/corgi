package link

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/sourcegraph/conc"
)

func (l *linker) LoadImports(ctx context.Context) {
	(&importLoader{}).load(ctx, l, l.logger)
}

type importLoader struct {
	reportMut sync.Mutex
}

func (loader *importLoader) load(ctx context.Context, l *linker, logger *slog.Logger) {
	logger = logger.WithGroup("import_loader")
	logger.Info("Loading imports")

	if l.importer == nil {
		loader.localOnlyModeCheck(l, logger)
		return
	}

	loader.loadImports(ctx, l, logger)
}

func (loader *importLoader) localOnlyModeCheck(l *linker, logger *slog.Logger) {
	logger.Info("Running in local-only mode, not allowed to use imports")

	logger.Debug("Checking for imports in package")
	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Checking file")

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
			l.report(&diagnostic.Diagnostic{
				Message:     "local-only mode: file contains imports",
				Primary:     primaries,
				Explanation: "In local-only mode, files are not allowed to make any imports.",
			})
		}
	}
}

func (loader *importLoader) loadImports(ctx context.Context, l *linker, logger *slog.Logger) {
	logger.Info("Concurrently loading imports")

	var wg conc.WaitGroup

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Loading imports for file")

		for _, imp := range f.Imports {
			if !imp.Explicit() || imp.Path == "" {
				continue
			}

			logger := logger.With(
				slog.String("import", imp.Path),
				slog.String("import_pos", imp.AST.Start().String()))
			logger.Debug("Loading import")

			wg.Go(func() {
				loader.loadImport(ctx, l, logger, f, imp)
			})
		}
	}

	logger.Info("Waiting for all goroutines to finish")
	wg.Wait()
	logger.Info("Finished loading imports")
}

func (loader *importLoader) loadImport(ctx context.Context, l *linker, logger *slog.Logger, f *file.File, imp *file.Import) {
	logger = logger.With(
		slog.String("import", imp.Path),
		slog.String("import_pos", imp.AST.Start().String()))
	logger.Info("Loading import")

	var d diagnostic.List
	var err error
	imp.Package, d, err = l.importer(ctx, imp.Path)
	if len(d) > 0 || err != nil {
		imp.LoadedWithErrors = true

		loader.reportMut.Lock()

		if err != nil {
			logger.Error("Failed to load import", slog.String("error", err.Error()))
			l.report(&diagnostic.Diagnostic{
				Message: "import: failed to load package",
				Cause:   err,
				Primary: []diagnostic.Annotation{
					anno.Node(f, imp.AST, "failed to load import"),
				},
			})
		}
		if len(d) > 0 {
			l.report(d...)
			logger.Error("Import contains errors", slog.String("err", d.Short()))
		}

		loader.reportMut.Unlock()
	} else {
		logger.Info("Successfully loaded import")
	}
}
