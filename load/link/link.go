// Package link links implements a linker for corgi files.
// It resolves imports and links component calls.
package link

import (
	"context"
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

type linker struct {
	p           *file.Package
	logger      *slog.Logger
	importer    Importer
	diagnostics diagnostic.List

	reportedMissingImports map[*file.File]map[namespace]bool
	dotImports             map[*file.File][]*file.Import // file -> dot imports
}

type (
	namespace             = string
	importPath            = string
	attributeSelector     = string
	fullAttributeSelector = string
	componentName         = string
	elementName           = string
	fullElementName       = string
)

type (
	Options struct {
		// Logger is the logger used by the linker.
		//
		// Default: no logging
		Logger *slog.Logger

		// Importer loads imports.
		// If not specified, the linker will run in local-only mode, where only
		// files must not make any imports.
		// This does not affect the use of built-in package.
		//
		// The linker does not cache results of the Importer on its own.
		// In a package where n files import the same package, the linker will
		// call the importer n times for that package.
		// It is highly recommend to implement some sort of caching logic.
		//
		// If you're not using the high-level load package, but are using the
		// linker directly, refrain from setting the Importer to
		// [github.com/mavolin/corgi/v2/load.Load], as that will not use any
		// caching logic.
		//
		// Default: nil
		Importer Importer
	}

	Importer func(ctx context.Context, path importPath) (*file.Package, diagnostic.List, error)
)

func (o *Options) applyDefaults() {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
}

// Link links the given package, linking all component calls, element
// references, and attribute references.
//
// It loads the minimal set of imports required to link the package and detect
// all collisions of corgi symbols.
//
// If you wish to include symbols from a corgi standard library, you should
// call [file.Symbols.AddBuiltinImport] before calling this function.
// The builtin package must not contain any exported symbols.
func Link(ctx context.Context, p *file.Package, o Options) diagnostic.List {
	o.applyDefaults()

	logger := o.Logger.With(
		slog.String("module", p.Module),
		slog.String("path_in_module", p.PathInModule))

	file.BuildSymbols(p)

	l := &linker{
		p:                      p,
		logger:                 logger,
		importer:               o.Importer,
		diagnostics:            make(diagnostic.List, 0, 128),
		reportedMissingImports: make(map[*file.File]map[namespace]bool, len(p.Files)),
	}
	for _, f := range p.Files {
		l.reportedMissingImports[f] = make(map[namespace]bool)
		if imps := filterDotImports(f); len(imps) > 0 {
			l.dotImports[f] = imps
		}
	}

	l.CheckImportCycles(ctx)
	ctx = addToImportersGraph(ctx, p)
	l.CheckDuplicateDotImports()
	l.CheckExplicitBuiltinImport()
	l.LoadImports(ctx)
	l.CheckImportNamespaceCollisions()
	l.CheckDotImportComponentCollisions()
	l.CheckDotImportElementSpecCollisions()
	l.CheckDotImportAttributeSpecCollisions()

	l.CheckComponentsCollisions()
	l.LinkComponentCalls()

	l.CheckAttributeSpecCollisions()
	l.CheckAttributeRuleCollisions()
	l.CheckAttributeRuleCollisions()
	l.LinkAttributeReferences()

	l.CheckElementSpecCollisions()
	l.LinkElementReferences()

	if len(l.diagnostics) > 0 {
		return slices.Clip(l.diagnostics)
	}
	return nil
}

func (l *linker) report(d ...*diagnostic.Diagnostic) {
	l.diagnostics = append(l.diagnostics, d...)
}

func (l *linker) reportMissingImport(f *file.File, namespace string, d *diagnostic.Diagnostic) {
	if l.reportedMissingImports[f][namespace] {
		return
	}

	l.reportedMissingImports[f][namespace] = true
	l.report(d)
}

func filterDotImports(f *file.File) []*file.Import {
	imps := make([]*file.Import, 0, 8)
	for _, imp := range f.Imports {
		if imp.Explicit() && imp.Alias != "." && imp.Package != nil {
			imps = append(imps, imp)
		}
	}
	return imps
}

type importersGraphKey struct{}

func addToImportersGraph(ctx context.Context, p *file.Package) context.Context {
	importers, _ := ctx.Value(importersGraphKey{}).([]*file.Package)
	return context.WithValue(ctx, importersGraphKey{}, append(importers, p))
}

// importersGraph returns the linear graph outlining the importers of a package,
// the last node being the package itself and the first node being the root
// package.
func importersGraph(ctx context.Context) []*file.Package {
	importers, _ := ctx.Value(importersGraphKey{}).([]*file.Package)
	return importers
}
