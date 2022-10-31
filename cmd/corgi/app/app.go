package app

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/mavolin/corgi/corgi"
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/internal/meta"
	"github.com/mavolin/corgi/writer"
)

func Run(args []string) error {
	ver := meta.Version
	if meta.Commit != meta.UnknownCommit {
		ver += " (" + meta.Commit + ")"
	}

	app := &cli.App{
		Name:  "corgi",
		Usage: "Generate Go functions from corgi files",
		Description: "This is the compiler for the corgi template language.\n\n" +
			"https://github.com/mavolin/corgi",
		Version:   ver,
		ArgsUsage: "<input file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "package",
				Aliases:  []string{"p"},
				Usage:    "set the name of the package to generate into; not required when using go generate",
				EnvVars:  []string{"GOPACKAGE"},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:      "resource",
				Aliases:   []string{"r"},
				Usage:     "add `DIR` to the list of resource directories",
				TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  "ignorecorgi",
				Usage: "don't use the $projectRoot/corgi resource directory",
			},
			&cli.StringFlag{
				Name:        "filetype",
				Aliases:     []string{"t"},
				Usage:       "overwrite the file type of the file (html, xhtml, xml)",
				DefaultText: "",
				Value:       "",
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "overwrite the name of the generated file",
				DefaultText: "corgi_file.corgi.go",
			},
			&cli.BoolFlag{
				Name:  "nofmt",
				Usage: "don't format the output and remove unused imports",
			},
			&cli.BoolFlag{
				Name:  "get",
				Usage: "go get github.com/mavolin/corgi before generating the function",
			},
		},
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		Action:               run,
	}

	return app.Run(args)
}

func run(ctx *cli.Context) error {
	a, err := parseArgs(ctx)
	if err != nil {
		return err
	}

	if a.Get {
		goGetCorgi()
	}

	ph := corgi.File(".", a.File, a.FileContents)

	if a.FileType != file.TypeUnknown {
		ph.WithFileType(a.FileType)
	}

	for _, rs := range a.ResourceSources {
		ph.WithResourceSource(rs)
	}

	f, err := ph.Parse()
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	w := writer.New(f, a.Package)

	out, err := os.Create(a.OutputFile)
	if err != nil {
		return errors.Wrap(err, "could not create output file")
	}

	if err = w.Write(out); err != nil {
		return err
	}

	if err = out.Close(); err != nil {
		return errors.Wrap(err, "could not close output file")
	}

	if !a.NoFmt {
		goImports(a)
	}

	return nil
}

func goGetCorgi() {
	goget := exec.Command("go", "get", "github.com/mavolin/corgi")
	goget.Stderr = os.Stderr

	if err := goget.Run(); err != nil {
		fmt.Println("couldn't go get corgi:", err.Error(),
			"; please do it yourself if you haven't already: `go get github.com/mavolin/corgi`")
	}
}

func goImports(args *args) {
	goimports := exec.Command("goimports", "-w", args.OutputFile) //nolint:gosec

	err := goimports.Run()
	if err == nil {
		return
	}

	if errors.Is(err, exec.ErrNotFound) {
		fmt.Println("goimports could not be found, but is needed to remove unused imports; " +
			"install using `go get golang.org/x/tools/cmd/goimports@latest`")
		return
	}

	fmt.Println(
		"could not format output; this could mean that there is an erroneous Go expression in your template):",
		err.Error())
}
