package command

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

var (
	indent = strings.Repeat(" ", 2)
	spacer = strings.Repeat(" ", 3)
)

// Help prints the help message for the command to the given writer.
func (c *Cmd) Help(w io.Writer) {
	if len(c.Commands) > 0 {
		c.groupHelp(w)
	} else {
		c.commandHelp(w)
	}
}

func (c *Cmd) groupHelp(w io.Writer) {
	var maxCmdWidth int
	for _, c := range c.Commands {
		if len(c.Name) > maxCmdWidth {
			maxCmdWidth = len(c.Name)
		}
	}

	fmt.Fprintln(w, widthLim(0, c.LongDescription))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, indent, c.CommandName(), "<command> [arguments...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, c := range c.Commands {
		fmt.Fprint(w, indent)
		fmt.Fprint(w, c.Name)
		fmt.Fprint(w, spacer)
		fmt.Fprint(w, strings.Repeat(" ", maxCmdWidth-len(c.Name)))
		fmt.Fprintln(w, widthLim(len(indent)+maxCmdWidth+len(spacer), c.ShortDescription))
	}
}

func (c *Cmd) commandHelp(w io.Writer) {
	name := c.CommandName()

	var hasFlags bool
	c.flags.VisitAll(func(*flag.Flag) { hasFlags = true })

	var maxArgWidth int
	for _, usage := range c.ArgUsages {
		maxArgWidth = max(maxArgWidth, len(usage[0]))
	}

	var maxFlagWidth int
	c.flags.VisitAll(func(f *flag.Flag) {
		h, _ := flagHeader(f)
		if len(h) > maxFlagWidth {
			maxFlagWidth = len(h)
		}
	})

	usageIndent := len(indent) + len(name)
	if hasFlags {
		usageIndent += len(" [flags]")
	}
	if maxArgWidth > 0 { // if there is at least one argument example
		usageIndent += len(" ") + maxArgWidth
	}
	usageIndent += len(spacer)

	fmt.Fprintln(w, widthLim(0, widthLim(0, c.LongDescription)))
	fmt.Fprintln(w)
	if len(c.ArgUsages) == 1 {
		fmt.Fprintln(w, "Usage:")
	} else {
		fmt.Fprintln(w, "Usages:")
	}
	for _, usage := range c.ArgUsages {
		fmt.Fprint(w, indent)
		fmt.Fprint(w, name)
		if !hasFlags {
			fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, " [flags] ")
		}
		fmt.Fprint(w, usage[0])
		fmt.Fprint(w, strings.Repeat(" ", maxArgWidth-len(usage[0])))
		fmt.Fprint(w, spacer)
		fmt.Fprintln(w, widthLim(usageIndent, usage[1]))
	}
	if hasFlags {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Flags:")
		c.flags.VisitAll(func(f *flag.Flag) {
			header, usage := flagHeader(f)
			fmt.Fprint(w, indent)
			fmt.Fprint(w, header)
			fmt.Fprint(w, spacer)
			fmt.Fprint(w, strings.Repeat(" ", maxFlagWidth-len(header)))
			fmt.Fprintln(w, widthLim(len(indent)+maxFlagWidth+len(spacer), usage))
		})
	}
}

func flagHeader(f *flag.Flag) (header, usage string) {
	var b strings.Builder

	b.WriteByte('-')
	b.WriteString(f.Name)

	name, usage := flag.UnquoteUsage(f)
	if name != "" {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(name))
	}

	return b.String(), usage
}

func widthLim(nIndent int, s string) string {
	indent := strings.Repeat(" ", nIndent)
	width := max(20, Width-nIndent)

	var out strings.Builder

	var end int
	for i := 0; i < len(s); {
		if i >= width || s[i] == '\n' {
			switch {
			case s[i] == '\n':
				out.WriteString(s[:i])
				i += len("\n")
			case end == 0:
				out.WriteString(s[:i])
			default:
				out.WriteString(s[:end])
				i = end + len(" ")
			}
			out.WriteByte('\n')
			out.WriteString(indent)
			s = s[i:]
			i, end = 0, 0
			continue
		}
		if s[i] == ' ' {
			end = i
		}
		i++
	}
	out.WriteString(s)

	return out.String()
}
