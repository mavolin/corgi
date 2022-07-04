package app

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/resource"
)

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
