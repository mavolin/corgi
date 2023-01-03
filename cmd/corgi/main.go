package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/mavolin/corgi/corgi"
	"github.com/mavolin/corgi/writer"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	ph := corgi.File(".", InFile, In)

	for _, rs := range ResourceSources {
		ph.WithResourceSource(rs)
	}

	f, err := ph.Parse()
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	w := writer.New(f, Package)

	if UseStdout {
		return writeStdout(w)
	}

	return writeFile(w)
}

func writeFile(w *writer.Writer) error {
	out, err := os.Create(OutFile)
	if err != nil {
		return fmt.Errorf("could not create output file: %w", err)
	}

	if err := w.Write(out); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	if RunGoImports {
		runGoImports()
	}

	return nil
}

func writeStdout(w *writer.Writer) error {
	if !RunGoImports { // fast path
		if err := w.Write(os.Stdout); err != nil {
			return err
		}

		return nil
	}

	// create a temporary file to write to, so we can run goimports on it
	out, err := os.CreateTemp("", "corgi")
	if err != nil {
		return fmt.Errorf("could not create temporary file, but need to run goimports: %w", err)
	}
	defer os.Remove(out.Name())

	if err := w.Write(out); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	if RunGoImports {
		runGoImports()
	}

	f, err := os.Open(out.Name())
	if err != nil {
		return fmt.Errorf("could not re-open temporary file: %w", err)
	}

	if _, err := io.Copy(os.Stdout, f); err != nil {
		return fmt.Errorf("could not copy temporary file to stdout: %w", err)
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func runGoImports() {
	goimports := exec.Command("goimports", "-w", OutFile) //nolint:gosec

	err := goimports.Run()
	if err == nil {
		return
	}

	if errors.Is(err, exec.ErrNotFound) {
		fmt.Fprint(os.Stderr, "goimports could not be found, but is needed to remove unused imports; "+
			"generated function might still work, if there are no unused imports\n"+
			"install goimports using `go get golang.org/x/tools/cmd/goimports@latest`\n"+
			"if you don't require formatting, use the -nofmt flag")
		return
	}

	fmt.Fprint(os.Stderr,
		"could not format output; this could mean that there is an erroneous Go expression in your template: ",
		err.Error())
}
