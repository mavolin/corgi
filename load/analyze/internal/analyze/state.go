package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// AnalyzeState analyzes the state variables of the package.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeState() {
	logger := z.Logger.WithGroup("state")
	logger.Debug("Analyzing state variables")

	for _, s := range z.P.State {
		logger := logger.With(
			slog.String("file", s.File.Name),
			slog.String("name", s.Name().Name))

		z.InferStateType(logger, s)
	}
}

// InferStateType infers the type of the given state variable if it is not
// explicitly typed.
//
// Depends on Checks: None
//
// Sets Fields:
//   - State.InferredType
//
// Depends on Fields: None
func (z *analyzer) InferStateType(logger *slog.Logger, s *file.State) {
	if s.AST.Type != nil {
		return
	}

	s.InferredType, _ = file.InferType(s.File, s.Value())
	if s.InferredType != "" {
		return
	}

	s.AnalyzedWithErrors = true
	logger.Error("Unable to infer type")
	z.Report(&diagnostic.Diagnostic{
		Message: "state: unable to infer type",
		Primary: []diagnostic.Annotation{
			anno.Node(s.File, s.Name(), "this state variable has no explicit type,\n"+
				"and no type could be inferred from the default"),
		},
		Hints: []diagnostic.Hint{
			{
				Hint:    "Give this variable an explicit type.",
				Example: "`" + s.Name().Name + " foo = ...`",
			},
		},
	})
}
