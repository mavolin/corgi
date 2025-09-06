package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
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

		d := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.ImportPath,
		})

		should.Equal(t, len(d), 0)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		builtinPkg := createPackage("builtin")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, "", builtinPkg.ImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.ImportPath,
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "explicit import of builtin package")
		}
	})
}
