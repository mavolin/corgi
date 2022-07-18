// Package link links implements a linker for corgi files.
// It resolves imports and links mixin calls.
// Furthermore, it validates that there are no namespace collisions from uses
// or from redeclared namespaces.
package link

import (
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/link/element"
	"github.com/mavolin/corgi/corgi/link/imports"
	"github.com/mavolin/corgi/corgi/link/mixin"
	"github.com/mavolin/corgi/corgi/link/use"
	"github.com/mavolin/corgi/corgi/parse"
	"github.com/mavolin/corgi/corgi/resource"
)

type Linker struct {
	resourceSources []resource.Source
	resourceFiles   []file.File

	f    *file.File
	mode parse.Mode
}

// New creates a new *Linker that links the given file.
func New(f *file.File, mode parse.Mode) *Linker {
	return &Linker{f: f, mode: mode}
}

// AddResourceSource adds a resource source to the linker.
//
// The linker will use it to find files referenced through extend, use, and
// include directives.
func (l *Linker) AddResourceSource(src resource.Source) {
	l.resourceSources = append(l.resourceSources, src)
}

// Link links the file and performs validation.
func (l *Linker) Link() error {
	fl := newFileLinker(l.f, l.resourceSources...)
	if err := fl.link(); err != nil {
		return err
	}

	ir := imports.NewResolver(l.f)
	if err := ir.Resolve(); err != nil {
		return err
	}

	unc := use.NewNamespaceChecker(*l.f)
	if err := unc.Check(); err != nil {
		return err
	}

	mc := mixin.NewChecker(l.mode, *l.f)
	if err := mc.Check(); err != nil {
		return err
	}

	ml := mixin.NewCallLinker(l.f, l.resourceFiles...)
	if err := ml.Link(); err != nil {
		return err
	}

	mcc := mixin.NewCallChecker(*l.f)
	if err := mcc.Check(); err != nil {
		return err
	}

	ac := element.NewAndChecker(*l.f)
	if err := ac.Check(); err != nil {
		return err
	}

	scc := element.NewSelfClosingChecker(*l.f)
	if err := scc.Check(); err != nil {
		return err
	}

	return nil
}
