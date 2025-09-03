package file

import (
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

type Attribute struct {
	//
	// BUILD SYMBOLS

	AST       ast.Attribute
	Reference *AttributeReference

	//
	// ANALYZER

	// Analyzed indicates whether the AttributeReference has been analyzed,
	// albeit with errors.
	Analyzed bool

	// Value is the value of the attribute.
	Value ResolvedAttributeValue

	// Forwarded indicates whether the attribute reference is forwarded to the
	// component calling the component containing it.
	//
	//    comp woof() {
	//      &(bark=...)
	//    }
	//
	// In the above example, the attribute reference bark is forwarded to the
	// component calling woof.
	Forwarded Analysis[bool]
	// ContainingElements are all elements this attribute is placed on.
	// If the attribute is forwarded, this list is incomplete: It only contains
	// the elements up to the component containing the attribute.
	//
	// A nil/empty slice indicates that the attribute reference is fully
	// forwarded.
	//
	// This list only contains the elements that influence the type of the
	// attribute, which is usually the desired behavior.
	// Since forwarded attributes must be explicitly typed, this list would not
	// contain elements from Woof in the below example, since they are
	// irrelevant to the type of bark:
	//    :Woof {
	//      :Bark(data-bark=myVar) // Bark forwards the attributes it receives
	//    }
	//
	// The pointer to the slice has no significance and is just there to
	// satisfy the comparable constraint of Analysis.
	// It is never nil.
	ContainingElements Analysis[*[]ContainingElement]

	// Type is the type of the attribute, resolved from the containing elements.
	//
	// A failed analysis indicates conflicting type values, e.g. if the
	// attribute is placed on multiple elements that specify different types.
	//
	// A type of [attrtype.Unknown] indicates that no type could be determined.
	// A [attrtype.Unknown] is only allowed, if the attribute has a constant
	// value.
	Type Analysis[attrtype.Type]
}

func (a *Attribute) Constant() bool {
	switch val := a.Value.(type) {
	case ConstantBoolAttributeValue:
		return true
	case TextAttributeValue:
		return val.Constant()
	default:
		return false
	}
}

// ============================================================================
// Attribute Value
// ======================================================================================

type (
	// ResolvedAttributeValue is either a [ConstantBoolAttributeValue],
	// [ExpressionBoolAttributeValue], [UntypedAttributeValue], or
	// [TextAttributeValue].
	ResolvedAttributeValue interface {
		_attributeValue()
	}

	ConstantBoolAttributeValue   bool
	ExpressionBoolAttributeValue ast.Expression
	UntypedAttributeValue        ast.Expression

	// TextAttributeValue is a sequence of constant and dynamic parts.
	TextAttributeValue     []TextAttributeValuePart
	TextAttributeValuePart interface {
		_textualAttributeValuePart()
	}
	ConstantTextAttributeValuePart      string
	ExpressionTextAttributeValuePart    ast.Expression
	ComponentCallTextAttributeValuePart ast.ComponentCall
)

var (
	_ ResolvedAttributeValue = ConstantBoolAttributeValue(false)
	_ ResolvedAttributeValue = (*ExpressionBoolAttributeValue)(nil)
	_ ResolvedAttributeValue = (*UntypedAttributeValue)(nil)
	_ ResolvedAttributeValue = (TextAttributeValue)(nil)

	_ TextAttributeValuePart = ConstantTextAttributeValuePart("")
	_ TextAttributeValuePart = (*ExpressionTextAttributeValuePart)(nil)
)

func (ConstantBoolAttributeValue) _attributeValue()    {}
func (*ExpressionBoolAttributeValue) _attributeValue() {}
func (*UntypedAttributeValue) _attributeValue()        {}
func (TextAttributeValue) _attributeValue()            {}

func (ConstantTextAttributeValuePart) _textualAttributeValuePart()       {}
func (*ExpressionTextAttributeValuePart) _textualAttributeValuePart()    {}
func (*ComponentCallTextAttributeValuePart) _textualAttributeValuePart() {}

func (v TextAttributeValue) Constant() bool {
	if len(v) != 1 {
		return false
	}
	_, ok := v[0].(ConstantTextAttributeValuePart)
	return ok
}
