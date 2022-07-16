// Package link links implements a linker for corgi files.
// It resolves imports and links mixin calls.
// Furthermore, it validates that there are no namespace collisions from uses
// or from redeclared namespaces.
package link

import (
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/parse"
	"github.com/mavolin/corgi/corgi/resource"
)

type Linker struct {
	rSources []resource.Source
	rFiles   []file.File

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
	l.rSources = append(l.rSources, src)
}

// Link links the file and performs validation.
func (l *Linker) Link() error {
	if err := l.linkFile(); err != nil {
		return err
	}

	if err := l.resolveImports(); err != nil {
		return err
	}

	if err := l.checkNamespaceCollisions(); err != nil {
		return err
	}

	if err := l.checkMixins(); err != nil {
		return err
	}

	if err := l.linkMixinCalls(); err != nil {
		return err
	}

	if err := l.checkElements(); err != nil {
		return err
	}

	if err := l.checkExtendBlocks(); err != nil {
		return err
	}

	return nil
}
