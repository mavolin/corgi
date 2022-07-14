package app

import (
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/mavolin/corgi/corgi"
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/internal/meta"
	"github.com/mavolin/corgi/writer"
)

var app = &cli.App{
	Name:  "corgi",
	Usage: "Generate Go functions from corgi files",
	Description: "This is the compiler for the corgi template language.\n\n" +
		"https://github.com/mavolin/corgi",
	Version:   meta.Version,
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
			Name:        "filename",
			Aliases:     []string{"f"},
			Usage:       "overwrite the name of the generated file",
			DefaultText: "corgi_file.corgi.go",
		},
		&cli.BoolFlag{
			Name:  "nofmt",
			Usage: "don't format the output",
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

func Run(args []string) error {
	return app.Run(args)
}

func run(ctx *cli.Context) error {
	//goland:noinspection GoBoolExpressions
	if meta.Version == meta.DevelopVersion {
		log.Println("you're running a development version of corgi, to get the stable release, " +
			"run `go get -u github.com/mavolin/corgi/cmd/corgi@latest`")
	}

	args, err := parseArgs(ctx)
	if err != nil {
		return err
	}

	if args.Get {
		goGetCorgi()
	}

	ph := corgi.File(".", args.File, args.FileContents)

	if args.FileType != file.TypeUnknown {
		ph.WithFileType(args.FileType)
	}

	for _, rs := range args.ResourceSources {
		ph.WithResourceSource(rs)
	}

	f, err := ph.Parse()
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	w := writer.New(f, args.Package)

	out, err := os.Create(args.OutputFile)
	if err != nil {
		return errors.Wrap(err, "could not create output file")
	}

	if err := w.Write(out); err != nil {
		return err
	}

	log.Println("generated", args.OutputFile)

	if !args.NoFmt {
		format(args)
	}

	return out.Close()
}

func goGetCorgi() {
	log.Println("generated functions import github.com/mavolin/corgi, I'm go getting it for you")

	goget := exec.Command("go", "get", "github.com/mavolin/corgi")
	goget.Stderr = os.Stderr

	if err := goget.Run(); err != nil {
		log.Println("couldn't go get corgi:", err.Error())
		log.Println("please do it yourself if you haven't already: go get github.com/mavolin/corgi")
	}
}

func format(args *args) {
	gofmt := exec.Command("gofmt", "-w", args.OutputFile) //nolint:gosec
	if err := gofmt.Run(); err == nil {
		log.Println("formatted output")
	} else {
		log.Println("could not format output "+
			"(this could mean that there is an erroneous Go expression in your template):",
			err.Error())
	}
}
