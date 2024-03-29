// Package file provides an AST for corgi files.
package file

// File represents a parsed corgi file.
type File struct {
	// METADATA
	//

	Type Type

	// Name is the name of the file.
	//
	// If this is the main file, it will be just the file's name.
	//
	// If this is an extended, included, or used file, this will be the path
	// of the specified file in its source.
	Name string

	// Module is the path/name of the Go module providing this file.
	//
	// This won't be set for main and include files.
	Module string
	// PathInModule is the path to the file in the Go module, relative to the
	// module root.
	//
	// It is always specified as a forward slash separated path.
	//
	// This won't be set for main and include files.
	PathInModule string
	// AbsolutePath is the resolved absolute path to the file.
	//
	// It is specified using the system's separator.
	AbsolutePath string

	// Library is the library this file belongs to, if any.
	Library *Library
	// DirLibrary provides the library files located in the same directory as
	// this main, include, or template file.
	//
	// Not filled for library files.
	DirLibrary *Library

	//
	// FILE CONTENTS
	//

	// Raw contains the raw input file, as it was parsed.
	Raw string
	// Lines are the lines of Raw, stripped of their CRLF/LF line endings.
	Lines []string

	// TopLevelComments contains all comments made before the first scope item.
	TopLevelComments []CorgiComment

	// Extend is the file that this file extends, if any.
	Extend *Extend

	// Imports is a list of imports made by this file, in the order they
	// appear.
	Imports []Import

	// Uses is a list of used libraries, in the order they appear.
	Uses []Use

	// GlobalCode is the code that is written above the output function.
	GlobalCode []Code

	// Func is the function header.
	// It is always present for main files, i.e. those files that are given
	// to the corgi command.
	Func *Func

	// Scope is the global scope.
	Scope Scope
}

type Type uint8

const (
	TypeMain Type = iota + 1
	TypeTemplate
	TypeLibraryFile
	TypeInclude
)

// A Scope represents a level of indentation.
// Every mixin available inside a scope is also available in its child scopes.
type Scope []ScopeItem

// ScopeItem represents an item in a scope.
type ScopeItem interface {
	_typeScopeItem()
	Poser
}

// Position indicates the position where a token was encountered.
type Position struct {
	Line int
	Col  int
}

var InvalidPosition = Position{0, 0}

type Poser interface {
	Pos() Position
}

func (p Position) Pos() Position {
	return p
}
