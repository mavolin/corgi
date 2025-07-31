package check

import (
	"fmt"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckComponentCalls() {
	logger := ch.Logger.WithGroup("component_calls")
	logger.Debug("Checking component calls")

	for _, f := range ch.P.Files {
		logger := logger.With(slog.String("file", f.Name))

		for _, cc := range f.ComponentCalls {
			logger := logger.With(
				slog.String("call_package", cc.Component.File.Package.Module+"/"+cc.Component.File.Package.PathInModule),
				slog.String("call_name", cc.Component.AST.Header.Name.Name),
				slog.String("call_pos", cc.AST.Start().String()))

			ch.CheckComponentArgsExist(logger, cc)
			ch.CheckNoDuplicateComponentArgs(logger, cc)
			ch.CheckRequiredComponentParamsSet(logger, cc)
			ch.CheckComponentAcceptsAttributes(logger, cc)
			if ch.CheckComponentCallBody(logger, cc) {
				ch.CheckUnreachableWiths(logger, cc)
				ch.CheckWithNotLooped(logger, cc)
			}

			ch.CheckRequiredBlocksAreSet(logger, cc)
		}
	}
}

func (ch *checker) CheckComponentCallBody(logger *slog.Logger, cc *file.ComponentCall) (ok bool) {
	ok = true
	logger = logger.WithGroup("body")

	sc, _ := cc.AST.Body.(*ast.Scope)
	if sc == nil {
		return false
	}

	walk.WalkT(sc, func(ctx *walk.ContextT[ast.ScopeNode]) error {
		switch ctx.Node.(type) {
		case *ast.Conditional:
		case *ast.Switch:
		case *ast.And:
		case *ast.For:
		case *ast.With:
			return walk.NoDive
		case *ast.ComponentCall:
			return walk.NoDive
		default:
			ok = false
			logger.Error("Illegal node in component call body")
			ch.Report(&diagnostic.Diagnostic{
				Message: "component call body: use of illegal node",
				Primary: []diagnostic.Annotation{
					anno.Position(cc.File, ctx.Node.Start(), fmt.Sprintf("cannot use %T here", ctx.Node)),
				},
				Hints: []diagnostic.Hint{
					{Hint: "Did you mean to use a default block shorthand?", Example: "`:foo() _{ ... }"},
					{Hint: "Wrap this inside a with statement", Example: "`with myBlock { ... }"},
				},
			})
		}

		return nil
	})
	return ok
}

func (ch *checker) CheckUnreachableWiths(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("unreachable_withs")

	if len(cc.BlockSetters) == 0 {
		return
	}

	sc, _ := cc.AST.Body.(*ast.Scope)
	if sc == nil {
		return
	}

	conditionalWiths := make(map[identifier][]*ast.With)
	topLevelWiths := make(map[identifier][]*ast.With)

	walk.WalkT(sc, func(ctx *walk.ContextT[*ast.With]) error {
		if len(ctx.Parents) == 0 {
			topLevelWiths[ctx.Node.Name()] = append(topLevelWiths[ctx.Node.Name()], ctx.Node)
		} else {
			conditionalWiths[ctx.Node.Name()] = append(conditionalWiths[ctx.Node.Name()], ctx.Node)
		}

		return walk.NoDive
	}, walk.DontDiveAny(&ast.ComponentCall{}))

	for _, tws := range topLevelWiths {
		cws := conditionalWiths[tws[0].Name()]
		if len(tws) <= 1 && len(cws) == 0 {
			continue
		}

		last := tws[len(tws)-1]

		primaries := make([]diagnostic.Annotation, 1, len(tws)+len(cws))
		primaries[0] = anno.Range(cc.File, last.Start(), last.Identifier.End(), "overwrites all of the above")

		for _, tw := range tws[:len(tws)-1] {
			primaries = append(primaries, anno.Node(cc.File, tw, "never actually used"))
		}
		for _, cw := range cws {
			primaries = append(primaries, anno.Node(cc.File, cw, "never actually used"))
		}

		logger.With(slog.String("name", last.Name())).
			Error("Unreachable withs")

		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: unreachable withs",
			Primary: primaries,
			Explanation: "The marked `with` blocks are never actually used, " +
				"because the last `with` always takes precedence over all the previous ones.",
		})
	}
}

func (ch *checker) CheckWithNotLooped(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("with_not_looped")

	if len(cc.BlockSetters) == 0 {
		return
	}

	sc, _ := cc.AST.Body.(*ast.Scope)
	if sc == nil {
		return
	}

	walk.WalkT(cc.AST.Body, func(ctx *walk.ContextT[*ast.With]) error {
		if len(ctx.Parents) == 0 {
			return nil
		}
		forLoop, _ := ctx.Parents[len(ctx.Parents)-1].Node.(*ast.For)
		if forLoop == nil {
			return nil
		}

		logger.Error("Component call: looped with")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: looped with",
			Primary: []diagnostic.Annotation{
				anno.Position(cc.File, ctx.Node.Start(), "only the with block from the very last iteration is ever used"),
			},
			Secondary: []diagnostic.Annotation{
				anno.Node(cc.File, forLoop, "in this for loop"),
			},
		})
		return walk.NoDive
	}, walk.DontDiveAny(&ast.ComponentCall{}))
}

func (ch *checker) CheckNoDuplicateComponentArgs(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("no_duplicate_args")

	if cc.AST.Header.Arguments == nil || len(cc.AST.Header.Arguments.List) <= 1 {
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
	logger.Debug("Checking that all component call arguments exist")

	if !cc.Component.File.Package.Analyzed || cc.Component.AnalyzedWithErrors {
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
					Context:    anno.ContextLines(cc.AST.Start(), cc.AST.Header.End()),
					Annotation: "component defines no parameter `" + name + "`",
				}),
			},
		})
		reported[name] = true
	}
}

func (ch *checker) CheckRequiredComponentParamsSet(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("required_params_set")

	if !cc.Component.File.Package.Analyzed || cc.Component.AnalyzedWithErrors {
		logger.Debug("Component analyzed with errors, skipping check")
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

func (ch *checker) CheckRequiredBlocksAreSet(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("required_blocks_set")

	if !cc.Component.File.Package.Analyzed || cc.Component.AnalyzedWithErrors {
		return
	}

	for _, block := range cc.Component.Blocks {
		logger := logger.With(slog.String("block", block.Name))
		if !block.Required {
			continue
		}

		if cc.BlockSetterByName(block.Name) != nil {
			continue
		}

		logger.Error("Required block not set")
		ch.Report(&diagnostic.Diagnostic{
			Message: "required block not set",
			Primary: []diagnostic.Annotation{
				anno.Node(cc.File, cc.AST.Header.Name, "requires block `"+block.Name+"` to be set"),
			},
			Explanation: "This component requires that the `" + block.Name + "` block is always set.\n" +
				"You can set a block using a with clause.",
			Docs: "component-call",
		})
	}
}

func (ch *checker) CheckComponentAcceptsAttributes(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("component_accepts_attributes")

	if !cc.Component.File.Package.Analyzed || cc.Component.AnalyzedWithErrors {
		return
	}

	acceptsAndPlaceholder := cc.Component.FirstIncludedAndPlaceholder(cc) != nil
	if acceptsAndPlaceholder {
		return
	}

	var attributeArgs []ast.Attribute
	if cc.AST.Header.Arguments != nil {
		attributeArgs = make([]ast.Attribute, 0, len(cc.AST.Header.Arguments.List))
		for _, arg := range cc.AST.Header.Arguments.List {
			attr, _ := arg.(ast.Attribute)
			if attr != nil {
				attributeArgs = append(attributeArgs, attr)
			}
		}
	}

	hasAttributes := cc.FirstAnd != nil || len(attributeArgs) > 0
	if !hasAttributes {
		return
	}

	primaries := make([]diagnostic.Annotation, 1, 2+len(attributeArgs))
	primaries[0] = anno.Node(cc.File, cc.AST.Header.Name, "this component does not accept any attributes")
	if cc.FirstAnd != nil {
		primaries = append(primaries, anno.Node(cc.File, cc.FirstAnd, "but you hand it attributes here"))
	}
	for _, attr := range attributeArgs {
		primaries = append(primaries, anno.Anno(cc.File, anno.Annotation{
			Context:    anno.ContextLines(cc.AST.Header.Start(), cc.AST.Header.End()),
			Highlight:  anno.HighlightNode(attr),
			Annotation: "but you pass it an attribute here",
		}))
	}

	logger.Error("Component does not accept attributes")
	diag := &diagnostic.Diagnostic{
		Message: "component call: component does not accept attributes",
		Primary: primaries,
		Explanation: "Components need to specify an &-placeholder somewhere in their body " +
			"for them to accept attributes. Since this component does not specify any " +
			"(or you have overwritten all block defaults that contain one), " +
			"you cannot hand attributes to it.",
		Docs: "attribute-placeholder",
	}
	couldAcceptAttributes := cc.Component.FirstIncludedAndPlaceholder(nil) != nil
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
