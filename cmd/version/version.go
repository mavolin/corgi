// Package version provides the version subcommand.
package version

import (
	"fmt"
	"os"

	"github.com/mavolin/corgi/v2/cmd/command"
	buildmeta "github.com/mavolin/corgi/v2/internal/meta"
)

var (
	meta = command.Meta{
		Name:             "version",
		ShortDescription: "Display the version of the CLI.",
		LongDescription:  "Display the version of the CLI.",
		ArgUsages: []string{
			"display the version of the CLI",
		},
	}

	Command = command.Command[*command.NoFlags](meta, nil, run)
)

func run(_ *command.Cmd, _ *command.NoFlags, args []string) {
	if len(args) != 0 {
		fmt.Fprintln(os.Stderr, "expected no arguments")
		os.Exit(2)
	}

	fmt.Printf("corgi version %s\n", buildmeta.Version)
}
