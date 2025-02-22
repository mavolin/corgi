package main

import (
	"os"

	"github.com/mavolin/corgi/v2/cmd/corgi/command"
	"github.com/mavolin/corgi/v2/cmd/corgi/compile"
	"github.com/mavolin/corgi/v2/cmd/corgi/debug"
	fmtcmd "github.com/mavolin/corgi/v2/cmd/corgi/fmt"
	"github.com/mavolin/corgi/v2/cmd/corgi/help"
)

//goland:noinspection GoImportUsedAsName
func main() {
	command.Group(command.Meta{
		Name: "corgi",
		LongDescription: "The CLI for the corgi html templating language.\n" +
			"\n" +
			"GitHub: github.com/mavolin/corgi",
	}, help.Command, fmtcmd.Command, compile.Command, debug.Group).
		Run(os.Args[1:])
}
