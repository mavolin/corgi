// Package link links implements a linker for corgi files.
// It resolves imports and links component calls.
package link

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileerr"
)

type Linker struct {
	importer Importer
}

type Importer func(path string) (*file.Package, error)

// New creates a new *Linker that uses the passed load.
func New(imp Importer) *Linker {
	return &Linker{importer: imp}
}

// Link concurrently links the passed package, filling the p's Components,
// ComponentCalls, and each of p's File's Imports.
//
// It only sets File, Source, and Component fields of the
// Component/ComponentCall.
//
// The returned error is always of type [fileerr.List].
func (l *Linker) Link(p *file.Package) error {
	ctx := &context{errs: make(fileerr.List, 0, 128)}

	imports := collectImports(ctx, p)
	checkDuplicateImports(ctx, p)
	collectComponents(ctx, p)
	collectComponentCalls(ctx, p)

	linkLocalComponentCalls(ctx, p)
	loadImports(ctx, p, l.importer, imports)
	linkExternalComponentCalls(ctx, p)

	return ctx.error()
}

type context struct {
	errs fileerr.List
}

func (ctx *context) err(err *fileerr.Error) {
	ctx.errs = append(ctx.errs, err)
}

func (ctx *context) error() error {
	return ctx.errs.AsError()
}
