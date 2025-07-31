package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckImportNamespaceCollisions(t *testing.T) {
	t.Parallel()

	t.Run("no collisions", func(t *testing.T) {
		t.Parallel()

		pkgA := createPackage("pkg/a")
		pkgB := createPackage("pkg/b")

		mainPkg := createPackage("main")
		mainFile := createFile(mainPkg, "main.corgi")

		createImport(mainFile, nil, "", pkgA.ImportPath)
		createImport(mainFile, nil, "", pkgB.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(pkgA, pkgB),
		})
		if !should.Equal(t, 0, len(ds)) {
			t.Log(ds.Short())
		}
	})

	t.Run("collision", func(t *testing.T) {
		const wantMessage = "import collision"

		tests := []struct {
			name                 string
			aliasA, packageNameA string
			aliasB, packageNameB string
		}{
			{
				name:   "same alias",
				aliasA: "same", packageNameA: "a",
				aliasB: "same", packageNameB: "b",
			}, {
				name:         "same package name",
				packageNameA: "same",
				packageNameB: "same",
			}, {
				name:   "alias and package name",
				aliasA: "alias", packageNameA: "a",
				packageNameB: "alias",
			},
		}

		for _, c := range tests {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				pkgA := createPackage(c.packageNameA)
				pkgB := createPackage(c.packageNameB)

				mainPkg := createPackage("main")
				mainFile := createFile(mainPkg, "main.corgi")

				createImport(mainFile, nil, c.aliasA, pkgA.ImportPath)
				createImport(mainFile, nil, c.aliasB, pkgB.ImportPath)

				ds := Link(context.Background(), mainPkg, Options{
					Importer: ImporterFor(pkgA, pkgB),
				})
				if should.Equal(t, 1, len(ds)) {
					if !should.Equal(t, wantMessage, ds[0].Message) {
						t.Log(ds[0].Short())
					}
				} else {
					t.Log(ds.Short())
				}
			})
		}
	})
}
