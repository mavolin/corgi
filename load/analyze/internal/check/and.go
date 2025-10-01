package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckAnd(logger *slog.Logger, f *file.File, _ []*walk.Context, a *ast.And) {
	logger = logger.WithGroup("and").
		With(slog.String("and_pos", a.Start().String()))

	ch.CheckAnd_NoEmptyAttributeList(logger, f, a)
	ch.CheckAnd_ContainsOnlyAttributes(logger, f, a)
}

// ============================================================================
// No Empty Attribute List, But Present Parentheses
// ======================================================================================

func (ch *checker) CheckAnd_NoEmptyAttributeList(logger *slog.Logger, f *file.File, a *ast.And) {
	logger = logger.WithGroup("no_empty_attribute_list")

	if a.Attributes == nil {
		return
	} else if len(a.Attributes.List) > 0 {
		return
	}

	logger.Error("Empty attribute list")
	ch.Report(&diagnostic.Diagnostic{
		Message: "and: empty attribute list",
		Primary: []diagnostic.Annotation{
			anno.Node(f, a.Attributes, "remove this empty attribute list"),
		},
		Hints: []diagnostic.Hint{
			{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
		},
	})
}

// ============================================================================
// Contains Only Attributes, No Component Arguments
// ======================================================================================

func (ch *checker) CheckAnd_ContainsOnlyAttributes(logger *slog.Logger, f *file.File, a *ast.And) {
	logger = logger.WithGroup("contains_only_attributes")

	if a.Attributes == nil {
		return
	}

	for _, arg := range a.Attributes.List {
		logger := logger.With(slog.String("pos", arg.Start().String()))

		if _, ok := arg.(ast.Attribute); ok {
			continue
		}

		logger.Error("Attribute list contains non-attribute")
		ch.Report(&diagnostic.Diagnostic{
			Message: "and: attribute list contains non-attribute",
			Primary: []diagnostic.Annotation{
				anno.Node(f, arg, "this is not an attribute"),
			},
			Hints: []diagnostic.Hint{
				{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
			},
		})
	}
}
