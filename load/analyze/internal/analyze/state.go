package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// AnalyzeStates analyzes the state variables of the package.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeStates() {
	logger := z.Logger.WithGroup("state")
	logger.Debug("Analyzing state variables")

	for _, s := range z.P.State {
		z.AnalyzeState(logger, s)
	}
}

// AnalyzeState analyzes the given state variable.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeState(logger *slog.Logger, s *file.State) {
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

// AnalyzeState_InferredType infers the type of the given state variable if it is not
// explicitly typed.
//
// Depends on Checks: None
//
// Sets Fields:
//   - State.InferredType
//
// Depends on Fields: None
func (z *analyzer) AnalyzeState_InferredType(logger *slog.Logger, s *file.State) {
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
