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

	// Analyzed indicates whether the Attribute has been analyzed,
	// albeit with errors.
	Analyzed bool

	// Value is the value of the attribute.
	//
	// Nil for &-placeholders.
	Value ResolvedValue

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
	// The pointer to the slice has no significance and is just there to
	// satisfy the comparable constraint of Analysis.
	// It is never nil.
	ContainingElements Analysis[*[]ast.ContainingElement]

	// ContainingElementSpecs are the unique specs of all containing elements,
	// including those containing the attribute indirectly.
	//
	// The pointer to the slice has no significance and is just there to
	// satisfy the comparable constraint of Analysis.
	// It is never nil.
	ContainingElementSpecs Analysis[*[]*ElementSpec]

	// Type is the type of the attribute, resolved from the containing elements.
	//
	// A failed analysis indicates conflicting type values, e.g. if the
	// attribute is placed on multiple elements that specify different types.
	//
	// A type of [attrtype.Unknown] indicates that no type could be determined.
	// This is only allowed when the attribute has a constant value and has no
	// type set for any of its containing elements.
	//
	// # Type Resolution
	//
	// The algorithm to determine the type is as follows:
	//   if attribute is explicitly typed:
	//       return explicit type
	//   if attribute is partially forwarded:
	//       fail
	//   if attribute is fully forwarded or attribute.Reference.Spec is nil:
	//       require attribute to be constant
	//       return attrtype.Unknown
	//   let t be attrtype.Unknown
	//   for each element spec in attribute.ContainingElementSpecs:
	//       if attribute.Reference.Spec has definition for element spec:
	//           set t to attribute type for that element spec
	//           break
	//   if t is attrtype.Unknown:
	//       require attribute to be constant
	//       return attrtype.Unknown
	//   for each element spec in attribute.ContainingElementSpecs:
	//       let rule = attribute.Reference.Spec.RuleFor(element spec)
	//       require rule to be non-nil and rule.Type to be t
	//       otherwise fail
	Type Analysis[attrtype.Type]
}

// ============================================================================
// Attribute Value
// ======================================================================================

type (
	// ResolvedValue is either a [ConstantBool],
	// [BoolExpression], [UndeterminedExpression], or
	// [Text].
	ResolvedValue interface {
		Constant() bool
		_attributeValue()
	}

	ConstantBool           bool
	BoolExpression         ast.Expression
	UndeterminedExpression ast.Expression

	// Text is a sequence of constant and dynamic parts.
	Text     []TextPart
	TextPart interface {
		_textualAttributeValuePart()
	}
	ConstantPart      string
	ExpressionPart    ast.Expression
	ComponentCallPart ast.ComponentCall
)

var (
	_ ResolvedValue = ConstantBool(false)
	_ ResolvedValue = (*BoolExpression)(nil)
	_ ResolvedValue = (*UndeterminedExpression)(nil)
	_ ResolvedValue = (Text)(nil)

	_ TextPart = ConstantPart("")
	_ TextPart = (*ExpressionPart)(nil)
	_ TextPart = (*ComponentCallPart)(nil)
)

func (ConstantBool) _attributeValue()            {}
func (ConstantBool) Constant() bool              { return true }
func (*BoolExpression) _attributeValue()         {}
func (*BoolExpression) Constant() bool           { return false }
func (*UndeterminedExpression) _attributeValue() {}
func (*UndeterminedExpression) Constant() bool   { return false }
func (Text) _attributeValue()                    {}

func (v Text) Constant() bool {
	if len(v) != 1 {
		return false
	}
	_, ok := v[0].(ConstantPart)
	return ok
}

func (ConstantPart) _textualAttributeValuePart()       {}
func (*ExpressionPart) _textualAttributeValuePart()    {}
func (*ComponentCallPart) _textualAttributeValuePart() {}
