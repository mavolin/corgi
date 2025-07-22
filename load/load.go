// Package load provides high-level APIs to load corgi packages.
//
// It is the glue between its lower-level subpackages parse, link, and analyze.
package load

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/load/analyze"
	"github.com/mavolin/corgi/v2/load/link"
	"github.com/mavolin/corgi/v2/load/parse"
)

//go:generate go run github.com/mavolin/corgi/v2/tools/codegen/stdlibdetector

const Ext = ".corgi"

type (
	// An importPath is the import path of a package, like
	// "github.com/mavolin/corgi/v2/load".
	importPath = string
	// A filesystemPath is a path to a directory on the filesystem, using the
	// operating system's native path separator.
	filesystemPath = string

	Reader interface {
		// ReadImport reads the package located at the passed import path.
		//
		// It must be concurrency-safe.
		ReadImport(ctx context.Context, path importPath) (*Package, error)
	}

	Package struct {
		// Module is the path/name of the Go module providing this package.
		//
		// It is always specified as a forward slash separated path.
		//
		// The module path needn't necessarily correspond to the import path
		// the reader was invoked with, for example if the import path is
		// symbolic.
		// The most common case of that is the corgi stdlib, which is
		// imported as "corgi/*", but is actually provided by the
		// "github.com/mavolin/corgi/v2" Go module.
		Module string
		// PathInModule is the path to the package in the Go module,
		// relative to the module root.
		//
		// It is always specified as a forward slash separated path.
		//
		// Like Module, this might differ from the import path, if the
		// import path is symbolic.
		// In that case, PathInModule should be the actual path to the
		// directory providing the package, such as "std/fmt" instead of
		// "corgi/fmt".
		PathInModule string

		Files []File
	}

	File struct {
		Name string
		Raw  string
	}

	// ComputeFunc loads the package at the given import path and returns the
	// result.
	//
	// Loading is considered failed if compute returns a non-empty
	// diagnostic.List and/or a non-nil error.
	//
	// A nil return without any errors is valid and indicates that the package
	// contains no corgi files.
	// It is guaranteed, that if a non-nil package is returned, it contains at
	// least one file.
	//
	// Both return values indicate an error has occurred, each with its own
	// purpose:
	//
	// Errors included in the diagnostic.List are errors attached to an AST
	// node, typically errors from the parser/linker/analyzer.
	//
	// The regular error is used for errors outside the AST, such as IO
	// errors.
	// The linker will attach that error, if present, to the respective
	// import spec as part of the linker's returned diagnostics.
	ComputeFunc func(context.Context) (*file.Package, diagnostic.List, error)

	Cache interface {
		// Import returns the cached import for the given import path.
		//
		// If the import isn't cached, it invokes compute and returns its
		// result instead, optionally caching it.
		Import(ctx context.Context, path importPath, compute ComputeFunc) (*file.Package, diagnostic.List, error)
		// Preload informs the cache that the parser found the given import path
		// while parsing and that it subsequently might be used by the linker
		// later.
		//
		// If the cache already contains the import, it is expected to do
		// nothing.
		// If it doesn't, however, the cache may preload the import using the
		// compute function at its own discretion.
		// This means, it is also allowed to ignore the call, for example if
		// it deems the import to be irrelevant.
		// [Load] and friends already filter out stdlib imports, so caches need
		// not implement logic for that.
		//
		// Preload must not block.
		Preload(ctx context.Context, path importPath, compute ComputeFunc)
	}
)

type Options struct {
	// Logger optionally logs the loading process.
	Logger *slog.Logger

	// Cache is an optional cache.
	//
	// Default: NewMemoizer()
	Cache Cache

	// BuiltinPath is the import path of the corgi builtin package.
	//
	// Set to the NoBuiltin constant to disable the builtin package.
	//
	// Default: "corgi/builtin"
	BuiltinPath string
}

const NoBuiltin = "no builtin"

func (o *Options) applyDefaults() {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
	if o.Cache == nil {
		o.Cache = NewMemoizer()
	}
	if o.BuiltinPath == "" {
		o.BuiltinPath = "corgi/builtin"
	} else if o.BuiltinPath == NoBuiltin {
		o.BuiltinPath = ""
	}
}

type loader struct {
	logger  *slog.Logger
	reader  Reader
	cache   Cache
	builtin *file.Package
}

// Load loads the passed package.
//
// A nil return without any errors is valid and indicates that the package
// contains no corgi files.
// Load may also return with a non-nil package with no files, if it was unable
// to ascertain whether the package contains corgi files without reading it.
//
// Load returns two kinds of errors:
// Through the diagnostic.List, errors that are attached to AST nodes;
// And through the regular error, errors that occurred outside of
// parsing/linking/analyzing, such as IO errors.
// If either the list is non-empty or the error is non-nil, Load has failed.
// Beware, however, that Load is capable of recovering from errors and as such,
// may return a non-nil package _and_ a non-empty list of errors.
//
// To see an example of how to use Load, see the [Directory] function.
func Load(ctx context.Context, impPath importPath, r Reader, o Options) (*file.Package, diagnostic.List, error) {
	o.applyDefaults()

	l := &loader{
		logger: o.Logger,
		reader: r,
		cache:  o.Cache,
	}

	logger := o.Logger
	logger.Info("Loading package tree", slog.String("root", impPath))
	defer func(start time.Time) {
		logger.Info("Loaded entire tree", slog.Duration("took", time.Since(start)))
	}(time.Now())

	if o.BuiltinPath != "" {
		var d diagnostic.List
		var err error
		l.builtin, d, err = l.loadCachedImport(ctx, logger, o.BuiltinPath)
		if len(d) > 0 || err != nil {
			return nil, d, err
		}
	}

	return l.loadCachedImport(ctx, logger, impPath)
}

func (l *loader) loadCachedImport(ctx context.Context, logger *slog.Logger, impPath importPath) (*file.Package, diagnostic.List, error) {
	logger = l.logger.With(slog.String("import", impPath))

	var noCacheHit bool
	p, d, err := l.cache.Import(ctx, impPath, func(ctx context.Context) (*file.Package, diagnostic.List, error) {
		logger.Info("Loading package", slog.Bool("cache_hit", false))
		noCacheHit = true
		return l.loadUncachedImport(ctx, logger, impPath)
	})
	if !noCacheHit {
		logger.Info("Loading package", slog.Bool("cache_hit", true))
	}
	return p, d, err
}

func (l *loader) loadUncachedImport(ctx context.Context, logger *slog.Logger, imp importPath) (*file.Package, diagnostic.List, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	data, err := l.reader.ReadImport(ctx, imp)
	if err != nil {
		logger.Error("Reading import", slog.String("err", err.Error()))
		return nil, nil, err
	}

	if len(data.Files) == 0 {
		return nil, nil, nil
	}

	p := &file.Package{
		Module:       data.Module,
		PathInModule: data.PathInModule,
		ImportPath:   imp,
		Files:        make([]*file.File, len(data.Files)),
	}
	if len(p.Files) == 0 {
		return p, nil, nil
	}

	logger = logger.With(
		slog.String("module", p.Module),
		slog.String("path_in_module", p.PathInModule))

	parseErrs := l.parse(ctx, logger, p, data.Files)

	l.buildSymbols(logger, p)
	l.addBuiltinImport(logger, p)

	// the linker can recover from parser errors
	linkErrs := l.link(ctx, logger, p)

	// the analyzer can recover from linker errors, but not from parser errors
	if len(parseErrs) > 0 {
		l.logger.Info("Returning with parse/link errors")
		return p, append(parseErrs, linkErrs...), nil
	}

	analyzeErrs := l.analyze(logger, p)
	if len(linkErrs) > 0 || len(analyzeErrs) > 0 {
		return p, append(linkErrs, analyzeErrs...), nil
	}

	l.logger.Info("Returning without errors")
	return p, nil, nil
}

func (l *loader) parse(ctx context.Context, logger *slog.Logger, p *file.Package, files []File) diagnostic.List {
	logger = logger.WithGroup("parse")
	logger.Info("Parsing package", slog.Int("n_files", len(p.Files)))

	errsChan := make(chan diagnostic.List)

	o := parse.Options{
		Preloader: l.newPreloadHook(ctx, logger),
	}

	for i, fileData := range files {
		go func() {
			logger := logger.With(slog.String("name", fileData.Name))

			logger.Info("Parsing file")
			f, err := parse.Parse(fileData.Raw, o)
			if f == nil {
				f = new(file.File)
			}

			f.Package = p
			f.Name = fileData.Name
			p.Files[i] = f

			if err != nil {
				logger.Error("Completed with errors", slog.String("err", err.Short()))
			} else {
				logger.Info("Done")
			}

			errsChan <- err
		}()
	}

	errs := make(diagnostic.List, 128)
	for range len(files) {
		fileErrs := <-errsChan
		if len(fileErrs) > 0 {
			errs = append(errs, fileErrs...)
		}
	}

	logger.Info("All files parsed")
	return slices.Clip(errs)
}

func (l *loader) buildSymbols(logger *slog.Logger, p *file.Package) {
	logger = logger.WithGroup("symbols")

	logger.Info("Building symbols")
	file.BuildSymbols(p)
	logger.Info("Done building symbols")
}

func (l *loader) addBuiltinImport(logger *slog.Logger, p *file.Package) {
	if l.builtin == nil {
		return
	}

	logger.Info("Adding builtin import")
	for _, f := range p.Files {
		f.AddBuiltinImport(file.BuiltinAlias, l.builtin)
	}
}

func (l *loader) link(ctx context.Context, logger *slog.Logger, p *file.Package) diagnostic.List {
	logger = logger.WithGroup("link")

	logger.Info("Linking package")
	errs := link.Link(ctx, p, link.Options{
		Logger:   logger,
		Importer: l.importHook,
	})
	if len(errs) > 0 {
		logger.Error("Completed with errors", slog.String("err", errs.Short()))
	} else {
		logger.Info("Done linking")
	}

	return errs
}

func (l *loader) analyze(logger *slog.Logger, p *file.Package) diagnostic.List {
	logger = logger.WithGroup("analyze")

	logger.Info("Analyzing package")
	errs := analyze.Analyze(p, analyze.Options{
		Logger: logger,
	})
	if len(errs) > 0 {
		logger.Error("Completed with errors", slog.String("err", errs.Short()))
	} else {
		logger.Info("Done analyzing")
	}

	return errs
}

// newPreloadHook returns a preloader for use alongside a cache.
func (l *loader) newPreloadHook(ctx context.Context, logger *slog.Logger) func(importPath) {
	logger = logger.WithGroup("preload_hook")
	return func(impPath importPath) {
		logger := logger.With(slog.String("import", impPath))
		logger.Info("Received preload request")

		if isStdlib(impPath) {
			logger.Info("Discarding Go stdlib import")
			return
		}

		l.cache.Preload(ctx, impPath, func(ctx context.Context) (*file.Package, diagnostic.List, error) {
			logger.Info("Cache called compute: Preloading package")
			return l.loadUncachedImport(ctx, logger, impPath)
		})
	}
}

func (l *loader) importHook(ctx context.Context, impPath importPath) (*file.Package, diagnostic.List, error) {
	logger := l.logger.
		WithGroup("import_hook").
		With(slog.String("import", impPath))

	if isStdlib(impPath) {
		logger.Info("Immediately returning nil package for stdlib import without loading")
		return nil, nil, nil
	}

	var noCacheHit bool
	p, d, err := l.cache.Import(ctx, impPath, func(ctx context.Context) (*file.Package, diagnostic.List, error) {
		logger.Info("Loading package", slog.Bool("cache_hit", false))
		noCacheHit = true
		return l.loadUncachedImport(ctx, logger, impPath)
	})
	if !noCacheHit {
		logger.Info("Loading package", slog.Bool("cache_hit", true))
	}
	return p, d, err
}
