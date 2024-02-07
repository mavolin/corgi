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
	imports := collectImports(p)
	duplErr := checkDuplicateImports(p)
	compErr := collectComponents(p)
	collectComponentCalls(p)

	linkLocalComponentCalls(p)
	impErr := loadImports(p, l.importer, imports)
	linkErr := linkExternalComponentCalls(p)

	return fileerr.Join(duplErr, impErr, compErr, linkErr)
}
