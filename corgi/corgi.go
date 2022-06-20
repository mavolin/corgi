// Package corgi provides parsing for corgi files.
package corgi

import (
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/file/minify"
	"github.com/mavolin/corgi/corgi/link"
	"github.com/mavolin/corgi/corgi/parse"
	"github.com/mavolin/corgi/corgi/resource"
)

type ParseHelper struct {
	source string
	name   string
	in     string

	ftype file.Type

	rSources []resource.Source
}

// File creates a new *ParseHelper for a file from the given source and name,
// with the passed contents.
func File(source, name string, contents string) *ParseHelper {
	return &ParseHelper{source: source, name: name, in: contents}
}

// WithResourceSource adds the given resource.Source to the ParseHelper.
//
// It returns the ParseHelper itself to allow chaining.
func (h *ParseHelper) WithResourceSource(src resource.Source) *ParseHelper {
	h.rSources = append(h.rSources, src)
	return h
}

// WithFileType sets the file type of the file.
//
// It returns the ParseHelper itself to allow chaining.
func (h *ParseHelper) WithFileType(t file.Type) *ParseHelper {
	h.ftype = t
	return h
}

// Parse parses the file, minifies it, and then links it.
func (h *ParseHelper) Parse() (*file.File, error) {
	p := parse.New(parse.ModeMain, parse.ContextRegular, h.source, h.name, h.in)
	f, err := p.Parse()
	if err != nil {
		return nil, err
	}

	f.Type = h.ftype
	if f.Type == file.TypeUnknown {
		root := f

		for root.Extend != nil {
			root = &root.Extend.File
		}

		f.Type = root.Type
		if f.Type == file.TypeUnknown {
			f.Type = file.TypeHTML
		}
	}

	minify.Minify(f)

	l := link.New(f, parse.ModeMain)

	for _, src := range h.rSources {
		l.AddResourceSource(src)
	}

	if err = l.Link(); err != nil {
		return nil, err
	}

	return f, nil
}
