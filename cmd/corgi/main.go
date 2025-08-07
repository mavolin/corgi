package main

import (
	"os"

	"github.com/mavolin/corgi/v2/cmd/corgi/command"
	"github.com/mavolin/corgi/v2/cmd/corgi/compile"
	"github.com/mavolin/corgi/v2/cmd/corgi/debug"
	fmtcmd "github.com/mavolin/corgi/v2/cmd/corgi/fmt"
	"github.com/mavolin/corgi/v2/cmd/corgi/help"
	"github.com/mavolin/corgi/v2/cmd/corgi/version"
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
