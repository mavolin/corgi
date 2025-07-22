package check

import (
	"log/slog"
	"unicode"
	"unicode/utf8"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (ch *checker) CheckComponents() {
	logger := ch.Logger.WithGroup("components")
	logger.Debug("Checking components")

	for _, c := range ch.P.Components {
		logger := logger.With(
			slog.String("file", c.File.Name),
			slog.String("comp", c.Header().Name.Name),
			slog.String("comp_pos", c.Start().String()))

		ch.CheckAliasDoesntOverwriteRequiredParams(logger, c)
		ch.CheckDuplicateComponentParams(logger, c)
		ch.CheckReservedComponentNames(logger, c)

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
// Alias Doesn't Overwrite Required Parameter of Child Component
// ======================================================================================

func (ch *checker) CheckAliasDoesntOverwriteRequiredParams(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("alias_doesnt_overwrite_required_params")

	if c.AliasAST == nil || c.AnalyzedWithErrors {
		return
	}

	aliasedCall := c.File.ComponentCallByNode(c.AliasAST.ComponentCall)
	if aliasedCall == nil || aliasedCall.Component == nil || aliasedCall.Component.AnalyzedWithErrors {
		return
	}

	aliasedComponent := aliasedCall.Component
	logger = logger.With(
		slog.String("aliased_component", aliasedCall.AST.Header.Name.Full()),
		slog.String("aliased_package", aliasedComponent.File.Package.ImportPath),
		slog.String("aliased_file", aliasedComponent.File.Name))

	providedParams := make(map[string]bool)
	if aliasedCall.AST.Header != nil {
		for _, arg := range aliasedCall.AST.Header.Arguments.Args {
			carg, _ := arg.(*ast.ComponentArgument)
			if carg != nil && carg.Name != nil {
				providedParams[carg.Name.Name] = true
			}
		}
	}

	// All required parameters must either be set by the aliased component call
	// or inherited by the alias.
	for _, aliasedParam := range aliasedComponent.Parameters {
		if !aliasedParam.Required() {
			continue
		}

		name := aliasedParam.AST.Name.Name
		logger := logger.With(slog.String("required_param", name))

		// aliased component call sets this parameter
		if providedParams[name] {
			continue
		}

		for _, param := range c.AliasAST.Header.Parameters.Params {
			if name != param.Name.Name {
				continue
			}

			logger.Error("Alias overwrites required parameter of aliased component")
			ch.Report(&diagnostic.Diagnostic{
				Message: "component alias overwrites required parameter",
				Primary: []diagnostic.Annotation{
					anno.Anno(c.File, anno.Annotation{
						Highlight:  anno.HighlightNode(param),
						Context:    anno.ContextLines(c.Start(), param.End()),
						Annotation: "overwrites required parameter of same name on aliased component",
					}),
				},
				Secondary: []diagnostic.Annotation{
					anno.Node(aliasedComponent.File, aliasedParam.AST, "required parameter defined here"),
				},
				Hints: []diagnostic.Hint{
					{Hint: "Rename this parameter to avoid the conflict."},
					{Hint: "Set this parameter in the component call to the aliased component: "},
				},
			})

			break // no need to check other parameters with the same name
		}
	}
}

// ============================================================================
// Duplicate Component Parameter Names
// ======================================================================================

func (ch *checker) CheckDuplicateComponentParams(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("duplicate_params")

	if c.AnalyzedWithErrors {
		return
	} else if len(c.Parameters) < 2 {
		return
	}

	reported := make(map[string]bool, len(c.Parameters))
	dupls := make([]*file.ComponentParameter, 0, len(c.Parameters))
	for i, a := range c.Parameters {
		// Only consider parameters defined by this component to avoid
		// unnecessary noise
		if a.Component != c {
			continue
		} else if reported[a.AST.Name.Name] {
			continue
		}

		dupls = dupls[:0]
		for _, b := range c.Parameters[i:] {
			if a.AST.Name.Name == b.AST.Name.Name {
				dupls = append(dupls, b)
			}
		}

		if len(dupls) == 0 {
			continue
		}

		primaries := make([]diagnostic.Annotation, 1, 1+len(dupls))
		primaries[0] = anno.Node(c.File, a.AST, "first defined here")
		for _, dupl := range dupls {
			primaries = append(primaries, anno.Node(c.File, dupl.AST, "then again here"))
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

	name := c.Header().Name.Name
	if name != "ctx" {
		return
	}

	logger.Error("Component uses reserved name")
	ch.Report(&diagnostic.Diagnostic{
		Message: "component uses reserved name",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, c.DefinedAST.Header.Name, "`"+name+"` is a reserved name"),
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
				anno.NRunes(c.File, c.Header().Name.Start(), 1, "this letter must not be uppercase"),
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
				anno.NRunes(c.File, c.Header().Name.Start(), 1, "cannot use an underscore as first letter"),
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
