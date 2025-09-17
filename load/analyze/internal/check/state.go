package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (ch *checker) CheckState() {
	logger := ch.Logger.WithGroup("state")
	logger.Info("Checking state variables")

	ch.CheckStateDuplicates(logger)

	for _, s := range ch.P.State {
		logger := logger.With(
			slog.String("file", string(s.File.Name)),
			slog.String("name", s.Name().Name))

		ch.CheckStateUnexported(logger, s)
	}
}

func (ch *checker) CheckStateDuplicates(logger *slog.Logger) {
	logger = logger.WithGroup("duplicates")

	if len(ch.P.State) <= 1 {
		logger.Debug("One or no state variables, skipping")
		return
	}

	reported := make(map[string]bool)
	dupls := make([]*file.State, 0, len(ch.P.State)-1)

	for ai, a := range ch.P.State[:len(ch.P.State)-1] {
		logger := logger.With(
			slog.String("file", string(a.File.Name)),
			slog.String("name", a.Name().Name))

		if reported[a.Name().Name] {
			continue
		}

		dupls = dupls[:0]

		for _, b := range ch.P.State[ai:] {
			if a.Name().Name != b.Name().Name {
				continue
			}

			dupls = append(dupls, b)
			reported[b.Name().Name] = true
		}

		if len(dupls) > 0 {
			logger.Error("Found duplicate state variables")

			primaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
			primaries[0] = anno.Node(a.File, a.Name(), "first defined here")
			for _, b := range dupls[1:] {
				primaries = append(primaries, anno.Node(b.File, b.Name(), "duplicate"))
			}

			ch.Report(&diagnostic.Diagnostic{
				Message: "state variable defined multiple times",
				Primary: primaries,
			})
		}
	}
}

func (ch *checker) CheckStateUnexported(logger *slog.Logger, s *file.State) {
	logger = logger.WithGroup("unexported")

	if file.IsExported(file.Identifier(s.Name().Name)) {
		logger.Error("state variable is exported")
		ch.Report(&diagnostic.Diagnostic{
			Message: "exported state variable",
			Primary: []diagnostic.Annotation{
				anno.Node(s.File, s.Name(), "state variables must not be exported"),
			},
			Hints: []diagnostic.Hint{{Hint: "Start the name with a lowercase letter."}},
		})
	}
}
