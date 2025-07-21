package ast

import "slices"

// ============================================================================
// Package Directive
// ======================================================================================

type PackageDirective struct {
	Package *Position
	Name    *Identifier // package name
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

var (
	_ Node        = (*Import)(nil)
	_ Highlighter = (*Import)(nil)
)

func (imp *Import) Start() Position {
	if imp.Import != nil {
		return *imp.Import
	} else if imp.LParen != nil {
		return *imp.LParen
	}
	for _, spec := range imp.Specs {
		if spec != nil {
			return spec.Start()
		}
	}
	if imp.RParen != nil {
		return *imp.RParen
	}
	return Position{}
}

func (imp *Import) End() Position {
	if imp.RParen != nil {
		return deltaPos(*imp.RParen, len(")"))
	}
	for _, spec := range slices.Backward(imp.Specs) {
		if spec != nil {
			return spec.End()
		}
	}
	if imp.LParen != nil {
		return deltaPos(*imp.LParen, len("("))
	} else if imp.Import != nil {
		return deltaPos(*imp.Import, len("import"))
	}
	return Position{}
}

func (imp *Import) Walk(w func(Node)) {
	for _, spec := range imp.Specs {
		if spec != nil {
			w(spec)
		}
	}
}

func (imp *Import) Highlight() (start, end Position) {
	start, end = imp.Start(), imp.End()
	if imp.LParen == nil && start.Line >= end.Line-3 {
		return start, end
	}
	if imp.Import != nil {
		return *imp.Import, deltaPos(*imp.Import, len("import"))
	}
	return start, end
}

func (*Import) _node() {}

// ============================================================================
// Import Spec
// ======================================================================================

type ImportSpec struct {
	// Alias is the alias of the import, if any.
	Alias *Identifier
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
