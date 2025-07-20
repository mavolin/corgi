package analyze

import (
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
)

// AggregateComponentCallData aggregates the component call data for all
// component calls.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AggregateComponentCallData() {
	logger := z.Logger.WithGroup("component_calls.aggregate")
	logger.Info("Aggregating component call data")

	for _, f := range z.P.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Aggregating from file")

		for _, cc := range f.ComponentCalls {
			logger := logger.With(
				slog.String("call_name", cc.Component.Header().Name.Name),
				slog.String("call_pos", cc.AST.Start().String()))
			logger.Debug("Aggregating component call data")

			z.AggregateWiths(logger, cc)
		}
	}
}

// AggregateWiths aggregates the withs of the given component call.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.Withs
//   - ComponentCalls.Withs.Name
//   - ComponentCalls.Withs.Instances
//   - ComponentCalls.Withs.Instances.Group
//   - ComponentCalls.Withs.Instances.AST
//
// Depends on Fields: None
func (z *analyzer) AggregateWiths(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("withs")
	logger.Debug("Aggregating withs")

	scope, _ := cc.AST.Body.(*ast.Scope)
	if scope == nil {
		logger.Debug("Call has no body or no scope, skipping")
		return
	}

	cc.Withs = make([]*file.With, 0, 16)

	walk.WalkT(scope, func(ctx *walk.ContextT[*ast.With]) error {
		instance := &file.WithInstance{AST: ctx.Node}

		name := ctx.Node.Block()

		group := cc.WithByName(name)
		if group != nil {
			instance.Group = group
			group.Instances = append(group.Instances, instance)
			return walk.NoDive
		}

		group = &file.With{
			Name:      name,
			Instances: make([]*file.WithInstance, 1, 8),
		}
		instance.Group = group
		group.Instances[0] = instance
		return walk.NoDive
	}, walk.DontDiveAny(&ast.ComponentCall{}))

	cc.Withs = slices.Clip(cc.Withs)
	for _, group := range cc.Withs {
		group.Instances = slices.Clip(group.Instances)
	}
}
