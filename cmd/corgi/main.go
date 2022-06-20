package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/mavolin/corgi/corgi"
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/resource"
	"github.com/mavolin/corgi/internal/meta"
	"github.com/mavolin/corgi/writer"
)

func main() {
	app := &cli.App{
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
				DefaultText: "html",
				Value:       "",
			},
			&cli.StringFlag{
				Name:        "filename",
				Aliases:     []string{"f"},
				Usage:       "overwrite the filename of the generated file",
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

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx *cli.Context) error {
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

// ============================================================================
// Args
// ======================================================================================

type args struct {
	FileType        file.Type
	Package         string
	ResourceSources []resource.Source

	Get   bool
	NoFmt bool

	File         string
	FileContents string

	OutputFile string
}

func parseArgs(ctx *cli.Context) (*args, error) {
	a := args{Package: ctx.String("package")}

	switch ctx.String("filetype") {
	case "":
		a.FileType = file.TypeUnknown
	case "html":
		a.FileType = file.TypeHTML
	case "xhtml":
		a.FileType = file.TypeXHTML
	case "xml":
		a.FileType = file.TypeXML
	default:
		return nil, fmt.Errorf("invalid file type: %s", ctx.String("filetype"))
	}

	a.ResourceSources = append(a.ResourceSources, resource.NewFSSource(".", os.DirFS(".")))

	if !ctx.Bool("ignorecorgi") {
		filesys, name, err := corgiFS()
		if err != nil {
			return nil, err
		}

		a.ResourceSources = append(a.ResourceSources, resource.NewFSSource(name, filesys))
	}

	for _, path := range ctx.StringSlice("resource") {
		_, err := os.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("resource directory does not exist: %s", path)
			}

			return nil, errors.Wrap(err, path)
		}

		a.ResourceSources = append(a.ResourceSources, resource.NewFSSource(path, os.DirFS(path)))
	}

	a.Get = ctx.Bool("get")
	a.NoFmt = ctx.Bool("nofmt")

	a.File = ctx.Args().Get(0)

	f, err := os.Open(a.File)
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file")
	}

	a.FileContents = string(data)

	a.OutputFile = ctx.String("filename")
	if a.OutputFile == "" {
		a.OutputFile = filepath.Dir(a.File) + "/" + a.File + ".go"
	}

	return &a, nil
}

func corgiFS() (fs.FS, string, error) {
	pwd, err := filepath.Abs(".")
	if err != nil {
		return nil, "", err
	}

	abs := pwd

	for abs != "/" {
		dir, err := os.ReadDir(abs)
		if err != nil {
			return nil, "", err
		}

		for _, e := range dir {
			if e.IsDir() {
				continue
			}

			if e.Name() == "go.mod" {
				for _, e := range dir {
					if e.IsDir() && e.Name() == "corgi" {
						rel, err := filepath.Rel(pwd, abs)
						if err != nil {
							return nil, "", err
						}

						return os.DirFS(filepath.Join(abs, "corgi")), rel, nil
					}
				}

			}
		}

		abs = filepath.Dir(abs)
	}

	return nil, "", fmt.Errorf("no go.mod found")
}
