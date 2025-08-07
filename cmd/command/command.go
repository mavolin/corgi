package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type (
	Cmd struct {
		Parent *Cmd
		Meta
		Commands []*Cmd
		flags    *flag.FlagSet
		run      func(args []string)
	}

	Meta struct {
		Name             string
		ArgUsages        []string
		ShortDescription string
		LongDescription  string
	}

	Flags interface {
		comparable
		Bind(s *flag.FlagSet)
	}
)

// NoFlags is a no-op implementation of the Flags interface.
type NoFlags struct{}

func (NoFlags) Bind(*flag.FlagSet) {}

// Group creates a new command group with the given metadata and subcommands.
func Group(meta Meta, cmds ...*Cmd) *Cmd {
	c := &Cmd{
		Meta:     meta,
		Commands: cmds,
	}
	for _, sub := range cmds {
		sub.Parent = c
	}
	return c
}

// Command creates a new command with the given metadata, flags, and run function.
//
// To indicate no flags, use the [NoFlags] type.
func Command[F Flags](meta Meta, flags F, run func(cmd *Cmd, flags F, args []string)) *Cmd {
	c := &Cmd{
		Meta:  meta,
		flags: flag.NewFlagSet(meta.Name, flag.ExitOnError),
	}

	c.flags.Usage = func() {
		c.HelpNotice(os.Stderr)
	}

	var flagZero F
	if flags != flagZero {
		flags.Bind(c.flags)
	}

	c.run = func(args []string) {
		_ = c.flags.Parse(args) // we're using ExitOnError
		run(c, flags, c.flags.Args())
	}
	return c
}

func (c *Cmd) Run(args []string) {
	if c.run != nil {
		c.run(args)
		return
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "no subcommand provided")
		c.HelpNotice(os.Stderr)
		return
	}

	for _, cmd := range c.Commands {
		if args[0] == cmd.Name {
			cmd.Run(args[1:])
			return
		}
	}

	fmt.Fprintln(os.Stderr, "unknown subcommand", args[0])
	c.HelpNotice(os.Stderr)
}

// Chain returns the chain of commands from the root command to this command.
func (c *Cmd) Chain() []*Cmd {
	cmds := make([]*Cmd, 0, 8)
	cmd := c
	for cmd != nil {
		cmds = append(cmds, cmd)
		cmd = cmd.Parent
	}

	slices.Reverse(cmds)
	return cmds
}

// CommandName returns the name of this, and it's parent commands, including
// the root command.
func (c *Cmd) CommandName() string {
	chain := c.Chain()

	var n int
	for _, cmd := range chain {
		n += len(cmd.Name) + len(" ")
	}

	var sb strings.Builder
	sb.Grow(n - len(" ")) // one space too many

	for i, cmd := range chain {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(cmd.Name)
	}
	return sb.String()
}

// SubcommandName returns the name of this, and it's parent commands, excluding
// the root command.
func (c *Cmd) SubcommandName() string {
	chain := c.Chain()[1:] // skip the root command

	var n int
	for _, cmd := range chain {
		n += len(cmd.Name) + len(" ")
	}

	var sb strings.Builder
	sb.Grow(n - len(" ")) // one space too many

	for i, cmd := range chain {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(cmd.Name)
	}
	return sb.String()
}

// HelpNotice prints a notice about how to get help for the command.
func (c *Cmd) HelpNotice(w io.Writer) {
	fmt.Fprintf(w, "Run '%s help %s' for help.\n", os.Args[0], c.SubcommandName())
}

var indent = strings.Repeat(" ", 3)

// Help prints the help message for the command to the given writer.
func (c *Cmd) Help(w io.Writer) {
	fmt.Fprintln(w, c.LongDescription)
	fmt.Fprintln(w)

	name := c.CommandName()

	if len(c.Commands) > 0 {
		fmt.Fprintln(w, "Help:")
		fmt.Fprintln(w, indent, name, " <command> [arguments...]")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Commands:")

		var maxCmdWidth int
		for _, c := range c.Commands {
			if len(c.Name) > maxCmdWidth {
				maxCmdWidth = len(c.Name)
			}
		}

		for _, c := range c.Commands {
			fmt.Fprintln(w, indent, c.Name, "  ", strings.Repeat(" ", maxCmdWidth-len(c.Name)), c.ShortDescription)
		}
		return
	}

	if len(c.ArgUsages) == 1 {
		fmt.Fprintln(w, "Help:")
	} else {
		fmt.Fprintln(w, "Usages:")
	}
	for _, u := range c.ArgUsages {
		fmt.Fprint(w, indent)
		if c.flags == nil {
			fmt.Fprintln(w, name, " ", u)
		} else {
			fmt.Fprintln(w, name, " [flags] ", u)
		}
	}

	var maxWidth int
	c.flags.VisitAll(func(f *flag.Flag) {
		h, _ := flagHeader(f)
		if len(h) > maxWidth {
			maxWidth = len(h)
		}
	})

	if maxWidth == 0 {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	c.flags.VisitAll(func(f *flag.Flag) {
		header, usage := flagHeader(f)
		fmt.Fprintln(w, indent, header, "  ", strings.Repeat(" ", maxWidth-len(header)), indented(len(indent)+maxWidth+len("  "), usage))
	})
}

func flagHeader(f *flag.Flag) (string, string) {
	var b strings.Builder

	b.WriteByte('-')
	b.WriteString(f.Name)

	name, usage := flag.UnquoteUsage(f)
	if len(name) > 0 {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(name))
	}

	return b.String(), usage
}

func indented(indent int, s string) string {
	idt := strings.Repeat(" ", indent)
	return strings.ReplaceAll(s, "\n", "\n"+idt)
}
