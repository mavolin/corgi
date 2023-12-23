// Package file provides an AST for corgi files.
package file

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

		//
		// FILES
		//

		// Files are the corgi files this directory consists of.
		Files []*File
	}

	PackageInfo struct {
		// CorgiVersion is the version of the corgi compiler used to compile
		// the files in this package.
		CorgiVersion string
		// HasState indicates whether this package contains state variables.
		HasState bool

		Components []PackageComponentInfo
	}
	PackageComponentInfo struct {
		Name   string
		Params []PackageComponentParamInfo
		ComponentInfo
	}
	PackageComponentParamInfo struct {
		Name       string
		Type       string
		IsSafeType bool
		HasDefault bool
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

		ImportComments []CorgiComment

		// Imports is a list of imports made by this file, in the order they
		// appear.
		Imports []Import

		// Scope is the global scope.
		Scope Scope
	}
)

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
