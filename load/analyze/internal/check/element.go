package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (ch *checker) CheckElement(logger *slog.Logger, f *file.File, _ []*walk.Context, e *ast.Element) {
	logger = logger.WithGroup("element").With(
		slog.String("element", e.Header.Name.Name.Name),
		slog.String("element_pos", e.Header.Name.Start().String()))
	logger.Debug("Checking element")

	ch.CheckElementNoEmptyAttributeList(logger, f, e)
	ch.CheckAttributeListContainsOnlyAttributes(logger, f, e)
}

func (ch *checker) CheckElementNoEmptyAttributeList(logger *slog.Logger, f *file.File, e *ast.Element) {
	logger = logger.WithGroup("no_empty_attribute_list")
	logger.Debug("Checking for empty attribute list")

	if e.Header.Attributes == nil {
		logger.Debug("No attribute list, skipping")
		return
	} else if len(e.Header.Attributes.Args) > 0 {
		logger.Debug("At least one attribute, skipping")
		return
	}

	logger.Error("Empty attribute list")
	ch.Report(&diagnostic.Diagnostic{
		Message: "element: empty attribute list",
		Primary: []diagnostic.Annotation{
			anno.Node(f, e.Header.Attributes, "remove this empty attribute list"),
		},
		Hints: []diagnostic.Hint{
			{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
		},
	})
}

func (ch *checker) CheckAttributeListContainsOnlyAttributes(logger *slog.Logger, f *file.File, e *ast.Element) {
	logger = logger.WithGroup("attribute_list_contains_only_attributes")
	logger.Debug("Checking that attribute list contains only attributes")

	if e.Header.Attributes == nil {
		logger.Debug("No attributes, skipping")
		return
	}

	for _, arg := range e.Header.Attributes.Args {
		logger := logger.With(slog.String("pos", arg.Start().String()))
		logger.Debug("Checking argument")

		if _, ok := arg.(ast.Attribute); ok {
			logger.Debug("Argument is an attribute")
			continue
		}

		logger.Error("Attribute list contains non-attribute")
		ch.Report(&diagnostic.Diagnostic{
			Message: "element: attribute list contains non-attribute",
			Primary: []diagnostic.Annotation{
				anno.Node(f, arg, "this is not an attribute"),
			},
			Hints: []diagnostic.Hint{
				{
					Hint: "If this is supposed to be a component call, " +
						"remember to add a colon (`:`) before the component name.",
				},
			},
		})
	}
}
