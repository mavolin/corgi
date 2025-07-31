package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckExplicitBuiltinImport(t *testing.T) {
	t.Parallel()

	t.Run("pass", func(t *testing.T) {
		t.Parallel()

		builtinPkg := createPackage("builtin")
		createFile(builtinPkg, "builtin.corgi")

		mainPkg := createPackage("main")
		createFile(mainPkg, "main.corgi")

		ds := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.ImportPath,
		})

		if !should.Equal(t, 0, len(ds)) {
			t.Log(ds.Short())
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		builtinPkg := createPackage("builtin")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, "", builtinPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.ImportPath,
		})

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "explicit import of builtin package", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})
}
