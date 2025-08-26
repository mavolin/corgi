// Package analyze implements the analysis of a corgi package.
//
// It establishes that certain preconditions are met, that are necessary to
// run the analysis.
// After that it runs the actual analysis.
// A positive outcome is required to run the postchecks.
package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/load/analyze/internal/context"
)

type analyzer struct {
	*context.Context
	Logger                 *slog.Logger
	analyzedComponentCalls map[*file.ComponentCall]bool
}

func Analyze(ctx *context.Context) {
	var numCCs int
	for _, f := range ctx.P.Files {
		numCCs += len(f.ComponentCalls)
	}

	z := &analyzer{
		Context:                ctx,
		Logger:                 ctx.Logger.WithGroup("analysis"),
		analyzedComponentCalls: make(map[*file.ComponentCall]bool, numCCs),
	}
	z.Logger.Info("Running analysis")

	if z.CheckPackageNamesMatch() {
		z.SetPackageName()
	}

	z.AnalyzeState()
	z.AnalyzeElementSpecs()

	z.AnalyzeComponents()
	z.AnalyzeComponentCalls()
	z.AnalyzeAttributeReferences()

	ctx.P.Analyzed = true
	for _, f := range ctx.P.Files {
		f.Analyzed = true
	}
}
