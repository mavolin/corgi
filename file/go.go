package file

// GoIdent represents a Go identifier.
type GoIdent struct {
	Ident string
	Position
}

// GoType represents the name or definition of a Go type.
type GoType struct {
	Type string
	Position
}
