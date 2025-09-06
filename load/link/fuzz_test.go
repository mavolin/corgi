package link

import (
	"bytes"
	"context"
	"log/slog"
	"path"
	"sync"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load/internal/fuzzdata"
	"github.com/mavolin/corgi/v2/load/parse"
)

var bufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// FuzzLink tests that the linker doesn't crash or hang on arbitrary input.
func FuzzLink(f *testing.F) {
	fuzzdata.AddBaseCorpus(f)

	f.Fuzz(func(t *testing.T, data string) {
		out := bufferPool.Get().(*bytes.Buffer)
		out.Reset()
		defer bufferPool.Put(out)
		logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))

		// We're just checking that Parse doesn't panic, so we ignore the return values
		fl, err := parse.Parse(data, parse.Options{})
		if err != nil {
			t.Log("=== Parser Errors ===")
			t.Log(err.Pretty(diagnostic.PrettyOptions{}))
			t.Log()
		}

		pkg := &file.Package{
			ImportPath: "linkfuzz",
			Name:       "linkfuzz",
			Files:      []*file.File{fl},
		}
		fl.Package = pkg

		defer t.Log(out.String())

		out.WriteString("=== Local Only Mode ===\n")
		d := Link(t.Context(), pkg, Options{Logger: logger})
		if len(d) > 0 {
			// invalid diagnostics
			should.NotPanic(t, func() { d.Pretty(diagnostic.PrettyOptions{}) })
		}

		out.WriteString("=== Failing Importer ===\n")
		d = Link(t.Context(), pkg, Options{
			Logger: logger,
			Importer: func(context.Context, importPath) (*file.Package, diagnostic.List, error) {
				return nil, nil, nil
			},
		})
		if len(d) > 0 {
			// invalid diagnostics
			should.NotPanic(t, func() { d.Pretty(diagnostic.PrettyOptions{}) })
		}

		out.WriteString("=== Successful Importer ===\n")
		d = Link(t.Context(), pkg, Options{
			Logger: logger,
			Importer: func(_ context.Context, imp importPath) (*file.Package, diagnostic.List, error) {
				pkg := &file.Package{
					ImportPath: imp,
					Name:       path.Base(imp),
				}
				return pkg, nil, nil
			},
		})
		if len(d) > 0 {
			// invalid diagnostics
			should.NotPanic(t, func() { d.Pretty(diagnostic.PrettyOptions{}) })
		}
	})
}
