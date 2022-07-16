// Package compile provides an utility function to compile files in preparation
// tests.
package compile

import (
	"log"
	"regexp"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/cmd/corgi/app"
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/test/internal/voidwriter"
)

type Options struct {
	// FileType overwrites the file type of the file to compile.
	FileType file.Type
	// OutName overwrites the name of the output file.
	OutName string
	// Format calls gofmt on the output file.
	Format bool

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

	args := []string{"-p", callingPackage(t, o)}

	if o.OutName != "" {
		args = append(args, "-f", o.OutName)
	}

	if !o.Format {
		args = append(args, "-nofmt")
	}

	switch o.FileType {
	case file.TypeHTML:
		args = append(args, "-t", "html")
	case file.TypeXHTML:
		args = append(args, "-t", "xhtml")
	case file.TypeXML:
		args = append(args, "-t", "xml")
	}

	args = append(args, name)

	err := app.Run(append([]string{"corgi"}, args...))

	require.NoErrorf(t, err, "failed to compile %s:\n%s", name, err)
}

// packageRegexp is a regular expression that allows extraction of the package
// name from the return of runtime.FuncForPC(pc).Name().
var packageRegexp = regexp.MustCompile(`^(?:[^/]+/)*(?P<package>[^.]+)(?:\.[^.]+)?\.[^.]+$`)

func callingPackage(t *testing.T, o Options) string {
	t.Helper()

	if o.Package != "" {
		return o.Package
	}

	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		require.Fail(t, "could not determine calling function")
	}

	callerName := runtime.FuncForPC(pc).Name()
	return packageRegexp.FindStringSubmatch(callerName)[1]
}
