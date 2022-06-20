package link

import (
	"path/filepath"

	"github.com/mavolin/corgi/corgi/file"
)

// resolveImports resolves the resolveImports of the file.
func (l *Linker) resolveImports() error {
	if l.f.Extend != nil {
		for _, imp := range l.f.Extend.File.Imports {
			if err := l.addImport(imp); err != nil {
				return err
			}
		}
	}

	for _, use := range l.f.Uses {
		for _, uf := range use.Files {
			for _, imp := range uf.Imports {
				if err := l.addImport(imp); err != nil {
					return err
				}
			}
		}
	}

	return l.resolveIncludeImports(l.f.Scope)
}

func (l *Linker) resolveIncludeImports(s file.Scope) error {
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Include:
			ci, ok := itm.Include.(file.CorgiInclude)
			if !ok {
				break
			}

			for _, imp := range ci.File.Imports {
				if err := l.addImport(imp); err != nil {
					return err
				}
			}
		case file.Block:
			if err := l.resolveIncludeImports(itm.Body); err != nil {
				return err
			}
		case file.Element:
			if err := l.resolveIncludeImports(itm.Body); err != nil {
				return err
			}
		case file.If:
			if err := l.resolveIncludeImports(itm.Then); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := l.resolveIncludeImports(ei.Then); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := l.resolveIncludeImports(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.IfBlock:
			if err := l.resolveIncludeImports(itm.Then); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := l.resolveIncludeImports(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, c := range itm.Cases {
				if err := l.resolveIncludeImports(c.Then); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := l.resolveIncludeImports(itm.Default.Then); err != nil {
					return err
				}
			}
		case file.For:
			if err := l.resolveIncludeImports(itm.Body); err != nil {
				return err
			}
		case file.While:
			if err := l.resolveIncludeImports(itm.Body); err != nil {
				return err
			}
		case file.Mixin:
			if err := l.resolveIncludeImports(itm.Body); err != nil {
				return err
			}
		case file.MixinCall:
			if err := l.resolveIncludeImports(itm.Body); err != nil {
				return err
			}
		}
	}

	return nil
}

// addImport checks if this file's imports allow an import of the passed
// alias and path.
//
// If there is already an import for the same package under the same checkNamespaceCollisions,
// addImport does not add the import.
//
// If the alias is '.' or '_', addImport adds the import if the path is not
// already imported.
//
// If another package uses the same alias and that alias is not '.' or '_',
// addImport returns an *ImportNamespaceError.
//
// In all other cases it adds the import.
func (l *Linker) addImport(imp file.Import) (err error) {
	impNamespace := resolveNamespace(imp)

	if imp.Alias == "." || imp.Alias == "_" {
		for _, cmp := range l.f.Imports {
			if cmp.Path == imp.Path {
				return nil
			}
		}

		l.f.Imports = append(l.f.Imports, imp)

		return nil
	}

	for i, cmp := range l.f.Imports {
		cmpNamespace := resolveNamespace(cmp)
		if cmpNamespace == "_" {
			l.f.Imports[i].Alias = imp.Alias
			return nil
		}

		if impNamespace != cmpNamespace {
			continue
		}

		// both have no alias, or both have the same alias and both have the
		// same path -> don't add this import, it already exists
		if imp.Path == cmp.Path {
			return nil
		}

		// checkNamespaceCollisions collision :(
		return &ImportNamespaceError{
			Source:      l.f.Source,
			File:        l.f.Name,
			Line:        imp.Line,
			Col:         imp.Col,
			OtherSource: cmp.Source,
			OtherFile:   cmp.File,
			OtherLine:   cmp.Line,
			OtherCol:    cmp.Col,
			Namespace:   impNamespace,
		}
	}

	l.f.Imports = append(l.f.Imports, imp)

	return nil
}

func resolveNamespace(imp file.Import) string {
	namespace := string(imp.Alias)

	if namespace == "" {
		namespace = filepath.Base(imp.Path)
	}

	return namespace
}
