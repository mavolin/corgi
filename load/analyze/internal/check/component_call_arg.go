package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/switches"
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
	} else if len(cc.ComponentArguments) <= 1 {
		return
	}

	args := make(map[file.Identifier][]*file.ComponentArgument, len(cc.ComponentArguments))
	for _, arg := range cc.ComponentArguments {
		if arg.Parameter == nil {
			// already reported by linker
			continue
		}

		args[arg.Name] = append(args[arg.Name], arg)
	}

	for _, dupls := range args {
		if len(dupls) < 2 {
			continue
		}

		logger := logger.With(
			slog.String("arg_name", string(dupls[0].Name)))

		primaries := make([]diagnostic.Annotation, len(dupls))
		for i, dupl := range dupls {
			primaries[i] = anno.Node(cc.File, dupl.AST, "set here")
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
	}

	reported := make(map[file.Identifier]bool)
	for _, arg := range cc.ComponentArguments {
		if reported[arg.Name] {
			continue
		} else if arg.Parameter != nil {
			continue
		}

		logger.Error("Component call argument does not exist",
			slog.String("arg_pos", arg.AST.Start().String()),
			slog.String("arg_name", string(arg.Name)))
		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: argument does not exist",
			Primary: []diagnostic.Annotation{
				anno.Anno(cc.File, anno.Annotation{
					Highlight:  anno.HighlightNode(arg.AST),
					Context:    anno.ContextNode(cc.AST.Header),
					Annotation: "component defines no parameter `" + string(arg.Name) + "`",
				}),
			},
		})
		reported[arg.Name] = true
	}
}

func (ch *checker) CheckRequiredComponentParamsSet(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("required_params_set")

	if cc.Component == nil {
		return
	}

	for _, param := range cc.Component.Parameters {
		if !param.Required() {
			continue
		}

		logger := logger.With(slog.String("param", string(param.Name)))

		arg := cc.ComponentArgumentForParameter(param)
		if arg != nil {
			continue
		}

		// parameter is not set
		logger.Error("Required parameter not set")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: required parameter not set",
			Primary: []diagnostic.Annotation{
				anno.Node(cc.File, cc.AST.Header.Name, "requires parameter `"+string(param.Name)+"` to be set"),
			},
			Explanation: "Parameters with no default, " +
				"like `" + string(param.Name) + "`, " +
				"are required to be set in every component call.",
		})
	}
}

func (ch *checker) CheckComponentAcceptsAttributes(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("component_accepts_attributes")

	if cc.Component == nil {
		return
	}

	if !cc.AcceptsAttributes().False() {
		return
	} else if !cc.ReceivesAttributes.True() && !cc.ReceivesAndPlaceholder.True() {
		return
	}

	var primaries []diagnostic.Annotation
	if cc.ReceivesAttributes.True() {
		chain := cc.ReceivedAttributeChain()
		primaries = make([]diagnostic.Annotation, len(chain))
		for i, n := range chain[:len(chain)-1] {
			primaries[i] = anno.Node(cc.File, n, "through this component call")
		}
		primaries[len(chain)-1] = anno.Node(cc.File, chain[len(chain)-1], "you hand it attributes here")
	} else {
		chain := cc.ReceivedAndPlaceholderChain()
		primaries = make([]diagnostic.Annotation, len(chain))
		for i, n := range chain[:len(chain)-1] {
			primaries[i] = anno.Node(cc.File, n, "through this component call")
		}
		primaries[len(chain)-1] = anno.Node(cc.File, chain[len(chain)-1], "you hand it attributes here")
	}

	logger.Error("Component does not accept attributes")
	diag := &diagnostic.Diagnostic{
		Message: "component call: component does not accept attributes",
		Primary: []diagnostic.Annotation{
			anno.Node(cc.File, cc.ReceivesAttributes.Reason(), "but you hand it attributes here"),
		},
		Explanation: "Components need to specify an &-placeholder somewhere in their body " +
			"for them to accept attributes. Since this component does not specify any " +
			"you cannot hand attributes to it.",
		Docs: "attribute-placeholder",
	}
	couldAcceptAttributes := cc.Component.CouldAcceptAttributes.True()
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
	}

	for _, arg := range cc.ComponentArguments {
		if arg.Parameter == nil {
			return
		}

		if arg.Parameter.AttributeType.Failed() {
			continue
		}

		typ := arg.Parameter.AttributeType.Result()
		if typ != attrtype.Unsafe && typ != attrtype.UnsafeBool {
			continue
		}

		txt, _ := arg.Value.(file.Text)
		if txt == nil {
			return
		}
		for _, n := range txt {
			ok := switches.TextPartR(n,
				func(*file.ComponentCallPart) bool { return false },
				func(file.ConstantPart) bool { return true },
				func(*file.ExpressionPart) bool { return false })
			if ok {
				continue
			}

			logger.Error("Unsafe argument contains interpolation")
			ch.Report(&diagnostic.Diagnostic{
				Message: "component call: unsafe*-typed argument contains interpolation",
				Primary: []diagnostic.Annotation{
					anno.Node(cc.File, arg.AST, "cannot use interpolation here"),
				},
				Secondary: []diagnostic.Annotation{
					anno.Anno(cc.Component.File, anno.Annotation{
						Context:    anno.ContextNode(cc.Component.AST.Header),
						Highlight:  anno.HighlightNode(arg.Parameter.AST),
						Annotation: "typed as `" + typ.String() + "`",
					}),
				},
				Hints: []diagnostic.Hint{
					{
						Hint: "If you are sure, the value you are constructing is safe, " +
							"consult package `safe` to construct a trusted value.",
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
