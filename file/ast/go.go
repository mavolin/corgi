package ast

// Types representing their Go counterparts.

// ============================================================================
// Name
// ======================================================================================

// FullIdentifier is a pointer to either an [Identifier] or a [QualifiedIdentifier].
//
// Not actually in the Go spec, but for our convenience.
type FullIdentifier interface {
	Node
	Full() string
	_fullIdent()
}

// if this is changed, change the comment above
var (
	_ FullIdentifier = (*Identifier)(nil)
	_ FullIdentifier = (*QualifiedIdentifier)(nil)
)

// ============================================================================
// Name
// ======================================================================================

// Identifier is a Go identifier.
type Identifier struct {
	Name     string
	Position *Position
}

var _ FullIdentifier = (*Identifier)(nil)

func (ident *Identifier) Start() Position {
	if ident.Position != nil {
		return *ident.Position
	}
	return NoPosition
}

func (ident *Identifier) End() Position {
	if ident.Position != nil {
		return deltaPos(*ident.Position, len([]rune(ident.Name)))
	}
	return NoPosition
}
func (ident *Identifier) Walk(func(Node)) {}
func (ident *Identifier) Full() string    { return ident.Name }

func (*Identifier) _node()      {}
func (*Identifier) _fullIdent() {}

// ============================================================================
// Qualified Name
// ======================================================================================

// QualifiedIdentifier is a qualified Go identifier.
type QualifiedIdentifier struct {
	Package *Identifier
	Dot     *Position
	Name    *Identifier
}

var _ FullIdentifier = (*QualifiedIdentifier)(nil)

func (ident *QualifiedIdentifier) Start() Position {
	if ident.Package != nil {
		if start := ident.Package.Start(); start != NoPosition {
			return start
		}
	}
	if ident.Dot != nil {
		return *ident.Dot
	}
	if ident.Name != nil {
		if start := ident.Name.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (ident *QualifiedIdentifier) End() Position {
	if ident.Name != nil {
		if end := ident.Name.End(); end != NoPosition {
			return end
		}
	}
	if ident.Dot != nil {
		return deltaPos(*ident.Dot, len("."))
	}
	if ident.Package != nil {
		if end := ident.Package.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (ident *QualifiedIdentifier) Walk(w func(Node)) {
	if ident.Package != nil {
		w(ident.Package)
	}
	if ident.Name != nil {
		w(ident.Name)
	}
}

func (ident *QualifiedIdentifier) Full() string {
	if ident.Package != nil {
		if ident.Dot != nil {
			if ident.Name != nil {
				return ident.Package.Name + "." + ident.Name.Name
			}
			return ident.Package.Name + "."
		}
		if ident.Name != nil {
			return ident.Package.Name + " " + ident.Name.Name
		}
		return ident.Package.Name
	}
	if ident.Dot != nil {
		if ident.Name != nil {
			return "." + ident.Name.Name
		}
		return "."
	}
	if ident.Name != nil {
		return ident.Name.Name
	}
	return ""
}

func (*QualifiedIdentifier) _node()      {}
func (*QualifiedIdentifier) _fullIdent() {}
