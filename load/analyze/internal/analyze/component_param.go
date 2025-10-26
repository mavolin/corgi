package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/meta"
)

func (z *analyzer) AnalyzeComponent_Parameters(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("parameters")

	for _, param := range c.Parameters {
		z.AnalyzeComponentParameter(logger, c, param)
	}
}

func (z *analyzer) AnalyzeComponentParameter(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	logger = logger.With(
		slog.String("param", string(param.Name)),
		slog.String("param_pos", param.AST.Name.Start().String()))

	z.AnalyzeComponentParameter_AttributeType_AttributeName(logger, c, param)
	z.AnalyzeComponentParameter_InferredType(logger, c, param)
}

// ============================================================================
// Infer Type From Attribute Type
// ======================================================================================

type (
	componentParameter_AttributeType struct{}
	componentParameter_AttributeName struct{}
)

func (z *analyzer) AnalyzeComponentParameter_AttributeType_AttributeName(
	logger *slog.Logger, c *file.Component, param *file.ComponentParameter,
) {
	defer z.Ran(c, componentParameter_AttributeType{})
	defer z.Ran(c, componentParameter_AttributeName{})
	logger = logger.WithGroup("attr_type_param")

	param.AttributeType.SetZero()
	param.AttributeName.SetZero()

	if param.AST.Type == nil || param.AST.Type.Parsed == nil {
		return
	}

	t, _ := param.AST.Type.Parsed.(*ast.AttributeType)
	if t == nil {
		return
	}

	if t.Name.Type == attrtype.String {
		logger.Error("Use of string attribute type as component parameter type")
		z.Report(&diagnostic.Diagnostic{
			Message: "component parameter: use of string attribute type",
			Primary: []diagnostic.Annotation{
				anno.Node(c.File, param.AST.Type, "`'string` is equivalent to `string` in every way"),
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
					anno.Node(c.File, t, "you need to specify an attribute name in brackets"),
				},
				Examples: []diagnostic.Example{{Example: "`'unsafe[myattr]"}},
				Explanation: "The `unsafe` and `unsafeBool` attribute types must always be tied to a specific attribute. " +
					"That way, it can be ensured that an unsafe value deemed safe for one attribute isn't used for another" +
					"kind of attribute, where it might not be safe.",
			})
			param.AttributeName.SetFailed()
		} else {
			param.AttributeName.SetResult(t.Attribute.Name)
		}
	}

	if t.Name.Type != nil {
		param.AttributeType.SetResult(t.Name.Type)
	} else {
		param.AttributeType.SetFailed()
	}
}

// ============================================================================
// Infer Type From Default
// ======================================================================================

type componentParameter_InferredType struct{}

func (z *analyzer) AnalyzeComponentParameter_InferredType(logger *slog.Logger, c *file.Component, param *file.ComponentParameter) {
	defer z.Ran(c, componentParameter_InferredType{})

	param.InferredType.SetZero()

	attributeType := z.componentParameter_AttributeType(param)
	if attributeType.Failed() || attributeType.NotZero() {
		z.AnalyzeComponentParameter_InferredType_fromAttributeType(logger, c, param)
	} else {
		z.AnalyzeComponentParameter_InferredType_fromDefault(logger, c, param)
	}
}

func (z *analyzer) AnalyzeComponentParameter_InferredType_fromAttributeType(
	logger *slog.Logger, c *file.Component, param *file.ComponentParameter,
) {
	attributeType := z.componentParameter_AttributeType(param)
	if attributeType.Failed() {
		param.InferredType.SetFailed()
		return
	}

	inferredType := z.attrtypeToSafeType(c.File, attributeType.Result())
	if inferredType != "" {
		param.InferredType.SetResult(inferredType)
		return
	}

	logger.Error("Use of unknown attribute type as component parameter type")
	explanation := "This error most likely occurred, because the parser was extended to support a new attribute type, " +
		"but the analyzer was not updated to support it.\n" +
		"\n" +
		"This is a bug, please open an issue."
	if !meta.CLI {
		explanation = "This error can occur in one of two ways:\n" +
			"Most likely, at some place in the program, " +
			"the value for this attribute type was set to an illegal value.\n" +
			"It could also be that the parser was extended to support a new attribute type, " +
			"but the analyzer was not updated to support it.\n" +
			"\n" +
			"In case of the latter: This is a bug, please open an issue."
	}

	var name *ast.AttributeTypeName
	if t, _ := param.AST.Type.Parsed.(*ast.AttributeType); t != nil {
		name = t.Name
	}

	var primary diagnostic.Annotation
	if name != nil {
		primary = anno.Node(c.File, name, "unknown attribute type")
	} else {
		primary = anno.Node(c.File, param.AST.Type, "unknown attribute type")
	}
	logger.Error("component parameter: use of unknown attribute type")
	z.Report(&diagnostic.Diagnostic{
		Type:        diagnostic.InternalError,
		Message:     "component parameter: use of unknown attribute type",
		Primary:     []diagnostic.Annotation{primary},
		Explanation: explanation,
	})
}

func (z *analyzer) attrtypeToSafeType(f *file.File, t attrtype.Type) file.Type {
	switch t {
	case attrtype.Bool:
		return "bool"
	case attrtype.Int:
		return "int"
	case attrtype.Float:
		return "float64"
	case attrtype.String:
		return "string"
	case attrtype.Identifier:
		return z.SafeImport(f).Qualifier.QualifiedType("Identifier")
	case attrtype.Datetime:
		return z.HTMLImport(f).Qualifier.QualifiedType("Datetime")
	case attrtype.URL:
		return z.SafeImport(f).Qualifier.QualifiedType("URL")
	case attrtype.ResourceURL:
		return z.SafeImport(f).Qualifier.QualifiedType("ResourceURL")
	case attrtype.Srcset:
		return z.SafeImport(f).Qualifier.QualifiedType("Srcset")
	case attrtype.Text:
		return "string"
	case attrtype.CSS:
		return z.SafeImport(f).Qualifier.QualifiedType("CSSDeclarations")
	case attrtype.JS:
		return z.SafeImport(f).Qualifier.QualifiedType("Script")
	}

	switch t := t.(type) {
	case attrtype.SpaceList:
		et := z.attrtypeToSafeType(f, t.Element)
		if et == "" {
			return ""
		}
		return "[]" + et
	case attrtype.CommaList:
		et := z.attrtypeToSafeType(f, t.Element)
		if et == "" {
			return ""
		}
		return "[]" + et
	}

	return ""
}

func (z *analyzer) AnalyzeComponentParameter_InferredType_fromDefault(
	logger *slog.Logger, c *file.Component, param *file.ComponentParameter,
) {
	logger = logger.WithGroup("infer_type_from_default").
		With(slog.String("param", string(param.Name)),
			slog.String("param_pos", param.AST.Name.Start().String()))

	if param.AST.Type != nil {
		param.InferredType.SetZero()
		return
	}

	if param.AST.Default == nil {
		param.InferredType.SetFailed()
		logger.Error("No default value set for untyped component parameter")
		z.Report(&diagnostic.Diagnostic{
			Message: "component parameter: neither type nor default value set",
			Primary: []diagnostic.Annotation{
				anno.Node(c.File, param.AST.Name, "unable to determine type of this parameter"),
			},
			Explanation: "Every parameter must either have an explicit type, or a default value " +
				"from which the type can be inferred.",
			Hints: []diagnostic.Hint{
				{
					Hint:    "Add an explicit type to this parameter.",
					Example: "`" + string(param.Name) + " MyType`",
				}, {
					Hint:    "Set a default value from which the type can be inferred.",
					Example: "`" + string(param.Name) + " = 123`",
				},
			},
		})
		return
	}

	t, _ := InferType(c.File, param.AST.Default)
	if t != "" {
		param.InferredType.SetResult(t)
	}

	param.InferredType.SetFailed()
	logger.Error("Unable to infer type from default value")
	z.Report(&diagnostic.Diagnostic{
		Message: "component parameter: unable to infer type from default value",
		Primary: []diagnostic.Annotation{
			anno.Node(c.File, param.AST.Default, "cannot infer type of this expression"),
			anno.Node(c.File, param.AST.Name, "has no explicit type"),
		},
		Explanation: "A component parameter's type is only optional, if the type can be " +
			"inferred from the default value. If that is not possible, specify the type as you" +
			"normally would behind the parameter's name.",
		Examples: []diagnostic.Example{
			{Example: "`" + string(param.Name) + " MyType = ...`"},
		},
	})
}
