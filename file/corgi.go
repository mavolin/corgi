package file

// ============================================================================
// String
// ======================================================================================

type String struct {
	Quote    byte
	Contents string

	Position
}

// ============================================================================
// Ident
// ======================================================================================

// Ident represents a corgi identifier.
type Ident struct {
	Ident string
	Position
}

// ============================================================================
// CorgiComment
// ======================================================================================

// CorgiComment represents a corgi comment, i.e. a comment that is not printed.
type CorgiComment struct {
	// Lines are the lines of the comment.
	//
	// Empty lines may be excluded.
	Lines []CorgiCommentLine
	Position
}

type CorgiCommentLine struct {
	Comment string
	Position
}

var _ ScopeItem = CorgiComment{}

func (CorgiComment) _typeScopeItem() {}
