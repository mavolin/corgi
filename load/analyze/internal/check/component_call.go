package check

import (
	"fmt"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/switches"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckComponentCalls() {
	logger := ch.Logger.WithGroup("component_calls")
	logger.Debug("Checking component calls")

	for _, f := range ch.Pkg.Files {
		logger := logger.With(slog.String("file", string(f.Name)))

		for _, cc := range f.ComponentCalls {
			logger := logger.With(
				slog.String("call_name", cc.AST.Header.Name.Full()),
				slog.String("call_pos", cc.AST.Start().String()))

			ch.CheckComponentCall_Body(logger, cc)
			ch.CheckComponentCall_UnreachableWiths(logger, cc)
			ch.CheckComponentCall_WithNotLooped(logger, cc)

			ch.CheckComponentCall_RequiredBlocksAreSet(logger, cc)
		}
	}
}

func (ch *checker) CheckComponentCall_Body(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("body")

	var scope *ast.Scope
	switches.ComponentCallBody(cc.AST.Body,
		func(*ast.DefaultBlockShorthand) {},
		func(s *ast.Scope) { scope = s })
	if scope == nil {
		return
	}

	walk.WalkT(scope, func(w *walk.ContextT[ast.ScopeNode]) walk.Action {
		switch w.Node.(type) {
		case *ast.Conditional:
		case *ast.Switch:
		case *ast.And:
		case *ast.For:
		case *ast.With:
			return walk.NoDive
		case *ast.ComponentCall:
			return walk.NoDive
		default:
			logger.Error("Illegal node in component call body")
			ch.Report(&diagnostic.Diagnostic{
				Message: "component call body: use of illegal node",
				Primary: []diagnostic.Annotation{
					anno.Position(cc.File, w.Node.Start(), fmt.Sprintf("cannot use %T here", w.Node)),
				},
				Hints: []diagnostic.Hint{
					{Hint: "Did you mean to use a default block shorthand?", Example: "`:foo() _{ ... }"},
					{Hint: "Wrap this inside a with statement", Example: "`with myBlock { ... }"},
				},
			})
		}

		return walk.Continue
	})
}

// ============================================================================
// Unreachable Withs
// ======================================================================================

func (ch *checker) CheckComponentCall_UnreachableWiths(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("unreachable_withs")

	if len(cc.BlockSetters) == 0 {
		return
	}

	var scope *ast.Scope
	switches.ComponentCallBody(cc.AST.Body,
		func(*ast.DefaultBlockShorthand) {},
		func(s *ast.Scope) { scope = s })
	if scope == nil {
		return
	}

	conditionalWiths := make(map[identifier][]*ast.With)
	topLevelWiths := make(map[identifier][]*ast.With)

	walk.WalkT(scope, func(w *walk.ContextT[*ast.With]) walk.Action {
		if len(w.Parents) == 0 {
			topLevelWiths[w.Node.Name()] = append(topLevelWiths[w.Node.Name()], w.Node)
		} else {
			conditionalWiths[w.Node.Name()] = append(conditionalWiths[w.Node.Name()], w.Node)
		}

		return walk.NoDive
	}, walk.DontDive[*ast.ComponentCall]())

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

// ============================================================================
// Withs Must Not Be Looped
// ======================================================================================

func (ch *checker) CheckComponentCall_WithNotLooped(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("with_not_looped")

	if len(cc.BlockSetters) == 0 {
		return
	}

	var scope *ast.Scope
	switches.ComponentCallBody(cc.AST.Body,
		func(*ast.DefaultBlockShorthand) {},
		func(s *ast.Scope) { scope = s })
	if scope == nil {
		return
	}

	walk.WalkT(cc.AST.Body, func(w *walk.ContextT[*ast.With]) walk.Action {
		if len(w.Parents) == 0 {
			return walk.Continue
		}
		forLoop, _ := w.Parents[len(w.Parents)-1].Node.(*ast.For)
		if forLoop == nil {
			return walk.Continue
		}

		logger.Error("Component call: looped with")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component call: looped with",
			Primary: []diagnostic.Annotation{
				anno.Position(cc.File, w.Node.Start(), "only the with block from the very last iteration is ever used"),
			},
			Secondary: []diagnostic.Annotation{
				anno.Node(cc.File, forLoop, "in this for loop"),
			},
		})
		return walk.NoDive
	}, walk.DontDive[*ast.ComponentCall]())
}

// ============================================================================
// Required Blocks Are Set
// ======================================================================================

func (ch *checker) CheckComponentCall_RequiredBlocksAreSet(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("required_blocks_set")

	if cc.Component == nil {
		return
	}

	for _, block := range cc.Component.Blocks {
		if block.Required().Equal(false) {
			continue
		} else if cc.BlockSetterByName(block.Name) != nil {
			continue
		}

		logger.Error("Required block not set", slog.String("block", string(block.Name)))
		ch.Report(&diagnostic.Diagnostic{
			Message: "required block not set",
			Primary: []diagnostic.Annotation{
				anno.Node(cc.File, cc.AST.Header.Name, "requires block `"+string(block.Name)+"` to be set"),
			},
			Explanation: "This component requires that the `" + string(block.Name) + "` block is always set.\n" +
				"You can set a block using a with clause.",
			Docs: "component-call",
		})
	}
}
