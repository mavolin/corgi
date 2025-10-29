package load

import (
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/internal/meta"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestModuleReader_LocalImportPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		symbolicImports []SymbolicImportMapping
		dir             filesystemPath
		want            file.CorgiImportPath
	}{
		{
			name: "simple",
			dir:  "",
			want: meta.Module + "/load",
		}, {
			name: "exact symbolic import",
			symbolicImports: []SymbolicImportMapping{
				{"woof/bark", "github.com/mavolin/corgi/v2/load"},
			},
			dir:  "",
			want: "woof/bark",
		}, {
			name: "nested symbolic import",
			symbolicImports: []SymbolicImportMapping{
				{"woof", "github.com/mavolin/corgi/v2"},
			},
			dir:  "",
			want: "woof/load",
		}, {
			name: "specific symbolic import",
			symbolicImports: []SymbolicImportMapping{
				{"bark", "github.com/mavolin/corgi/v2"},
				{"woof", "github.com/mavolin/corgi/v2/load"},
			},
			dir:  "",
			want: "woof",
		}, {
			name: "stdlib",
			dir:  "../std/html",
			want: "corgi/html",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			r, err := NewModuleReader(t.Context(), "", ModuleReaderOptions{
				SymbolicImports: c.symbolicImports,
			})
			if !should.NoError(t, err) {
				return
			}

			got, err := r.LocalImportPath(c.dir)
			if should.NoError(t, err) {
				should.Equal(t, got, c.want)
			}
		})
	}
}

func TestModuleReader_ReadImport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		symbolicImports []SymbolicImportMapping
		imp             file.CorgiImportPath
	}{
		{
			name: "simple",
			imp:  meta.Module + "/load/testdata/ReadImport",
		}, {
			name: "exact symbolic import",
			symbolicImports: []SymbolicImportMapping{
				{"woof/bark", meta.Module + "/load/testdata/ReadImport"},
			},
			imp: "woof/bark",
		}, {
			name: "nested symbolic import",
			symbolicImports: []SymbolicImportMapping{
				{"woof", meta.Module + "/load/testdata"},
			},
			imp: "woof/ReadImport",
		}, {
			name: "specific symbolic import",
			symbolicImports: []SymbolicImportMapping{
				{"bark", meta.Module + "/load"},
				{"woof", meta.Module + "/load/testdata/ReadImport"},
			},
			imp: "woof",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			r, err := NewModuleReader(t.Context(), "", ModuleReaderOptions{
				SymbolicImports: c.symbolicImports,
			})
			if !should.NoError(t, err) {
				return
			}

			p, err := r.ReadImport(t.Context(), c.imp)
			if !should.NoError(t, err) {
				return
			}

			should.Equal(t, p.Module, meta.Module)
			should.Equal(t, p.PathInModule, "load/testdata/ReadImport")
			if should.Equal(t, len(p.Files), 1) {
				should.Equal(t, p.Files[0].Name, "test.corgi")
				should.Equal(t, p.Files[0].Raw, "package readimport_test")
			} else {
				t.Logf("files: %+v", p.Files)
			}
		})
	}
}
