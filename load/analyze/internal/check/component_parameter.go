package check

import (
	"log/slog"
	"unicode"
	"unicode/utf8"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (ch *checker) CheckComponentParameters(logger *slog.Logger, c *file.Component) {
	for _, param := range c.Parameters {
		logger := logger.With(
			slog.String("param", string(param.Name)),
			slog.String("param_pos", param.AST.Name.Start().String()))

		ch.CheckComponentParameter_NoReservedName(logger, c, param)
		ch.CheckComponentParameter_NoUpperName(logger, c, param)
		ch.CheckComponentParameter_NoUnderscoreName(logger, c, param)
	}
}

// ============================================================================
// No Uppercase Names
// ======================================================================================

func (ch *checker) CheckComponentParameter_NoUpperName(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	logger = logger.WithGroup("no_upper_names")

	r, _ := utf8.DecodeRuneInString(string(param.Name))
	if unicode.IsUpper(r) {
		logger.Error("Component parameter uses uppercase name")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component parameter: use of uppercase name",
			Primary: []diagnostic.Annotation{
				anno.NRunes(c.File, c.AST.Header.Name.Start(), 1, "this letter must not be uppercase"),
			},
			Hints: []diagnostic.Hint{{Hint: "Rename this parameter."}},
		})
	}
}

// ============================================================================
// No Parameter Starts With an Underscore
// ======================================================================================

func (ch *checker) CheckComponentParameter_NoUnderscoreName(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	logger = logger.WithGroup("no_underscore_names")

	if param.Name[0] == '_' {
		logger.Error("Component parameter uses name starting with an underscore")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component parameter: use of name with underscore-prefix",
			Primary: []diagnostic.Annotation{
				anno.NRunes(c.File, c.AST.Header.Name.Start(), 1, "cannot use an underscore as first letter"),
			},
			Hints: []diagnostic.Hint{{Hint: "Rename this parameter."}},
		})
	}
}

// ============================================================================
// No Parameter Uses Reserved Name
// ======================================================================================

func (ch *checker) CheckComponentParameter_NoReservedName(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	logger = logger.WithGroup("no_reserved_param_name")

	name := param.Name
	if name != "ctx" {
		return
	}

	logger.Error("Component parameter uses reserved name")
	ch.Report(&diagnostic.Diagnostic{
		Message: "component parameter uses reserved name",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, param.AST.Name, "`"+string(name)+"` is a reserved name"),
		},
		Hints: []diagnostic.Hint{{Hint: "Rename this parameter."}},
	})
}
