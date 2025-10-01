package check

import (
	"log/slog"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckStatement(logger *slog.Logger, f *file.File, parents []*walk.Context, s *ast.Statement) {
	logger = logger.WithGroup("statements").
		With("statement_pos", s.Start().String())

	ch.CheckStatement_NoGoto(logger, parents, f, s)
}

// ============================================================================
// No Goto Statements
// ======================================================================================

func (ch *checker) CheckStatement_NoGoto(logger *slog.Logger, parents []*walk.Context, f *file.File, s *ast.Statement) {
	logger = logger.WithGroup("no_goto")

	if s.Parsed != nil {
		return
	} else if len(s.Nodes) != 1 {
		return
	}

	gc, _ := s.Nodes[0].(*ast.GoCode)
	if gc == nil {
		return
	}

	if strings.HasPrefix(gc.Code, "goto") {
		logger.Error("Illegal use of Goto statement")
		explanation := "`goto`-Statements cannot be placed at the top level of a file."
		if len(parents) > 0 {
			if _, ok := parents[0].Node.(*ast.Component); ok {
				explanation = "`goto`-Statements disrupt the control flow of the program " +
					"and can lead to inconsistencies in template-internal state."
			}
		}
		ch.Report(&diagnostic.Diagnostic{
			Message: "use of goto statement",
			Primary: []diagnostic.Annotation{
				anno.Node(f, gc, "remove this goto statement"),
			},
			Explanation: explanation,
		})
	}
}
