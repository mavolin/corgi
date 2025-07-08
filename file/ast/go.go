package ast

import (
	"strconv"
	"strings"
)

// Types representing their Go counterparts.

// ============================================================================
// Identifier
// ======================================================================================

// FullIdent is a pointer to either an [Ident] or a [QualifiedIdent].
//
// Not actually in the Go spec, but for our convenience.
type FullIdent interface {
	Node
	Full() string
	_fullIdent()
}

// if this is changed, change the comment above
var (
	_ FullIdent = (*Ident)(nil)
	_ FullIdent = (*QualifiedIdent)(nil)
)

// ======================================= Ident ========================================

// Ident is a Go identifier.
type Ident struct {
	Ident    string
	Position *Position
}

var _ FullIdent = (*Ident)(nil)

func (ident *Ident) Start() Position {
	if ident.Position != nil {
		return *ident.Position
	}
	return Position{}
}

func (ident *Ident) End() Position {
	if ident.Position != nil {
		return deltaPos(*ident.Position, len(ident.Ident))
	}
	return Position{}
}
func (ident *Ident) Walk(func(Node)) {}
func (ident *Ident) Full() string    { return ident.Ident }

func (*Ident) _node()      {}
func (*Ident) _fullIdent() {}

// ================================== Qualified Ident ===================================

// QualifiedIdent is a qualified Go identifier.
type QualifiedIdent struct {
	Package *Ident
	Dot     *Position
	Name    *Ident
}

var _ FullIdent = (*QualifiedIdent)(nil)

func (ident *QualifiedIdent) Start() Position {
	switch {
	case ident.Package != nil:
		return ident.Package.Start()
	case ident.Dot != nil:
		return *ident.Dot
	case ident.Name != nil:
		return ident.Name.Start()
	}
	return Position{}
}

func (ident *QualifiedIdent) End() Position {
	switch {
	case ident.Name != nil:
		return ident.Name.End()
	case ident.Dot != nil:
		return deltaPos(*ident.Dot, len("."))
	case ident.Package != nil:
		return ident.Package.End()
	}
	return Position{}
}

func (ident *QualifiedIdent) Walk(w func(Node)) {
	if ident.Package != nil {
		w(ident.Package)
	}
	if ident.Name != nil {
		w(ident.Name)
	}
}

func (ident *QualifiedIdent) Full() string {
	if ident.Package != nil {
		if ident.Dot != nil {
			if ident.Name != nil {
				return ident.Package.Ident + "." + ident.Name.Ident
			}
			return ident.Package.Ident + "."
		}
		if ident.Name != nil {
			return ident.Package.Ident + " " + ident.Name.Ident
		}
		return ident.Package.Ident
	}
	if ident.Dot != nil {
		if ident.Name != nil {
			return "." + ident.Name.Ident
		}
		return "."
	}
	if ident.Name != nil {
		return ident.Name.Ident
	}
	return ""
}

func (*QualifiedIdent) _node()      {}
func (*QualifiedIdent) _fullIdent() {}

// ============================================================================
// Static String
// ======================================================================================

// StaticString is a string literal with no interpolation, equivalent to the
// regular Go string literal.
type StaticString struct {
	Open     *Position
	Quote    rune
	Contents string
	Close    *Position
}

var _ Node = (*StaticString)(nil)

func (s *StaticString) Start() Position {
	if s.Open != nil {
		return *s.Open
	} else if s.Close != nil {
		return *s.Close
	}
	return Position{}
}

func (s *StaticString) End() Position {
	if s.Close != nil {
		return deltaPos(*s.Close, len(`"`))
	} else if s.Open != nil {
		if s.Quote == '"' {
			return deltaPos(*s.Open, len(`"`)+len(s.Contents))
		}

		i := strings.LastIndexByte(s.Contents, '\n')
		if i < 0 {
			return deltaPos(*s.Open, len(`"`)+len(s.Contents))
		}

		lines := strings.Count(s.Contents[:i], "\n")
		return Position{
			Line: s.Open.Line + lines,
			Col:  len(s.Contents[i:]) + 1,
		}
	}

	return Position{}
}
func (s *StaticString) Walk(func(Node)) {}

func (s *StaticString) Quoted() string {
	return string(s.Quote) + s.Contents + string(s.Quote)
}

func (s *StaticString) Unquote() string {
	if s.Quote == '`' {
		return s.Contents
	}

	unq, err := strconv.Unquote(`"` + s.Contents + `"`)
	if err != nil {
		return ""
	}

	return unq
}

func (*StaticString) _node() {}
