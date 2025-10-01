package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckType(logger *slog.Logger, f *file.File, parents []*walk.Context, t *ast.Type) {
	logger = logger.WithGroup("type").
		With(slog.String("type_pos", t.Start().String()))

	ch.CheckType_AttributeTypeAliasOnlyOnComponentParams(logger, f, parents, t)
}

// ============================================================================
// Attribute Type-Alias Only Used On Component Parameters
// ======================================================================================

func (ch *checker) CheckType_AttributeTypeAliasOnlyOnComponentParams(logger *slog.Logger, f *file.File, parents []*walk.Context, t *ast.Type) {
	logger = logger.WithGroup("attribute_type-alias_only_on_component_params")

	attributeType, _ := t.Parsed.(*ast.AttributeType)
	if attributeType == nil {
		return
	} else if len(parents) == 0 {
		return
	} else if _, isParam := parents[0].Node.(*ast.ComponentParameter); isParam {
		return
	}

	logger.Error("Attribute type-alias can only be used on component parameters")
	ch.Report(&diagnostic.Diagnostic{
		Message: "attribute type-alias used outside of component parameter",
		Primary: []diagnostic.Annotation{
			anno.Node(f, attributeType, "cannot use an attribute type-alias here"),
		},
		Explanation: "The attribute type-alias is meant to be used on component parameters, " +
			"to indicate that a parameter is used as an attribute.\n" +
			"Using it outside of a component parameter is pointless.",
		Hints: []diagnostic.Hint{
			{Hint: "Consider using the `safe.X` equivalent instead."},
		},
	})
}
