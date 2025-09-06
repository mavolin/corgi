package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckDuplicateDotImports(t *testing.T) {
	t.Parallel()

	t.Run("no duplicates", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkgA := createPackage("pkg/a")
		pkgB := createPackage("pkg/b")

		mainPkg := createPackage("main")
		mainFile := createFile(mainPkg, "main.corgi")
		createImport(mainFile, &start, ".", pkgA.ImportPath)
		createImport(mainFile, &start, ".", pkgB.ImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(pkgA, pkgB),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
	})

	t.Run("duplicate dot imports", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkgA := createPackage("pkg/a")

		mainPkg := createPackage("main")
		mainFile := createFile(mainPkg, "main.corgi")

		createImport(mainFile, &start, ".", pkgA.ImportPath)
		createImport(mainFile, &start, ".", pkgA.ImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(pkgA),
		})

		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "duplicated dot imports")
		}
	})
}
