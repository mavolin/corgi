package check

import (
	"fmt"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckArguments(logger *slog.Logger, f *file.File, _ []*walk.Context, a *ast.Arguments) {
	logger = logger.WithGroup("arguments").
		With(slog.String("arguments_pos", a.Start().String()))

	for _, arg := range a.List {
		logger := logger.With(
			slog.String("arg_pos", arg.Start().String()),
			slog.String("arg_type", fmt.Sprintf("%T", arg)))

		switch arg := arg.(type) {
		case ast.Attribute:
			ch.CheckAttribute(logger, f, f.AttributeByNode(arg))
		}
	}
}
