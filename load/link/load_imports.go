package link

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/sourcegraph/conc"
)

func (l *linker) LoadImports(ctx context.Context) {
	(&importLoader{
		marked: make(map[importPath]*markedImport),
	}).load(ctx, l, l.logger)
}

type (
	importLoader struct {
		marked map[importPath]*markedImport
	}

	markedImport struct {
		p         *file.Package
		dotImport bool
	}
)

func (loader *importLoader) load(ctx context.Context, l *linker, logger *slog.Logger) {
	logger = l.logger.WithGroup("import_loader")
	logger.Info("Loading imports")

	loader.collectImports(l, logger)
	if !loader.localOnlyModeCheck(l, logger) {
		return
	}

	loader.loadImports(ctx, l, logger)
	loader.setImports(l, logger)
}

// ============================================================================
// Collect
// ======================================================================================

func (loader *importLoader) collectImports(l *linker, logger *slog.Logger) {
	logger = logger.WithGroup("collector")
	logger.Info("Collecting imports that need to be loaded")
	for _, f := range l.p.Files {
		loader.collectImportsFromFile(l, logger, f)
	}

	if l.builtin != nil {
		// In case the builtin package was illegally explicitly imported,
		// unmark it for loading.
		loader.marked[l.builtin.ImportPath] = nil
	}
}

func (loader *importLoader) collectImportsFromFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.With(slog.String("file", f.Name))
	logger.Debug("Checking file")

	loader.collectImportsForComponentCalls(l, logger, f)
	loader.collectImportsForElementReferences(l, logger, f)
	loader.collectImportsForAttributeReferences(l, logger, f)
}

func (loader *importLoader) collectImportsForComponentCalls(_ *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("component_calls")
	logger.Debug("Processing component calls")

	var needDotImports bool
	for _, cc := range f.ComponentCalls {
		if cc == nil || cc.AST.Header == nil || cc.AST.Header.Name == nil {
			continue
		}

		logger := logger.With(slog.String("pos", cc.AST.Start().String()))

		switch ident := cc.AST.Header.Name.(type) {
		case *ast.Ident:
			if ident != nil {
				logger.Debug("Processing component call", slog.String("name", ident.Ident))
				needDotImports = needDotImports || file.IsExported(ident.Ident)
			}
		case *ast.QualifiedIdent:
			if ident.Package != nil {
				logger := logger.With(slog.String("pos", cc.AST.Start().String()))
				if ident.Name != nil {
					logger = logger.With(slog.String("name", ident.Package.Ident+"."+ident.Name.Ident))
					logger.Debug("Processing component call")
				} else {
					logger = logger.With(slog.String("name", ident.Package.Ident+"."))
					logger.Warn("Processing component call with malformed qualified ident: package readable, marking for loading regardless")
				}

				imp := f.ImportByNamespace(ident.Package.Ident)
				if imp == nil {
					logger.Warn("Component call references unknown package", slog.String("package", ident.Package.Ident))
					continue
				}

				if impPath := imp.ImportPath(); impPath != "" {
					logger.Debug("Marking package for loading", slog.String("import", impPath))
					loader.markImport(impPath, false)
				} else {
					logger.Warn("Have no import path for package, due to syntax error", slog.String("package", ident.Package.Ident))
				}
			}
		}
	}
	if needDotImports {
		logger.Info("Found at least one non-qualified call to an exported component, loading dot imports, if there are any")
		for _, imp := range f.Symbols.Imports {
			if imp.AST != nil && imp.AST.Alias != nil && imp.AST.Alias.Ident == "." {
				logger.Debug("Marking package for loading", slog.String("import", imp.ImportPath()))
				loader.markImport(imp.ImportPath(), true)
			}
		}
	}
}

func (loader *importLoader) collectImportsForElementReferences(_ *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("element_references")
	logger.Debug("Processing element references")

	var needDotImports bool
	for _, ref := range f.ElementReferences {
		if ref == nil {
			continue
		} else if ref.AST.Package == nil {
			logger.Debug("Processing element reference", slog.String("element", ref.AST.Name.Name))
			needDotImports = true
			continue
		}

		logger := logger.With(slog.String("pos", ref.AST.Start().String()))
		if ref.AST.Name != nil {
			logger = logger.With(slog.String("element", ref.AST.Package.Ident+"."+ref.AST.Name.Name))
			logger.Debug("Processing element reference")
		} else {
			logger = logger.With(slog.String("element", ref.AST.Package.Ident+"."))
			logger.Warn("Processing element reference with malformed qualified ident: package readable, marking for loading regardless")
		}

		imp := f.ImportByNamespace(ref.AST.Package.Ident)
		if imp == nil {
			logger.Warn("Element reference references unknown package", slog.String("package", ref.AST.Package.Ident))
			continue
		}

		if p := imp.ImportPath(); p != "" {
			logger.Debug("Marking package for loading", slog.String("import", p))
			loader.markImport(p, false)
		} else {
			logger.Warn("Have no import path for package, due to syntax error", slog.String("package", ref.AST.Package.Ident))
		}
	}
	if needDotImports {
		logger.Info("Found at least one non-qualified reference to an element, loading dot imports, if there are any")
		for _, imp := range f.Symbols.Imports {
			if imp.AST != nil && imp.AST.Alias != nil && imp.AST.Alias.Ident == "." {
				logger.Debug("Marking package for loading", slog.String("import", imp.ImportPath()))
				loader.markImport(imp.ImportPath(), true)
			}
		}
	}
}

func (loader *importLoader) collectImportsForAttributeReferences(_ *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("attribute_references")
	logger.Debug("Processing attribute references")

	var needDotImports bool
	for _, ref := range f.AttributeReferences {
		if ref == nil {
			continue
		} else if ref.AST.Package == nil {
			logger.Debug("Processing attribute reference", slog.String("attribute", ref.AST.Name.Name))
			needDotImports = true
			continue
		}

		needDotImports = needDotImports || ref.AST.Name != nil

		logger := logger.With(slog.String("pos", ref.AST.Start().String()))
		if ref.AST.Name != nil {
			logger = logger.With(slog.String("attribute", ref.AST.Package.Ident+"."+ref.AST.Name.Name))
			logger.Debug("Processing attribute reference")
		} else {
			logger = logger.With(slog.String("attribute", ref.AST.Package.Ident+"."))
			logger.Warn("Processing attribute reference with malformed qualified ident: package readable, marking for loading regardless")
		}

		imp := f.ImportByNamespace(ref.AST.Package.Ident)
		if imp == nil {
			logger.Warn("Attribute reference references unknown package", slog.String("package", ref.AST.Package.Ident))
			continue
		}

		if p := imp.ImportPath(); p != "" {
			logger.Debug("Marking package for loading", slog.String("import", p))
			loader.markImport(p, false)
		} else {
			logger.Warn("Have no import path for package, due to syntax error", slog.String("package", ref.AST.Package.Ident))
		}
	}
	if needDotImports {
		logger.Info("Found at least one non-qualified reference to an attribute, loading dot imports, if there are any")
		for _, imp := range f.Symbols.Imports {
			if imp.AST != nil && imp.AST.Alias != nil && imp.AST.Alias.Ident == "." {
				logger.Debug("Marking package for loading", slog.String("import", imp.ImportPath()))
				loader.markImport(imp.ImportPath(), true)
			}
		}
	}
}

// ============================================================================
// Load
// ======================================================================================

func (loader *importLoader) localOnlyModeCheck(l *linker, logger *slog.Logger) (ok bool) {
	if l.importer != nil {
		return true
	}

	logger.Info("Running in local-only mode, not loading imports")

	ok = true
	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))

		for _, imp := range f.Symbols.Imports {
			if imp == nil {
				continue
			}

			mark := loader.marked[imp.ImportPath()]
			if mark == nil {
				continue
			}

			logger := logger.With(slog.String("import", imp.ImportPath()))
			logger.Error("Cannot load import: running in local-only mode")

			if mark.dotImport {
				l.report(&diagnostic.Diagnostic{
					Message: "cannot load dot import: running in local-only mode",
					Primary: []diagnostic.Annotation{anno.Node(f, imp.AST.Path, "imported here")},
					Explanation: "The use of dot imports is not supported in local-only mode.\n" +
						"Although the linker might be able to fully link the file without loading the import, " +
						"it won't be able to detect collisions with local symbols that would yield an error had " +
						"the local-only mode not been enabled.",
					Hints: []diagnostic.Hint{
						{
							Hint:    "Remove the dot import in favor of a qualified import.",
							Example: "`. \"woof\"` -> `\"woof\"` or `bark \"woof\"`",
						},
					},
				})
			} else {
				l.report(&diagnostic.Diagnostic{
					Message: "cannot load import: running in local-only mode",
					Primary: []diagnostic.Annotation{anno.Node(f, imp.AST.Path, "imported here")},
					Explanation: "This import could not be loaded, as the linker is running in local-only mode, " +
						"but needs to be loaded, as it is used in the file by at least one component call, " +
						"element reference, or attribute reference.",
				})
			}

			ok = false
		}
	}

	return ok
}

func (loader *importLoader) loadImports(ctx context.Context, l *linker, logger *slog.Logger) {
	logger.Info("Concurrently loading imports", slog.Int("n_imports", len(loader.marked)))

	var mut sync.Mutex
	var wg conc.WaitGroup

	for impPath := range loader.marked {
		wg.Go(func() {
			pkg, dl := loader.loadImport(ctx, l, logger, impPath)
			mut.Lock()
			loader.marked[impPath].p = pkg
			l.report(dl...)
			mut.Unlock()
		})
	}

	logger.Info("Waiting for all goroutines to finish")
	wg.Wait()
	logger.Info("Finished loading imports")
}

func (loader *importLoader) loadImport(ctx context.Context, l *linker, logger *slog.Logger, impPath importPath) (*file.Package, diagnostic.List) {
	logger = logger.With(slog.String("import", impPath))
	logger.Info("Loading import")
	pkg, dl := l.importer(ctx, impPath)

	if dl != nil {
		logger.Error("Failed to load import", slog.String("err", dl.Short()))
	} else {
		logger.Info("Successfully loaded import")
	}

	return pkg, dl
}

func (loader *importLoader) setImports(l *linker, logger *slog.Logger) {
	logger.Debug("Setting loaded packages")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Setting file")

		for _, imp := range f.Symbols.Imports {
			p := imp.ImportPath()
			if p != "" {
				logger := logger.With(slog.String("import", p))
				mark := loader.marked[p]
				if mark == nil || mark.p == nil {
					logger.Debug("Import not loaded, setting to nil", slog.Bool("marked", mark != nil))
					imp.Package = nil
				} else {
					logger.Debug("Setting import")
					imp.Package = mark.p
				}
			}
		}
	}
}

// ============================================================================
// Helpers
// ======================================================================================

func (loader *importLoader) markImport(p importPath, dotImport bool) {
	if imp := loader.marked[p]; imp != nil {
		if !dotImport {
			imp.dotImport = false
		}
		return
	}

	loader.marked[p] = &markedImport{dotImport: dotImport}
}
