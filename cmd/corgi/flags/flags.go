package flags

import (
	"flag"
	"os/exec"

	"github.com/fatih/color"
)

type ParseFlags struct {
	Color   bool
	Verbose bool
}

func (f *ParseFlags) Bind(s *flag.FlagSet) {
	s.BoolVar(&f.Color, "color", !color.NoColor,
		"Whether to colorize error output.\n"+
			"Enabled by default, if stdout is a terminal, $NO_COLOR == \"\", and $TERM != \"dumb\".")
	s.BoolVar(&f.Verbose, "verbose", false, "Print verbose output explaining what is currently being done.")
}

type LoadFlags struct {
	ParseFlags
	GoExec string
}

func (f *LoadFlags) Bind(s *flag.FlagSet) {
	f.ParseFlags.Bind(s)

	defaultGoExecPath, err := exec.LookPath("go")
	if err != nil {
		defaultGoExecPath = ""
	}

	s.StringVar(&f.GoExec, "go", defaultGoExecPath, "`Path` to the Go executable.\n"+
		"This is used to download dependencies and get their locations in the Go module cache.\n"+
		"Not needed if there are no external dependencies.\n"+
		"Defaults to the Go executable in $PATH.")
}
