package link

import (
	"log/slog"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

type attrDefinitionInfo struct {
	qualifiedName string // without prefix
	htmlName      string // with prefix
	wildcard      bool
}

func (i attrDefinitionInfo) qualifiedSelector() string {
	if i.wildcard {
		return i.qualifiedName + "*"
	}
	return i.qualifiedName
}

func (i attrDefinitionInfo) htmlNameSelector() string {
	if i.wildcard {
		return i.htmlName + "*"
	}
	return i.htmlName
}

func attrSpecInfo(attr *file.AttributeSpec) *attrDefinitionInfo {
	if attr.AST.Selector == nil {
		return nil
	}

	sel, _ := attr.AST.Selector.(*ast.BasicAttributeSelector)
	if sel == nil || sel.Name == "" {
		return nil
	}

	var info attrDefinitionInfo
	info.qualifiedName = strings.ToLower(sel.Name)
	if attr.Definition != nil && attr.Definition.Prefix != nil {
		info.htmlName = strings.ToLower(attr.Definition.Prefix.Name) + info.qualifiedName
	} else {
		info.htmlName = info.qualifiedName
	}
	info.wildcard = sel.Wildcard

	return &info
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
