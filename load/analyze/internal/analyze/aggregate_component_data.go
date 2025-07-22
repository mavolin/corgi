package analyze

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

// AggregateComponentData aggregates all component data.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AggregateComponentData() {
	logger := z.Logger.WithGroup("components.aggregate")
	logger.Debug("Aggregating component data")

	for _, c := range z.P.Components {
		logger := logger.With(
			slog.String("file", c.File.Name),
			slog.String("comp", c.Header().Name.Name),
			slog.String("comp_pos", c.Start().String()))

		z.CheckForeignAlias(logger, c)
		z.CheckCircularAlias(logger, c)
	}

	z.AggregateParameters(logger)
	z.AggregateBlocks(logger)
}

// ============================================================================
// Check for Circular Aliases
// ======================================================================================

// CheckCircularAlias checks that the given component, if it is an alias, is
// not an alias of itself, either directly or indirectly through other aliases.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) CheckCircularAlias(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("checks.circular_alias")

	if c.AliasAST == nil {
		return
	}

	// Check that no root component calls itself.
	//
	// Take, for example:
	// 	alias Foo Bar()
	//  alias Bar Fooz()
	//  alias Fooz Bar()
	//
	// Then we only report when iterating over Bar itself, even though Foo
	// obviously also contains a circular alias.
	// That way we don't flood the user with errors with the same root cause.

	chain := make([]*file.Component, 1, 8)
	chain[0] = c

	aliasOf := c.File.ComponentCallByNode(c.AliasAST.ComponentCall)
	for aliasOf != nil {
		aliasComp := aliasOf.Component
		if aliasComp == nil { //nolint:gocritic
			break
		} else if aliasComp.AliasAST == nil {
			break
		} else if aliasComp.AnalyzedWithErrors {
			// probably a subchain of a circular alias that we already
			// reported
			break
		} else if slices.Contains(chain[1:], aliasComp) {
			c.AnalyzedWithErrors = true
			break
		}

		if aliasComp != c {
			aliasOf = aliasComp.File.ComponentCallByNode(aliasComp.AliasAST.ComponentCall)
			chain = append(chain, aliasComp)
			continue
		}

		var secondaries []diagnostic.Annotation
		if len(chain) > 1 {
			secondaries = make([]diagnostic.Annotation, 1, len(chain))
			secondaries[0] = anno.Node(c.File, c.AliasAST.ComponentCall,
				"1: aliases `"+c.AliasAST.ComponentCall.Header.Name.Full()+"`")
			for i, c := range chain[1:] {
				secondaries = append(secondaries,
					anno.Anno(c.File, anno.Annotation{
						Context:    anno.ContextLines(c.AliasAST.Start(), c.AliasAST.ComponentCall.Header.Name.End()),
						Highlight:  anno.HighlightNode(c.AliasAST.Header.Name),
						Annotation: fmt.Sprint(i+2, ": `"+c.AliasAST.Header.Name.Name+"` is an alias of `"+c.AliasAST.ComponentCall.Header.Name.Full()+"`"),
					}))
			}
		} else {
			secondaries = []diagnostic.Annotation{
				anno.Node(c.File, c.AliasAST.ComponentCall.Header, "aliases itself directly"),
			}
		}

		logger.Error("Found circular alias")
		name := c.Header().Name.Name
		z.Report(&diagnostic.Diagnostic{
			Message: "circular alias",
			Primary: []diagnostic.Annotation{
				anno.Node(c.File, c.AliasAST.Header.Name, "this component is an alias of itself"),
			},
			Secondary: secondaries,
			Explanation: "This component is an alias of itself, either directly or indirectly " +
				"through other aliases. " +
				"To solve this error, break the chain so that `" + name + "` doesn't alias itself.",
		})

		for _, c := range chain {
			c.AnalyzedWithErrors = true
		}
	}
}

// ============================================================================
// Check Foreign Alias
// ======================================================================================

// CheckForeignAlias checks that the given component, if it is an alias, is
// not an alias of a component from another package.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) CheckForeignAlias(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("checks.foreign_alias")

	if c.AliasAST == nil {
		return
	}

	cc := c.File.ComponentCallByNode(c.AliasAST.ComponentCall)
	if cc == nil || cc.Component == nil {
		return
	} else if cc.Component.File.Package == c.File.Package {
		return
	}

	c.AnalyzedWithErrors = true

	logger.Error("Found foreign alias")
	z.Report(&diagnostic.Diagnostic{
		Message: "foreign alias",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, c.AliasAST.Header.Name, "is an alias"),
			anno.Node(c.File, c.AliasAST.ComponentCall, "calls a component from another package"),
		},
	})
}

// ============================================================================
// Aggregate Parameters
// ======================================================================================

// AggregateParameters aggregates the parameters of all components.
//
// Depends on Checks:
//   - CheckCircularAlias - To ensure that we don't aggregate parameters of
//     components that are aliases of themselves, which would lead to an
//     infinite loop.
//   - CheckForeignAlias - Sensible requirement so aliases don't break when
//     dependencies are updated.
//
// Sets Fields:
//   - Components.Parameters
//   - Components.Parameters.AST
//   - Components.Parameters.Component
//
// Depends on Fields: None
func (z *analyzer) AggregateParameters(logger *slog.Logger) {
	logger = logger.WithGroup("parameters")

	// Our strategy to aggregate component parameters:
	//  For as long as we aggregated at least one component the last iteration:
	//    1. If this is a defined component, aggregate it. This can be done without
	//       any dependencies.
	//    2. If this is an alias component, aggregate it if the component we are
	//       aliasing has already been aggregated.
	//
	// Ideally, we can aggregate all components in two iterations, assuming that
	// aliases always reference a defined component, which should be the
	// "more than average" case.
	// Even more ideally, if aliases are always placed below the component they
	// are aliasing, we can aggregate all components in one iteration.

	n := len(z.P.Components)
	comps := make([]*file.Component, n)
	copy(comps, z.P.Components)

	prevN := n + 1
	var i int
	for prevN != n {
		prevN = n
		logger.Debug("Starting new iteration of aggregating component parameters",
			slog.Int("iteration", i),
			slog.Int("remaining_components", n))

		for ci, c := range comps {
			if c == nil {
				continue
			}

			if c.AnalyzedWithErrors {
				comps[ci] = nil
				n--
				continue
			}

			if c.DefinedAST != nil {
				z.aggregateDefinedComponentParameters(c)
				n--
				comps[ci] = nil
				continue
			}

			cc := c.File.ComponentCallByNode(c.AliasAST.ComponentCall)
			switch {
			case cc.Component == nil:
				fallthrough
			case cc.Component.File.Package != c.File.Package:
				fallthrough
			case cc.Component.AnalyzedWithErrors:
				c.AnalyzedWithErrors = true
				comps[ci] = nil
				n--
				continue
			case cc.Component.Parameters == nil:
				continue
			}

			z.aggregateAliasComponentParameters(c, cc)
			comps[ci] = nil
			n--
		}

		i++
	}

	if n == 0 {
		return
	}

	for _, c := range comps {
		if c == nil {
			continue
		}

		c.AnalyzedWithErrors = true

		logger := logger.With(
			slog.String("file", c.File.Name),
			slog.String("component", c.Header().Name.Name),
			slog.String("pos", c.Start().String()))

		// should've been caught by CheckCircularAlias
		logger.Error("Failed to aggregate component parameters")
		z.Report(&diagnostic.Diagnostic{
			Type:    diagnostic.InternalError,
			Message: "component alias: failed to aggregate parameters",
			Primary: []diagnostic.Annotation{
				anno.Node(c.File, c.Header().Name, "could not aggregate this component's parameters"),
			},
			Explanation: "This error should've been caught earlier with a better error message. " +
				"Please open an issue.",
		})
	}
}

func (z *analyzer) aggregateDefinedComponentParameters(c *file.Component) {
	if c.DefinedAST.Header == nil {
		return
	} else if c.DefinedAST.Header.Parameters == nil {
		return
	}

	c.Parameters = make([]*file.ComponentParameter, len(c.DefinedAST.Header.Parameters.List))
	for i, param := range c.DefinedAST.Header.Parameters.List {
		c.Parameters[i] = &file.ComponentParameter{
			AST:       param,
			Component: c,
		}
	}
}

func (z *analyzer) aggregateAliasComponentParameters(c *file.Component, cc *file.ComponentCall) {
	parent := cc.Component

	params := make([]*file.ComponentParameter, 0, len(parent.Parameters)+len(c.AliasAST.Header.Parameters.List))
Params:
	for _, parentParam := range parent.Parameters {
		if cc.AST.Header.Arguments != nil {
			for _, arg := range cc.AST.Header.Arguments.List {
				carg, _ := arg.(*ast.ComponentArgument)
				if carg != nil && carg.Name.Name == parentParam.AST.Name.Name {
					continue Params // skip this parameter, not inherited by c
				}
			}
		}
		if c.AliasAST.Header.Parameters != nil {
			for _, childParam := range c.AliasAST.Header.Parameters.List {
				if childParam.Name.Name == parentParam.AST.Name.Name {
					continue Params // skip this parameter, overridden by c
				}
			}
		}

		params = append(params, parentParam)
	}

	if c.AliasAST.Header.Parameters != nil {
		for _, param := range c.AliasAST.Header.Parameters.List {
			params = append(params, &file.ComponentParameter{
				AST:       param,
				Component: c,
			})
		}
	}

	c.Parameters = slices.Clip(params)
}

// ============================================================================
// Aggregate Blocks
// ======================================================================================

// AggregateBlocks aggregates the blocks of all components.
//
// Depends on Checks:
//   - CheckCircularAlias - To ensure that we don't aggregate blocks
//     of components that are aliases of themselves, which would lead to an
//     infinite loop.
//   - CheckForeignAlias - Sensible requirement so aliases don't break when
//     dependencies are updated.
//
// Sets Fields:
//   - Components.Blocks
//   - Components.Blocks.Component
//   - Components.Blocks.Name
//   - Components.Blocks.Instances
//   - Components.Blocks.Instances.Group
//   - Components.Blocks.Instances.AST
//   - Components.Blocks.Instances.ChildOf
//   - Components.Blocks.Instances.Default
//   - Components.Blocks.Instances.Default.AST
//
// Depends on Fields:
//   - ComponentCalls.Withs
//   - ComponentCalls.Withs.Name
func (z *analyzer) AggregateBlocks(logger *slog.Logger) {
	logger = logger.WithGroup("blocks")

	// Our strategy to aggregate component blocks is essentially the same as
	// for parameters:
	//  For as long as we aggregated at least one component the last iteration:
	//    1. If this is a defined component, aggregate it. This can be done
	//    	 without any dependencies.
	//    2. If this is an alias component, aggregate it if the aliased
	//       component has already been aggregated.
	//
	// Ideally, we can aggregate all components in two iterations, assuming that
	// aliases always reference a defined component, which should be the "more than
	// average" case. Even more ideally, if aliases are always placed below the
	// component they are aliasing, we can aggregate all components in one
	// iteration.

	n := len(z.P.Components)
	comps := make([]*file.Component, n)
	copy(comps, z.P.Components)

	prevN := n + 1
	var i int
	for prevN > n {
		prevN = n
		logger.Debug("Starting new iteration of aggregating component blocks",
			slog.Int("iteration", i),
			slog.Int("remaining_components", n))

		for ci, c := range comps {
			if c == nil {
				continue
			}

			if c.AnalyzedWithErrors {
				comps[ci] = nil
				n--
				continue
			}

			if c.DefinedAST != nil {
				z.aggregateDefinedComponentBlocks(c)
				n--
				comps[ci] = nil
				continue
			}

			cc := c.File.ComponentCallByNode(c.AliasAST.ComponentCall)
			switch {
			case cc.Component == nil:
				fallthrough
			case cc.Component.File.Package != c.File.Package:
				fallthrough
			case cc.Component.AnalyzedWithErrors:
				c.AnalyzedWithErrors = true
				comps[ci] = nil
				n--
				continue
			case cc.Component.Blocks == nil:
				continue
			}

			z.aggregateAliasComponentBlocks(c, cc)
			comps[ci] = nil
			n--
		}

		i++
	}

	if n == 0 {
		return
	}

	for _, c := range comps {
		if c == nil {
			continue
		}

		c.AnalyzedWithErrors = true

		logger := logger.With(
			slog.String("file", c.File.Name),
			slog.String("comp", c.Header().Name.Name),
			slog.String("comp_pos", c.Start().String()))

		logger.Error("Failed to aggregate component blocks")
		z.Report(&diagnostic.Diagnostic{
			Message: "component alias: failed to aggregate blocks",
			Primary: []diagnostic.Annotation{
				anno.Node(c.File, c.Header().Name, "could not aggregate this component's blocks"),
			},
			Explanation: "This error should've been caught earlier with a better error message. " +
				"Please open an issue.",
		}) // should've been caught by CheckCircularAlias
	}
}

func (z *analyzer) aggregateDefinedComponentBlocks(c *file.Component) {
	scope, _ := c.DefinedAST.Body.(*ast.Scope)
	if scope == nil {
		return
	}

	c.Blocks = make([]*file.Block, 0, 12) // 12 seems like a sensible max

	walk.WalkT(scope, func(ctx *walk.ContextT[*ast.Block]) error {
		parent := walk.Closest[*ast.Block](ctx.Parents)

		instance := &file.BlockInstance{
			AST:     ctx.Node,
			ChildOf: nil,
			Default: file.BlockInstanceDefault{
				AST: ctx.Node.Default,
			},
		}
		if parent != nil {
			// Walk is depth-first, so the parent block is already aggregated.
			instance.ChildOf = c.BlockInstanceByNode(parent)
		}

		for _, block := range c.Blocks {
			if block.Name == ctx.Node.Name() {
				instance.Group = block
				block.Instances = append(block.Instances, instance)
				return nil
			}
		}

		group := &file.Block{
			Component: c,
			Name:      ctx.Node.Name(),
			Instances: make([]*file.BlockInstance, 0, 4),
		}
		instance.Group = group
		group.Instances = append(group.Instances, instance)
		c.Blocks = append(c.Blocks, group)

		return nil
	})

	c.Blocks = slices.Clip(c.Blocks)
	for _, block := range c.Blocks {
		block.Instances = slices.Clip(block.Instances)
	}
}

func (z *analyzer) aggregateAliasComponentBlocks(c *file.Component, cc *file.ComponentCall) {
	if cc.AST.Body == nil { // fast path
		c.Blocks = make([]*file.Block, len(cc.Component.Blocks))
		copy(c.Blocks, cc.Component.Blocks)
		return
	}

	parent := cc.Component

	c.Blocks = make([]*file.Block, 0, len(parent.Blocks)+8)

	// Add all blocks that are not set by the component call.
	for _, parentBlock := range parent.Blocks {
		with := cc.WithByName(parentBlock.Name)
		if with == nil {
			c.Blocks = append(c.Blocks, parentBlock)
		}
	}

	// Now collect all newly declared blocks.
	// If a block with the same name as an inherited block is declared,
	// it overrides the inherited block.

	scope, _ := cc.AST.Body.(*ast.Scope)
	if scope == nil {
		return
	}

	walk.WalkT(scope, func(ctx *walk.ContextT[*ast.Block]) error {
		// Three cases:
		//  1. Block overwrites an inherited block, replace it.
		//  2. A new block instance for an already encountered block, add it
		//     to the existing block.
		//  3. The first instance of a new block, create a new block group.

		instance := &file.BlockInstance{
			AST: ctx.Node,
			Default: file.BlockInstanceDefault{
				AST: ctx.Node.Default,
			},
		}
		parentBlock := walk.Closest[*ast.Block](ctx.Parents)
		if parentBlock != nil {
			// Walk is depth-first, so the parent block is already aggregated.
			instance.ChildOf = c.BlockInstanceByNode(parentBlock)
		}

		for i, group := range c.Blocks {
			if group.Name != ctx.Node.Name() {
				continue
			}

			if group.Component == c { // Case 2
				instance.Group = group
				group.Instances = append(group.Instances, instance)
				return nil
			}

			// Case 1
			group = &file.Block{
				Component: c,
				Name:      ctx.Node.Name(),
				Instances: make([]*file.BlockInstance, 1, 8),
			}
			instance.Group = group
			group.Instances[0] = instance
			c.Blocks[i] = group
			return nil
		}

		// Case 3
		group := &file.Block{
			Component: c,
			Name:      ctx.Node.Name(),
			Instances: make([]*file.BlockInstance, 1, 8),
		}
		instance.Group = group
		group.Instances[0] = instance
		c.Blocks = append(c.Blocks, group)
		return nil
	})
}
