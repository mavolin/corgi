package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// AnalyzeComponentParameters analyzes the parameters of the given component.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields:
//   - Components.Parameters
//   - Components.Parameters.AST
func (z *analyzer) AnalyzeComponentParameters(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("parameters")

	for _, param := range c.Parameters {
		if param.Component != c {
			// Only analyze parameter that belong to the component
			continue
		}

		logger := logger.With(
			slog.String("param", param.AST.Name.Name),
			slog.String("param_pos", param.AST.Name.Start().String()))

		z.AnalyzeAttrTypeComponentParam(logger, param)
		z.InferTypeFromComponentParamDefault(logger, param)
	}
}

// ============================================================================
// Infer Type From Attribute Type Component Parameters
// ======================================================================================

// AnalyzeAttrTypeComponentParam analyzes the given component parameter to
// infer its type, if it uses an attribute type as its type.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Parameters.AttributeType
//   - Components.Parameters.AttributeName
//
// Depends on Fields: None
func (z *analyzer) AnalyzeAttrTypeComponentParam(logger *slog.Logger, param *file.ComponentParameter) {
	logger = logger.WithGroup("attr_type_param")

	if param.AST.Type == nil || param.AST.Type.Parsed == nil {
		return
	}

	t, _ := param.AST.Type.Parsed.(*ast.AttributeType)
	if t == nil {
		return
	}

	if t.Name.Type == attrtype.Innocuous {
		logger.Error("Use of innocuous attribute type as component parameter type")
		z.Report(&diagnostic.Diagnostic{
			Message: "component parameter: use of innocuous attribute type",
			Primary: []diagnostic.Annotation{
				anno.Node(param.Component.File, param.AST.Type, "`'innocuous` is equivalent to `string` in every way"),
			},
			Hints: []diagnostic.Hint{{Hint: "If you want to accept any printable value, use `escape.Printable`."}},
		})
		return
	}

	if t.Name.Type == attrtype.Unsafe || t.Name.Type == attrtype.UnsafeBool {
		if t.Attribute == nil {
			logger.Error("Use of unsafe attribute type as component parameter type without explicit attribute name")
			z.Report(&diagnostic.Diagnostic{
				Message: "component parameter: use of unsafe attribute type without explicit attribute name",
				Primary: []diagnostic.Annotation{
					anno.Node(param.Component.File, t, "you need to specify an attribute name in brackets"),
				},
				Examples: []diagnostic.Example{{Example: "`'unsafe[myattr]"}},
				Explanation: "The `unsafe` and `unsafeBool` attribute types must always be tied to a specific attribute. " +
					"That way, it can be ensured that an unsafe value deemed safe for one attribute isn't used for another" +
					"kind of attribute, where it might not be safe.",
			})
		} else {
			param.AttributeName = t.Attribute.Name
		}
	}

	param.AttributeType = t.Name.Type
	param.InferredType = z.SafeImport(param.Component.File).Namespace + "."
	switch t.Name.Type {
	case attrtype.Unsafe:
		param.InferredType += "Unsafe"
	case attrtype.UnsafeBool:
		param.InferredType += "UnsafeBool"
	case attrtype.Bool:
		param.InferredType += "Bool"
	case attrtype.Innocuous:
		// already handled above
	case attrtype.Text:
		param.InferredType = "string"
	case attrtype.CSS:
		param.InferredType += "CSS"
	case attrtype.JS:
		param.InferredType += "JS"
	case attrtype.URL:
		param.InferredType += "URL"
	case attrtype.URLList:
		param.InferredType += "URLList"
	case attrtype.ResourceURL:
		param.InferredType += "ResourceURL"
	case attrtype.Srcset:
		param.InferredType += "Srcset"
	case attrtype.Unknown:
		fallthrough
	default:
		logger.Error("Use of unknown attribute type as component parameter type")
		z.Report(&diagnostic.Diagnostic{
			Type:    diagnostic.InternalError,
			Message: "component parameter: use of unknown attribute type",
			Primary: []diagnostic.Annotation{
				anno.Node(param.Component.File, t.Name, "unknown attribute type"),
			},
			Explanation: "This error can occur in one of two ways:\n" +
				"If you are not running the corgi CLI, most likely, at some place in the program, " +
				"the value for this attribute type was set to an illegal value.\n" +
				"It could also be that the parser was extended to support a new attribute type, " +
				"but the analyzer was not updated to support it.\n\n" +
				"In any case: If you are running the corgi CLI, please open an issue, this is a bug.",
		})
	}
}

// ============================================================================
// Infer Type From Default
// ======================================================================================

// InferTypeFromComponentParamDefault infers the type of the given component
// parameter from its default value, if it has one.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Parameters.InferredType
//
// Depends on Fields: None
func (z *analyzer) InferTypeFromComponentParamDefault(logger *slog.Logger, param *file.ComponentParameter) {
	logger = logger.WithGroup("infer_type_from_default").
		With(slog.String("param", param.AST.Name.Name),
			slog.String("param_pos", param.AST.Name.Start().String()))

	if param.AST.Type != nil {
		return
	} else if param.InferredType != "" {
		return
	}

	if param.AST.Default == nil {
		logger.Error("No default value set for untyped component parameter")
		z.Report(&diagnostic.Diagnostic{
			Message: "component parameter: neither type nor default value set",
			Primary: []diagnostic.Annotation{
				anno.Node(param.Component.File, param.AST.Name, "unable to determine type of this parameter"),
			},
			Explanation: "Every parameter must either have an explicit type, or a default value " +
				"from which the type can be inferred.",
			Hints: []diagnostic.Hint{
				{
					Hint:    "Add an explicit type to this parameter.",
					Example: "`" + param.AST.Name.Name + " MyType`",
				}, {
					Hint:    "Set a default value from which the type can be inferred.",
					Example: "`" + param.AST.Name.Name + " = 123`",
				},
			},
		})
		return
	}

	param.InferredType, _ = file.InferType(param.Component.File, param.AST.Default)
	if param.InferredType == "" {
		logger.Error("Unable to infer type from default value")
		z.Report(&diagnostic.Diagnostic{
			Message: "component parameter: unable to infer type from default value",
			Primary: []diagnostic.Annotation{
				anno.Node(param.Component.File, param.AST.Name, "has no explicit type"),
				anno.Node(param.Component.File, param.AST.Default, "cannot infer type of this expression"),
			},
			Explanation: "A component parameter's type is only optional, if the type can be " +
				"inferred from the default value. If that is not possible, specify the type as you" +
				"normally would behind the parameter's name.",
			Examples: []diagnostic.Example{
				{Example: "`" + param.AST.Name.Name + " MyType = ...`"},
			},
		})
	}
}
