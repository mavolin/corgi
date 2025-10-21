package link

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
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

		var start ast.Position
		builtinPkg := createPackage("builtin")
		builtinF := createFile(builtinPkg, "builtin.corgi")
		createComponent(builtinF, &start, "test")

		p := createPackage("test")
		f := createFile(p, "test.corgi")
		comp := createComponent(f, &start, "test")
		call := createComponentCall(f, &start, "", comp.Name)

		d := Link(context.Background(), p, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.CorgiImportPath,
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
		if !should.True(t, comp == call.Component) {
			t.Log(cmp.Diff(comp, call.Component))
		}
	})

	t.Run("builtin", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		builtinPkg := createPackage("builtin")
		builtinF := createFile(builtinPkg, "builtin.corgi")
		comp := createComponent(builtinF, &start, "test")

		p := createPackage("test")
		f := createFile(p, "test.corgi")
		call := createComponentCall(f, &start, "", comp.Name)

		d := Link(context.Background(), p, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.CorgiImportPath,
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
		if !should.True(t, comp == call.Component) {
			t.Log(cmp.Diff(comp, call.Component))
		}
	})

	importTests := []struct {
		name                      string
		packageNameDiffersFromDir bool
		alias                     file.Qualifier
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

			var start ast.Position
			importedPkg := createPackage("imported")
			if c.packageNameDiffersFromDir {
				importedPkg.Name += "pkg"
			}
			importedF := createFile(importedPkg, "imported.corgi")
			importedComp := createComponent(importedF, &start, "Test")

			qualifier := importedPkg.Name
			if c.alias == "." {
				qualifier = ""
			} else if c.alias != "" {
				qualifier = c.alias
			}

			mainPkg := createPackage("main")
			mainFile := createFile(mainPkg, "main.corgi")
			createImport(mainFile, &start, c.alias, importedPkg.CorgiImportPath)
			call := createComponentCall(mainFile, &start, qualifier, importedComp.Name)

			d := Link(context.Background(), mainPkg, Options{
				Importer: ImporterFor(importedPkg),
			})

			t.Log(d.Pretty(diagnostic.PrettyOptions{}))
			should.Equal(t, len(d), 0)
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
		setup   func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error)
	}{
		{
			name:    "unresolved unqualified call",
			message: "component call: unresolved reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createComponentCall(f, &start, "", "Test")

				return p, nil, nil
			},
		}, {
			name:    "builtin not loaded",
			message: "failed to load builtin package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createComponentCall(f, &start, "", "test")

				return p, nil, map[file.CorgiImportPath]error{
					builtinPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				importedPkg := createPackage("imported")
				importedPkg.PackageSymbols = nil
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, &start, "", importedPkg.CorgiImportPath)
				createComponentCall(f, &start, "imported", "Test")

				return p, []*file.Package{importedPkg},
					map[file.CorgiImportPath]error{
						imp.CorgiPath: errors.New("stub error"),
					}
			},
		}, {
			name:    "dot import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, &start, ".", "imported")
				createComponentCall(f, &start, "", "Test")

				return p, nil, map[file.CorgiImportPath]error{
					imp.CorgiPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "qualified call to unexported component",
			message: "component call: cannot call unexported component",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				importedPkg := createPackage("imported")
				importedF := createFile(importedPkg, "imported.corgi")
				comp := createComponent(importedF, &start, "test")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createComponentCall(f, &start, importedPkg.Name, comp.Name)

				return p, []*file.Package{importedPkg}, nil
			},
		}, {
			name:    "qualified call to unknown package",
			message: "component call: unresolved reference to package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createComponentCall(f, &start, "unknown", "Test")

				return p, nil, nil
			},
		}, {
			name:    "qualified call to unknown component",
			message: "component call: unresolved reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createImport(f, &start, "", importedPkg.CorgiImportPath)
				createComponentCall(f, &start, importedPkg.Name, "Test")

				return p, []*file.Package{importedPkg}, nil
			},
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, packages, errs := c.setup()

			packagesMap := make(map[file.CorgiImportPath]*file.Package, len(packages))
			for _, pkg := range packages {
				packagesMap[pkg.CorgiImportPath] = pkg
			}

			importer := &mockImporter{packages: packagesMap, errors: errs}
			o := Options{Importer: importer.Import}
			if _, ok := packagesMap[builtinPath]; ok {
				o.BuiltinPath = builtinPath
			} else if _, ok := errs[builtinPath]; ok {
				o.BuiltinPath = builtinPath
			}
			d := Link(context.Background(), p, o)

			t.Log(d.Pretty(diagnostic.PrettyOptions{}))
			if should.Equal(t, len(d), 1) {
				should.True(t, d[0].Message == c.message)
			}
		})
	}
}

func TestLinker_linkBlockSetterBlocks(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		t.Run("named block", func(t *testing.T) {
			t.Parallel()

			var start ast.Position
			p := createPackage("test")
			f := createFile(p, "test.corgi")
			comp := createComponent(f, &start, "Test")
			block := createBlock(comp, "content")

			call := createComponentCall(f, &start, "", comp.Name)
			blockSetter := createBlockSetter(call, block.Name)
			createWith(blockSetter, &start)

			d := Link(context.Background(), p, Options{})
			t.Log(d.Pretty(diagnostic.PrettyOptions{}))

			should.Equal(t, len(d), 0)

			if !should.True(t, comp == call.Component) {
				t.Log(cmp.Diff(comp, call.Component))
			}

			if !should.True(t, block == blockSetter.Block) {
				t.Log(cmp.Diff(block, blockSetter.Block))
			}
		})

		t.Run("default block", func(t *testing.T) {
			t.Parallel()

			var start ast.Position
			p := createPackage("test")
			f := createFile(p, "test.corgi")

			comp := createComponent(f, &start, "Test")
			block := createBlock(comp, "")

			call := createComponentCall(f, &start, "", comp.Name)
			blockSetter := createBlockSetter(call, block.Name)

			d := Link(context.Background(), p, Options{})

			t.Log(d.Pretty(diagnostic.PrettyOptions{}))
			should.Equal(t, len(d), 0)

			if !should.True(t, comp == call.Component) {
				t.Log(cmp.Diff(comp, call.Component))
			}

			if !should.True(t, blockSetter.Block == block) {
				t.Log(cmp.Diff(block, blockSetter.Block))
			}
		})
	})

	t.Run("unknown block", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		p := createPackage("test")
		f := createFile(p, "test.corgi")

		comp := createComponent(f, &start, "Test")
		createBlock(comp, "sidebar")

		call := createComponentCall(f, &start, "", comp.Name)
		blockSetter := createBlockSetter(call, "nonexistent")
		createWith(blockSetter, &start)

		d := Link(context.Background(), p, Options{})

		should.True(t, comp == call.Component)

		should.True(t, blockSetter.Block == nil)
		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.True(t, d[0].Message == "component call: block setter references unknown block")
		}
	})
}

func Test_linkComponentArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params []file.Identifier
		args   []file.Identifier
		error  string
	}{
		{
			name:   "all arguments match",
			params: []file.Identifier{"foo", "bar"},
			args:   []file.Identifier{"foo", "bar"},
		}, {
			name:   "argument missing parameter",
			params: []file.Identifier{"foo"},
			args:   []file.Identifier{"foo", "baz"},
			error:  "component call: argument: unresolved reference",
		}, {
			name:   "no arguments",
			params: []file.Identifier{"foo"},
			args:   []file.Identifier{},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			var start ast.Position
			p := createPackage("test")
			f := createFile(p, "test.corgi")
			comp := createComponent(f, &start, "Test")
			for _, param := range c.params {
				createParameter(comp, &start, param)
			}
			cc := createComponentCall(f, &start, "", comp.Name)
			for _, arg := range c.args {
				createArgument(cc, &start, arg)
			}

			ds := Link(context.Background(), p, Options{})
			t.Log(ds.Pretty(diagnostic.PrettyOptions{}))

			should.Equal(t, len(cc.ComponentArguments), len(c.args)) // argument count mismatch
			for _, arg := range cc.ComponentArguments {
				if slices.Contains(c.params, arg.Name) {
					if should.NotEqual(t, arg.Parameter, nil) { // argument should be linked
						should.Equal(t, arg.Parameter.Name, arg.Name) // linked to wrong parameter
					}
				} else {
					should.Equal(t, arg.Parameter, nil) // argument should not be linked
				}
			}

			if len(ds) == 0 {
				should.Equal(t, len(c.error), 0) // no error expected
			} else {
				should.Equal(t, len(ds), 1) // only one error expected
				if len(c.error) > 0 {
					should.Equal(t, ds[0].Message, c.error)
				}
			}
		})
	}
}
