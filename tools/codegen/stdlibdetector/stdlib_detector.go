package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	packageName = "load"
	outFile     = "stdlib_detector.go"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		return errors.New("GOROOT is not set")
	}

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
		return fmt.Errorf("walking GOROOT/src: %w", err)
	}

	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("opening output file: %w", err)
	}
	defer f.Close()

	fmt.Fprintln(f, "package", packageName)
	fmt.Fprintln(f)
	fmt.Fprintln(f, "// isStdlib heuristically detects already known Go stdlib packages, so that")
	fmt.Fprintln(f, "// we don't unnecessarily preload them.")
	fmt.Fprintln(f, "func isStdlib(path string) bool {")
	fmt.Fprintln(f, "\tswitch path {")
	for _, pkg := range packages {
		fmt.Fprintf(f, "\tcase %q:\n", pkg)
	}
	fmt.Fprintln(f, "\tdefault:")
	fmt.Fprintln(f, "\t\treturn false")
	fmt.Fprintln(f, "\t}")
	fmt.Fprintln(f, "\treturn true")
	fmt.Fprintln(f, "}")

	return nil
}
