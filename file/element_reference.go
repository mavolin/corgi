package file

import "github.com/mavolin/corgi/v2/file/ast"

type ElementReference struct {
	//
	// BUILD SYMBOLS

	AST *ast.ElementReference

	// Qualifier is the qualifier of the element reference, if the reference
	// is qualified.
	Qualifier Qualifier
	// QualifiableName is the identifier used to refer to the element in a
	// qualified reference.
	// Set to the empty string if the reference is unqualified.
	QualifiableName CanonicalQualifiableElementName

	// UnqualifiedName is the html name of the element, in ascii-lowercase,
	// if the reference is unqualified.
	UnqualifiedName CanonicalElementName

	//
	// LINKER

	// Linked indicates whether the ElementReference has been seen by the
	// linker, and it attempted to link it.
	//
	// If true, but Spec is nil, the linker encountered an error while
	// linking the ElementReference.
	Linked bool

	// Spec is the spec providing the type of the Element.
	Spec *ElementSpec
}

func (ref *ElementReference) Qualified() bool {
	return ref.AST.Package != nil || ref.AST.Dot != nil
}
