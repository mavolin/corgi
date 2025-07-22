package link

import (
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
)

type attrDefinitionInfo struct {
	name     string // without prefix
	fullName string // with prefix
	wildcard bool
}

func (i attrDefinitionInfo) selector() string {
	if i.wildcard {
		return i.name + "*"
	}
	return i.name
}

func (i attrDefinitionInfo) fullSelector() string {
	if i.wildcard {
		return i.fullName + "*"
	}
	return i.fullName
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
	info.name = strings.ToLower(sel.Name)
	if attr.Definition != nil && attr.Definition.Prefix != nil {
		info.fullName = strings.ToLower(attr.Definition.Prefix.Name) + info.name
	} else {
		info.fullName = info.name
	}
	info.wildcard = sel.Wildcard

	return &info
}
