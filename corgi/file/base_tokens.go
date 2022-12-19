package file

// This file contains corgi's base types

type String struct {
	Unquoted string
	Raw      string

	Pos
}

// Ident represents a corgi identifier.
type Ident struct {
	Ident string
	Pos
}
