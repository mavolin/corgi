package link

import (
	"sync"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/sourcegraph/conc"
)

type fileImport struct {
	file *file.File
	*file.Import
}

func collectImports(p *file.Package) map[string][]*fileImport {
	imports := make(map[string][]*fileImport)

	for _, f := range p.Files {
		if f.Scope == nil {
			continue
		}

		for _, itm := range f.Scope.Items {
			imp, _ := itm.(*ast.Import)
			if imp == nil {
				continue
			}

			if f.Imports == nil {
				// 95% of all files have a single import group containing all imports,
				// so this is a pretty solid guess
				f.Imports = make([]*file.Import, 0, len(imp.Imports)+1)
			}

			for _, impItm := range imp.Imports {
				spec, _ := impItm.(*ast.ImportSpec)
				if spec == nil {
					continue
				}

				fImp := &fileImport{file: f, Import: &file.Import{Source: spec}}
				f.Imports = append(f.Imports, fImp.Import)

				if spec.Path == nil {
					continue
				}

				path := fImp.Path()
				pImp := imports[path]
				if pImp == nil {
					pImp = make([]*fileImport, 1, len(p.Files))
					pImp[0] = fImp
					imports[path] = pImp
				} else {
					pImp = append(pImp, fImp)
					imports[path] = pImp
				}
			}
		}

		f.Imports = f.Imports[:len(f.Imports):len(f.Imports)]
	}

	return imports
}

func checkDuplicateImports(p *file.Package) fileerr.List {
	errs := make(fileerr.List, 0, 48)

	duplNamespace := make([]*file.Import, 0, 8)
	duplPath := make([]*file.Import, 0, 8)

	for _, f := range p.Files {
		if len(f.Imports) <= 1 {
			continue
		}

		ignore := make(map[*file.Import]struct{}, len(f.Imports))

		for ai, a := range f.Imports[:len(f.Imports)-1] {
			if a.Source.Path == nil {
				continue
			}
			if _, ok := ignore[a]; ok {
				continue
			}

			duplNamespace = duplNamespace[:0]
			duplPath = duplPath[:0]

			for _, b := range f.Imports[ai:] {
				if b.Source.Path == nil {
					continue
				}

				if a.Path() == b.Path() {
					duplPath = append(duplPath, a, b)
					ignore[b] = struct{}{}
				} else if a.Namespace() == b.Namespace() {
					duplNamespace = append(duplNamespace, a, b)
					ignore[b] = struct{}{}
				}
			}

			if len(duplPath) > 0 {
				start := a.Source.Path.Start
				if a.Source.Alias != nil {
					start = a.Source.Alias.Position
				}

				err := &fileerr.Error{
					Message: "package imported twice",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      start,
						ToEOL:      true,
						Annotation: "imported for the first time here",
					}),
					HintAnnotations: make([]fileerr.Annotation, len(duplPath)),
					Suggestions: []fileerr.Suggestion{
						{Suggestion: "only keep one import and remove the others"},
					},
				}
				for i, dupl := range duplPath {
					start = dupl.Source.Path.Start
					if dupl.Source.Alias != nil {
						start = dupl.Source.Alias.Position
					}
					err.HintAnnotations[i] = anno.Anno(f, anno.Annotation{
						Start:      start,
						ToEOL:      true,
						Annotation: "then again here",
					})
				}
				errs = append(errs, err)
			}
			if len(duplNamespace) > 0 {
				start := a.Source.Path.Start
				if a.Source.Alias != nil {
					start = a.Source.Alias.Position
				}

				err := &fileerr.Error{
					Message: "import collision",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      start,
						ToEOL:      true,
						Annotation: "namespace `" + a.Namespace() + "` used for the first time here",
					}),
					HintAnnotations: make([]fileerr.Annotation, len(duplNamespace)),
					Suggestions: []fileerr.Suggestion{
						{Suggestion: "use an import alias"},
					},
				}
				for i, dupl := range duplNamespace {
					start = dupl.Source.Path.Start
					if dupl.Source.Alias != nil {
						start = dupl.Source.Alias.Position
					}
					err.HintAnnotations[i] = anno.Anno(f, anno.Annotation{
						Start:      start,
						ToEOL:      true,
						Annotation: "then again here",
					})
				}
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func loadImports(p *file.Package, importer Importer, imports map[string] /*path*/ []*fileImport) fileerr.List {
	var wg conc.WaitGroup
	var mut sync.Mutex

	errs := make(fileerr.List, 0, len(imports))
	for _, fImps := range imports {
		ref := fImps[0]
		namespace := ref.Namespace()

		for _, cc := range p.ComponentCalls {
			// Only load the import if we actually need it.
			// Local component calls are linked before loading imports, so that
			// we only need to load a dot import, if it's actually used.
			if cc.Component != nil || cc.Namespace() != namespace || !cc.Exported() {
				continue
			}

			wg.Go(func() {
				pkg, err := importer(ref.Path())

				mut.Lock()
				defer mut.Unlock()

				if err != nil {
					if lerr := fileerr.As(err); lerr != nil {
						errs = append(errs, lerr...)
					} else {
						start := ref.Source.Path.Start
						if ref.Source.Alias != nil {
							start = ref.Source.Alias.Position
						}

						errs = append(errs, &fileerr.Error{
							Message: "failed to load import",
							ErrorAnnotation: anno.Anno(fImps[0].file, anno.Annotation{
								Start:      start,
								ToEOL:      true,
								Annotation: err.Error(),
							}),
							Cause: err,
						})
					}
					return
				}

				for _, fImp := range fImps {
					fImp.Package = pkg
				}
			})
		}
	}

	wg.Wait()

	if len(errs) == 0 {
		return nil
	}
	return errs
}
