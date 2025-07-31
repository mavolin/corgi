package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckDuplicateDotImports(t *testing.T) {
	t.Parallel()

	t.Run("no duplicates", func(t *testing.T) {
		t.Parallel()

		pkgA := createPackage("pkg/a")
		pkgB := createPackage("pkg/b")

		mainPkg := createPackage("main")
		mainFile := createFile(mainPkg, "main.corgi")
		createImport(mainFile, nil, ".", pkgA.ImportPath)
		createImport(mainFile, nil, ".", pkgB.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(pkgA, pkgB),
		})
		if !should.Equal(t, 0, len(ds)) {
			t.Log(ds.Short())
		}
	})

	t.Run("duplicate dot imports", func(t *testing.T) {
		t.Parallel()

		pkgA := createPackage("pkg/a")

		mainPkg := createPackage("main")
		mainFile := createFile(mainPkg, "main.corgi")

		createImport(mainFile, nil, ".", pkgA.ImportPath)
		createImport(mainFile, nil, ".", pkgA.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(pkgA),
		})

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "duplicated dot imports", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})
}
