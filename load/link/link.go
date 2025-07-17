// Package link links implements a linker for corgi files.
// It resolves imports and links component calls.
package link

import (
	"context"
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/set"
)

type linker struct {
	p           *file.Package
	logger      *slog.Logger
	importer    Importer
	builtin     *file.Package
	diagnostics diagnostic.List

	stringSet *set.SliceSet[string]

	reportedMissingImports map[*file.File]*set.SliceSet[importPath]
}

type (
	importPath      = string
	componentName   = string
	elementName     = string
	fullElementName = string
)

type (
	Options struct {
		// Logger is the logger used by the linker.
		//
		// Default: no logging
		Logger *slog.Logger

		// Importer loads imports.
		// If not specified, the linker will run in local-only mode, where only
		// local component calls are allowed.
		// For every external component call, the linker will return an error.
		//
		// Default: nil
		Importer Importer

		// Builtin is the builtin package of the corgi standard library.
		//
		// All symbols in this package are always available, even if not
		// imported, but can be shadowed by local or dot-imported symbols.
		//
		// If a builtin package is specified, the linker will report an error
		// if that same package is explicitly imported.
		//
		// Default: nil
		Builtin *file.Package
	}

	Importer func(ctx context.Context, path importPath) (*file.Package, diagnostic.List)
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
		builtin:                o.Builtin,
		diagnostics:            make(diagnostic.List, 0, 128),
		stringSet:              set.NewSliceSet[importPath](32),
		reportedMissingImports: make(map[*file.File]*set.SliceSet[importPath], len(p.Files)),
	}
	for _, f := range p.Files {
		l.reportedMissingImports[f] = set.NewSliceSet[importPath](len(f.Symbols.Imports))
	}

	l.CheckImportCycles(ctx)
	ctx = addToImportersGraph(ctx, p)
	l.CheckImportNamespaceCollisions(ctx)
	l.CheckDuplicateDotImports(ctx)
	l.CheckExplicitBuiltinImport(ctx)
	l.LoadImports(ctx)
	l.CheckDotImportCollisions(ctx)
	l.CheckLocalDotImportCollisions(ctx)

	l.CheckDuplicateComponents(ctx)
	l.LinkComponentCalls(ctx)

	l.CheckDuplicateAttributeDefinitions(ctx)
	l.CheckDuplicateAttributeDefinitionElementTypes(ctx)
	l.CheckDuplicateAttributeDefinitionElementSelectors(ctx)
	l.LinkAttributeReferences(ctx)

	l.CheckDuplicateElementDefinitions(ctx)
	l.LinkElementReferences(ctx)

	if len(l.diagnostics) > 0 {
		return slices.Clip(l.diagnostics)
	}
	return nil
}

func (l *linker) report(d ...*diagnostic.Diagnostic) {
	l.diagnostics = append(l.diagnostics, d...)
}

func (l *linker) takeStringSet() *set.SliceSet[string] {
	l.stringSet.Clear()
	return l.stringSet
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
