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
	"github.com/mavolin/corgi/v2/load/analyze/internal/candidate"
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

	for _, f := range z.Pkg.Files {
		logger := logger.With(slog.String("file", string(f.Name)))
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

	z.AnalyzeAttribute_Value(f, attr)
	z.AnalyzeAttribute_Forwarded(f, parents, attr)
	z.AnalyzeAttribute_Receivers(f, parents, attr)
	z.AnalyzeAttribute_ReceivingElementSpecs(f, attr)
	z.AnalyzeAttribute_Type(logger, f, attr)

	attr.Analyzed = true
}

// ============================================================================
// Value
// ======================================================================================

// AnalyzeAttribute_Value sets the Value field on the given attribute.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeAttribute_Value(f *file.File, attr *file.Attribute) {
	switches.Attribute(attr.AST,
		func(*ast.AndPlaceholder) {},
		func(attrAST *ast.ClassShorthand) { attr.Value = z.classShorthandToAttributeValue(attrAST) },
		func(attrAST *ast.IDShorthand) { attr.Value = z.shorthandToResolvedValue(nil, attrAST.ID) },
		func(attrAST *ast.NamedAttribute) { attr.Value = z.namedAttributeToResolvedValue(f, attrAST) })
}

// ============================================================================
// Forwarded
// ======================================================================================

// AnalyzeAttribute_Forwarded determines whether the given attribute
// reference is forwarded out of the component or not.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Attributes.Forwarded
//
// Depends on Fields:
//   - Components.Blocks.Forwarded
//   - Components.Blocks.Instances.Forwarded
func (z *analyzer) AnalyzeAttribute_Forwarded(f *file.File, parents []*walk.Context, attr *file.Attribute) {
	attr.Forwarded.SetResult(true)

	i := len(parents) - 1
	for i >= 0 {
		candidate.SwitchAttributeReceiver(parents[i].Node,
			func(*ast.Element) {
				attr.Forwarded.SetResult(false)
			},
			func(parent *ast.ComponentCall) {
				cc := f.ComponentCallByNode(parent)
				forwardsReceivedAttributes := cc.ForwardsAcceptedAttributes()
				if forwardsReceivedAttributes.Failed() {
					// Continue checking: if the attribute has another element as
					// parent, we can still be sure it's not forwarded.
					attr.Forwarded.SetFailed()
				} else if forwardsReceivedAttributes.False() {
					attr.Forwarded.SetResult(false)
				}
			},
			func(parent ast.BlockSetter) {
				ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
				if ccI < 0 {
					attr.Forwarded.SetFailed()
					return
				}

				ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
				cc := f.ComponentCallByNode(ccAST)

				s := cc.BlockSetterByNode(parent)
				if s == nil || s.Block == nil {
					// Continue checking: if the attribute has another element as
					// parent, we can still be sure it's not forwarded.
					attr.Forwarded.SetFailed()
				} else if s.Block.Forwarded().False() {
					attr.Forwarded.SetResult(false)
				}
				i = ccI // continue with the parent of the component call
			})
		if attr.Forwarded.Equal(false) {
			return
		}
		i--
	}
}

// ============================================================================
// Receivers
// ======================================================================================

// AnalyzeAttribute_Receivers calculates the containing elements
// of the given attribute.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Attributes.Receivers
//
// Depends on Fields:
//   - ComponentCalls.ElementsWithAndPlaceholder
//   - Components.Blocks.Receivers
//   - Components.Blocks.Forwarded
func (z *analyzer) AnalyzeAttribute_Receivers(f *file.File, parents []*walk.Context, attr *file.Attribute) {
	var receivers []ast.AttributeReceiver

	i := len(parents) - 1
	for i >= 0 {
		done := candidate.SwitchAttributeReceiverR(parents[i].Node,
			func(parent *ast.Element) bool {
				receivers = append(receivers, parent)
				return true
			},
			func(parent *ast.ComponentCall) bool {
				cc := f.ComponentCallByNode(parent)
				forwardsReceivedAttributes := cc.ForwardsAcceptedAttributes()
				if forwardsReceivedAttributes.Failed() || cc.ElementsWithAndPlaceholder.Failed() {
					attr.Receivers.SetFailed()
					return true
				}

				if cc.ElementsWithAndPlaceholder.Result().Len() > 0 {
					receivers = append(receivers, (*ast.AndPlaceholderAttributeReceiver)(parent))
				}
				return forwardsReceivedAttributes.False()
			},
			func(parent ast.BlockSetter) bool {
				ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
				if ccI < 0 {
					attr.Receivers.SetFailed()
					return true
				}

				ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
				cc := f.ComponentCallByNode(ccAST)

				s := cc.BlockSetterByNode(parent)
				if s == nil || s.Block == nil {
					attr.Receivers.SetFailed()
					return true
				}
				for _, instance := range s.Block.Instances {
					if instance.ContainingElements.Failed() {
						attr.Receivers.SetFailed()
						return true
					}
				}

				receivers = append(receivers, &ast.BlockSetterContainingElement{
					ComponentCall: ccAST,
					BlockSetter:   parent,
				})
				if s.Block.Forwarded().False() {
					return true
				}
				i = ccI // continue with the parent of the component call
				return false
			})
		if done {
			break
		}

		i--
	}

	if !attr.Receivers.Failed() {
		receivers = slices.Clip(receivers)
		attr.Receivers.SetResult(file.SliceRefFrom(receivers))
	}
}

// ============================================================================
// Receiving Element Specs
// ======================================================================================

// AnalyzeAttribute_ReceivingElementSpecs calculates the containing element
// specs of the given attribute.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Attributes.ReceivingElementSpecs
//
// Depends on Fields:
//   - Attributes.Receivers
//   - ComponentCalls.ElementSpecsWithAndPlaceholder
//   - Components.Blocks.ReceivingElementSpecs
func (z *analyzer) AnalyzeAttribute_ReceivingElementSpecs(f *file.File, attr *file.Attribute) {
	if attr.Receivers.Failed() {
		attr.ReceivingElementSpecs.SetFailed()
		return
	}

	containingElements := attr.Receivers.Result().Get()
	if len(containingElements) == 0 {
		attr.ReceivingElementSpecs.SetResult(file.NilSliceRef[*file.ElementSpec]())
		return
	}

	specSet := make(map[*file.ElementSpec]struct{}, len(containingElements))
	for _, e := range containingElements {
		switches.AttributeReceiver(e,
			func(e *ast.AndPlaceholderAttributeReceiver) {
				cc := f.ComponentCallByNode((*ast.ComponentCall)(e))
				if cc.ElementSpecsWithAndPlaceholder.Failed() {
					attr.ReceivingElementSpecs.SetFailed()
					return
				}

				for _, spec := range cc.ElementSpecsWithAndPlaceholder.Result().Get() {
					specSet[spec] = struct{}{}
				}
			},
			func(e *ast.BlockSetterContainingElement) {
				cc := f.ComponentCallByNode(e.ComponentCall)
				s := cc.BlockSetterByNode(e.BlockSetter)
				if s == nil || s.Block == nil {
					attr.ReceivingElementSpecs.SetFailed()
					return
				}

				for _, instance := range s.Block.Instances {
					if instance.ContainingElementSpecs.Failed() {
						attr.ReceivingElementSpecs.SetFailed()
						return
					}

					for _, spec := range instance.ContainingElementSpecs.Result().Get() {
						specSet[spec] = struct{}{}
					}
				}
			},
			func(e *ast.Element) {
				ref := f.ElementReferenceByNode(e.Header.Name)
				if ref.Spec == nil {
					attr.ReceivingElementSpecs.SetFailed()
					return
				}
				specSet[ref.Spec] = struct{}{}
			})
		if attr.ReceivingElementSpecs.Failed() {
			return
		}
	}

	specs := make([]*file.ElementSpec, 0, len(specSet))
	for spec := range specSet {
		specs = append(specs, spec)
	}
	attr.ReceivingElementSpecs.SetResult(file.SliceRefFrom(specs))
}

// ============================================================================
// Type
// ======================================================================================

// AnalyzeAttribute_Type determines the type of the given attribute.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Attributes.Type
//
// Depends on Fields:
//   - Attributes.Forwarded
//   - Attributes.Receivers
func (z *analyzer) AnalyzeAttribute_Type(logger *slog.Logger, f *file.File, attr *file.Attribute) {
	logger = logger.WithGroup("type")

	attr.Type.SetResult(nil)

	z.analyzeAttribute_Type_explicit(logger, f, attr)
	if attr.Type.Failed() || attr.Type.Result() != nil {
		return
	}

	z.analyzeAttribute_Type_inferred(logger, f, attr)
}

func (z *analyzer) analyzeAttribute_Type_explicit(logger *slog.Logger, f *file.File, attr *file.Attribute) {
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

func (z *analyzer) analyzeAttribute_Type_inferred(logger *slog.Logger, f *file.File, attr *file.Attribute) {
	if attr.Forwarded.Equal(true) {
		partial := !attr.Receivers.Failed() && attr.Receivers.Result().Len() > 0
		if partial {
			attr.Type.SetFailed()
			logger.Error("Untyped attribute")
			z.Report(&diagnostic.Diagnostic{
				Message: "attribute: unable to determine type: partially outside of an element",
				Primary: []diagnostic.Annotation{
					anno.Node(f, attr.AST, "neither always inside an element nor explicitly typed"),
				},
				Explanation: "As part of corgi's security model, forwarded attributes must be typed." +
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

		attr.Type.SetResult(nil)
		if attr.Value.Constant() {
			return
		}

		logger.Error("Untyped attribute")
		z.Report(&diagnostic.Diagnostic{
			Message: "attribute: unable to determine type: outside of an element",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attr.AST, "neither inside an element nor explicitly typed"),
			},
			Explanation: "As part of the security model, non-constant attributes must be typed. " +
				"For example, the `href` attribute placed on an `<a>` element is defined as `url`.\n" +
				"Since this attribute is forwarded out of the component, it must be assigned an explicit type " +
				"or hold a constant value.",
			Hints: []diagnostic.Hint{
				{
					Hint:    "Explicitly type the attribute.",
					Example: "`data-woof='url(myVar)`",
				}, {
					Hint:    "Set this attribute to a constant value.",
					Example: "`data-woof=\"bark\"`",
				},
			},
			Docs: "attribute-type",
		})
		return
	}

	if attr.ReceivingElementSpecs.Failed() || attr.Forwarded.Failed() {
		attr.Type.SetFailed()
		return
	}

	containingElementSpecs := attr.ReceivingElementSpecs.Result().Get()
	if len(containingElementSpecs) == 0 {
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
		attr.Type.SetResult(nil)
		if attr.Value.Constant() {
			return
		}

		logger.Error("Attribute not defined")
		z.Report(&diagnostic.Diagnostic{
			Message: "attribute: unable to determine type: attribute not defined",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attr.AST, "unable to determine type"),
			},
			Explanation: "As part of the security model, non-constant attributes must be typed.\n" +
				"You can type an attribute using one of two ways:\n" +
				"Either explicitly type the attribute, e.g. `data-foo='url(myVar)`, " +
				"define the attribute for the elements it is attached to, " +
				"or set it to a constant value.\n" +
				"For example, the `href` attribute placed on an `<a>` element is defined as `url`.",
			Hints: []diagnostic.Hint{
				{
					Hint:    "Explicitly type the attribute.",
					Example: "`data-woof='url(myVar)`",
				}, {
					Hint:    "Define the attribute for the elements it is attached to.",
					Example: "`attr woof { div url }`",
				}, {
					Hint:    "Set this attribute to a constant value.",
					Example: "`data-woof=\"bark\"`",
				},
			},
			Docs: "attribute-type",
		})
		return
	}

	attrSpec := attr.Reference.Spec.Result()

	refSpec := containingElementSpecs[0]
	refRule := attrSpec.RuleFor(refSpec)

	typSeen := make(map[attrtype.Type]bool)
	var refTyp attrtype.Type
	if refRule != nil {
		refTyp = refRule.Type.Type
		typSeen[refTyp] = true
	}

	var secondaries []diagnostic.Annotation
	for _, spec := range containingElementSpecs[1:] {
		rule := attrSpec.RuleFor(spec)
		if rule == nil && refTyp == nil {
			continue
		} else if rule != nil && refTyp == rule.Type.Type {
			continue
		}

		// report every attr type once
		if rule != nil && typSeen[rule.Type.Type] {
			continue
		}

		if len(secondaries) == 0 {
			if refRule == nil {
				secondaries = append(secondaries,
					anno.Node(attrSpec.File, attrSpec.AST.Selector, "not defined for `"+refSpec.StylizedHTMLName+"`"))
			} else {
				secondaries = append(secondaries,
					anno.Node(attrSpec.File, refRule.Type, "defined as `"+refTyp.String()+"` for `"+refSpec.StylizedHTMLName+"`"))
			}
		}

		if rule == nil {
			secondaries = append(secondaries,
				anno.Node(attrSpec.File, attrSpec.AST.Selector, "not defined for `"+spec.StylizedHTMLName+"`"))
		} else {
			secondaries = append(secondaries,
				anno.Node(attrSpec.File, rule.Type, "defined as `"+rule.Type.Type.String()+"` for `"+spec.StylizedHTMLName+"`"))
			typSeen[rule.Type.Type] = true
		}
	}

	if len(secondaries) > 0 {
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

	attr.Type.SetResult(refTyp)
	if refTyp != nil || attr.Value.Constant() {
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
