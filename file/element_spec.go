package file

import (
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

type ElementSpec struct {
	//
	// BUILD SYMBOLS

	AST        *ast.ElementSpec
	Definition *ast.ElementDefinition
	File       *File

	// StylizedHTMLName is the html name of the element, in the casing found in
	// the corgi source.
	StylizedHTMLName string
	HTMLName         CanonicalElementName
	QualifiableName  CanonicalQualifiableElementName

	//
	// ANALYZE

	// Analyzed indicates whether the ElementSpec has been analyzed,
	// albeit with errors.
	Analyzed bool

	// Circular indicates that the element is defined by referencing itself.
	Circular bool

	Type Analysis[elemtype.Type]
}
