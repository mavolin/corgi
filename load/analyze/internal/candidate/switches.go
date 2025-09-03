package candidate

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/switches"
)

// ============================================================================
// And Placeholder Writer
// ======================================================================================

func SwitchAndPlaceholderWriter(n ast.Node,
	AndPlaceholder func(*ast.AndPlaceholder),
	ComponentCall func(*ast.ComponentCall),
	BlockSetter func(ast.BlockSetter),
) {
	switch n := n.(type) {
	case *ast.AndPlaceholder:
		AndPlaceholder(n)
	case *ast.ComponentCall:
		ComponentCall(n)
	case ast.BlockSetter:
		BlockSetter(n)
	}
}

func init() { //nolint:gochecknoinits
	if false {
		// If a compilation error occurs here, change the functions above.
		switches.AndPlaceholderWriter(nil,
			func(*ast.AndPlaceholder) {},
			func(*ast.ComponentCallAndPlaceholderWriter) {})
	}
}

// ============================================================================
// Attribute Writer
// ======================================================================================

func SwitchAttributeWriter(n ast.Node,
	ClassShorthand func(*ast.ClassShorthand),
	ComponentCall func(*ast.ComponentCall),
	IDShorthand func(*ast.IDShorthand),
	NamedAttribute func(*ast.NamedAttribute),
) {
	switch n := n.(type) {
	case *ast.ClassShorthand:
		ClassShorthand(n)
	case *ast.ComponentCall:
		ComponentCall(n)
	case *ast.IDShorthand:
		IDShorthand(n)
	case *ast.NamedAttribute:
		NamedAttribute(n)
	}
}

func init() { //nolint:gochecknoinits
	if false {
		// If a compilation error occurs here, change the functions above.
		switches.AttributeWriter(nil,
			func(*ast.ClassShorthand) {},
			func(call *ast.ComponentCall) {},
			func(*ast.IDShorthand) {},
			func(*ast.NamedAttribute) {})
	}
}

// ============================================================================
// Attribute Inhibitor
// ======================================================================================

func SwitchAttributeInhibitor(n ast.Node,
	Block func(*ast.Block),
	BlockSetter func(ast.BlockSetter),
	CharacterEscape func(*ast.CharacterEscape),
	CharacterReference func(*ast.CharacterReference),
	ComponentCall func(*ast.ComponentCall),
	Doctype func(*ast.Doctype),
	Element func(*ast.Element),
	ExpressionInterpolation func(*ast.ExpressionInterpolation),
	RawElement func(*ast.RawElement),
	Text func(*ast.Text),
) {
	switch n := any(n).(type) {
	case *ast.Block:
		Block(n)
	case ast.BlockSetter:
		BlockSetter(n)
	case *ast.CharacterEscape:
		CharacterEscape(n)
	case *ast.CharacterReference:
		CharacterReference(n)
	case *ast.ComponentCall:
		ComponentCall(n)
	case *ast.Doctype:
		Doctype(n)
	case *ast.Element:
		Element(n)
	case *ast.ExpressionInterpolation:
		ExpressionInterpolation(n)
	case *ast.RawElement:
		RawElement(n)
	case *ast.Text:
		Text(n)
	}
}

func init() { //nolint:gochecknoinits
	if false {
		// If a compilation error occurs here, change the functions above.
		switches.AttributeInhibitor(nil,
			func(*ast.Block) {},
			func(*ast.BlockSetterAttributeInhibitor) {},
			func(*ast.CharacterEscape) {},
			func(*ast.CharacterReference) {},
			func(*ast.ComponentCall) {},
			func(*ast.Doctype) {},
			func(*ast.Element) {},
			func(*ast.ExpressionInterpolation) {},
			func(*ast.RawElement) {},
			func(*ast.Text) {})
	}
}
