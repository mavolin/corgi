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
	Value AttributeValue

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
	// ContainingElements are all elements containing this block instance.
	// If Forwarded is true, the list is not absolute: It would need to be
	// extended with the containing elements of the component call.
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
	case TextualAttributeValue:
		return val.Constant()
	default:
		return false
	}
}

// ============================================================================
// Attribute Value
// ======================================================================================

type (
	// AttributeValue is either a [ConstantBoolAttributeValue],
	// [DynamicBoolAttributeValue], [UntypedAttributeValue], or
	// [TextualAttributeValue].
	AttributeValue interface {
		_attributeValue()
	}

	ConstantBoolAttributeValue bool
	DynamicBoolAttributeValue  ast.Expression
	UntypedAttributeValue      ast.Expression

	// TextualAttributeValue is a sequence of constant and dynamic parts.
	TextualAttributeValue     []TextualAttributeValuePart
	TextualAttributeValuePart interface {
		_textualAttributeValuePart()
	}
	ConstantTextualAttributeValuePart string
	DynamicTextualAttributeValuePart  ast.Expression
)

var (
	_ AttributeValue = ConstantBoolAttributeValue(false)
	_ AttributeValue = (*DynamicBoolAttributeValue)(nil)
	_ AttributeValue = (*UntypedAttributeValue)(nil)
	_ AttributeValue = (TextualAttributeValue)(nil)

	_ TextualAttributeValuePart = ConstantTextualAttributeValuePart("")
	_ TextualAttributeValuePart = (*DynamicTextualAttributeValuePart)(nil)
)

func (ConstantBoolAttributeValue) _attributeValue() {}
func (*DynamicBoolAttributeValue) _attributeValue() {}
func (*UntypedAttributeValue) _attributeValue()     {}
func (TextualAttributeValue) _attributeValue()      {}

func (ConstantTextualAttributeValuePart) _textualAttributeValuePart() {}
func (*DynamicTextualAttributeValuePart) _textualAttributeValuePart() {}

func (v TextualAttributeValue) Constant() bool {
	if len(v) != 1 {
		return false
	}
	_, ok := v[0].(ConstantTextualAttributeValuePart)
	return ok
}
