package link

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_LinkComponentCalls(t *testing.T) {
	t.Parallel()

	t.Run("success", testLinker_LinkComponentCalls_success)
	t.Run("failure", testLinker_LinkComponentCalls_failure)
}

func testLinker_LinkComponentCalls_success(t *testing.T) {
	t.Parallel()

	t.Run("local", func(t *testing.T) {
		t.Parallel()

		builtinPkg := createPackage("builtin")
		builtinF := createFile(builtinPkg, "builtin.corgi")
		createComponent(builtinF, nil, "test")

		p := createPackage("test")
		f := createFile(p, "test.corgi")
		comp := createComponent(f, nil, "test")
		call := createComponentCall(f, nil, "", comp.AST.Header.Name.Name)

		ds := Link(context.Background(), p, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.ImportPath,
		})

		if !should.Equal(t, 0, len(ds)) {
			t.Log(ds.Short())
		}
		if !should.True(t, comp == call.Component) {
			t.Log(cmp.Diff(comp, call.Component))
		}
	})

	t.Run("builtin", func(t *testing.T) {
		t.Parallel()

		builtinPkg := createPackage("builtin")
		builtinF := createFile(builtinPkg, "builtin.corgi")
		comp := createComponent(builtinF, nil, "test")

		p := createPackage("test")
		f := createFile(p, "test.corgi")
		call := createComponentCall(f, nil, "", comp.AST.Header.Name.Name)

		ds := Link(context.Background(), p, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.ImportPath,
		})

		if !should.Equal(t, 0, len(ds)) {
			t.Log(ds.Short())
		}
		if !should.True(t, comp == call.Component) {
			t.Log(cmp.Diff(comp, call.Component))
		}
	})

	importTests := []struct {
		name                      string
		packageNameDiffersFromDir bool
		alias                     string
	}{
		{
			name: "qualified/package name/matches directory",
		}, {
			name:                      "qualified/package name/doesn't match directory",
			packageNameDiffersFromDir: true,
		}, {
			name:  "qualified/alias",
			alias: "myalias",
		}, {
			name:  "dot import",
			alias: ".",
		},
	}
	for _, c := range importTests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			importedPkg := createPackage("imported")
			if c.packageNameDiffersFromDir {
				importedPkg.Name += "pkg"
			}
			importedF := createFile(importedPkg, "imported.corgi")
			importedComp := createComponent(importedF, nil, "Test")

			namespace := importedPkg.Name
			if c.alias == "." {
				namespace = ""
			} else if c.alias != "" {
				namespace = c.alias
			}

			mainPkg := createPackage("main")
			mainFile := createFile(mainPkg, "main.corgi")
			createImport(mainFile, nil, c.alias, importedPkg.ImportPath)
			call := createComponentCall(mainFile, nil, namespace, importedComp.AST.Header.Name.Name)

			ds := Link(context.Background(), mainPkg, Options{
				Importer: ImporterFor(importedPkg),
			})

			if !should.Equal(t, 0, len(ds)) {
				t.Log(ds.Short())
			}
			if !should.True(t, importedComp == call.Component) {
				t.Log(cmp.Diff(importedComp, call.Component))
			}
		})
	}
}

func testLinker_LinkComponentCalls_failure(t *testing.T) {
	t.Parallel()

	const builtinPath = "builtin"

	tests := []struct {
		name    string
		message string
		setup   func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error)
	}{
		{
			name:    "unresolved unqualified call",
			message: "component call: unresolved reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createComponentCall(f, nil, "", "Test")

				return p, nil, nil
			},
		}, {
			name:    "builtin not loaded",
			message: "failed to load builtin package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createComponentCall(f, nil, "", "test")

				return p, nil, map[importPath]error{
					builtinPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, nil, "", importedPkg.ImportPath)
				createComponentCall(f, nil, "imported", "Test")

				return p, []*file.Package{importedPkg},
					map[importPath]error{
						imp.Path: errors.New("stub error"),
					}
			},
		}, {
			name:    "dot import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, nil, ".", "imported")
				createComponentCall(f, nil, "", "Test")

				return p, nil, map[importPath]error{
					imp.Path: errors.New("stub error"),
				}
			},
		}, {
			name:    "qualified call to unexported component",
			message: "component call: cannot call unexported component",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				importedPkg := createPackage("imported")
				importedF := createFile(importedPkg, "imported.corgi")
				comp := createComponent(importedF, nil, "test")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createComponentCall(f, nil, importedPkg.Name, comp.AST.Header.Name.Name)

				return p, []*file.Package{importedPkg}, nil
			},
		}, {
			name:    "qualified call to unknown package",
			message: "component call: unresolved reference to package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createComponentCall(f, nil, "unknown", "Test")

				return p, nil, nil
			},
		}, {
			name:    "qualified call to unknown component",
			message: "component call: unresolved reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createImport(f, nil, "", importedPkg.ImportPath)
				createComponentCall(f, nil, importedPkg.Name, "Test")

				return p, []*file.Package{importedPkg}, nil
			},
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, packages, errs := c.setup()

			packagesMap := make(map[importPath]*file.Package, len(packages))
			for _, pkg := range packages {
				packagesMap[pkg.ImportPath] = pkg
			}

			importer := &mockImporter{packages: packagesMap, errors: errs}
			o := Options{Importer: importer.Import}
			if _, ok := packagesMap[builtinPath]; ok {
				o.BuiltinPath = builtinPath
			} else if _, ok := errs[builtinPath]; ok {
				o.BuiltinPath = builtinPath
			}
			ds := Link(context.Background(), p, o)

			if should.Equal(t, 1, len(ds)) {
				if !should.True(t, ds[0].Message == c.message) {
					t.Log(ds[0].Short())
				}
			} else {
				t.Log(ds.Short())
			}
		})
	}
}
