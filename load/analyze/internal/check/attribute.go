package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/switches"
)

func (ch *checker) CheckAttribute(logger *slog.Logger, f *file.File, attr *file.Attribute) {
	switch attrAST := attr.AST.(type) {
	case *ast.NamedAttribute:
		name := attr.Reference.HTMLName()
		ch.CheckClassAlwaysInnocuous(logger, f, attr, name, attrAST)
		ch.CheckNoInterpolationInUnsafeAttribute(logger, f, attr, attrAST)
		ch.CheckNonBoolAttributeSpecifiedAsBool(logger, f, attr, attrAST)
		ch.CheckBoolAttributeSetToNonBoolExpression(logger, f, attr, attrAST)
		ch.CheckSuperfluousAttributeNameOnAttributeType(logger, f, attrAST)
	}
}

// ============================================================================
// Class Attribute is Always Typed as Innocuous
// ======================================================================================

func (ch *checker) CheckClassAlwaysInnocuous(
	logger *slog.Logger, f *file.File, attr *file.Attribute, name file.Analysis[string], attrAST *ast.NamedAttribute,
) {
	logger = logger.WithGroup("class_not_typed")

	if name.Failed() || name.Result() != "class" {
		return
	}

	isBool := switches.ResolvedAttributeValueR(attr.Value,
		func(file.ConstantBoolAttributeValue) bool { return true },
		func(*file.ExpressionBoolAttributeValue) bool { return true },
		func(file.TextAttributeValue) bool { return false },
		func(*file.UntypedAttributeValue) bool { return false })
	if isBool {
		logger.Error("class attribute incorrectly typed")
		ch.Report(&diagnostic.Diagnostic{
			Message: "attribute: `class` incorrectly typed",
			Primary: []diagnostic.Annotation{
				anno.Node(f, attrAST, "should be `innocuous`, but set to a bool value"),
			},
			Explanation: "The `class` attribute must always be typed as `innocuous`, so that class shorthands " +
				"work as expected.",
		})
		return
	}

	if attr.Type.Failed() {
		return
	}
	typ := attr.Type.Result()
	if typ == attrtype.Innocuous {
		return
	} else if typ == attrtype.Unknown {
		ok := switches.ResolvedAttributeValueR(attr.Value,
			func(file.ConstantBoolAttributeValue) bool { return false },
			func(*file.ExpressionBoolAttributeValue) bool { return false },
			// If the value is constant, we can use it as a class attribute.
			// If the value is not, we already captured an error elsewhere
			// asserting that the attribute must be typed
			func(file.TextAttributeValue) bool { return true },
			func(*file.UntypedAttributeValue) bool { return true })
		if ok {
			return
		}
	}

	logger.Error("class attribute incorrectly typed")

	// Generate a good error message

	if attrAST.Value != nil {
		tval, _ := attrAST.Value.(*ast.TypedAttributeValue)
		if tval != nil {
			ch.Report(&diagnostic.Diagnostic{
				Message: "attribute: `class` incorrectly typed",
				Primary: []diagnostic.Annotation{
					anno.Node(f, tval.Type, "should be `innocuous`, but is `"+typ.String()+"`"),
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
				anno.Node(spec.File, rule.Type, "but defined here as `"+typ.String()+"`"),
			}
		} else {
			secondaries = []diagnostic.Annotation{
				anno.Node(spec.File, spec.AST, "but defined here as `"+typ.String()+"`"),
			}
		}
		ch.Report(&diagnostic.Diagnostic{
			Message: "attribute: `class` incorrectly typed",
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
			anno.Node(f, attrAST.Name, "typed by analyzer as `"+typ.String()+"`"),
		},
		Explanation: "Using the result of the analysis, which is safe, but the error following this will be confusing.\n" +
			"This might've occurred because the AST was extended, but the analyzer wasn't subsequently updated (correctly).\n" +
			"\n" +
			"You should not see this error, please open an issue, this is a bug in the analyzer.",
	})

	ch.Report(&diagnostic.Diagnostic{
		Message: "attribute: `class` incorrectly typed",
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

func (ch *checker) CheckNoInterpolationInUnsafeAttribute(logger *slog.Logger, f *file.File, attr *file.Attribute, attrAST *ast.NamedAttribute) {
	logger = logger.WithGroup("no_interpolation_in_unsafe")

	if !attr.Type.Equal(attrtype.Unsafe) {
		return
	}

	ok := switches.ResolvedAttributeValueR(attr.Value,
		func(file.ConstantBoolAttributeValue) bool { return true },    // different error
		func(*file.ExpressionBoolAttributeValue) bool { return true }, // different error
		func(val file.TextAttributeValue) bool { return val.Constant() },
		func(*file.UntypedAttributeValue) bool { return false })
	if ok {
		return
	}

	expr := ch.expressionFromAttributeValue(attrAST.Value)
	if expr == nil || len(expr.Nodes) != 1 {
		return
	}

	s, _ := expr.Nodes[0].(*ast.String)
	if s == nil {
		return
	}

	for _, n := range s.Contents {
		ok := switches.StringNodeR(n,
			func(*ast.BadInterpolation) bool { panic("analyzer called with parser errors") },
			func(*ast.CharacterEscape) bool { return true },
			func(*ast.CharacterReference) bool { return true },
			func(*ast.ComponentCallInterpolation) bool { return false },
			func(*ast.ExpressionInterpolation) bool { return false },
			func(*ast.StringText) bool { return true })
		if ok {
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

func (ch *checker) CheckNonBoolAttributeSpecifiedAsBool(logger *slog.Logger, f *file.File, attr *file.Attribute, attrAST *ast.NamedAttribute) {
	logger = logger.WithGroup("non_bool_attribute_specified_as_bool")

	if attr.Type.Failed() {
		return
	}

	typ := attr.Type.Result()
	if typ == attrtype.Bool || typ == attrtype.UnsafeBool || typ == attrtype.Unknown {
		return
	}

	ok := switches.ResolvedAttributeValueR(attr.Value,
		func(file.ConstantBoolAttributeValue) bool { return false },
		func(*file.ExpressionBoolAttributeValue) bool { return false },
		func(file.TextAttributeValue) bool { return true },
		func(*file.UntypedAttributeValue) bool { return true })
	if ok {
		return
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
		Message: "non-bool attribute set to `bool` value",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attrAST.Name, "set to a bool value"),
		},
		Secondary: secondaries,
	})
}

// ============================================================================
// Bool Attribute Set to Non-bool Expression
// ======================================================================================

func (ch *checker) CheckBoolAttributeSetToNonBoolExpression(logger *slog.Logger, f *file.File, attr *file.Attribute, attrAST *ast.NamedAttribute) {
	logger = logger.WithGroup("bool_attribute_set_to_non_bool_expression")

	ok := switches.ResolvedAttributeValueR(attr.Value,
		func(file.ConstantBoolAttributeValue) bool { return true },
		func(*file.ExpressionBoolAttributeValue) bool { return true },
		func(file.TextAttributeValue) bool { return false },
		func(*file.UntypedAttributeValue) bool { return true })
	if ok {
		return
	}

	if attr.Type.Failed() {
		return
	}

	typ := attr.Type.Result()
	if typ != attrtype.Bool && typ != attrtype.UnsafeBool {
		return
	}

	var secondaries []diagnostic.Annotation
	if attr.Reference.Spec.NotZero() {
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
			anno.Node(f, attrAST.Value, "this expression does not yield a `bool`, but is used for a bool attribute"),
		},
		Secondary: secondaries,
	})
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

func (ch *checker) expressionFromAttributeValue(v ast.AttributeValue) *ast.Expression {
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
// Constantly False Boolean Attribute
// ======================================================================================

func (ch *checker) CheckConstantlyFalseBooleanAttribute(logger *slog.Logger, f *file.File, attr *file.Attribute, attrAST *ast.NamedAttribute) {
	logger = logger.WithGroup("constantly_false_boolean_attribute")

	c, ok := attr.Value.(file.ConstantBoolAttributeValue)
	if !ok || bool(c) {
		return
	}

	logger.Error("Constantly false boolean attribute")
	ch.Report(&diagnostic.Diagnostic{
		Message: "attribute: constantly `false` value",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attrAST, "this attribute is always false"),
		},
		Explanation: "A boolean attribute that is always `false` is pointless, " +
			"because it is never printed in the output HTML.",
	})
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
