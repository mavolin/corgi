package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (z *analyzer) AnalyzeStates() {
	logger := z.Logger.WithGroup("state")
	logger.Debug("Analyzing state variables")

	for _, s := range z.Pkg.State {
		z.AnalyzeState(logger, s)
	}
}

func (z *analyzer) AnalyzeState(logger *slog.Logger, s *file.State) {
	if s.Analyzed {
		return
	}

	logger = logger.With(
		slog.String("file", string(s.File.Name)),
		slog.String("name", string(s.Name())),
		slog.String("pos", s.AST.Start().String()))

	z.AnalyzeState_InferredType(logger, s)

	s.Analyzed = true
}

// ============================================================================
// Inferred Type
// ======================================================================================

type state_InferredType struct{}

func (z *analyzer) AnalyzeState_InferredType(logger *slog.Logger, s *file.State) {
	defer z.Ran(s, state_InferredType{})

	if s.AST.Type != nil {
		s.InferredType.SetZero()
		return
	}

	t, _ := InferType(s.File, s.ValueNode())
	if t != "" {
		s.InferredType.SetResult(t)
		return
	}

	s.InferredType.SetFailed()
	logger.Error("Unable to infer type")
	z.Report(&diagnostic.Diagnostic{
		Message: "state: unable to infer type",
		Primary: []diagnostic.Annotation{
			anno.Node(s.File, s.NameNode(), "this state variable has no explicit type,\n"+
				"and no type could be inferred from the default"),
		},
		Hints: []diagnostic.Hint{
			{
				Hint:    "Give this variable an explicit type.",
				Example: "`" + string(s.Name()) + " foo = ...`",
			},
		},
	})
}
