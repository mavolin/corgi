package ast

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/k0kubun/pp"
	"github.com/mavolin/corgi/v2/cmd/command"
	"github.com/mavolin/corgi/v2/cmd/flags"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/load/link"
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

	Link bool
}

func (f *Flags) Bind(s *flag.FlagSet) {
	f.ParseFlags.Bind(s)

	s.BoolVar(&f.Link, "link", false, "link the AST")
}

func run(_ *command.Cmd, f *Flags, args []string) {
	var data []byte
	switch {
	case len(args) == 0:
		var err error
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read from stdin: %v\n", err)
			os.Exit(1)
		}
	case len(args) == 1:
		var err error
		data, err = os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read file %q: %v\n", args[0], err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "expected at most 1 argument")
		os.Exit(2)
	}

	fi, parseErrs := parse.Parse(string(data), parse.Options{})

	var linkErrs []*diagnostic.Diagnostic
	if f.Link {
		p := &file.Package{
			PathInModule: "stdin",
			Files:        []*file.File{fi},
		}
		fi.Package = p
		linkErrs = link.Link(context.Background(), p, link.Options{
			Logger: f.Logger,
		})
	}

	if fi != nil {
		pp.Println(fi)
	}

	_ = os.Stdout.Close() // so that errs appear at the bottom

	if len(parseErrs) > 0 {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "=== Parse Errors ===")
		fmt.Fprintln(os.Stderr, parseErrs.Pretty(diagnostic.PrettyOptions{
			Color: f.Color,
			Width: flags.Width,
		}))
	}
	if len(linkErrs) > 0 {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "=== Link Errors ===")
		fmt.Fprintln(os.Stderr, diagnostic.List(linkErrs).Pretty(diagnostic.PrettyOptions{
			Color: f.Color,
			Width: flags.Width,
		}))
	}
}
