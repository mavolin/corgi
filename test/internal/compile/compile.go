// Package compile provides an utility function to compile files in preparation
// tests.
package compile

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mavolin/corgi"
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/test/internal/voidwriter"
	"github.com/mavolin/corgi/write"
)

type Options struct {
	// Package sets the name of the package in which the generated function
	// will be placed.
	//
	// If not set, the name of the directory in which the calling file is
	// located, will be used.
	Package string

	AllowedFilters []string
}

func init() {
	log.SetOutput(voidwriter.Writer)
}

// Compile compiles the file at the given path using the corgi compiler
// available under os.Args[1].
func Compile(t *testing.T, name string, o Options) {
	t.Helper()

	if o.Package == "" {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get workdir: %s", err.Error())
		}

		o.Package = filepath.Base(wd)
	}

	f, err := corgi.LoadMain(name, corgi.LoadOptions{})
	if err != nil {
		if lerr := corgierr.As(err); lerr != nil {
			t.Fatalf(lerr.Pretty(corgierr.PrettyOptions{Colored: true}))
			return
		}

		t.Fatalf("parse: %s", err)
		return
	}

	w := write.New(write.Options{
		AllowedFilters: o.AllowedFilters,
	})

	file, err := os.Create(name + ".go")
	if err != nil {
		t.Fatalf("could not create output file: %s", err)
		return
	}

	goimports := exec.Command("goimports")
	goimports.Stdout = file
	var stderr bytes.Buffer
	goimports.Stderr = &stderr
	genOut, err := goimports.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create pipe:")
	}

	if err := goimports.Start(); err != nil {
		t.Fatalf("failed to start goimports: %s", err.Error())
	}

	if err = w.GenerateFile(genOut, o.Package, f); err != nil {
		t.Fatalf("could not write to output file: %s", err)
		return
	}

	genOut.Close()

	if err := goimports.Wait(); err != nil {
		if stderr.Len() > 0 {
			err = errors.New(stderr.String())
		}
		t.Fatalf("failed to wait for goimports: %s", err.Error())
	}

	if err = file.Close(); err != nil {
		t.Errorf("could not close output file: %s", err)
	}
}
