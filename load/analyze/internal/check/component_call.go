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
	logger.Info("Checking component calls")

	for _, f := range ch.P.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Checking file")

		for _, cc := range f.ComponentCalls {
			logger := logger.With(
				slog.String("call_package", cc.Component.File.Package.Module+"/"+cc.Component.File.Package.PathInModule),
				slog.String("call_name", cc.Component.Header().Name.Ident),
				slog.String("call_pos", cc.AST.Start().String()))
			logger.Debug("Checking component call")

			ch.CheckComponentArgsExist(logger, cc)
			ch.CheckNoDuplicateComponentArgs(logger, cc)
			ch.CheckRequiredComponentParamsSet(logger, cc)
			ch.CheckComponentAcceptsAttributes(logger, cc)
			if ch.CheckComponentCallBody(logger, cc) {
				ch.CheckUnreachableWiths(logger, cc)
				ch.CheckWithNotLooped(logger, cc)
			}

			ch.CheckRequiredBlocksAreSet(logger, cc)
			ch.CheckBlockExists(logger, cc)
		}
	}
}

func (ch *checker) CheckComponentCallBody(logger *slog.Logger, cc *file.ComponentCall) (ok bool) {
	ok = true
	logger = logger.WithGroup("body")
	logger.Debug("Checking for valid body")

	bt, _ := cc.AST.Body.(*ast.BracketText)
	if bt != nil {
		logger.Error("Bracket text used as component call body")
		ch.Report(&diagnostic.Diagnostic{
			Message: "bracket text used as component call body",
			Primary: []diagnostic.Annotation{
				anno.Position(cc.File, cc.AST.Body.Start(), "cannot use a bracket text here"),
			},
			Hints: []diagnostic.Hint{
				{Hint: "Did you mean to use an underscore block shorthand?", Example: "`_[ ... ]"},
			},
		})
		return false
	}

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
					{Hint: "Did you mean to use a underscore block shorthand?", Example: "`:foo() _{ ... }"},
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
	logger.Debug("Checking for unreachable withs")

	// We don't need successful analysis to run this check.
	// But if it ran, we can safely check if this component call has any withs at all,
	// before we do an unnecessary walk.
	if !cc.AnalyzedWithErrors && len(cc.Withs) == 0 {
		logger.Debug("Component call has no withs, skipping")
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
			topLevelWiths[ctx.Node.Name.Ident] = append(topLevelWiths[ctx.Node.Name.Ident], ctx.Node)
		} else {
			conditionalWiths[ctx.Node.Name.Ident] = append(conditionalWiths[ctx.Node.Name.Ident], ctx.Node)
		}

		return walk.NoDive
	}, walk.DontDiveAny(&ast.ComponentCall{}))

	for _, tws := range topLevelWiths {
		cws := conditionalWiths[tws[0].Name.Ident]
		if len(tws) <= 1 && len(cws) == 0 {
			continue
		}

		last := tws[len(tws)-1]

		primaries := make([]diagnostic.Annotation, 1, len(tws)+len(cws))
		primaries[0] = anno.Range(cc.File, *last.With, last.Name.End(), "overwrites all of the above")

		for _, tw := range tws[:len(tws)-1] {
			primaries = append(primaries, anno.Range(cc.File, *tw.With, tw.Name.End(), "never actually used"))
		}
		for _, cw := range cws {
			primaries = append(primaries, anno.Range(cc.File, *cw.With, cw.Name.End(), "never actually used"))
		}

		logger.
			With(slog.String("name", last.Name.Ident)).
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
	logger.Debug("Checking that withs are not looped")

	// We don't need successful analysis to run this check.
	// But if it ran, we can safely check if this component call has any withs at all,
	// before we do an unnecessary walk.
	if !cc.AnalyzedWithErrors && len(cc.Withs) == 0 {
		logger.Debug("Component call has no withs, skipping")
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
				anno.Node(cc.File, ctx.Node, "in this for loop"),
			},
		})
		return walk.NoDive
	}, walk.DontDiveAny(&ast.ComponentCall{}))
}

func (ch *checker) CheckNoDuplicateComponentArgs(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("no_duplicate_args")
	logger.Debug("Checking for duplicate component call arguments")

	if cc.AST.Header.Arguments == nil || len(cc.AST.Header.Arguments.Args) <= 1 {
		logger.Debug("One or no arguments, skipping")
		return
	}
	args := cc.AST.Header.Arguments.Args

	reported := ch.TakeStringSet()
	dupls := make([]*ast.ComponentArgument, 0, len(args)-1)

	for ai, a := range args[:len(args)-1] {
		aArg, _ := a.(*ast.ComponentArgument)
		if aArg == nil {
			continue
		}
		aName := aArg.Name.Ident
		if reported.Contains(aName) {
			continue
		} else if cc.Component.ParameterByName(aName) == nil {
			// non-existent arguments are handled by CheckComponentArgsExist
			continue
		}

		logger := logger.With(
			slog.String("arg_pos", aArg.Start().String()),
			slog.String("arg_name", aName))

		for _, b := range args[ai:] {
			bArg, _ := b.(*ast.ComponentArgument)
			if bArg != nil && aName == bArg.Name.Ident {
				dupls = append(dupls, bArg)
			}
		}

		if len(dupls) > 0 {
			primaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
			primaries[0] = anno.Node(cc.File, aArg.Name, "first set here")
			for _, b := range dupls {
				primaries = append(primaries, anno.Node(cc.File, b.Name, "then here again"))
			}

			logger.Error("Found duplicate component call argument")
			ch.Report(&diagnostic.Diagnostic{
				Message: "component call: argument specified twice",
				Primary: primaries,
			})
			reported.Add(aName)
			dupls = dupls[:0] // reset slice
		}
	}
}

func (ch *checker) CheckComponentArgsExist(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("args_exist")
	logger.Debug("Checking that all component call arguments exist")

	if cc.Component.AnalyzedWithErrors {
		logger.Debug("Component analyzed with errors, skipping check")
		return
	} else if cc.AST.Header.Arguments == nil || len(cc.AST.Header.Arguments.Args) == 0 {
		logger.Debug("Component call has no arguments, skipping")
		return
	}

	reported := ch.TakeStringSet()
	for _, arg := range cc.AST.Header.Arguments.Args {
		carg, _ := arg.(*ast.ComponentArgument)
		if carg == nil {
			continue
		}

		name := carg.Name.Ident
		if reported.Contains(name) {
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
				anno.Node(cc.File, carg.Name, "this component defines no argument `"+name+"`"),
			},
			Secondary: []diagnostic.Annotation{
				anno.Anno(cc.Component.File, anno.Annotation{
					Annotation: "referenced component",
					Highlight:  anno.HighlightNode(cc.Component.Header().Name),
					Context:    anno.ContextLines(cc.Component.Header().Start(), cc.Component.Header().End()),
				}),
			},
		})
		reported.Add(name)
	}
}

func (ch *checker) CheckRequiredComponentParamsSet(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("required_params_set")
	logger.Debug("Checking required component parameters are set")

	if cc.Component.AnalyzedWithErrors {
		logger.Debug("Component analyzed with errors, skipping check")
		return
	}

Params:
	for _, param := range cc.Component.Parameters {
		if !param.Required() {
			continue
		}

		logger := logger.With(slog.String("param", param.AST.Name.Ident))
		logger.Debug("Checking parameter")

		for _, arg := range cc.AST.Header.Arguments.Args {
			carg, _ := arg.(*ast.ComponentArgument)
			if carg == nil {
				continue
			}

			if carg.Name.Ident == param.AST.Name.Ident {
				logger.Debug("Required parameter is set", slog.String("arg_pos", carg.Start().String()))
				continue Params // parameter is set
			}
		}

		// parameter is not set
		logger.Error("Required parameter not set")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: required parameter not set",
			Primary: []diagnostic.Annotation{
				anno.Node(cc.File, cc.AST.Header.Name, "requires parameter `"+param.AST.Name.Ident+"` to be set"),
			},
			Explanation: "Parameters with no default, " +
				"like `" + param.AST.Name.Ident + "`, " +
				"are required to be set in every component call.",
		})
	}
}

func (ch *checker) CheckRequiredBlocksAreSet(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("required_blocks_set")
	logger.Debug("Checking required blocks are set")

	if cc.Component.AnalyzedWithErrors {
		logger.Debug("Component analyzed with error, skipping check")
		return
	}

	for _, block := range cc.Component.Blocks {
		logger := logger.With(slog.String("block", block.Name))
		logger.Debug("Checking block")
		if !block.Required {
			logger.Debug("Block is not required, skipping")
			continue
		}

		if cc.WithByName(block.Name) != nil {
			logger.Debug("Required block is set")
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

func (ch *checker) CheckBlockExists(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("block_exists")
	logger.Debug("Checking that all blocks specified in with clauses exists")

	if cc.Component.AnalyzedWithErrors {
		logger.Debug("Component analyzed with errors, skipping check")
		return
	}

	for _, with := range cc.Withs {
		logger := logger.With(slog.String("with_name", with.Name))
		logger.Debug("Checking with block")

		if cc.Component.BlockByName(with.Name) != nil {
			logger.Debug("With block exists")
			continue
		}

		logger.Error("With block does not exist")
		primaries := make([]diagnostic.Annotation, len(with.Instances))
		for i, inst := range with.Instances {
			primaries[i] = anno.Node(cc.File, inst.AST.Name, "unknown block")
		}
		ch.Report(&diagnostic.Diagnostic{
			Message:     "component call: with: block does not exist",
			Primary:     primaries,
			Explanation: "This component does not define a block named `" + with.Name + "`.",
		})
	}
}

func (ch *checker) CheckComponentAcceptsAttributes(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("component_accepts_attributes")
	logger.Debug("Checking that if call hands attributes, component accepts attributes")

	if cc.Component.AnalyzedWithErrors {
		logger.Debug("Component analyzed with errors, skipping check")
		return
	}

	if cc.FirstPlaceholderAnd == nil {
		if cc.AST.Header.Arguments != nil {
			for _, arg := range cc.AST.Header.Arguments.Args {
				attr, _ := arg.(ast.Attribute)
				if attr != nil {
					goto HasAttributes
				}
			}
		}

		logger.Debug("Component call does not hand attributes, skipping check")
		return
	}

HasAttributes:
	if cc.Component.FirstIncludedAndPlaceholder(cc) != nil {
		return
	}

	logger.Error("Component does not accept attributes")
	diag := &diagnostic.Diagnostic{
		Message: "component call: component does not accept attributes",
		Primary: []diagnostic.Annotation{
			anno.Node(cc.File, cc.AST.Header.Name, "this component does not accept any attributes"),
			anno.Node(cc.File, cc.FirstPlaceholderAnd, "but you hand it attributes here"),
		},
		Explanation: "Components need to specify an &-placeholder somewhere in their body " +
			"for them to accept attributes. Since this component does not specify any " +
			"(or you have overwritten all block defaults that contain one), " +
			"you cannot hand attributes to it.",
		Docs: "attribute-placeholder",
	}
	if cc.Component.FirstIncludedAndPlaceholder(nil) != nil {
		diag.Hints = []diagnostic.Hint{
			{
				Hint: "The only &-placeholders of this component are specified in defaults of blocks, " +
					"that you are overwriting. " +
					"Perhaps, you could add the attributes in your block body directly, " +
					"to achieve the same result?",
			},
		}
	}
	ch.Report(diag)
}
