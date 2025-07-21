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
	logger.Info("Checking components")

	for _, c := range ch.P.Components {
		logger := logger.With(
			slog.String("file", c.File.Name),
			slog.String("comp", c.Header().Name.Name),
			slog.String("comp_pos", c.Start().String()))
		logger.Debug("Checking component")

		ch.CheckAliasDoesntOverwriteRequiredParams(logger, c)
		ch.CheckDuplicateComponentParams(logger, c)
		ch.CheckReservedComponentNames(logger, c)

		for _, param := range c.Parameters {
			logger := logger.With(
				slog.String("param", param.AST.Name.Name),
				slog.String("param_pos", param.AST.Name.Start().String()))
			logger.Debug("Checking parameter")

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
	logger.Debug("Checking that alias doesn't overwrite required parameters of child component")

	if c.AliasAST == nil {
		logger.Debug("Component is not an alias, skipping")
		return
	} else if c.AnalyzedWithErrors {
		logger.Debug("Component analyzed with errors, skipping")
		return
	}

	cc := c.File.ComponentCallByNode(c.AliasAST.ComponentCall)
	logger = logger.With(
		slog.String("aliased_component_package", cc.Component.File.Package.ImportPath),
		slog.String("aliased_component_file", cc.Component.File.Name),
		slog.String("aliased_component", cc.AST.Header.Name.Full()),
		slog.String("aliased_component_pos", cc.Component.Start().String()))
	if cc.Component.AnalyzedWithErrors {
		logger.Debug("Aliased component analyzed with errors, skipping")
		return
	}

	// All required parameters must either be set by the alias' component call
	// or inherited by the alias.
AliasedParams:
	for _, aliasedParam := range cc.Component.Parameters {
		if !aliasedParam.Required() {
			continue
		}

		logger := logger.With(
			slog.String("aliased_param", aliasedParam.AST.Name.Name),
			slog.String("aliased_param_pos", aliasedParam.AST.Name.Start().String()))
		logger.Debug("Checking required parameter")

		if cc.AST.Header.Arguments != nil {
			for _, arg := range cc.AST.Header.Arguments.Args {
				carg, _ := arg.(*ast.ComponentArgument)
				if carg == nil {
					continue
				}
				if aliasedParam.AST.Name.Name == carg.Name.Name {
					logger.Debug("Parameter set by component call, skipping")
					continue AliasedParams
				}
			}
		}

		if c.AliasAST.Header.Parameters != nil {
			for _, param := range c.AliasAST.Header.Parameters.Params {
				if aliasedParam.AST.Name.Name != param.Name.Name {
					continue
				}

				logger.Error("Alias overwrites required parameter of aliased component")
				ch.Report(&diagnostic.Diagnostic{
					Message: "component alias overwrites required parameter",
					Primary: []diagnostic.Annotation{
						anno.Anno(c.File, anno.Annotation{
							Highlight:  anno.HighlightNode(param),
							Context:    anno.ContextLines(c.Start(), param.End()),
							Annotation: "overwrites required parameter of the same name of aliased component",
						}),
					},
					Hints: []diagnostic.Hint{
						{Hint: "Rename this parameter."},
						{
							Hint: "Set the parameter of the same name in the component call to the aliased component, " +
								"so you fulfill the requirement constraint of the aliased component.",
						},
					},
				})
			}
		}
	}
}

// ============================================================================
// Duplicate Component Parameter Names
// ======================================================================================

func (ch *checker) CheckDuplicateComponentParams(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("duplicate_params")
	logger.Debug("Checking for duplicate component parameter names")

	if c.AnalyzedWithErrors {
		logger.Debug("Component analyzed with errors, skipping")
		return
	} else if len(c.Parameters) < 2 {
		logger.Debug("Component has less than two parameters, skipping")
		return
	}

	reported := ch.TakeStringSet()
	dupls := make([]*file.ComponentParameter, 0, len(c.Parameters))
	for i, a := range c.Parameters {
		// Only consider parameters defined by this component to avoid
		// unnecessary noise
		if a.Component != c {
			continue
		} else if reported.Contains(a.AST.Name.Name) {
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
		primaries[0] = anno.Node(c.File, a.AST.Name, "first defined here")
		for _, dupl := range dupls {
			primaries = append(primaries, anno.Node(c.File, dupl.AST.Name, "then again here"))
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
	logger.Debug("Checking that component is not using a reserved name")

	name := c.Header().Name.Name
	if name != "ctx" {
		return
	}

	logger.Error("Component uses reserved name")
	ch.Report(&diagnostic.Diagnostic{
		Message: "component uses reserved name",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, c.Header().Name, "this component is named `"+name+"`, which is a reserved name"),
		},
		Hints: []diagnostic.Hint{{Hint: "Rename this component."}},
	})
}

// ============================================================================
// Uppercase Component Parameter Names
// ======================================================================================

func (ch *checker) CheckUpperComponentParamName(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	logger = logger.WithGroup("no_upper_names")
	logger.Debug("Checking that parameter is not using an uppercase name")

	r, _ := utf8.DecodeRuneInString(param.AST.Name.Name)
	if unicode.IsUpper(r) {
		logger.Error("Component parameter uses uppercase name")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component parameter: use of uppercase name",
			Primary: []diagnostic.Annotation{
				anno.NChars(c.File, c.Header().Name.Start(), 1, "this letter must not be uppercase"),
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
	logger.Debug("Checking that parameter is not using a name starting with an underscore")

	if param.AST.Name.Name[0] == '_' {
		logger.Error("Component parameter uses name starting with an underscore")
		ch.Report(&diagnostic.Diagnostic{
			Message: "component parameter: use of name with underscore-prefix",
			Primary: []diagnostic.Annotation{
				anno.NChars(c.File, c.Header().Name.Start(), 1, "cannot use an underscore as first letter"),
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
	logger.Debug("Checking that parameter doesn't use a reserved name")

	name := param.AST.Name.Name
	if name != "ctx" {
		return
	}

	logger.Error("Component parameter uses reserved name")
	ch.Report(&diagnostic.Diagnostic{
		Message: "component parameter uses reserved name",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, c.Header().Name, "this parameter is named `"+name+"`, which is a reserved name"),
		},
		Hints: []diagnostic.Hint{{Hint: "Rename this parameter."}},
	})
}
