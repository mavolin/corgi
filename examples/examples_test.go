package examples

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load"
)

func TestExamples(t *testing.T) {
	t.Parallel()

	dir, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read examples directory: %v", err)
	}

	for _, entry := range dir {
		if !entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			var logs bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logs, nil))
			p, d, err := load.Directory(t.Context(), filepath.Join(".", entry.Name()), load.Options{
				Logger: logger,
			})
			if should.NoError(t, err) { // failed to load example
				if len(d) > 0 {
					t.Error(logs.String())
					t.Error(d.Pretty(diagnostic.PrettyOptions{}))
				}
				should.NotEqual(t, p, nil) // got nil package
				should.True(t, len(p.Files) == 1)
			}
		})
	}
}
