package ast

import "strconv"

// ============================================================================
// Ident
// ======================================================================================

// Ident represents a Go identifier.
type Ident struct {
	Ident    string
	Position Position
}

var _ Node = (*Ident)(nil)

func (ident *Ident) Pos() Position { return ident.Position }
func (ident *Ident) End() Position { return deltaPos(ident.Position, len(ident.Ident)) }

func (*Ident) _node() {}

// ============================================================================
// Type
// ======================================================================================

// Type represents the name or definition of a Go type.
type Type struct {
	Type     string
	Position Position
}

var _ Node = (*Type)(nil)

func (t *Type) Pos() Position { return t.Position }
func (t *Type) End() Position { return deltaPos(t.Position, len(t.Type)) }

func (*Type) _node() {}

// ============================================================================
// Static String
// ======================================================================================

// StaticString represents a string literal without any interpolation.
type StaticString struct {
	Open     Position
	Quote    rune
	Contents string
	Close    *Position
}

var _ Node = (*StaticString)(nil)

func (s *StaticString) Pos() Position { return s.Open }
func (s *StaticString) End() Position {
	if s.Close != nil {
		return *s.Close
	}

	return deltaPos(s.Open, len(`"`)+len(s.Contents))
}

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
