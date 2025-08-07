package main

import (
	"os"

	"github.com/mavolin/corgi/v2/cmd/command"
	"github.com/mavolin/corgi/v2/cmd/compile"
	"github.com/mavolin/corgi/v2/cmd/debug"
	fmtcmd "github.com/mavolin/corgi/v2/cmd/fmt"
	"github.com/mavolin/corgi/v2/cmd/help"
	"github.com/mavolin/corgi/v2/cmd/version"
)

var (
	meta = command.Meta{
		Name: "corgi",
		LongDescription: "The CLI for the corgi html templating language.\n" +
			"\n" +
			"GitHub: github.com/mavolin/corgi",
	}

	Group = command.Group(meta,
		help.Command, compile.Command, debug.Group, fmtcmd.Command, version.Command)
)

func main() {
	Group.Run(os.Args[1:])
}
