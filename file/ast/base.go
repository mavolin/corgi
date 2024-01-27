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

var _ Poser = (*Ident)(nil)

func (ident *Ident) Pos() Position { return ident.Position }

// ============================================================================
// Type
// ======================================================================================

// Type represents the name or definition of a Go type.
type Type struct {
	Type     string
	Position Position
}

var _ Poser = (*Type)(nil)

func (t *Type) Pos() Position { return t.Position }

// ============================================================================
// Static String
// ======================================================================================

// StaticString represents a string literal without any interpolation.
type StaticString struct {
	Start    Position
	Quote    rune
	Contents string
	End      *Position
}

var _ Poser = (*StaticString)(nil)

func (s *StaticString) Pos() Position { return s.Start }

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
