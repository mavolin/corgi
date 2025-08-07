package flags

import (
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"strconv"

	"github.com/fatih/color"
	"github.com/lmittmann/tint"
	"golang.org/x/term"
)

// Width is the width of the terminal, or a sensible default.
//
// Minimums and maximums are enforced, never being less than 40 or more than
// 120.
var Width = func() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 100
	}
	return max(min(width, 120), 40)
}()

type ParseFlags struct {
	Color   bool
	Verbose int
	Logger  *slog.Logger
}

func (f *ParseFlags) Bind(s *flag.FlagSet) {
	f.Logger = slog.New(slog.DiscardHandler)

	s.BoolVar(&f.Color, "color", !color.NoColor,
		"Whether to colorize error output.\n"+
			"Enabled by default, if stdout is a terminal, $NO_COLOR == \"\", and $TERM != \"dumb\".")
	s.BoolFunc("verbose",
		"Print verbose output explaining what is currently being done. Specify twice or set to 2, to enable debug logging.",
		func(s string) error {
			switch s {
			case "1":
				f.Verbose = 1
			case "2":
				f.Verbose = 2
			default:
				v, err := strconv.ParseBool(s)
				if err != nil {
					return err
				}
				if v {
					f.Verbose++
				} else {
					f.Verbose = 0
				}
			}

			if f.Verbose == 1 {
				f.Logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
					Level:   slog.LevelInfo,
					NoColor: !f.Color,
				}))
			} else if f.Verbose >= 2 {
				f.Logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
					Level:   slog.LevelDebug,
					NoColor: !f.Color,
				}))
			}
			return nil
		})
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
