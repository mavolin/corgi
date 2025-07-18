package context

import (
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/set"
)

type Context struct {
	P           *file.Package
	Logger      *slog.Logger
	diagnostics diagnostic.List
	stringSet   set.Set[string]
}

func New(p *file.Package, logger *slog.Logger) *Context {
	return &Context{
		P:           p,
		Logger:      logger,
		diagnostics: make(diagnostic.List, 0, 32),
		stringSet:   set.NewSliceSet[string](32),
	}
}

func (ctx *Context) TakeStringSet() set.Set[string] {
	ctx.stringSet.Clear()
	return ctx.stringSet
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
