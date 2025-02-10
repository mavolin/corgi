package ast

import "slices"

// ============================================================================
// Package Directive
// ======================================================================================

type PackageDirective struct {
	Package *Position
	Name    *Ident // package name
}

var _ Node = (*PackageDirective)(nil)

func (d *PackageDirective) Start() Position {
	if d.Package != nil {
		return *d.Package
	}
	return d.Name.Start()
}
func (d *PackageDirective) End() Position {
	if d.Name != nil {
		return d.Name.End()
	} else if d.Package != nil {
		return deltaPos(*d.Package, len("package"))
	}
	return Position{}
}
func (d *PackageDirective) Walk(w func(Node)) {
	if d.Name != nil {
		w(d.Name)
	}
}

func (*PackageDirective) _node() {}

// ============================================================================
// Import
// ======================================================================================

type Import struct {
	Import *Position
	LParen *Position // nil if this is a single-line import
	Specs  []*ImportSpec
	RParen *Position // nil if this is a single-line import
}

var _ Node = (*Import)(nil)

func (i *Import) Start() Position {
	if i.Import != nil {
		return *i.Import
	} else if i.LParen != nil {
		return *i.LParen
	}
	for _, spec := range i.Specs {
		if spec != nil {
			return spec.Start()
		}
	}
	if i.RParen != nil {
		return *i.RParen
	}
	return Position{}
}
func (i *Import) End() Position {
	if i.RParen != nil {
		return deltaPos(*i.RParen, len(")"))
	}
	for _, spec := range slices.Backward(i.Specs) {
		if spec != nil {
			return spec.End()
		}
	}
	if i.LParen != nil {
		return deltaPos(*i.LParen, len("("))
	} else if i.Import != nil {
		return deltaPos(*i.Import, len("import"))
	}
	return Position{}
}
func (i *Import) Walk(w func(Node)) {
	for _, spec := range i.Specs {
		if spec != nil {
			w(spec)
		}
	}
}

func (*Import) _node() {}

// ============================================================================
// Import Spec
// ======================================================================================

type ImportSpec struct {
	// Alias is the alias of the import, if any.
	Alias *Ident
	Path  *StaticString
}

var _ Node = (*ImportSpec)(nil)

func (s *ImportSpec) Start() Position {
	if s.Alias != nil {
		return s.Alias.Start()
	} else if s.Path != nil {
		return s.Path.Start()
	}
	return Position{}
}
func (s *ImportSpec) End() Position {
	if s.Path != nil {
		return s.Path.End()
	} else if s.Alias != nil {
		return s.Alias.End()
	}
	return Position{}
}
func (s *ImportSpec) Walk(w func(Node)) {
	if s.Alias != nil {
		w(s.Alias)
	}
	if s.Path != nil {
		w(s.Path)
	}
}

func (*ImportSpec) _node() {}
