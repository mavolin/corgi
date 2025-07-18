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
	logger.Debug("Checking statement")

	ch.CheckNoGoto(logger, parents, f, s)
	ch.CheckFallthroughOnlyInSwitch(logger, parents, f, s)
}

func (ch *checker) CheckNoGoto(logger *slog.Logger, parents []*walk.Context, f *file.File, s *ast.Statement) {
	logger = logger.WithGroup("no_goto")
	logger.Debug("Checking that no goto statement is used")

	if s.Parsed != nil {
		logger.Debug("Parsed statement, skipping")
		return
	} else if len(s.Nodes) != 1 {
		logger.Debug("Multiple code nodes, skipping")
		return
	}

	gc, _ := s.Nodes[0].(*ast.GoCode)
	if gc == nil {
		logger.Debug("No GoCode, skipping")
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

func (ch *checker) CheckFallthroughOnlyInSwitch(logger *slog.Logger, parents []*walk.Context, f *file.File, s *ast.Statement) {
	logger = logger.WithGroup("fallthrough_only_in_switch").
		With("statement_pos", s.Start().String())
	logger.Debug("Checking that fallthrough is only used in switch statements")

	if s.Parsed == nil {
		logger.Debug("Not a ParsedStatement, skipping")
		return
	}
	ft, _ := s.Parsed.(*ast.Fallthrough)
	if ft == nil {
		logger.Debug("Not a Fallthrough, skipping")
		return
	}

	sw := walk.Closest[*ast.Case](parents)
	if sw != nil {
		return
	}

	logger.Error("Illegal use of fallthrough statement outside of switch case")
	ch.Report(&diagnostic.Diagnostic{
		Message: "fallthrough statement outside of switch case",
		Primary: []diagnostic.Annotation{
			anno.Node(f, ft, "`fallthrough`s can only be used inside a switch case"),
		},
	})
}
