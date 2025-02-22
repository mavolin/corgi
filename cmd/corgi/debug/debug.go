package debug

import (
	"github.com/mavolin/corgi/v2/cmd/corgi/command"
	"github.com/mavolin/corgi/v2/cmd/corgi/debug/ast"
)

var (
	meta = command.Meta{
		Name:             "debug",
		ShortDescription: "Commands to help in the development of corgi.",
		LongDescription: `The debug commands are meant to help in the development of corgi itself.

They provide commands to inspect the AST, the tokens, and the scopes of corgi files.`,
	}
	Group = command.Group(meta, ast.Command)
)
