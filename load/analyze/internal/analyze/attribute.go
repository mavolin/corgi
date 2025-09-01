package analyze

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

// AnalyzeAttributes analyzes all attribute in the package.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeAttributes() {
	logger := z.Logger.WithGroup("attribute_references")
	logger.Debug("Analyzing attribute references")

	for _, f := range z.P.Files {
		logger := logger.With(slog.String("file", f.Name))
		walk.WalkT(f.AST, func(w *walk.ContextT[ast.Attribute]) walk.Action {
			z.AnalyzeAttribute(logger, f, w.Parents, f.AttributeByNode(w.Node))
			return walk.Continue
		})
	}
}

// AnalyzeAttribute analyzes the given attribute.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeAttribute(logger *slog.Logger, f *file.File, parents []*walk.Context, attr *file.Attribute) {
	logger = logger.With(slog.String("attr_pos", attr.AST.Start().String()))

	z.AnalyzeAttributeValue(logger, f, attr)
	z.AnalyzeAttributeForwarded(f, parents, attr)
	z.AnalyzeAttributeContainingElements(f, parents, attr)
	z.AnalyzeAttributeType(logger, f, parents, attr)

	attr.Analyzed = true
}

// ============================================================================
// Value
// ======================================================================================

// AnalyzeAttributeValue sets the Value field on the given attribute.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeAttributeValue(logger *slog.Logger, f *file.File, attr *file.Attribute) {
	logger = logger.WithGroup("value")

	switch attrAST := attr.AST.(type) {
	case *ast.IDShorthand:
		attr.Value = z.shorthandToAttributeValue(nil, logger, f, attrAST.ID)
	case *ast.ClassShorthand:
		attr.Value = z.classShorthandToAttributeValue(logger, f, *attrAST)
	case *ast.NamedAttribute:
		z.namedAttributeValue(logger, f, attr, attrAST)
	}
}

func (z *analyzer) classShorthandToAttributeValue(logger *slog.Logger, f *file.File, s ast.ClassShorthand) file.TextAttributeValue {
	var n int
	for _, name := range s.Names {
		n += len(name)
	}

	v := make(file.TextAttributeValue, 0, n)

	for i, name := range s.Names {
		if i > 0 {
			last := v[len(v)-1]
			if c, _ := last.(file.ConstantTextAttributeValuePart); c != "" {
				v[len(v)-1] = c + " "
			} else {
				v = append(v, file.ConstantTextAttributeValuePart(" "))
			}
		}

		v = z.shorthandToAttributeValue(v, logger, f, name)
	}

	return slices.Clip(v)
}

func (z *analyzer) shorthandToAttributeValue(v file.TextAttributeValue, logger *slog.Logger, f *file.File, s ast.Shorthand) file.TextAttributeValue {
	v = slices.Grow(v, len(s))
	for i, n := range s {
		switch n := n.(type) {
		case *ast.ShorthandText:
			if i == 0 && len(v) > 0 {
				last := v[len(v)-1]
				if c, _ := last.(file.ConstantTextAttributeValuePart); c != "" {
					v[len(v)-1] = c + file.ConstantTextAttributeValuePart(n.Text)
					continue
				}
			}
			v = append(v, file.ConstantTextAttributeValuePart(n.Text))
		case *ast.ShorthandInterpolation:
			v = append(v, (*file.ExpressionTextAttributeValuePart)(n.Expression))
		default:
			logger.Error("Unknown shorthand node")
			z.Report(&diagnostic.Diagnostic{
				Type:    diagnostic.InternalError,
				Message: "unknown shorthand node",
				Primary: []diagnostic.Annotation{
					anno.Node(f, n, fmt.Sprintf("uknown shorthand node type %T", n)),
				},
			})
		}
	}
	return v
}

func (z *analyzer) namedAttributeValue(logger *slog.Logger, f *file.File, attr *file.Attribute, attrAST *ast.NamedAttribute) {
	if attrAST.Value == nil {
		attr.Value = file.ConstantBoolAttributeValue(true)
		return
	}

	expr := z.expressionFromAttributeValue(logger, f, attrAST.Value)
	if expr == nil {
		return
	}

	n0 := expr.Nodes[0]
	s, _ := n0.(*ast.String)
	if s != nil {
		attr.Value = z.stringToAttributeValue(logger, f, s)
		return
	}

	g, _ := n0.(*ast.GoCode)
	if g != nil {
		switch g.Code {
		case "true":
			attr.Value = file.ConstantBoolAttributeValue(true)
			return
		case "false":
			attr.Value = file.ConstantBoolAttributeValue(false)
			return
		}
	}

	typ, _ := InferType(f, expr)
	switch typ {
	case "bool":
		attr.Value = (*file.ExpressionBoolAttributeValue)(expr)
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "string":
		attr.Value = file.TextAttributeValue{(*file.ExpressionTextAttributeValuePart)(expr)}
	default:
		attr.Value = (*file.UntypedAttributeValue)(expr)
	}
	return
}

func (z *analyzer) stringToAttributeValue(logger *slog.Logger, f *file.File, s *ast.String) file.TextAttributeValue {
	v := make(file.TextAttributeValue, 0, len(s.Contents))

	var last file.ConstantTextAttributeValuePart
	for _, content := range s.Contents {
		switch content := content.(type) {
		case *ast.StringText:
			if last != "" {
				last += file.ConstantTextAttributeValuePart(content.Text)
				v[len(v)-1] = last
			} else {
				last = file.ConstantTextAttributeValuePart(content.Text)
				v = append(v, last)
			}
		case *ast.CharacterEscape:
			if last != "" {
				last += file.ConstantTextAttributeValuePart(content.Rune)
				v[len(v)-1] = last
			} else {
				last = file.ConstantTextAttributeValuePart(content.Rune)
				v = append(v, last)
			}
		case *ast.CharacterReference:
			if last != "" {
				last += file.ConstantTextAttributeValuePart(content.Chars)
				v[len(v)-1] = last
			} else {
				last = file.ConstantTextAttributeValuePart(content.Chars)
				v = append(v, last)
			}
		case *ast.ExpressionInterpolation:
			last = ""
			v = append(v, (*file.ExpressionTextAttributeValuePart)(content.Expression))
		case *ast.ComponentCallInterpolation:
			last = ""
			v = append(v, (*file.ComponentCallTextAttributeValuePart)(content.ComponentCall))
		default:
			logger.Error("Unknown string content node")
			z.Report(&diagnostic.Diagnostic{
				Type:    diagnostic.InternalError,
				Message: "unknown string content node",
				Primary: []diagnostic.Annotation{
					anno.Node(f, content, fmt.Sprintf("uknown string content node type %T", content)),
				},
				Explanation: "This most likely happened because the ast.StringNode sum type was extended.\n\n" +
					"This is a bug in the analyzer, please open an issue.",
			})
		}
	}

	return slices.Clip(v)
}

func (z *analyzer) expressionFromAttributeValue(logger *slog.Logger, f *file.File, v ast.AttributeValue) *ast.Expression {
	for {
		switch typed := v.(type) {
		case *ast.TypedAttributeValue:
			v = typed.Value
		case *ast.ExpressionAttributeValue:
			return (*ast.Expression)(typed)
		default:
			logger.Error("Attribute value is neither an expression nor typed attribute value")
			z.Report(&diagnostic.Diagnostic{
				Type:    diagnostic.InternalError,
				Message: "attribute value is neither an expression nor typed attribute value",
				Primary: []diagnostic.Annotation{
					anno.Node(f, v, "for this node"),
				},
				Explanation: "This most likely happened because the ast.AttributeValue sum type was extended.\n\n" +
					"This is a bug in the analyzer, please open an issue.",
			})
			return nil
		}
	}
}

// ============================================================================
// Forwarded
// ======================================================================================

// AnalyzeAttributeForwarded determines whether the given attribute
// reference is forwarded out of the component or not.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Attributes.Forwarded
//
// Depends on Fields:
//   - ComponentCalls.ForwardsReceivedAttributes
//   - Components.Blocks.Forwarded
//   - Components.Blocks.Instances.Forwarded
func (z *analyzer) AnalyzeAttributeForwarded(f *file.File, parents []*walk.Context, attr *file.Attribute) {
	comp := walk.Closest[*ast.Component](parents)
	if comp == nil {
		attr.Forwarded.SetFailed()
		return
	}

	attr.Forwarded.SetResult(true)

	i := len(parents) - 1
	for i >= 0 {
		parent := parents[i]
		switch parent := parent.Node.(type) {
		case *ast.ComponentCall: // we're filling the cc's &-placeholder
			cc := f.ComponentCallByNode(parent)
			if cc.ForwardsReceivedAttributes.Failed() {
				// Continue checking: if the attribute has another element as
				// parent, we can still be sure it's not forwarded.
				attr.Forwarded.SetFailed()
			} else if cc.ForwardsReceivedAttributes.False() {
				attr.Forwarded.SetResult(false)
				return
			}
		case ast.BlockSetter:
			ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
			if ccI < 0 {
				attr.Forwarded.SetFailed()
				continue
			}

			ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
			cc := f.ComponentCallByNode(ccAST)

			s := cc.BlockSetterByName(parent.Name())
			if s == nil || s.Block == nil {
				// Continue checking: if the attribute has another element as
				// parent, we can still be sure it's not forwarded.
				attr.Forwarded.SetFailed()
			} else if s.Block.Forwarded.False() {
				attr.Forwarded.SetResult(false)
				return
			}
			i = ccI - 1 // continue with the parent of the component call
			continue
		case *ast.Element:
			attr.Forwarded.SetResult(false)
			return
		}
		i--
	}
}

// ============================================================================
// Containing Elements
// ======================================================================================

// AnalyzeAttributeContainingElements calculates the containing elements
// of the given attribute.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Attributes.ContainingElements
//
// Depends on Fields:
//   - ComponentCalls.ForwardsReceivedAttributes
//   - ComponentCalls.ElementsWithAndPlaceholder
//   - Components.Blocks.Forwarded
//   - Components.Blocks.Instances.Forwarded
func (z *analyzer) AnalyzeAttributeContainingElements(f *file.File, parents []*walk.Context, attr *file.Attribute) {
	compAST := walk.Closest[*ast.Component](parents)
	if compAST == nil {
		attr.ContainingElements.SetFailed()
		return
	}
	comp := f.Package.ComponentByNode(compAST)

	var containingElements []file.ContainingElement

	i := len(parents) - 1
	for i >= 0 {
		parent := parents[i]
		switch parent := parent.Node.(type) {
		case *ast.ComponentCall: // we're filling the cc's &-placeholder
			cc := f.ComponentCallByNode(parent)
			if cc.ForwardsReceivedAttributes.Failed() || cc.ElementsWithAndPlaceholder.Failed() {
				attr.ContainingElements.SetFailed()
				return
			}

			containingElements = append(containingElements, *cc.ElementsWithAndPlaceholder.Result()...)
			if cc.ForwardsReceivedAttributes.False() {
				containingElements = slices.Clip(containingElements)
				attr.ContainingElements.SetResult(&containingElements)
				return
			}
		case ast.BlockSetter:
			ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
			if ccI < 0 {
				attr.ContainingElements.SetFailed()
				return
			}

			ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
			cc := f.ComponentCallByNode(ccAST)

			s := cc.BlockSetterByName(parent.Name())
			if s == nil || s.Block == nil || s.Block.ContainingElements.Failed() {
				attr.ContainingElements.SetFailed()
				return
			}

			containingElements = append(containingElements, *s.Block.ContainingElements.Result()...)

			if s.Block.Forwarded.False() {
				containingElements = slices.Clip(containingElements)
				attr.ContainingElements.SetResult(&containingElements)
				return
			}
			i = ccI - 1 // continue with the parent of the component call
			continue
		case *ast.Element:
			containingElements = append(containingElements, file.ContainingElement{
				Component: comp,
				Element:   comp.File.ElementReferenceByNode(parent.Header.Name),
			})
			containingElements = slices.Clip(containingElements)
			attr.ContainingElements.SetResult(&containingElements)
			return
		}
		i--
	}
}

// ============================================================================
// Type
// ======================================================================================

// AnalyzeAttributeType determines the type of the given attribute.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Attributes.Type
//
// Depends on Fields:
//   - Attributes.Forwarded
//   - Attributes.ContainingElements
func (z *analyzer) AnalyzeAttributeType(logger *slog.Logger, f *file.File, parents []*walk.Context, attr *file.Attribute) {
	logger = logger.WithGroup("type")

	attr.Type.SetResult(attrtype.Unknown)

	z.analyzeExplicitAttributeType(logger, f, attr)
	if attr.Type.Failed() || attr.Type.Result() != attrtype.Unknown {
		return
	}

	z.analyzeInferredAttributeType(logger, f, attr)
}

func (z *analyzer) analyzeExplicitAttributeType(logger *slog.Logger, f *file.File, attr *file.Attribute) {
	nAttr, _ := attr.AST.(*ast.NamedAttribute)
	if nAttr == nil {
		return
	}

	tav, _ := nAttr.Value.(*ast.TypedAttributeValue)
	if tav == nil {
		return
	}

	if nested, _ := tav.Value.(*ast.TypedAttributeValue); nested != nil {
		attr.Type.SetFailed()
		logger.Error("Nested typed attribute value")
		z.Report(&diagnostic.Diagnostic{
			Message: "attribute: nested typed attribute value",
			Primary: []diagnostic.Annotation{
				anno.Node(f, tav.Type, "nested typed attribute value"),
				anno.Node(f, nested.Type, "nested typed attribute value"),
			},
			Explanation: "There can only be one explicit type per attribute.",
		})
		return
	}

	attr.Type.SetResult(tav.Type.Name.Type)
}

func (z *analyzer) analyzeInferredAttributeType(logger *slog.Logger, f *file.File, attr *file.Attribute) {
	if attr.Forwarded.Equal(true) {
		partial := !attr.ContainingElements.Failed() && len(*attr.ContainingElements.Result()) > 0
		if partial {
			attr.Type.SetFailed()
			logger.Error("Untyped attribute")
			z.Report(&diagnostic.Diagnostic{
				Message: "attribute: unable to determine type: partially outside of an element",
				Primary: []diagnostic.Annotation{
					anno.Node(f, attr.AST, "neither always inside an element nor explicitly typed"),
				},
				Explanation: "As part of the security model, non-constant attributes must be typed." +
					"For example, the `href` attribute placed on an `<a>` element is defined as `url`.\n" +
					"While this attribute is sometimes attached to an element, there is at least one case " +
					"where it is forwarded out of the component, requiring explicit typing.",
				Hints: []diagnostic.Hint{
					{
						Hint:    "Explicitly type the attribute.",
						Example: "`data-woof='url(myVar)`",
					},
				},
				Docs: "attribute-type",
			})
			return
		}

		attr.Type.SetResult(attrtype.Unknown)
		if !attr.Constant() {
			logger.Error("Untyped attribute")
			z.Report(&diagnostic.Diagnostic{
				Message: "attribute: unable to determine type: outside of an element",
				Primary: []diagnostic.Annotation{
					anno.Node(f, attr.AST, "neither inside an element nor explicitly typed"),
				},
				Explanation: "As part of the security model, non-constant attributes must be typed. " +
					"For example, the `href` attribute placed on an `<a>` element is defined as `url`.\n" +
					"Since this attribute is forwarded out of the component, it must be assigned an explicit type.",
				Hints: []diagnostic.Hint{
					{
						Hint:    "Explicitly type the attribute.",
						Example: "`data-woof='url(myVar)`",
					},
				},
				Docs: "attribute-type",
			})
		}
		return
	}

	if attr.ContainingElements.Failed() || attr.Forwarded.Failed() {
		attr.Type.SetFailed()
		return
	}

	containingElements := *attr.ContainingElements.Result()
	if len(containingElements) == 0 {
		attr.Type.SetFailed()
		logger.Error("attribute not forwarded but not contained in any element")
		z.Report(&diagnostic.Diagnostic{
			Type:    diagnostic.InternalError,
			Message: "attribute: not forwarded but neither contained in any element",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attr.AST, "marked as not forwarded, but analysis shows no containing elements"),
			},
			Explanation: "You shouldn't see this error. Please report it.",
		})
		return
	}

	if attr.Reference.Spec.Failed() {
		attr.Type.SetFailed()
		return
	} else if attr.Reference.Spec.Result() == nil {
		attr.Type.SetResult(attrtype.Unknown)
		if !attr.Constant() {
			logger.Error("Attribute not defined")
			z.Report(&diagnostic.Diagnostic{
				Message: "attribute: unable to determine type: attribute not defined",
				Primary: []diagnostic.Annotation{
					anno.Node(f, attr.AST, "unable to determine type"),
				},
				Explanation: "As part of the security model, non-constant attributes must be typed.\n" +
					"You can type an attribute using one of two ways:\n" +
					"Either explicitly type the attribute, e.g. `data-foo='url(myVar)`, " +
					"or define the attribute for the elements it is attached to.\n" +
					"For example, the `href` attribute placed on an `<a>` element is defined as `url`.",
				Hints: []diagnostic.Hint{
					{
						Hint:    "Explicitly type the attribute.",
						Example: "`data-woof='url(myVar)`",
					}, {
						Hint:    "Define the attribute for the elements it is attached to.",
						Example: "`attr woof { div url }`",
					},
				},
				Docs: "attribute-type",
			})
		}
		return
	}

	attrSpec := attr.Reference.Spec.Result()

	var refRule *ast.AttributeRule
	var refElemName string
	var typ attrtype.Type

	e0 := containingElements[0]
	if e0.Element.Spec == nil {
		attr.Type.SetFailed()
		return
	}
	if e0.Element.AST.Package != nil {
		refElemName = e0.Element.AST.Package.Name + "." + e0.Element.AST.Name.Name
	} else {
		refElemName = e0.Element.AST.Name.Name
	}
	refRule = attrSpec.RuleFor(e0.Element.Spec)
	if refRule != nil {
		typ = refRule.Type.Type
	}

	for _, e := range containingElements[1:] {
		if e.Element.Spec == nil {
			attr.Type.SetFailed()
			return
		}

		rule := attrSpec.RuleFor(e.Element.Spec)
		if rule == nil && typ == attrtype.Unknown {
			continue
		} else if rule != nil && typ == rule.Type.Type {
			continue
		}

		var elemName string
		if e.Element.AST.Package != nil {
			elemName = e.Element.AST.Package.Name + "." + e.Element.AST.Name.Name
		} else {
			elemName = e.Element.AST.Name.Name
		}

		secondaries := make([]diagnostic.Annotation, 2)
		if refRule == nil {
			secondaries[0] = anno.Node(attrSpec.File, attrSpec.AST.Selector, "not defined for `"+refElemName+"`")
		} else {
			secondaries[0] = anno.Node(attrSpec.File, refRule.Type, "defined as `"+refRule.Type.Type.String()+"` for `"+refElemName+"`")
		}
		if rule == nil {
			secondaries[1] = anno.Node(attrSpec.File, attrSpec.AST.Selector, "not defined for `"+elemName+"`")
		} else {
			secondaries[1] = anno.Node(attrSpec.File, rule.Type, "defined as `"+rule.Type.Type.String()+"` for `"+elemName+"`")
		}

		attr.Type.SetFailed()
		logger.Error("Attribute has conflicting types in different elements")
		z.Report(&diagnostic.Diagnostic{
			Message: "attribute: unable to determine type: conflicting types",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attr.AST, "attached to different elements with conflicting types"),
			},
			Secondary:   secondaries,
			Explanation: "This attribute is attached to multiple elements that define the attribute with different types.",
			Hints: []diagnostic.Hint{
				{
					Hint:    "Explicitly type the attribute.",
					Example: "`data-woof='url(myVar)`",
				},
			},
		})
		return
	}

	attr.Type.SetResult(typ)
	if typ != attrtype.Unknown || attr.Constant() {
		return
	}

	logger.Error("Untyped attribute")
	z.Report(&diagnostic.Diagnostic{
		Message: "attribute: unable to determine type: attribute not defined for element",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attr.AST, "neither explicitly typed nor defined for the element"),
		},
		Explanation: "As part of the security model, non-constant attributes must be typed. " +
			"For example, the `href` attribute placed on an `<a>` element is defined as `url`.\n" +
			"This attribute is attached to an element that does not define it.",
		Hints: []diagnostic.Hint{
			{
				Hint:    "Explicitly type the attribute.",
				Example: "`data-woof='url(myVar)`",
			}, {
				Hint:    "Define the attribute for the elements it is attached to.",
				Example: "`attr woof { div url }`",
			},
		},
		Docs: "attribute-type",
	})
}
