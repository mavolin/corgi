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
	logger.Debug("Checking arguments")

	for _, arg := range a.Args {
		logger := logger.With(
			slog.String("arg_pos", arg.Start().String()),
			slog.String("arg_type", fmt.Sprintf("%T", arg)))
		logger.Debug("Checking argument")

		switch arg := arg.(type) {
		case *ast.NamedAttribute:
			ch.CheckClassAlwaysInnocuous(logger, f, arg)
			ch.CheckNoInterpolationInUnsafeAttribute(logger, f, arg)
			ch.CheckDefinedNonBoolAttributeSpecifiedAsBool(logger, f, arg)
			ch.CheckBoolAttributeSetToNonBoolExpression(logger, f, arg)
		}
	}
}

func (ch *checker) CheckClassAlwaysInnocuous(logger *slog.Logger, f *file.File, attr *ast.NamedAttribute) {
	logger = logger.WithGroup("class_not_typed")
	logger.Debug("Checking that class attribute is not typed")

	ref := f.AttributeReferenceByNode(attr.Name)
	switch {
	case ref.AnalyzedWithErrors:
		logger.Debug("Attribute analyzed with errors, skipping")
		return
	case ref.Type == attrtype.Innocuous:
		logger.Debug("Attribute is innocuous (and possibly not even a class attribute), skipping")
		return
	case ref.Name() != "class":
		logger.Debug("Attribute is not a class attribute, skipping")
		return
	}

	logger.Error("class attribute wrongly typed")

	// Generate a good error message

	if attr.Value != nil {
		tval, _ := attr.Value.(*ast.TypedAttributeValue)
		if tval != nil {
			if tval.Type.Name.Type != ref.Type {
				logger.Warn("Resolved type and explicit type do not match. " +
					"Analysis is correct, but the reported error will be confusing. " +
					"This shouldn't happen, please open an issue.")
			}

			ch.Report(&diagnostic.Diagnostic{
				Message: "class attribute wrongly typed",
				Primary: []diagnostic.Annotation{
					anno.Node(f, tval.Type, "should be `innocuous`, but is `"+tval.Type.Name.Type.String()+"`"),
				},
				Explanation: "The `class` attribute must always be typed as `innocuous`, so that class shorthands" +
					"work as expected.",
			})
			return
		}
	}

	if ref.Rule != nil {
		logger := logger.With(slog.String("rule_pos", ref.Rule.Start().String()))
		logger.Debug("Found attribute definition")

		if ref.Rule.Type.Type != ref.Type {
			logger.Warn("Resolved type and definition type do not match (found no explicit typing). " +
				"Analysis is correct, but the reported error will be confusing. " +
				"This shouldn't happen, please open an issue.")
			return
		}

		ch.Report(&diagnostic.Diagnostic{
			Message: "class attribute wrongly typed",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attr.Name, "not an `innocuous` attribute"),
				anno.Node(ref.Spec.File, ref.Rule, "but defined here as `"+ref.Rule.Type.Type.String()+"`"),
			},
			Explanation: "The `class` attribute must always be typed as `innocuous`, so that class shorthands" +
				" work as expected.",
		})
		return
	}

	// AST was extended (or someone called analyze before linking/with linker errors)
	logger.Warn("Attribute is neither explicitly typed not has a corresponding attribute definition. " +
		"Falling back to reporting just the incorrect resolved type. " +
		"This is safe, but shouldn't happen. Please open an issue.")

	ch.Report(&diagnostic.Diagnostic{
		Message: "class attribute wrongly typed",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attr.Value, "should be `innocuous`"),
		},
		Explanation: "The `class` attribute must always be typed as `innocuous`, so that class shorthands" +
			" work as expected.",
	})
}

func (ch *checker) CheckNoInterpolationInUnsafeAttribute(logger *slog.Logger, f *file.File, attr *ast.NamedAttribute) {
	logger = logger.WithGroup("no_interpolation_in_unsafe")
	logger.Debug("Checking for expression/component call interpolation in unsafe attribute")

	if attr.Value == nil {
		logger.Debug("Attribute has no value, skipping")
		return
	}

	ref := f.AttributeReferenceByNode(attr.Name)
	if ref.AnalyzedWithErrors {
		logger.Debug("Attribute analyzed with errors, skipping")
		return
	} else if ref.Type != attrtype.Unsafe {
		logger.Debug("Attribute is not unsafe, skipping")
		return
	}

	var expr *ast.Expression
	val := attr.Value
Loop:
	for {
		switch typed := val.(type) {
		case *ast.TypedAttributeValue:
			val = typed.Value
		case *ast.ExpressionAttributeValue:
			expr = (*ast.Expression)(typed)
			break Loop
		default:
			logger.Error("Attribute value is neither an expression nor typed attribute value")
			ch.Report(&diagnostic.Diagnostic{
				Message: "internal error: attribute value is neither an expression nor typed attribute value",
				Primary: []diagnostic.Annotation{
					anno.Node(f, attr.Value, "for this node"),
				},
				Explanation: "This is a bug in the analyzer, please open an issue.",
			})
		}
	}

	if len(expr.Nodes) != 1 {
		logger.Debug("Attribute value is not a single expression, skipping")
		return
	}

	s, _ := expr.Nodes[0].(*ast.String)
	if s == nil {
		logger.Debug("Not a string literal, skipping")
		return
	}

	for _, c := range s.Contents {
		switch c.(type) {
		case *ast.ExpressionInterpolation:
			logger.Error("Unsafe attribute contains interpolation")
			ch.Report(&diagnostic.Diagnostic{
				Message: "unsafe attribute contains interpolation",
				Primary: []diagnostic.Annotation{
					anno.Node(f, attr.Value, "cannot use interpolation here"),
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
		case *ast.ComponentCallInterpolation:
			logger.Error("Unsafe attribute contains interpolation")
			ch.Report(&diagnostic.Diagnostic{
				Message: "unsafe attribute contains interpolation",
				Primary: []diagnostic.Annotation{
					anno.Node(f, attr.Value, "cannot use interpolation here"),
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
}

func (ch *checker) CheckDefinedNonBoolAttributeSpecifiedAsBool(logger *slog.Logger, f *file.File, attr *ast.NamedAttribute) {
	logger = logger.WithGroup("defined_non_bool_attribute_specified_as_bool")
	logger.Debug("Checking that a defined attribute, declared as non-bool, isn't specified as a bool shorthand")

	if attr.Value != nil {
		logger.Debug("Not a bool shorthand, skipping")
		return
	}

	ref := f.AttributeReferenceByNode(attr.Name)
	if ref.AnalyzedWithErrors {
		logger.Debug("Attribute analyzed with errors, skipping")
		return
	} else if ref.Spec == nil {
		logger.Debug("Attribute is not a defined attribute, skipping")
		return
	}

	if ref.Type == attrtype.Bool || ref.Type == attrtype.UnsafeBool {
		logger.Debug("Attribute is a bool shorthand and of a bool type, skipping")
		return
	}

	logger.Error("Non-bool attribute set as using bool shorthand")
	ch.Report(&diagnostic.Diagnostic{
		Message: "non-bool attribute set using bool shorthand",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attr.Name, "this attribute is not a bool attribute and can henceforth not be set using a bool shorthand"),
		},
		Secondary: []diagnostic.Annotation{
			anno.Node(ref.Spec.File, ref.Rule, "attribute type defined here as `"+ref.Type.String()+"`"),
		},
	})
}

func (ch *checker) CheckBoolAttributeSetToNonBoolExpression(logger *slog.Logger, f *file.File, attr *ast.NamedAttribute) {
	logger = logger.WithGroup("bool_attribute_set_to_non_bool_expression")
	logger.Debug("Checking that a bool attribute isn't set to a non-bool-yielding expression")

	if attr.Value == nil {
		logger.Debug("Attribute has no value, skipping")
		return
	}

	ref := f.AttributeReferenceByNode(attr.Name)
	if ref.AnalyzedWithErrors {
		logger.Debug("Attribute analyzed with errors, skipping")
		return
	} else if ref.Type != attrtype.Bool && ref.Type != attrtype.UnsafeBool {
		logger.Debug("Attribute is not a bool attribute, skipping")
		return
	}

	var expr *ast.Expression
	val := attr.Value
Loop:
	for {
		switch typed := val.(type) {
		case *ast.TypedAttributeValue:
			val = typed.Value
		case *ast.ExpressionAttributeValue:
			expr = (*ast.Expression)(typed)
			break Loop
		default:
			logger.Error("Attribute value is neither an expression nor typed attribute value")
			ch.Report(&diagnostic.Diagnostic{
				Message: "internal error: attribute value is neither an expression nor typed attribute value",
				Primary: []diagnostic.Annotation{
					anno.Node(f, attr.Value, "for this node"),
				},
				Explanation: "This is a bug in the analyzer, please open an issue.",
			})
		}
	}

	t, _ := file.InferType(f, expr)
	// We can only judge types we know
	switch t {
	case "bool":
		logger.Debug("Expression yields a bool, skipping")
	case "int", "float", "string", "any", "interface{}":
		logger.Error("Bool attribute set to non-bool expression")
		ch.Report(&diagnostic.Diagnostic{
			Message: "bool attribute set to non-bool expression",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attr.Value, "this expression does not yield a bool, but is used for a bool attribute"),
			},
			Secondary: []diagnostic.Annotation{
				anno.Node(ref.Spec.File, ref.Rule, "attribute type defined here as `"+ref.Type.String()+"`"),
			},
		})
	default:
		logger.Debug("Expression yields an unknown type, skipping")
	}
}
