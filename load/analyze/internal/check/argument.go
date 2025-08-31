package check

import (
	"fmt"
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckArguments(logger *slog.Logger, f *file.File, _ []*walk.Context, a *ast.Arguments) {
	logger = logger.WithGroup("arguments").
		With(slog.String("arguments_pos", a.Start().String()))

	for _, arg := range a.List {
		logger := logger.With(
			slog.String("arg_pos", arg.Start().String()),
			slog.String("arg_type", fmt.Sprintf("%T", arg)))

		switch arg := arg.(type) {
		case *ast.NamedAttribute:
			ch.CheckClassAlwaysInnocuous(logger, f, arg)
			ch.CheckNoInterpolationInUnsafeAttribute(logger, f, arg)
			ch.CheckNonBoolAttributeSpecifiedAsBool(logger, f, arg)
			ch.CheckBoolAttributeSetToNonBoolExpression(logger, f, arg)
			ch.CheckSuperfluousAttributeNameOnAttributeType(logger, f, arg)
		}
	}
}

// ============================================================================
// Nested Typed Attribute Values
// ======================================================================================

func (ch *checker) CheckNestedTypedAttributeValues(logger *slog.Logger, f *file.File, a *ast.NamedAttribute) {
	logger = logger.WithGroup("nested_typed_attribute_values")

	if a.Value == nil {
		return
	}

	tval, _ := a.Value.(*ast.TypedAttributeValue)
	if tval == nil {
		return
	}

	start := tval.Value.Start()
	end := tval.Value.End()
	for {
		tval, _ = tval.Value.(*ast.TypedAttributeValue)
		if tval == nil {
			break
		}
		end = *tval.LParen
	}
	if start == end {
		return
	}

	logger.Error("Nested typed attribute values")
	ch.Report(&diagnostic.Diagnostic{
		Message: "attribute: nested typed attribute values",
		Primary: []diagnostic.Annotation{
			anno.Range(f, start, end, "expected only a single typed attribute value"),
		},
	})
}

// ============================================================================
// Class Attribute is Always Typed as Innocuous
// ======================================================================================

func (ch *checker) CheckClassAlwaysInnocuous(logger *slog.Logger, f *file.File, attrAST *ast.NamedAttribute) {
	logger = logger.WithGroup("class_not_typed")

	attr := f.AttributeByNode(attrAST)
	if attr.Type.Failed() || attr.Type.Result() == attrtype.Innocuous {
		return
	} else if htmlName := attr.Reference.HTMLName(); htmlName.Failed() || htmlName.Result() != "class" {
		return
	}

	if attr.Type.Result() == attrtype.Unknown && attr.StaticValue.NotZero() {
		if _, ok := attr.StaticValue.Result().(file.TextualAttributeValue); ok {
			return
		}

		logger.Error("class attribute wrongly typed")
		ch.Report(&diagnostic.Diagnostic{
			Message: "class attribute wrongly typed",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attrAST.Name, "should be `innocuous`"),
			},
			Secondary: []diagnostic.Annotation{
				anno.Node(f, attrAST, "set to a boolean value here"),
			},
			Explanation: "The `class` attribute must always be typed as `innocuous`, so that class shorthands " +
				"work as expected.",
		})
		return
	}

	logger.Error("class attribute wrongly typed")

	// Generate a good error message

	if attrAST.Value != nil {
		tval, _ := attrAST.Value.(*ast.TypedAttributeValue)
		if tval != nil {
			if tval.Type.Name.Type != attr.Type.Result() {
				ch.Report(&diagnostic.Diagnostic{
					Type:    diagnostic.InternalError,
					Message: "analyze.CheckClassAlwaysInnocuous: resolved type and explicit type do not match",
					Primary: []diagnostic.Annotation{
						anno.Node(f, attrAST.Name, "typed by analyzer as `"+attr.Type.Result().String()+"`"),
						anno.Node(f, tval.Type, "explicitly typed as `"+tval.Type.Name.Type.String()+"`"),
					},
					Explanation: "Using the result of the analysis, which is safe, but the error following this will be confusing.\n" +
						"\n" +
						"You should not see this error, please open an issue, this is a bug in the analyzer.",
				})
			}

			ch.Report(&diagnostic.Diagnostic{
				Message: "class attribute wrongly typed",
				Primary: []diagnostic.Annotation{
					anno.Node(f, tval.Type, "should be `innocuous`, but is `"+tval.Type.Name.Type.String()+"`"),
				},
				Explanation: "The `class` attribute must always be typed as `innocuous`, so that class shorthands " +
					"work as expected.",
			})
			return
		}
	}

	if attr.Reference.Spec.NotZero() {
		spec := attr.Reference.Spec.Result()

		var secondaries []diagnostic.Annotation
		if rule := singleAttributeRule(attr); rule != nil {
			secondaries = []diagnostic.Annotation{
				anno.Node(spec.File, rule.Type, "but defined here as `"+attr.Type.Result().String()+"`"),
			}
		} else {
			secondaries = []diagnostic.Annotation{
				anno.Node(spec.File, spec.AST, "but defined here as `"+attr.Type.Result().String()+"`"),
			}
		}
		ch.Report(&diagnostic.Diagnostic{
			Message: "class attribute wrongly typed",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attrAST.Name, "should be `innocuous`"),
			},
			Secondary: secondaries,
			Explanation: "The `class` attribute must always be typed as `innocuous`, so that class shorthands " +
				"work as expected.",
		})
		return
	}

	// AST types were extended?
	ch.Report(&diagnostic.Diagnostic{
		Type:    diagnostic.InternalError,
		Message: "analyze.CheckClassAlwaysInnocuous: attribute is neither explicitly typed nor attached to a definition",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attrAST.Name, "typed by analyzer as `"+attr.Type.Result().String()+"`"),
		},
		Explanation: "Using the result of the analysis, which is safe, but the error following this will be confusing.\n" +
			"This might've occurred because the AST was extended, but the analyzer wasn't subsequently updated (correctly).\n" +
			"\n" +
			"You should not see this error, please open an issue, this is a bug in the analyzer.",
	})

	ch.Report(&diagnostic.Diagnostic{
		Message: "class attribute wrongly typed",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attrAST.Value, "should be `innocuous`"),
		},
		Explanation: "The `class` attribute must always be typed as `innocuous`, so that class shorthands " +
			"work as expected.",
	})
}

// ============================================================================
// No Interpolation in an Unsafe Attribute Value
// ======================================================================================

func (ch *checker) CheckNoInterpolationInUnsafeAttribute(logger *slog.Logger, f *file.File, attrAST *ast.NamedAttribute) {
	logger = logger.WithGroup("no_interpolation_in_unsafe")

	if attrAST.Value == nil {
		return
	}

	attr := f.AttributeByNode(attrAST)
	if !attr.Type.Equal(attrtype.Unsafe) {
		return
	}

	expr := ch.expressionFromAttributeValue(logger, f, attrAST.Value)
	if expr == nil || len(expr.Nodes) != 1 {
		return
	}

	s, _ := expr.Nodes[0].(*ast.String)
	if s == nil {
		return
	}

	for _, n := range s.Contents {
		switch n.(type) {
		case *ast.ExpressionInterpolation:
		case *ast.ComponentCallInterpolation:
		default:
			continue
		}

		logger.Error("Unsafe attribute contains interpolation")
		ch.Report(&diagnostic.Diagnostic{
			Message: "unsafe attribute contains interpolation",
			Primary: []diagnostic.Annotation{
				anno.Node(f, n, "cannot use interpolation here"),
			},
			Hints: []diagnostic.Hint{
				{
					Hint: "If you are sure, the value you are constructing is safe, " +
						"wrap the entire expression in a `safe.TrustedUnsafe` call. " +
						"Make sure to read the documentation of `safe.TrustedUnsafe` before doing so!",
				},
			},
			Explanation: "Attributes marked as `unsafe` must not use any interpolation, " +
				"even if the expression you are interpolating is of type `safe.Unsafe`. " +
				"Either the entire attribute value must be a `safe.Unsafe` value, " +
				"or it must be a literal.",
		})
	}
}

// ============================================================================
// Non-bool Attribute Set Using Bool Shorthand
// ======================================================================================

func (ch *checker) CheckNonBoolAttributeSpecifiedAsBool(logger *slog.Logger, f *file.File, attrAST *ast.NamedAttribute) {
	logger = logger.WithGroup("defined_non_bool_attribute_specified_as_bool")

	attr := f.AttributeByNode(attrAST)
	if attr.Type.Failed() {
		return
	}

	typ := attr.Type.Result()
	if typ == attrtype.Bool || typ == attrtype.UnsafeBool || typ == attrtype.Unknown {
		return
	}

	// if this is not a bool shorthand, try to see if the expression yields a bool
	boolShorthand := attrAST.Value == nil
	if !boolShorthand {
		typ, _ := file.InferType(f, ch.expressionFromAttributeValue(logger, f, attrAST.Value))
		if typ != "bool" {
			return
		}
	}

	var secondaries []diagnostic.Annotation
	if attr.Reference.Spec.NotZero() {
		spec := attr.Reference.Spec.Result()
		if rule := singleAttributeRule(attr); rule != nil {
			secondaries = []diagnostic.Annotation{
				anno.Node(spec.File, rule, "attribute type defined here as `"+typ.String()+"`"),
			}
		} else {
			secondaries = []diagnostic.Annotation{
				anno.Node(spec.File, spec.AST, "attribute type defined here as `"+typ.String()+"`"),
			}
		}
	} else {
		ch.Report(&diagnostic.Diagnostic{
			Type:    diagnostic.InternalError,
			Message: "analyze.CheckNonBoolAttributeSpecifiedAsBool: attribute not attached to definition",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attrAST.Name, "typed by analyzer as `"+typ.String()+"`"),
			},
			Explanation: "Using the result of the analysis, which is safe, but the error following this will be confusing.\n" +
				"\n" +
				"Either Spec or Rule are not set, the attribute is not explicitly typed, but the analyzer still " +
				"managed to infer a type for it. " +
				"Most likely, the AST was extended, but the analyzer wasn't subsequently updated (correctly).\n" +
				"\n" +
				"You should not see this error, please open an issue, this is a bug in the analyzer.",
		})
	}

	logger.Error("Non-bool attribute set to bool value")
	ch.Report(&diagnostic.Diagnostic{
		Message: "non-bool attribute set to bool value",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attrAST.Name, "not a bool attribute"),
		},
		Secondary: secondaries,
	})
}

// ============================================================================
// Bool Attribute Set to Non-bool Expression
// ======================================================================================

func (ch *checker) CheckBoolAttributeSetToNonBoolExpression(logger *slog.Logger, f *file.File, attrAST *ast.NamedAttribute) {
	logger = logger.WithGroup("bool_attribute_set_to_non_bool_expression")

	if attrAST.Value == nil {
		return
	}

	attr := f.AttributeByNode(attrAST)
	if attr.Type.Failed() {
		return
	}

	typ := attr.Type.Result()
	if typ != attrtype.Bool && typ != attrtype.UnsafeBool {
		return
	}
	if attr.StaticValue.NotZero() {
		if _, ok := attr.StaticValue.Result().(file.BoolAttributeValue); ok {
			return
		}
	}

	expr := ch.expressionFromAttributeValue(logger, f, attrAST.Value)
	if expr == nil {
		return
	}

	t, _ := file.InferType(f, expr)
	// We can only judge types we know
	switch t {
	case "bool":
	case "int", "float", "string", "any", "interface{}":
		var secondaries []diagnostic.Annotation
		if tav, _ := attrAST.Value.(*ast.TypedAttributeValue); tav != nil {
			secondaries = []diagnostic.Annotation{
				anno.Node(f, tav.Type, "explicitly typed as `"+tav.Type.Name.Type.String()+"`"),
			}
		} else if attr.Reference.Spec.NotZero() {
			spec := attr.Reference.Spec.Result()
			if rule := singleAttributeRule(attr); rule != nil {
				secondaries = []diagnostic.Annotation{
					anno.Node(spec.File, rule.Type, "attribute type defined here as `"+typ.String()+"`"),
				}
			} else {
				secondaries = []diagnostic.Annotation{
					anno.Node(spec.File, spec.AST, "attribute type defined here as `"+typ.String()+"`"),
				}
			}
		}
		logger.Error("Bool attribute set to non-bool expression")
		ch.Report(&diagnostic.Diagnostic{
			Message: "bool attribute set to non-bool expression",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attrAST.Value, "this expression does not yield a bool, but is used for a bool attribute"),
			},
			Secondary: secondaries,
		})
	}
}

// singleAttributeRule returns the attribute rule that is used for all
// containing elements of the given attribute, if there is one.
func singleAttributeRule(attr *file.Attribute) *ast.AttributeRule {
	if !attr.Reference.Spec.NotZero() || attr.ContainingElements.Failed() {
		return nil
	}

	spec := attr.Reference.Spec.Result()

	var rule *ast.AttributeRule
	for _, elem := range *attr.ContainingElements.Result() {
		if elem.Element.Spec == nil {
			return nil
		} else if rule == nil {
			rule = spec.RuleFor(elem.Element.Spec)
		} else {
			rule2 := spec.RuleFor(elem.Element.Spec)
			if rule != rule2 {
				return nil
			}
		}
	}
	return rule
}

func (ch *checker) expressionFromAttributeValue(logger *slog.Logger, f *file.File, v ast.AttributeValue) *ast.Expression {
	var expr *ast.Expression
Loop:
	for {
		switch typed := v.(type) {
		case *ast.TypedAttributeValue:
			v = typed.Value
		case *ast.ExpressionAttributeValue:
			expr = (*ast.Expression)(typed)
			break Loop
		default:
			logger.Error("Attribute value is neither an expression nor typed attribute value")
			ch.Report(&diagnostic.Diagnostic{
				Type:    diagnostic.InternalError,
				Message: "attribute value is neither an expression nor typed attribute value",
				Primary: []diagnostic.Annotation{
					anno.Node(f, v, "for this node"),
				},
				Explanation: "This most likely happened because the ast.AttributeValue sum type was extended.\n\n" +
					"This is a bug in the analyzer, please open an issue.",
			})
		}
	}
	return expr
}

// ============================================================================
// Superfluous Attribute Name Attached To Attribute Type
// ======================================================================================

func (ch *checker) CheckSuperfluousAttributeNameOnAttributeType(logger *slog.Logger, f *file.File, a *ast.NamedAttribute) {
	logger = logger.WithGroup("superfluous_attribute_name_on_attribute_type")

	tv, _ := a.Value.(*ast.TypedAttributeValue)
	if tv == nil {
		return
	}

	if tv.Type.Attribute == nil {
		return
	}

	if tv.Type.Name.Type != attrtype.Unsafe && tv.Type.Name.Type != attrtype.UnsafeBool {
		logger.Error("Attribute name on non-unsafe type")
		ch.Report(&diagnostic.Diagnostic{
			Message: "attribute type: attribute name on non-unsafe type",
			Primary: []diagnostic.Annotation{
				anno.Range(f, *tv.Type.LBracket, *tv.Type.RBracket, "remove this attribute name"),
			},
			Hints: []diagnostic.Hint{
				{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
			},
			Explanation: "Attribute types other than `unsafe` and `unsafeBool` " +
				"do not need to be tied to a specific attribute.",
		})
		return
	}

	logger.Error("Superfluous attribute name on attribute type")
	ch.Report(&diagnostic.Diagnostic{
		Message: "attribute type: superfluous attribute name",
		Primary: []diagnostic.Annotation{
			anno.Range(f, *tv.Type.LBracket, *tv.Type.RBracket, "remove this attribute name"),
		},
		Hints: []diagnostic.Hint{
			{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
		},
		Explanation: "Attribute names need not be specified a second time " +
			"in brackets when writing a typed named attribute.",
	})
}
