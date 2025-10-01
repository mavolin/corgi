package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

// ============================================================================
// Break
// ======================================================================================

func (ch *checker) CheckBreak(logger *slog.Logger, f *file.File, parents []*walk.Context, b *ast.Break) {
	logger = logger.WithGroup("break").
		With(slog.String("break_pos", b.Start().String()))

	ch.CheckBreak_Allowed(logger, parents, f, b)
}

func (ch *checker) CheckBreak_Allowed(logger *slog.Logger, parents []*walk.Context, f *file.File, b *ast.Break) {
	logger = logger.WithGroup("break_allowed")

	if !walk.IsChildOf[*ast.For](parents) && !walk.IsChildOf[*ast.Switch](parents) {
		logger.Error("Break not in loop or switch")
		ch.Report(&diagnostic.Diagnostic{
			Message: "`break` used outside of loop and switch",
			Primary: []diagnostic.Annotation{
				anno.Node(f, b, "`break`s can only be used inside loops or switch statements"),
			},
		})
	}
}

// ============================================================================
// Continue
// ======================================================================================

func (ch *checker) CheckContinue(logger *slog.Logger, f *file.File, parents []*walk.Context, c *ast.Continue) {
	logger = logger.WithGroup("continue").
		With(slog.String("continue_pos", c.Start().String()))

	ch.CheckContinue_Allowed(logger, parents, f, c)
}

func (ch *checker) CheckContinue_Allowed(logger *slog.Logger, parents []*walk.Context, f *file.File, c *ast.Continue) {
	logger = logger.WithGroup("continue_allowed")

	if !walk.IsChildOf[*ast.For](parents) {
		logger.Error("Continue not in loop")
		ch.Report(&diagnostic.Diagnostic{
			Message: "`continue` used outside of loop",
			Primary: []diagnostic.Annotation{
				anno.Node(f, c, "`continue`s can only be used inside loops"),
			},
		})
	}
}

// ============================================================================
// Fallthrough
// ======================================================================================

func (ch *checker) CheckFallthrough(logger *slog.Logger, f *file.File, parents []*walk.Context, ft *ast.Fallthrough) {
	logger = logger.WithGroup("fallthrough").
		With(slog.String("fallthrough_pos", ft.Start().String()))

	ch.CheckFallthrough_Allowed(logger, parents, f, ft)
}

func (ch *checker) CheckFallthrough_Allowed(logger *slog.Logger, parents []*walk.Context, f *file.File, ft *ast.Fallthrough) {
	logger = logger.WithGroup("fallthrough_allowed")

	if !walk.IsChildOf[*ast.Switch](parents) {
		logger.Error("Fallthrough not in switch")
		ch.Report(&diagnostic.Diagnostic{
			Message: "`fallthrough` used outside of switch",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ft, "`fallthrough`s can only be used inside switch statements"),
			},
		})
	}
}
