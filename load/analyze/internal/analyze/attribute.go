package analyze

import (
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/switches"
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

	z.AnalyzeAttributeValue(f, attr)
	z.AnalyzeAttributeForwarded(f, parents, attr)
	z.AnalyzeAttributeContainingElements(f, parents, attr)
	z.AnalyzeAttributeType(logger, f, attr)

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
func (z *analyzer) AnalyzeAttributeValue(f *file.File, attr *file.Attribute) {
	switches.Attribute(attr.AST,
		func(*ast.AndPlaceholder) {},
		func(attrAST *ast.ClassShorthand) { attr.Value = z.classShorthandToAttributeValue(attrAST) },
		func(attrAST *ast.IDShorthand) { attr.Value = z.shorthandToAttributeValue(nil, attrAST.ID) },
		func(attrAST *ast.NamedAttribute) { attr.Value = z.namedAttributeToAttributeValue(f, attrAST) })
}

func (z *analyzer) classShorthandToAttributeValue(s *ast.ClassShorthand) file.Text {
	var n int
	for _, name := range s.Names {
		n += len(name)
	}

	v := make(file.Text, 0, n)

	for i, name := range s.Names {
		if i > 0 {
			last := v[len(v)-1]
			if c, _ := last.(file.ConstantPart); c != "" {
				v[len(v)-1] = c + " "
			} else {
				v = append(v, file.ConstantPart(" "))
			}
		}

		v = z.shorthandToAttributeValue(v, name)
	}

	return slices.Clip(v)
}

func (z *analyzer) shorthandToAttributeValue(v file.Text, s ast.Shorthand) file.Text {
	v = slices.Grow(v, len(s))
	for i, n := range s {
		switches.ShorthandNode(n,
			func(n *ast.ShorthandInterpolation) {
				v = append(v, (*file.ExpressionPart)(n.Expression))
			},
			func(n *ast.ShorthandText) {
				if i == 0 && len(v) > 0 {
					last := v[len(v)-1]
					if c, _ := last.(file.ConstantPart); c != "" {
						v[len(v)-1] = c + file.ConstantPart(n.Text)
						return
					}
				}
				v = append(v, file.ConstantPart(n.Text))
			})
	}
	return v
}

func (z *analyzer) namedAttributeToAttributeValue(f *file.File, attrAST *ast.NamedAttribute) file.ResolvedAttributeValue {
	if attrAST.Value == nil {
		return file.ConstantBool(true)
	}

	expr := z.expressionFromAttributeValue(attrAST.Value)

	n0 := expr.Nodes[0]
	switches.CodeNodeR(n0,
		func(*ast.BlockFunction) file.ResolvedAttributeValue { return nil },
		func(*ast.ComponentCall) file.ResolvedAttributeValue { return nil },
		func(gc *ast.GoCode) file.ResolvedAttributeValue {
			switch gc.Code {
			case "true":
				return file.ConstantBool(true)
			case "false":
				return file.ConstantBool(false)
			default:
				return nil
			}
		},
		func(s *ast.String) file.ResolvedAttributeValue {
			return z.stringToAttributeValue(s)
		},
		func(*ast.Ternary) file.ResolvedAttributeValue { return nil },
		func(*ast.ZeroCoalescing) file.ResolvedAttributeValue { return nil },
	)

	typ, _ := InferType(f, expr)
	switch typ {
	case "bool":
		return (*file.BoolExpression)(expr)
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "string":
		return file.Text{(*file.ExpressionPart)(expr)}
	default:
		return (*file.UndeterminedExpression)(expr)
	}
}

func (z *analyzer) stringToAttributeValue(s *ast.String) file.Text {
	v := make(file.Text, 0, len(s.Contents))

	var last file.ConstantPart
	for _, content := range s.Contents {
		switches.StringNode(content,
			func(content *ast.BadInterpolation) {
				panic("analyzer called with file with parse errors: " + content.Start().String())
			},
			func(content *ast.CharacterEscape) { addConstant(&v, &last, string(content.Rune)) },
			func(content *ast.CharacterReference) { addConstant(&v, &last, content.Chars) },
			func(content *ast.ComponentCallInterpolation) {
				last = ""
				v = append(v, (*file.ComponentCallPart)(content.ComponentCall))
			},
			func(content *ast.ExpressionInterpolation) {
				last = ""
				v = append(v, (*file.ExpressionPart)(content.Expression))
			},
			func(content *ast.StringText) { addConstant(&v, &last, content.Text) })
	}

	return slices.Clip(v)
}

func addConstant(v *file.Text, last *file.ConstantPart, s string) {
	if *last != "" {
		*last += file.ConstantPart(s)
		(*v)[len(*v)-1] = *last
	} else {
		*last = file.ConstantPart(s)
		*v = append(*v, *last)
	}
}

func (z *analyzer) expressionFromAttributeValue(v ast.AttributeValue) *ast.Expression {
	for {
		e := switches.AttributeValueR(v,
			func(eav *ast.ExpressionAttributeValue) *ast.Expression {
				return (*ast.Expression)(eav)
			},
			func(tav *ast.TypedAttributeValue) *ast.Expression {
				v = tav.Value
				return nil
			})
		if e != nil {
			return e
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
			compAST := walk.Closest[*ast.Component](parents)
			if compAST == nil {
				attr.ContainingElements.SetFailed()
				return
			}
			comp := f.Package.ComponentByNode(compAST)

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

	containingElements = slices.Clip(containingElements)
	attr.ContainingElements.SetResult(&containingElements)
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
func (z *analyzer) AnalyzeAttributeType(logger *slog.Logger, f *file.File, attr *file.Attribute) {
	logger = logger.WithGroup("type")

	attr.Type.SetResult(attrtype.Unknown)

	z.analyzeExplicitAttributeType(logger, f, attr)
	if attr.Type.Failed() || attr.Type.Result() != attrtype.Unknown {
		return
	}

	z.analyzeInferredAttributeType(logger, f, attr)
}

func (z *analyzer) analyzeExplicitAttributeType(logger *slog.Logger, f *file.File, attr *file.Attribute) {
	val := switches.AttributeR(attr.AST,
		func(*ast.AndPlaceholder) ast.AttributeValue { return nil },
		func(*ast.ClassShorthand) ast.AttributeValue { return nil },
		func(*ast.IDShorthand) ast.AttributeValue { return nil },
		func(na *ast.NamedAttribute) ast.AttributeValue { return na.Value })
	if val == nil {
		return
	}

	tav, _ := val.(*ast.TypedAttributeValue)
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
