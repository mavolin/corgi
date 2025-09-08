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

// BuiltinAlias is the alias used for the builtin package, if it is loaded by
// the linker.
const BuiltinAlias = "__corgi_builtin"

type linker struct {
	p           *file.Package
	logger      *slog.Logger
	importer    Importer
	diagnostics diagnostic.List
	builtinPath importPath

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
		// If not specified, the linker will run in local-only mode, where
		// files must not make any imports.
		//
		// The linker caches loaded packages and guarantees that from the root
		// package down, it will call the Importer at most once per unique
		// import path.
		//
		// The cache is shared through the context passed to the Importer, and
		// also carries information necessary to detect import cycles.
		// It is utmost important that when you call the linker in the importer
		// function, you pass it the context given to the importer function.
		// Otherwise, deadlocks may occur if there are import cycles.
		//
		// Caching can still be beneficial if you call the linker for multiple
		// different root packages that may share imports.
		//
		// If you're not using the high-level load package, but are using the
		// linker directly, refrain from setting the Importer to
		// [github.com/mavolin/corgi/v2/load.Load], as that will not use any
		// caching logic.
		//
		// The linker expects that the Importer respects the deadline of the
		// context passed to it.
		// That means, unless you know that you can complete the import in
		// short, constant time, you should check the context's deadline before
		// trying to load the package.
		//
		// Default: nil
		Importer Importer

		// BuiltinPath is the import path of the builtin package, if any.
		//
		// The package must not contain any explicit imports and must not
		// contain any exported components.
		//
		// If set, the linker will import the specified package as any other
		// explicit import, linking its symbols last in priority.
		// Unlike other imports, the linker will not report errors for
		// collisions between builtin symbols and local symbols.
		//
		// Furthermore, if set, the linker will report an error if the package
		// tries to explicitly import the builtin package.
		//
		// If set, the Importer must be set as well and no file in the package
		// must already have a builtin import.
		BuiltinPath importPath
	}

	Importer func(ctx context.Context, path importPath) (*file.Package, diagnostic.List, error)
)

func (o *Options) applyDefaults() {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}

	if o.BuiltinPath != "" && o.Importer == nil {
		panic("link.Options: BuiltinPath set, but Importer is nil")
	}
}

// Link links the given package, linking all component calls, element
// references, and attribute references, correctly setting [file.Import.Forward]
// on the required imports.
//
// If not already done, the linker will build the package's symbols.
//
// Instead of using [Options.BuiltinPath], you may also add an already loaded
// builtin package to each file's imports, which will then be linked according
// to the same rules as laid out in [Options.BuiltinPath]'s documentation.
// When adding a builtin package like this, you may still choose to run in
// local-only mode.
// Remember that to add a builtin package like this, you need to build the
// package's symbols beforehand.
func Link(ctx context.Context, p *file.Package, o Options) diagnostic.List {
	o.applyDefaults()

	logger := o.Logger.With(
		slog.String("module", p.Module),
		slog.String("path_in_module", p.PathInModule))

	if p.PackageSymbols == nil {
		file.BuildSymbols(p)
	}

	l := &linker{
		p:                      p,
		logger:                 logger,
		importer:               o.Importer,
		diagnostics:            make(diagnostic.List, 0, 128),
		builtinPath:            o.BuiltinPath,
		reportedMissingImports: make(map[*file.File]map[namespace]bool, len(p.Files)),
		dotImports:             make(map[*file.File][]*file.Import, len(p.Files)),
	}
	for _, f := range p.Files {
		l.reportedMissingImports[f] = make(map[namespace]bool)
		if imps := filterDotImports(f); len(imps) > 0 {
			l.dotImports[f] = imps
		}
	}

	l.CheckSelfImport()
	l.CheckImportCycles(ctx)
	ctx = addToImportersGraph(ctx, p)
	l.CheckDuplicateDotImports()
	l.LoadImports(ctx)
	l.CheckExplicitBuiltinImport()
	l.CheckImportNamespaceCollisions()
	l.CheckDotImportComponentCollisions()
	l.CheckDotImportElementSpecCollisions()
	l.CheckDotImportAttributeSpecCollisions()

	l.CheckComponentCollisions()
	l.LinkComponentCalls()

	l.CheckElementSpecCollisions()
	l.LinkElementReferences()

	l.CheckAttributeSpecCollisions()
	l.CheckAttributeRuleCollisions()
	l.LinkAttributeReferences()

	p.Linked = true
	for _, f := range p.Files {
		f.Linked = true
	}

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
	seen := make(map[importPath]bool, len(f.Imports))
	for _, imp := range f.Imports {
		if imp.Explicit() && imp.Alias == "." && !seen[imp.CorgiPath] {
			imps = append(imps, imp)
			seen[imp.CorgiPath] = true
		}
	}
	return imps
}

type importersGraphKey struct{}

func addToImportersGraph(ctx context.Context, p *file.Package) context.Context {
	importers, _ := ctx.Value(importersGraphKey{}).([]*file.Package)

	clone := make([]*file.Package, len(importers)+1)
	copy(clone, importers)
	clone[len(importers)] = p
	return context.WithValue(ctx, importersGraphKey{}, clone)
}

// importersGraph returns the linear graph outlining the importers of a package,
// the last node being the package itself and the first node being the root
// package.
func importersGraph(ctx context.Context) []*file.Package {
	importers, _ := ctx.Value(importersGraphKey{}).([]*file.Package)
	return importers
}
