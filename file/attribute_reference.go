package file

import "github.com/mavolin/corgi/v2/file/ast"

type AttributeReference struct {
	//
	// BUILD SYMBOLS

	AST *ast.AttributeReference

	//
	// LINKER

	// Linked indicates whether the AttributeReference has been seen by the
	// linker, and it attempted to link it.
	// If true, but Spec is nil, the linker encountered an error while
	// linking the AttributeReference.
	Linked bool

	// Spec is the spec declaring the attribute.
	//
	// Unlike element references, attribute references needn't have a spec:
	// Attributes can be explicitly typed, in which case the prior definition
	// of the attribute is optional.
	//
	// It is the analyzer's responsibility to report cases in which it expects
	// an attribute reference to have a spec, but it doesn't.
	Spec Analysis[*AttributeSpec] // may be nil
}

// HTMLName returns the name of the attribute.
func (r *AttributeReference) HTMLName() (a Analysis[string]) {
	// possibly has a prefix
	if r.AST.Package != nil {
		if r.Spec.Equal(nil) { // externally defined attribute, but no spec?
			a.SetFailed()
			return a
		}

		// prepend the prefix
		if r.Spec.Result().Definition.Prefix != nil {
			a.SetResult(r.Spec.Result().Definition.Prefix.Name + r.AST.Name.Name)
			return a
		}

		// fallthrough, no prefix
	}

	a.SetResult(r.AST.Name.Name)
	return a
}
