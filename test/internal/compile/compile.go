// Package compile provides an utility function to compile files in preparation
// tests.
package compile

import (
	"log"
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/corgi"
	"github.com/mavolin/corgi/corgi/resource"
	"github.com/mavolin/corgi/test/internal/voidwriter"
	"github.com/mavolin/corgi/writer"
)

type Options struct {
	// Package sets the name of the package in which the generated function
	// will be placed.
	//
	// If not set, the name of the directory in which the calling file is
	// located, will be used.
	Package string
}

func init() {
	log.SetOutput(voidwriter.Writer)
}

// Compile compiles the file at the given path using the corgi compiler
// available under os.Args[1].
func Compile(t *testing.T, name string, o Options) {
	t.Helper()

	in, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("could not read file: %s", err)
		return
	}

	f, err := corgi.File(".", name, string(in)).
		WithResourceSource(resource.NewFSSource(".", os.DirFS("."))).
		Parse()
	if err != nil {
		t.Fatalf("parse: %s", err)
		return
	}

	w := writer.New(f, o.Package)

	out, err := os.Create(name + ".go")
	if err != nil {
		t.Fatalf("could not create output file: %s", err)
		return
	}

	if err = w.Write(out); err != nil {
		t.Fatalf("could not write to output file: %s", err)
		return
	}

	if err = out.Close(); err != nil {
		t.Errorf("could not close output file: %s", err)
	}
}

// packageRegexp is a regular expression that allows extraction of the package
// name from the return of runtime.FuncForPC(pc).Names().
var packageRegexp = regexp.MustCompile(`^(?:[^/]+/)*(?P<package>[^.]+)(?:\.[^.]+)?\.[^.]+$`)

func callingPackage(t *testing.T) string {
	t.Helper()

	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		require.Fail(t, "could not determine calling function")
	}

	callerName := runtime.FuncForPC(pc).Name()
	return packageRegexp.FindStringSubmatch(callerName)[1]
}
