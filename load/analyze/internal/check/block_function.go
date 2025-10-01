package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckBlockFunction(logger *slog.Logger, f *file.File, parents []*walk.Context, bf *ast.BlockFunction) {
	logger = logger.WithGroup("block_function").
		With(slog.String("block_name", bf.Name()),
			slog.String("block_function_pos", bf.Start().String()))

	ch.CheckBlockFunction_BlockDefined(logger, f, parents, bf)
}

// ============================================================================
// Block Is Defined by the Component
// ======================================================================================

func (ch *checker) CheckBlockFunction_BlockDefined(logger *slog.Logger, f *file.File, parents []*walk.Context, bf *ast.BlockFunction) {
	logger = logger.WithGroup("block_defined")

	cAST, _ := parents[0].Node.(*ast.Component)
	if cAST == nil {
		return
	}

	c := f.Package.ComponentByNode(cAST)
	if c.BlockByName(file.Identifier(bf.Name())) != nil {
		return
	}

	logger.Error("Block function uses block not defined by the component")

	annoText := "component does not define a default block"
	if bf.BlockName != nil {
		annoText = "component does not define a block named `" + bf.BlockName.Name + "`"
	}
	ch.Report(&diagnostic.Diagnostic{
		Message: "block function: reference to undefined block",
		Primary: []diagnostic.Annotation{
			anno.Node(f, bf, annoText),
		},
	})
}
