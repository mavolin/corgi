package help

import (
	"fmt"
	"os"

	"github.com/mavolin/corgi/v2/cmd/command"
)

var (
	meta = command.Meta{
		Name: "help",
		ArgUsages: []string{
			"           list all commands",
			"<command>  display help for a command",
		},
		ShortDescription: "Display help for a command or the entire cli.",
		LongDescription: `Displays a list of all commands.

If a command is passed, its help message will be displayed instead.`,
	}

	Command = command.Command[*command.NoFlags](meta, nil, run)
)

func run(cmd *command.Cmd, _ *command.NoFlags, args []string) {
	c := cmd.Parent
args:
	for _, arg := range args {
		for _, sub := range c.Commands {
			if sub.Name == arg {
				c = sub
				continue args
			}
		}

		fmt.Fprintf(os.Stderr, "unknown command %q\n", arg)
	}

	c.Help(os.Stdout)
}
