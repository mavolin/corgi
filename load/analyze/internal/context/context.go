package context

import (
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

type Context struct {
	P           *file.Package
	Logger      *slog.Logger
	diagnostics diagnostic.List
}

func New(p *file.Package, logger *slog.Logger) *Context {
	return &Context{
		P:           p,
		Logger:      logger,
		diagnostics: make(diagnostic.List, 0, 32),
	}
}

// SafeImport returns the import for the safe package for the given file, or
// adds it if it does not exist yet.
func (ctx *Context) SafeImport(f *file.File) *file.Import {
	if imp := f.ImportByPath(file.SafeImport); imp != nil {
		return imp
	}

	alias := "__corgi_safe"
	for f.ImportByNamespace(alias) != nil {
		alias += "_"
	}

	imp := &file.Import{
		Alias:     alias,
		Path:      file.SafeImport,
		Namespace: alias,
		Forward:   true,
	}
	f.AddImport(imp)
	return imp
}

func (ctx *Context) Report(d *diagnostic.Diagnostic) {
	ctx.diagnostics = append(ctx.diagnostics, d)
}

func (ctx *Context) Diagnostics() diagnostic.List {
	if len(ctx.diagnostics) == 0 {
		return nil
	}
	return slices.Clip(ctx.diagnostics)
}
