package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (ch *checker) CheckComponentCallArguments(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("arguments")

	ch.CheckComponentArgsExist(logger, cc)
	ch.CheckNoDuplicateComponentArgs(logger, cc)
	ch.CheckRequiredComponentParamsSet(logger, cc)
	ch.CheckComponentAcceptsAttributes(logger, cc)
	ch.CheckNoInterpolationInUnsafeTypedArguments(logger, cc)
}

func (ch *checker) CheckNoDuplicateComponentArgs(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("no_duplicate_args")

	if cc.Component == nil {
		return
	} else if cc.AST.Header.Arguments == nil || len(cc.AST.Header.Arguments.List) <= 1 {
		return
	}

	args := make(map[string][]*ast.ComponentArgument, len(cc.AST.Header.Arguments.List))
	for _, arg := range cc.AST.Header.Arguments.List {
		carg, _ := arg.(*ast.ComponentArgument)
		if carg == nil {
			continue
		}
		if cc.Component.ParameterByName(carg.Name.Name) == nil {
			// non-existent arguments are handled by CheckComponentArgsExist
			continue
		}

		name := carg.Name.Name
		args[name] = append(args[name], carg)
	}

	for _, dupls := range args {
		if len(dupls) < 2 {
			continue
		}

		logger := logger.With(
			slog.String("arg_name", dupls[0].Name.Name))

		primaries := make([]diagnostic.Annotation, len(dupls))
		for i, dupl := range dupls {
			primaries[i] = anno.Node(cc.File, dupl, "set here")
		}

		logger.Error("Found duplicate component call argument")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: argument specified multiple times",
			Primary: primaries,
		})
	}
}

func (ch *checker) CheckComponentArgsExist(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("args_exist")

	if cc.Component == nil {
		return
	} else if cc.AST.Header.Arguments == nil || len(cc.AST.Header.Arguments.List) == 0 {
		return
	}

	reported := make(map[string]bool)
	for _, arg := range cc.AST.Header.Arguments.List {
		carg, _ := arg.(*ast.ComponentArgument)
		if carg == nil {
			continue
		}

		name := carg.Name.Name
		if reported[name] {
			continue
		} else if cc.Component.ParameterByName(name) != nil {
			continue
		}

		logger.Error("Component call argument does not exist",
			slog.String("arg_pos", carg.Start().String()),
			slog.String("arg_name", name))
		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: argument does not exist",
			Primary: []diagnostic.Annotation{
				anno.Anno(cc.File, anno.Annotation{
					Highlight:  anno.HighlightNode(carg.Name),
					Context:    anno.ContextNode(cc.AST.Header),
					Annotation: "component defines no parameter `" + name + "`",
				}),
			},
		})
		reported[name] = true
	}
}

func (ch *checker) CheckRequiredComponentParamsSet(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("required_params_set")

	if cc.Component == nil {
		return
	}

Params:
	for _, param := range cc.Component.Parameters {
		if !param.Required() {
			continue
		}

		logger := logger.With(slog.String("param", param.AST.Name.Name))

		for _, arg := range cc.AST.Header.Arguments.List {
			carg, _ := arg.(*ast.ComponentArgument)
			if carg == nil {
				continue
			}

			if carg.Name.Name == param.AST.Name.Name {
				continue Params // parameter is set
			}
		}

		// parameter is not set
		logger.Error("Required parameter not set")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: required parameter not set",
			Primary: []diagnostic.Annotation{
				anno.Node(cc.File, cc.AST.Header.Name, "requires parameter `"+param.AST.Name.Name+"` to be set"),
			},
			Explanation: "Parameters with no default, " +
				"like `" + param.AST.Name.Name + "`, " +
				"are required to be set in every component call.",
		})
	}
}

func (ch *checker) CheckComponentAcceptsAttributes(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("component_accepts_attributes")

	if cc.Component == nil {
		return
	}

	if cc.AcceptsAttributes.Equal(true) {
		return
	}

	if cc.FirstDelegatedAttributeWriter.Equal(nil) && cc.FirstDelegatedAndPlaceholderWriter.Equal(nil) {
		return
	}

	primaries := make([]diagnostic.Annotation, 1)
	if cc.FirstDelegatedAttributeWriter.NotZero() { // todo
		primaries[0] = anno.Node(cc.File, cc.FirstDelegatedAttributeWriter.Result, "but you hand it attributes here")
	} else {
		primaries[0] = anno.Node(cc.File, cc.FirstDelegatedAndPlaceholderWriter.Result, "but you hand it attributes here")
	}

	logger.Error("Component does not accept attributes")
	diag := &diagnostic.Diagnostic{
		Message: "component call: component does not accept attributes",
		Primary: []diagnostic.Annotation{
			anno.Node(cc.File, cc.FirstDelegatedAttributeWriter.Result, "but you hand it attributes here"),
		},
		Explanation: "Components need to specify an &-placeholder somewhere in their body " +
			"for them to accept attributes. Since this component does not specify any " +
			"you cannot hand attributes to it.",
		Docs: "attribute-placeholder",
	}
	couldAcceptAttributes := cc.Component.CouldAcceptAttributes.NotZero()
	if couldAcceptAttributes {
		diag.Hints = []diagnostic.Hint{
			{
				Hint: "The only &-placeholders of this component are specified in defaults of blocks, " +
					"that you are overwriting. " +
					"Perhaps, you could add the attributes in those blocks directly, " +
					"to achieve the same result?",
			},
		}
	}
	ch.Report(diag)
}

func (ch *checker) CheckNoInterpolationInUnsafeTypedArguments(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("no_interpolation_in_unsafe_typed_args")

	if cc.Component == nil {
		return
	} else if cc.AST.Header.Arguments == nil {
		return
	}

	for _, arg := range cc.AST.Header.Arguments.List {
		carg, _ := arg.(*ast.ComponentArgument)
		if carg == nil {
			continue
		}

		param := cc.Component.ParameterByName(carg.Name.Name)
		if param.AttributeType.Failed {
			continue
		} else if param.AttributeType.Result != attrtype.Unsafe && param.AttributeType.Result != attrtype.UnsafeBool {
			continue
		}

		s, _ := carg.Value.Nodes[0].(*ast.String)
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

			logger.Error("Unsafe argument contains interpolation")
			ch.Report(&diagnostic.Diagnostic{
				Message: "component call: unsafe*-typed argument contains interpolation",
				Primary: []diagnostic.Annotation{
					anno.Node(cc.File, n, "cannot use interpolation here"),
				},
				Secondary: []diagnostic.Annotation{
					anno.Anno(cc.Component.File, anno.Annotation{
						Context:    anno.ContextNode(cc.Component.AST.Header),
						Highlight:  anno.HighlightNode(param.AST),
						Annotation: "typed as `" + param.AttributeType.Result.String() + "`",
					}),
				},
				Hints: []diagnostic.Hint{
					{
						Hint: "If you are sure, the value you are constructing is safe, " +
							"wrap the entire expression in a `safe.TrustedUnsafe` call. " +
							"Make sure to read the documentation of `safe.TrustedUnsafe` before doing so!",
					},
				},
				Explanation: "Arguments marked as `unsafe` must not use any interpolation, " +
					"even if the expression you are interpolating is of type `safe.Unsafe`. " +
					"Either the entire argument value must be a `safe.Unsafe` value, " +
					"or it must be a literal.",
			})
		}
	}
}
