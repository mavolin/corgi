package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func qualifiableAttrSelector(spec *file.AttributeSpec) file.CanonicalQualifiableAttributeName {
	if spec.AST.Selector == nil {
		return ""
	}

	sel, _ := spec.AST.Selector.(*ast.BasicAttributeSelector)
	if sel == nil || sel.CanonicalName == "" {
		return ""
	}
	if sel.Wildcard {
		return file.CanonicalQualifiableAttributeName(sel.CanonicalName) + "*"
	}
	return file.CanonicalQualifiableAttributeName(sel.CanonicalName)
}

func htmlAttrName(spec *file.AttributeSpec) file.CanonicalAttributeName {
	if spec.AST.Selector == nil {
		return ""
	}

	sel, _ := spec.AST.Selector.(*ast.BasicAttributeSelector)
	if sel == nil || sel.CanonicalName == "" {
		return ""
	}
	if sel.Wildcard {
		return spec.Prefix + file.CanonicalAttributeName(sel.CanonicalName) + "*"
	}
	return spec.Prefix + file.CanonicalAttributeName(sel.CanonicalName)
}

func (l *linker) implicitImportCheck(logger *slog.Logger, f *file.File, imp *file.Import, node ast.Node, indefiniteArticle, name string) bool {
	if !imp.Implicit() {
		return true
	}

	logger.Error("Reference to implicitly loaded package")
	l.report(&diagnostic.Diagnostic{
		Message: "attempting to load " + name + " from internal package",
		Primary: []diagnostic.Annotation{
			anno.Node(f, node, "refusing to load from this package"),
		},
		Explanation: "The generated file imports certain helper packages that are not in the list of imports " +
			"of the respective corgi file. " +
			"Theses packages are not for use by humans, and you should depend on those imports " +
			"existing in future releases. " +
			"If you want to use " + indefiniteArticle + " " + name + " from this package, load the package explicitly in the " +
			"import section of your corgi file.",
	})
	return false
}
