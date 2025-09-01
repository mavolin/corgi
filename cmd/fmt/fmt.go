// Package fmt provides the fmt subcommand.
package fmt

import (
	"flag"

	"github.com/mavolin/corgi/v2/cmd/command"
	"github.com/mavolin/corgi/v2/cmd/command/flags"
)

var (
	meta = command.Meta{
		Name: "fmt",
		ArgUsages: [][2]string{
			{"", "read from stdin and write to stdout"},
			{"<file or dir...>", "overwrite listed files"},
			{"<dir>/...", "recursively format the dir and its subdirs"},
		},
		ShortDescription: "Format corgi files.",
		LongDescription: "The corgi equivalent of gofmt.\n" +
			"\n" +
			"If no files are specified, the input will be read from stdin and written to stdout.\n" +
			"\n" +
			"For each directory, the command formats all .corgi files in that directory. " +
			"Subdirectories remain untouched.\n" +
			"\n" +
			"Also accepts the special ./... argument, which recursively formats all .corgi " +
			"files in the present working directory and all of its subdirectories.",
	}

	Command = command.Command(meta, new(Flags), run)
)

type Flags struct {
	flags.ParseFlags

	stdout bool
	out    string
}

func (f *Flags) Bind(s *flag.FlagSet) {
	f.ParseFlags.Bind(s)

	s.StringVar(&f.out, "o", "",
		"The `path` of the output file to write to.\n"+
			"Can only be used if a single file or stdin is read.")
	s.BoolVar(&f.stdout, "stdout", false,
		"Output to stdout, even though an input file was passed.\n"+
			"Can only be used if a single file is read.")
}

func run(_ *command.Cmd, f *Flags, args []string) {
	return
}
