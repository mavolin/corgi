package analyze

import (
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
	logger.Debug("Aggregating component call data")

	for _, f := range z.P.Files {
		for _, cc := range f.ComponentCalls {
			z.AggregateWiths(cc)
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
func (z *analyzer) AggregateWiths(cc *file.ComponentCall) {
	scope, _ := cc.AST.Body.(*ast.Scope)
	if scope == nil {
		return
	}

	cc.Withs = make([]*file.With, 0, 16)

	walk.WalkT(scope, func(ctx *walk.ContextT[*ast.With]) error {
		instance := &file.WithInstance{AST: ctx.Node}

		name := ctx.Node.Name()

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
