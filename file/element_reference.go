package file

import "github.com/mavolin/corgi/v2/file/ast"

type ElementReference struct {
	//
	// BUILD SYMBOLS

	AST *ast.ElementReference

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

// HTMLName returns the name of the element.
//
// Can only be called after successful linking.
func (r *ElementReference) HTMLName() string {
	return r.Spec.HTMLName()
}
