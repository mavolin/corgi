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
	logger.Info("Aggregating component data")

	for _, c := range z.P.Components {
		logger := logger.With(
			slog.String("file", c.File.Name),
			slog.String("comp", c.Header().Name.Name),
			slog.String("comp_pos", c.Start().String()))
		logger.Debug("Aggregating component data")

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
	logger = logger.WithGroup("check.circular_alias")
	logger.Info("Checking that this is not a circular alias")

	if c.AliasAST == nil {
		logger.Debug("Not an alias, skipping")
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
			logger.Debug("Component in chain not found, skipping",
				slog.String("component_name", aliasOf.AST.Header.Name.Full()),
				slog.String("component_call_pos", aliasOf.AST.Start().String()))
			break
		} else if aliasComp.AliasAST == nil {
			logger.Debug("Component in chain is not an alias, no circular alias",
				slog.String("component_name", aliasComp.Header().Name.Full()),
				slog.String("component_call_pos", aliasOf.AST.Start().String()))
			break
		} else if aliasComp.AnalyzedWithErrors {
			// probably a subchain of a circular alias that we already
			// reported
			logger.Debug("Already analyzed with error, skipping")
			break
		} else if slices.Contains(chain[1:], aliasComp) {
			logger.Debug("Component calls a circular alias, this will be reported when the root alias is analyzed",
				slog.String("circular_alias_name", aliasComp.Header().Name.Full()))
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
// Aggregate Parameters
// ======================================================================================

// AggregateParameters aggregates the parameters of all components.
//
// Depends on Checks:
//   - CheckCircularAlias - To ensure that we don't aggregate parameters of
//     components that are aliases of themselves, which would lead to an
//     infinite loop.
//
// Sets Fields:
//   - Components.Parameters
//   - Components.Parameters.AST
//   - Components.Parameters.Component
//
// Depends on Fields: None
func (z *analyzer) AggregateParameters(logger *slog.Logger) {
	logger = logger.WithGroup("parameters")
	logger.Info("Aggregating all component parameters")

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

			logger := logger.With(
				slog.String("file", c.File.Name),
				slog.String("component", c.Header().Name.Name),
				slog.String("pos", c.Start().String()))
			logger.Debug("Trying to aggregate component parameters")

			if c.AnalyzedWithErrors {
				logger.Debug("Component has been analyzed with errors, skipping")
				comps[ci] = nil
				n--
				continue
			}

			if c.DefinedAST != nil {
				z.aggregateDefinedComponentParameters(logger, c)
				n--
				comps[ci] = nil
				continue
			}

			cc := c.File.ComponentCallByNode(c.AliasAST.ComponentCall)
			switch {
			case cc.Component == nil:
				logger.Debug("Component call is not linked, skipping")
				c.AnalyzedWithErrors = true
				comps[ci] = nil
				n--
				continue
			case cc.Component.AnalyzedWithErrors:
				logger.Debug("Aliased component has been analyzed with errors, skipping")
				c.AnalyzedWithErrors = true
				comps[ci] = nil
				n--
				continue
			case cc.Component.Parameters == nil:
				logger.Debug("Aliased component has not yet had its parameters aggregated, waiting for next iteration")
				continue
			}

			z.aggregateAliasComponentParameters(logger, c, cc)
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

func (z *analyzer) aggregateDefinedComponentParameters(logger *slog.Logger, c *file.Component) {
	logger.Debug("Component is a defined component, using defined parameter list")

	if c.DefinedAST.Header == nil {
		logger.Debug("Component has no header, skipping")
	} else if c.DefinedAST.Header.Parameters == nil {
		logger.Debug("Component has no parameters, skipping")
	}

	c.Parameters = make([]*file.ComponentParameter, len(c.DefinedAST.Header.Parameters.Params))
	for i, param := range c.DefinedAST.Header.Parameters.Params {
		c.Parameters[i] = &file.ComponentParameter{
			AST:       param,
			Component: c,
		}
	}
}

func (z *analyzer) aggregateAliasComponentParameters(logger *slog.Logger, c *file.Component, cc *file.ComponentCall) {
	logger.Debug("Component is an alias, resolving which parameters are available")

	parent := cc.Component

	params := make([]*file.ComponentParameter, 0, len(parent.Parameters)+len(c.AliasAST.Header.Parameters.Params))
Params:
	for _, parentParam := range parent.Parameters {
		if cc.AST.Header.Arguments != nil {
			for _, arg := range cc.AST.Header.Arguments.Args {
				carg, _ := arg.(*ast.ComponentArgument)
				if carg != nil && carg.Name.Name == parentParam.AST.Name.Name {
					continue Params // skip this parameter, not inherited by c
				}
			}
		}
		if c.AliasAST.Header.Parameters != nil {
			for _, childParam := range c.AliasAST.Header.Parameters.Params {
				if childParam.Name.Name == parentParam.AST.Name.Name {
					continue Params // skip this parameter, overridden by c
				}
			}
		}

		params = append(params, parentParam)
	}

	if c.AliasAST.Header.Parameters != nil {
		for _, param := range c.AliasAST.Header.Parameters.Params {
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
	logger.Info("Aggregating all component blocks")

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

			logger := logger.With(
				slog.String("file", c.File.Name),
				slog.String("comp", c.Header().Name.Name),
				slog.String("comp_pos", c.Start().String()))
			logger.Debug("Trying to aggregate component blocks")

			if c.AnalyzedWithErrors {
				logger.Debug("Component has been analyzed with errors, skipping")
				comps[ci] = nil
				n--
				continue
			}

			if c.DefinedAST != nil {
				z.aggregateDefinedComponentBlocks(logger, c)
				n--
				comps[ci] = nil
				continue
			}

			cc := c.File.ComponentCallByNode(c.AliasAST.ComponentCall)
			switch {
			case cc.Component == nil:
				logger.Debug("Component call is not linked, skipping")
				comps[ci] = nil
				n--
				continue
			case cc.Component.AnalyzedWithErrors:
				logger.Debug("Aliased component has been analyzed with errors, skipping")
				c.AnalyzedWithErrors = true
				comps[ci] = nil
				n--
				continue
			case cc.Component.Blocks == nil:
				logger.Debug("Aliased component has not yet had its blocks aggregated, waiting for next iteration")
				continue
			}

			z.aggregateAliasComponentBlocks(logger, c, cc)
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

func (z *analyzer) aggregateDefinedComponentBlocks(logger *slog.Logger, c *file.Component) {
	logger.Debug("Component is a defined component, using aggregating blocks from body")

	scope, _ := c.DefinedAST.Body.(*ast.Scope)
	if scope == nil {
		logger.Debug("Component body is not a scope, skipping aggregation")
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

		logger := logger.With(
			slog.String("block", ctx.Node.Name()),
			slog.String("block_pos", ctx.Node.Start().String()),
			slog.String("block_child_of", instance.ChildOf.Group.Name))
		logger.Debug("Aggregating block instance")

		for _, block := range c.Blocks {
			if block.Name == ctx.Node.Name() {
				instance.Group = block
				block.Instances = append(block.Instances, instance)
				logger.Debug("Block group already recorded, adding instance")
				return nil
			}
		}

		logger.Debug("Block group not recorded yet, creating new group")
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

func (z *analyzer) aggregateAliasComponentBlocks(logger *slog.Logger, c *file.Component, cc *file.ComponentCall) {
	logger.Debug("Component is an alias, resolving which blocks are available")

	if cc.AST.Body == nil { // fast path
		logger.Debug("Component call has no body, component inherits all blocks from aliased component")
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
