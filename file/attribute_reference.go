package file

import "github.com/mavolin/corgi/v2/file/ast"

type AttributeReference struct {
	//
	// BUILD SYMBOLS

	AST *ast.AttributeReference

	// Qualifier is the qualifier of the attribute reference, if the reference
	// is qualified.
	Qualifier Qualifier
	// QualifiableName is the identifier used to refer to the attribute in a
	// qualified reference.
	// Set to the empty string if the reference is unqualified.
	QualifiableName CanonicalQualifiableAttributeName

	// UnqualifiedName is the ascii-lowercased name of the attribute, if the
	// reference is unqualified.
	UnqualifiedName CanonicalAttributeName

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

	// HTMLName is the ascii-lowercased name of the attribute.
	HTMLName Analysis[CanonicalAttributeName]

	//
	// ANALYZER

	// Analyzed indicates whether the AttributeReference has been analyzed,
	// albeit with errors.
	Analyzed bool
}

func (ref *AttributeReference) Qualified() bool {
	return ref.AST.Package != nil || ref.AST.Dot != nil
}
