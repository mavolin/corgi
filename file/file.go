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

	// Module is the name of the Go module providing this file.
	Module string
	// ModulePath is the file in the Go module that provides this file,
	// relative to the module root.
	//
	// It is always specified as a forward slash separated path.
	ModulePath string
	// AbsolutePath is the resolved absolute path to the file.
	//
	// It is always specified as a forward slash separated path.
	AbsolutePath string

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
	//
	// After linking, this list will be appended by imports made by extended,
	// used, and included files.
	Imports []Import

	// Uses is a list of used library files, in the order they appear.
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
	TypeExtend
	TypeUse
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
