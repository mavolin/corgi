package file

import "strconv"

// ============================================================================
// Ident
// ======================================================================================

// Ident represents a Go identifier.
type Ident struct {
	Ident string
	Position
}

// ============================================================================
// Type
// ======================================================================================

// Type represents the name or definition of a Go type.
type Type struct {
	Type string
	Position
}

// ============================================================================
// Static String
// ======================================================================================

type StaticString struct {
	Start    Position
	Quote    rune
	Contents string
	End      Position
}

func (s StaticString) Pos() Position { return s.Start }

func (s StaticString) Quoted() string {
	return string(s.Quote) + s.Contents + string(s.Quote)
}

func (s StaticString) Unquote() string {
	if s.Quote == '`' {
		return s.Contents
	}

	unq, err := strconv.Unquote(`"` + s.Contents + `"`)
	if err != nil {
		return ""
	}

	return unq
}
