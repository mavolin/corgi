// Package check contains checks that do not influence the outcome of the
// analysis or that need to be run after analysis.
package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
	"github.com/mavolin/corgi/v2/load/analyze/internal/context"
)

type checker struct {
	*context.Context
	Logger      *slog.Logger
	BuiltinPath string
}

type (
	identifier = string
)

func Check(ctx *context.Context) {
	c := &checker{
		Context: ctx,
		Logger:  ctx.Logger.WithGroup("check"),
	}
	c.Logger.Debug("Running checks")

	c.CheckComponents()
	c.CheckComponentCalls()
	c.CheckState()
	c.CheckScope()
}

func (ch *checker) CheckScope() {
	logger := ch.Logger.WithGroup("scope")
	logger.Debug("Checking scope")

	for _, f := range ch.P.Files {
		logger := logger.With(slog.String("file", f.Name))

		walk.Walk(f.AST, func(ctx *walk.Context) error {
			switch n := ctx.Node.(type) {
			case *ast.And:
				ch.CheckAnd(logger, f, ctx.Parents, n)
			case *ast.Arguments:
				ch.CheckArguments(logger, f, ctx.Parents, n)
			case *ast.BlockFunction:
				ch.CheckBlockFunction(logger, f, ctx.Parents, n)
			case *ast.Break:
				ch.CheckBreak(logger, f, ctx.Parents, n)
			case *ast.Continue:
				ch.CheckContinue(logger, f, ctx.Parents, n)
			case *ast.Element:
				ch.CheckElement(logger, f, ctx.Parents, n)
			case *ast.Fallthrough:
				ch.CheckFallthrough(logger, f, ctx.Parents, n)
			case *ast.Statement:
				ch.CheckStatement(logger, f, ctx.Parents, n)
			case *ast.Type:
				ch.CheckAttributeTypeAliasOnlyOnComponentParams(logger, f, ctx.Parents, n)
			}
			return nil
		})
	}
}
