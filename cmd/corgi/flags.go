package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mavolin/corgi/corgi/resource"
	"github.com/mavolin/corgi/internal/meta"
)

var (
	// Flags

	Package         string
	ResourceSources []resource.Source

	RunGoImports bool

	OutFile   string
	UseStdout bool

	// Args

	InFile string
	In     string
)

func init() {
	var (
		showHelp    bool
		showVersion bool

		ignoreCorgi bool
	)

	flag.Usage = usage

	flag.BoolVar(&showHelp, "h", false, "show this help message")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.StringVar(&Package, "package", "",
		"the name of the package to generate into; not required when using go generate")
	flag.Func("r", "add `Path` to the list of resource sources", func(path string) error {
		_, err := os.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("resource directory does not exist: %w", path)
			}

			return fmt.Errorf("%s: %w", path, err)
		}

		ResourceSources = append(ResourceSources, resource.NewFSSource(path, os.DirFS(path)))
		return nil
	})
	flag.BoolVar(&RunGoImports, "nofmt", true, "do not run goimports on the generated file")
	flag.BoolVar(&ignoreCorgi, "ignorecorgi", false, "do not use the corgi resource source")
	flag.StringVar(&OutFile, "o", "", "write output to `File` instead of stdout (defaults to `my_corgi_file.corgi.go`)")
	flag.BoolVar(&UseStdout, "stdout", false, "write output to stdout instead of a file")

	flag.Parse()

	if showHelp {
		flag.Usage()
		return
	} else if showVersion {
		fmt.Println("corgi", version())
		return
	}

	if !ignoreCorgi {
		filesys, err := corgiFS()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(2)
		}

		ResourceSources = append([]resource.Source{resource.NewFSSource("corgi", filesys)}, ResourceSources...)
	}

	if Package == "" {
		Package = os.Getenv("GOPACKAGE")
		if Package == "" {
			fmt.Println("you must either specify a package name or use go generate")
			os.Exit(2)
		}
	}

	InFile = flag.Arg(0)
	if InFile == "" {
		fmt.Println("you must specify exactly one input file")
		os.Exit(2)
	}

	inBytes, err := os.ReadFile(InFile)
	if err != nil {
		fmt.Println("could not read input file:", err.Error())
		os.Exit(2)
	}

	In = string(inBytes)

	if OutFile != "" && UseStdout {
		fmt.Println("conflicting flags -o and -stdout")
		os.Exit(2)
	}

	if OutFile == "" && !UseStdout {
		OutFile = filepath.Base(flag.Arg(0)) + ".go"
	}
}

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), "corgi ", version())
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "This is the compiler for the corgi template language.")
	fmt.Fprintln(flag.CommandLine.Output(), "https://github.com/mavolin/corgi")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Usage: corgi [options] file.corgi")
	fmt.Fprintln(flag.CommandLine.Output())
	flag.PrintDefaults()
}

func version() string {
	ver := meta.Version
	if meta.Commit != meta.UnknownCommit {
		ver += " (" + meta.Commit + ")"
	}

	return ver
}

func corgiFS() (fs.FS, error) {
	pwd, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	abs := pwd

	for abs != "/" {
		dir, err := os.ReadDir(abs)
		if err != nil {
			return nil, err
		}

		for _, e := range dir {
			if e.IsDir() {
				continue
			}

			if e.Name() == "go.mod" {
				for _, e := range dir {
					if e.IsDir() && e.Name() == "corgi" {
						return os.DirFS(filepath.Join(abs, "corgi")), nil
					}
				}

				return nil, nil
			}
		}

		abs = filepath.Dir(abs)
	}

	return nil, nil
}
