package parse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mavolin/corgi/v2/file/diagnostic"
)

func Foo() {}

// FuzzParse tests that the parser doesn't crash or hang on arbitrary input.
func FuzzParse(f *testing.F) {
	f.Add("")
	f.Add("package")
	f.Add("package main")
	f.Add("package main\n\nimport \"std\"")
	f.Add("import foo \"fmt\"\n")
	f.Add("import \"example.com/mypackage\"\n")
	f.Add("comp Woof() {}")
	f.Add("comp Woof() {\n\tdiv [ Hello World! ]\n}")
	f.Add(":Bark(woof: 12)")
	f.Add(":Bark(woof: 12, #id, .class .another)")
	f.Add("span")
	f.Add("span(attr=12, bool, foo=\"string\")")
	f.Add("span(attr=12, bool, foo=\"string\") { }")
	f.Add("div [ Hello World! ]")
	f.Add("if")
	f.Add("if condition? {\n\tdiv\n}")
	f.Add("else")
	f.Add("if foo { } else { }")
	f.Add("for")
	f.Add("for i, v := range items {\n\tdiv\n}")
	f.Add("for i < 0 {\n\t&(attr)\n}")
	f.Add("for condition {\n\t:call()\n}")
	f.Add("state")
	f.Add("state foo = bar")
	f.Add("state foo myType = bar")
	f.Add("state foo, bar = baz, qux")
	f.Add("state (\n\tfoo = bar\n)")
	f.Add("var foo = bar")
	f.Add("var (\n\tmuffin = biscuit\n)")
	f.Add("const foo = bar")
	f.Add("const (\n\tpi = 3.14\n)")
	f.Add("foo := bar?.baz()?")
	f.Add("attr")
	f.Add("attr hx- foo { * bar }")
	f.Add("elem foo = bar")
	f.Add("elem foo js")
	f.Add("import \"std\"")
	f.Add(`"foo\n"`)
	f.Add("block(foo)")
	f.Add("- woof := bar")
	f.Add("#{woof}")
	f.Add("#:woof(arg: 12)")
	f.Add("!doctype(html)")
	f.Add("!raw")
	f.Add("!raw [ foo ]")
	f.Add("// comment")
	f.Add("/* comment */")
	f.Add("> div")
	f.Add("switch")
	f.Add("case")
	f.Add("default")
	f.Add("switch x {\ncase 1 {\n\tdiv\n}\ndefault {\n\tspan\n}\n}")
	f.Add("return")
	f.Add("fallthrough")
	f.Add("goto")
	f.Add("break")
	f.Add("continue")

	f.Add("comp A() {}\ncomp B() { :A() }")
	f.Add("import imp \"example.com/imp\"\ncomp C() { :imp.A() }")
	f.Add("import . \"example.com/imp\"\ncomp C() { :A() }")
	f.Add("import \"corgi/builtin\"\ncomp C() {}")
	f.Add("comp E() { div(#id .cls title=\"t\", bool) [ Hello ] }")
	f.Add("comp Base() { html { head { title { block title } } body { block body } } }\ncomp Page() : Base { with title [ T ] with body { span [ ok ] } }")
	f.Add("comp bigM(name string) { &(.font-size--big) }\ncomp Use() { p { :bigM(name: \"M\") } }")
	f.Add("elem Alert = div\ncomp Use() { Alert [ ok ] }")
	f.Add("attr hx- get { * url }\ncomp Use() { div(hx-get=\"/api\") }")
	f.Add("comp D() {}\ncomp D() {}\ncomp E() {}")
	f.Add("elem p = span\nelem P = div\ncomp Use() { p [ x ] }")
	f.Add("comp Loop(name string) { for i, _ := range []int{1,2,3} { div [ #{name} ] } }")
	f.Add("comp R() { !raw [ <div>ok</div> ] }")

	// Read all .corgi files in the project
	corgiFiles, err := recursivelyReadAll("../../", "*.corgi")
	if err != nil {
		f.Fatalf("failed to read .corgi files: %v", err)
	}

	for _, file := range corgiFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			f.Fatalf("failed to read file %s: %v", file, err)
		}
		f.Add(string(data))
	}

	f.Fuzz(func(t *testing.T, data string) {
		_, d := Parse(data, Options{})
		if len(d) > 0 {
			shouldValidDiagnostics(t, d)
		}
	})
}

func shouldValidDiagnostics(t *testing.T, ds diagnostic.List) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: diagnostic.List.Pretty: %v", r)
		}
	}()
	ds.Pretty(diagnostic.PrettyOptions{})
}

func recursivelyReadAll(dir string, pattern string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if !matched {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}
