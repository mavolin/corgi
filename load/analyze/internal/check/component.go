package check

import (
	"log/slog"
	"unicode"
	"unicode/utf8"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (ch *checker) CheckComponents() {
	logger := ch.Logger.WithGroup("components")
	logger.Debug("Checking components")

	for _, c := range ch.P.Components {
		logger := logger.With(
			slog.String("file", string(c.File.Name)),
			slog.String("comp", c.AST.Header.Name.Name),
			slog.String("comp_pos", c.AST.Start().String()))

		ch.CheckDuplicateComponentParams(logger, c)
		ch.CheckReservedComponentNames(logger, c)
		ch.CheckDataComponent(logger, c)

		for _, param := range c.Parameters {
			logger := logger.With(
				slog.String("param", param.AST.Name.Name),
				slog.String("param_pos", param.AST.Name.Start().String()))

			ch.CheckReservedComponentParamName(logger, c, param)
			ch.CheckUpperComponentParamName(logger, c, param)
			ch.CheckUnderscoreComponentParamName(logger, c, param)
		}
	}
}

// ============================================================================
// Duplicate Component Parameter Names
// ======================================================================================

func (ch *checker) CheckDuplicateComponentParams(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("duplicate_params")

	if len(c.Parameters) < 2 {
		return
	}

	params := make(map[string][]*file.ComponentParameter, len(c.Parameters))
	for _, param := range c.Parameters {
		params[param.AST.Name.Name] = append(params[param.AST.Name.Name], param)
	}

	for _, dupls := range params {
		if len(dupls) == 1 {
			continue
		}

		primaries := make([]diagnostic.Annotation, len(dupls))
		for i, dupl := range dupls {
			primaries[i] = anno.Node(c.File, dupl.AST, "defined here")
		}

		logger.Error("Component has duplicate parameter names")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component: parameter defined multiple times",
			Primary: primaries,
			Hints:   []diagnostic.Hint{{Hint: "Remove or rename all but one of these parameters."}},
		})
	}
}

// ============================================================================
// Reserved Component Names
// ======================================================================================

func (ch *checker) CheckReservedComponentNames(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("reserved_names")

	name := c.AST.Header.Name.Name
	if name != "ctx" {
		return
	}

	logger.Error("Component uses reserved name")
	ch.Report(&diagnostic.Diagnostic{
		Message: "component uses reserved name",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, c.AST.Header.Name, "`"+name+"` is a reserved name"),
		},
		Hints: []diagnostic.Hint{{Hint: "Rename this component."}},
	})
}

// ============================================================================
// Uppercase Component Parameter Names
// ======================================================================================

func (ch *checker) CheckUpperComponentParamName(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	logger = logger.WithGroup("no_upper_names")

	r, _ := utf8.DecodeRuneInString(param.AST.Name.Name)
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
// Component Parameter Starts With Underscore
// ======================================================================================

func (ch *checker) CheckUnderscoreComponentParamName(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	logger = logger.WithGroup("no_underscore_names")

	if param.AST.Name.Name[0] == '_' {
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
// Component Parameter Uses Reserved Name
// ======================================================================================

func (ch *checker) CheckReservedComponentParamName(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	logger = logger.WithGroup("reserved_param_name")

	name := param.AST.Name.Name
	if name != "ctx" {
		return
	}

	logger.Error("Component parameter uses reserved name")
	ch.Report(&diagnostic.Diagnostic{
		Message: "component parameter uses reserved name",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, param.AST.Name, "`"+name+"` is a reserved name"),
		},
		Hints: []diagnostic.Hint{{Hint: "Rename this parameter."}},
	})
}

// ============================================================================
// For Every Component Foo, No FooData Component Exists
// ======================================================================================

func (ch *checker) CheckDataComponent(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("no_data_component")

	name := c.Name
	if len(name) <= len("Data") || name[len(name)-len("Data"):] != "Data" {
		return
	}

	name = name[:len(name)-len("Data")] // remove "Data" suffix
	other := c.File.Package.ComponentByName(name)
	if other == nil {
		return
	}

	logger.Error("Component with suffix `Data` defined",
		slog.String("other_file", string(other.File.Name)),
		slog.String("other_comp", other.AST.Header.Name.Name),
		slog.String("other_comp_pos", other.AST.Start().String()))
	ch.Report(&diagnostic.Diagnostic{
		Message: "data component defined",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, c.AST.Header.Name, "cannot define component with suffix `Data`"),
		},
		Secondary: []diagnostic.Annotation{
			anno.Node(other.File, other.AST.Header.Name, "because another component without the `Data` suffix exists"),
		},
		Explanation: "For every component `Foo`, corgi creates a struct `FooData` used for calling the component " +
			"in Go code. Since `" + string(name) + "`'s generated data struct and `" + string(name) + "Data` would clash, " +
			"you cannot define a component called `" + string(name) + "Data`.",
		Hints: []diagnostic.Hint{{Hint: "Rename `" + string(name) + "Data`."}},
	})
}
