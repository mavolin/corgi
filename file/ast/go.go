package ast

import (
	"strconv"
	"strings"
)

// Types representing their Go counterparts.

// ============================================================================
// Identifier
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
// Identifier
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
	return Position{}
}

func (ident *Identifier) End() Position {
	if ident.Position != nil {
		return deltaPos(*ident.Position, len(ident.Name))
	}
	return Position{}
}
func (ident *Identifier) Walk(func(Node)) {}
func (ident *Identifier) Full() string    { return ident.Name }

func (*Identifier) _node()      {}
func (*Identifier) _fullIdent() {}

// ============================================================================
// Qualified Identifier
// ======================================================================================

// QualifiedIdentifier is a qualified Go identifier.
type QualifiedIdentifier struct {
	Package *Identifier
	Dot     *Position
	Name    *Identifier
}

var _ FullIdentifier = (*QualifiedIdentifier)(nil)

func (ident *QualifiedIdentifier) Start() Position {
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

func (ident *QualifiedIdentifier) End() Position {
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
