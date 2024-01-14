// Package file provides an AST for corgi files.
package file

import (
	"fmt"
)

type (
	Package struct {
		// METADATA
		//

		// Module is the path/name of the Go module providing this directory.
		Module string
		// PathInModule is the path to the directory in the Go module, relative
		// to the module root.
		//
		// It is always specified as a forward slash separated path.
		PathInModule string
		// AbsolutePath is the resolved absolute path to the directory.
		//
		// It is specified using the system's separator.
		AbsolutePath string

		Info *PackageInfo // set by loader after parsing

		//
		// FILES
		//

		// Files are the corgi files this directory consists of.
		Files []*File
	}

	// File represents a parsed corgi file.
	File struct {
		// METADATA
		//

		// Name is the name of the file.
		// of the specified file in its source.
		Name string

		// Module is the path/name of the Go module providing this file.
		Module string
		// PathInModule is the path to the file in the Go module, relative to the
		// module root.
		//
		// It is always specified as a forward slash separated path.
		PathInModule string
		// AbsolutePath is the resolved absolute path to the file.
		//
		// It is specified using the system's separator.
		AbsolutePath string

		// Package is the directory containing this file.
		Package *Package

		//
		// FILE CONTENTS
		//

		// Raw contains the raw input file, as it was parsed.
		Raw string
		// Lines are the lines of Raw, stripped of their CRLF/LF line endings.
		Lines []string

		PackageDoc       []CorgiComment
		PackageDirective PackageDirective

		// ResolvedImports contains a subset of the imports of the file that
		// are relevant for linking component calls.
		ResolvedImports []ResolvedImport // set by linker

		// Scope is the global scope.
		Scope Scope
	}

	ResolvedImport struct {
		Alias string
		Path  string

		Info *PackageInfo
		// Package is the package this import resolves to.
		// Users should not rely on this field being set, as it's usually only
		// set when a package is recompiled, or is compiled for the first time.
		//
		// All necessary information should be available in Info.
		Package *Package
	}
)

type Poser interface {
	Pos() Position
}

// Position indicates the position where a token was encountered.
type Position struct {
	Line int
	Col  int
}

var InvalidPosition = Position{0, 0}

func (p Position) Pos() Position {
	return p
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}
