package ast

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/k0kubun/pp"
	"github.com/mavolin/corgi/v2/cmd/corgi/command"
	"github.com/mavolin/corgi/v2/cmd/corgi/flags"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/load/parse"
)

var (
	meta = command.Meta{
		Name:             "ast",
		ShortDescription: "Inspect the AST of corgi files.",
		LongDescription:  `The ast command prints the AST of a corgi file.`,
	}
	Command = command.Command(meta, new(Flags), run)
)

type Flags struct {
	flags.ParseFlags
}

func (f *Flags) Bind(s *flag.FlagSet) {
	f.ParseFlags.Bind(s)
}

func run(_ *command.Cmd, f *Flags, args []string) {
	var data []byte
	if len(args) == 0 {
		var err error
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read from stdin: %v\n", err)
			os.Exit(1)
		}
	} else if len(args) == 1 {
		var err error
		data, err = os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read file %q: %v\n", args[0], err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintln(os.Stderr, "expected at most 1 argument")
		os.Exit(2)
	}

	fi, dl := parse.Parse(string(data), parse.Options{})
	if len(dl) > 0 {
		fmt.Fprintln(os.Stderr, dl.Pretty(diagnostic.PrettyOptions{
			Color: f.Color,
		}))
		fmt.Fprintln(os.Stderr)
	}

	pp.Println(fi)
}
