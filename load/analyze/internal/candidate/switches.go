package candidate

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/switches"
)

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
	switch n := n.(type) {
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

// ============================================================================
// Containing Element
// ======================================================================================

func SwitchContainingElement(n ast.Node,
	Element func(*ast.Element),
	ComponentCall func(*ast.ComponentCall),
	BlockSetter func(ast.BlockSetter),
) {
	switch n := n.(type) {
	case *ast.Element:
		Element(n)
	case *ast.ComponentCall:
		ComponentCall(n)
	case ast.BlockSetter:
		BlockSetter(n)
	}
}

func SwitchContainingElementR[T any](n ast.Node,
	Element func(*ast.Element) T,
	ComponentCall func(*ast.ComponentCall) T,
	BlockSetter func(ast.BlockSetter) T,
) T {
	switch n := n.(type) {
	case *ast.Element:
		return Element(n)
	case *ast.ComponentCall:
		return ComponentCall(n)
	case ast.BlockSetter:
		return BlockSetter(n)
	default:
		var zero T
		return zero
	}
}

func init() { //nolint:gochecknoinits
	if false {
		// If a compilation error occurs here, change the functions above.
		switches.ContainingElement(nil,
			func(*ast.AndPlaceholderContainingElement) {},
			func(*ast.BlockSetterContainingElement) {},
			func(*ast.Element) {})
	}
}
