package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	args, err := resolveArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := run(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type args struct {
	GOROOT  string
	Package string

	Out string
}

func resolveArgs() (*args, error) {
	var a args

	a.GOROOT = os.Getenv("GOROOT")
	if a.GOROOT == "" {
		return nil, errors.New("GOROOT is not set")
	}

	a.Package = os.Getenv("GOPACKAGE")
	if a.Package == "" {
		return nil, errors.New("GOPACKAGE is not set")
	}

	if len(os.Args) != 2 {
		return nil, fmt.Errorf("expected exactly one argument, got %d", len(os.Args)-1)
	}

	a.Out = os.Args[1]
	return &a, nil
}

func run(args *args) error {
	packages, err := collectPackages(args.GOROOT)
	if err != nil {
		return err
	}

	f, err := os.Create(args.Out)
	if err != nil {
		return fmt.Errorf("opening output file: %w", err)
	}
	defer f.Close()

	writeFile(f, args.Package, packages)
	return nil
}

func collectPackages(goroot string) ([]string, error) {
	packages := make([]string, 0, 256)

	err := fs.WalkDir(os.DirFS(filepath.Join(goroot, "src")), ".", func(p string, d fs.DirEntry, err error) error {
		if p == "." {
			return nil
		}
		if err != nil {
			return err
		}

		name := strings.TrimPrefix(p, "./")
		base := path.Base(name)
		if d.IsDir() {
			if base == "internal" || base == "vendor" || base == "testdata" {
				return fs.SkipDir
			}
			packages = append(packages, name)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking GOROOT/src: %w", err)
	}
	return packages, nil
}

func writeFile(w io.Writer, pkg string, stdlib []string) {
	fmt.Fprintln(w, "package", pkg)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// isStdlib heuristically detects already known Go stdlib packages, so that")
	fmt.Fprintln(w, "// we don't unnecessarily preload them.")
	fmt.Fprintln(w, "func isStdlib(path string) bool {")
	fmt.Fprintln(w, "\tswitch path {")
	for _, pkg := range stdlib {
		fmt.Fprintf(w, "\tcase %q:\n", pkg)
	}
	fmt.Fprintln(w, "\tdefault:")
	fmt.Fprintln(w, "\t\treturn false")
	fmt.Fprintln(w, "\t}")
	fmt.Fprintln(w, "\treturn true")
	fmt.Fprintln(w, "}")
}
