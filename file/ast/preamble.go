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
	if d.Name != nil {
		if start := d.Name.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (d *PackageDirective) End() Position {
	if d.Name != nil {
		if end := d.Name.End(); end != NoPosition {
			return end
		}
	}
	if d.Package != nil {
		return deltaPos(*d.Package, len("package"))
	}
	return NoPosition
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
	}
	if imp.LParen != nil {
		return *imp.LParen
	}
	for _, spec := range imp.Specs {
		if spec != nil {
			if start := spec.Start(); start != NoPosition {
				return start
			}
		}
	}
	if imp.RParen != nil {
		return *imp.RParen
	}
	return NoPosition
}

func (imp *Import) End() Position {
	if imp.RParen != nil {
		return deltaPos(*imp.RParen, len(")"))
	}
	for _, spec := range slices.Backward(imp.Specs) {
		if spec != nil {
			if end := spec.End(); end != NoPosition {
				return end
			}
		}
	}
	if imp.LParen != nil {
		return deltaPos(*imp.LParen, len("("))
	}
	if imp.Import != nil {
		return deltaPos(*imp.Import, len("import"))
	}
	return NoPosition
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
		if start := s.Alias.Start(); start != NoPosition {
			return start
		}
	}
	if s.Path != nil {
		if start := s.Path.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (s *ImportSpec) End() Position {
	if s.Path != nil {
		if end := s.Path.End(); end != NoPosition {
			return end
		}
	}
	if s.Alias != nil {
		if end := s.Alias.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
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
