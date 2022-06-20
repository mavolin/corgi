package writer

import (
	"strconv"

	"github.com/mavolin/corgi/corgi/file"
)

func (w *Writer) writePackage() error {
	return w.writeToFile("package " + w.packageName + "\n\n")
}

func (w *Writer) writeDoNotEdit() error {
	return w.writeToFile("// Code generated by github.com/mavolin/corgi/cmd/corgi. DO NOT EDIT.\n\n")
}

func (w *Writer) writeImports() error {
	if err := w.writeToFile("import (\n"); err != nil {
		return err
	}

	if err := w.writeToFile("_bytes \"bytes\"\n_io \"io\"\n"); err != nil {
		return err
	}

	err := w.writeToFile("_writeutil \"github.com/mavolin/corgi/pkg/writeutil\"\n")
	if err != nil {
		return err
	}

	for _, imp := range w.main.Imports {
		if err := w.writeToFile(strconv.Quote(imp.Path) + "\n"); err != nil {
			return err
		}
	}

	return w.writeToFile(")\n\n")
}

func (w *Writer) writeGlobalCode() error {
	if err := w.writeGlobalCodeFile(w.main, make(map[string]struct{})); err != nil {
		return err
	}

	return w.writeToFile("\n")
}

func (w *Writer) writeGlobalCodeFile(f *file.File, alreadyWritten map[string]struct{}) error {
	if f.Extend != nil {
		if err := w.writeGlobalCodeFile(&f.Extend.File, alreadyWritten); err != nil {
			return err
		}
	}

	for _, c := range f.GlobalCode {
		if err := w.writeToFile(c.Code + "\n"); err != nil {
			return err
		}
	}

	for _, use := range f.Uses {
		for _, uf := range use.Files {
			if _, ok := alreadyWritten[uf.Source+"/"+uf.Name]; ok {
				continue
			}

			for _, c := range uf.GlobalCode {
				if err := w.writeToFile(c.Code + "\n"); err != nil {
					return err
				}
			}

			if err := w.writeGlobalCodeScope(uf.Scope, alreadyWritten); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Writer) writeGlobalCodeScope(s file.Scope, alreadyWritten map[string]struct{}) error {
Items:
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Include:
			ci, ok := itm.Include.(file.CorgiInclude)
			if !ok {
				break
			}

			if _, ok := alreadyWritten[ci.File.Source+"/"+ci.File.Name]; ok {
				continue Items
			}

			for _, c := range ci.File.GlobalCode {
				if err := w.writeToFile(c.Code + "\n"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (w *Writer) writeFunc() error {
	if err := w.writeToFile("func " + string(w.main.Func.Name) + "(_w _io.Writer, "); err != nil {
		return err
	}

	// don't writeToFile the opening paren
	if err := w.writeToFile(w.main.Func.Params.Expression[1:]); err != nil {
		return err
	}

	return w.writeToFile(" (err error) ")
}
