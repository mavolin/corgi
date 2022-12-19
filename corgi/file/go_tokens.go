package file

// GoLiteral represents a Go literal.
//
// It is purely used for easier identification of the expected contents of a
// string.
type GoLiteral struct {
	Literal string
	Pos
}

// GoIdent represents a Go identifier.
//
// It is purely used for easier identification of the expected contents of a
// string.
type GoIdent struct {
	Ident string
	Pos
}
