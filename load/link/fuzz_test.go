package link

import (
	"context"
	"path"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load/parse"
)

// FuzzLink tests that the linker doesn't crash or hang on arbitrary input.
func FuzzLink(f *testing.F) {
	// Seed with some simple cases.
	f.Add("div", "example.com/imp", "comp A() {}", "")
	f.Add("import foo \"example.com/imp\"\ncomp A() {\n\t:B()\n}", "example.com/imp", "comp A() {}", "")
	f.Add("comp A() {\n\t:B()\n}", "example.com/imp", "", "comp B() {}")

	f.Add("comp A() {}\ncomp B() { :A() }", "", "", "")
	f.Add("import imp \"example.com/imp\"\ncomp C() { :imp.A() }", "example.com/imp", "comp A() {}", "")
	f.Add("import . \"example.com/imp\"\ncomp C() { :A() }", "example.com/imp", "comp A() {}", "")
	f.Add("import \"fuzz/builtin\"\ncomp C() {}", "", "", "comp X() {}")
	f.Add("comp E() { div(#id .cls title=\"t\", bool) [ Hello ] }", "", "", "")
	f.Add("comp Base() { html { head { title { block title } } body { block body } } }\ncomp Page() : Base { with title [ T ] with body { span [ ok ] } }", "", "", "")
	f.Add("comp bigM(name string) { &(.font-size--big) }\ncomp Use() { p { :bigM(name: \"M\") } }", "", "", "")
	f.Add("elem Alert = div\ncomp Use() { Alert [ ok ] }", "", "", "")
	f.Add("attr hx- get { * url }\ncomp Use() { div(hx-get=\"/api\") }", "", "", "")
	f.Add("comp D() {}\ncomp D() {}\ncomp E() {}", "", "", "")
	f.Add("elem p = span\nelem P = div\ncomp Use() { p [ x ] }", "", "", "")
	f.Add("comp Loop(name string) { for i, _ := range []int{1,2,3} { div [ #{name} ] } }", "", "", "")
	f.Add("comp R() { !raw [ <div>ok</div> ] }", "", "", "")
	f.Add("import . \"example.com/imp\"\nimport . \"example.com/imp\"\ncomp X() {}", "example.com/imp", "comp A() {}", "")
	f.Add("import \"a/foo\"\nimport \"b/foo\"\ncomp X() {}", "a/foo", "comp A() {}", "")
	f.Add("import __corgi_bad \"example.com/imp\"\ncomp X() {}", "example.com/imp", "comp A() {}", "")

	f.Fuzz(func(t *testing.T, data, impPath, impContent, builtinContent string) {
		fl, d := parse.Parse(data, parse.Options{})
		if d != nil {
			should.NotPanic(t, func() { d.Pretty(diagnostic.PrettyOptions{}) })
		}

		pkg := &file.Package{
			CorgiImportPath: "linkfuzz/imports",
			Name:            "linkfuzz_imports",
			Files:           []*file.File{fl},
		}
		fl.Package = pkg

		// Importer that provides one fuzzed import and an optional builtin.
		const builtinPath = "fuzz/builtin"
		importer := func(ctx context.Context, imp importPath) (*file.Package, diagnostic.List, error) {
			var raw string
			switch imp {
			case builtinPath:
				raw = builtinContent
			case impPath:
				raw = impContent
			default:
				return &file.Package{CorgiImportPath: imp, Name: path.Base(imp)}, nil, nil
			}

			ff, d := parse.Parse(raw, parse.Options{})
			if d != nil {
				should.NotPanic(t, func() { d.Pretty(diagnostic.PrettyOptions{}) })
			}

			p := &file.Package{CorgiImportPath: imp, Name: path.Base(imp), Files: []*file.File{ff}}
			ff.Package = p
			d = Link(ctx, p, Options{})
			return p, d, nil
		}

		// Run without builtin
		d = Link(t.Context(), pkg, Options{Importer: importer})
		if len(d) > 0 {
			should.NotPanic(t, func() { d.Pretty(diagnostic.PrettyOptions{}) })
		}

		// Run with builtin, if provided
		if builtinContent != "" {
			d = Link(t.Context(), pkg, Options{Importer: importer, BuiltinPath: builtinPath})
			if len(d) > 0 {
				should.NotPanic(t, func() { d.Pretty(diagnostic.PrettyOptions{}) })
			}
		}
	})
}
