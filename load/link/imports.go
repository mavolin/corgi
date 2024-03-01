package link

import (
	"sync"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/fileerr/anno"
	"github.com/sourcegraph/conc"
)

type fileImport struct {
	file *file.File
	*file.Import
}

func collectImports(_ *context, p *file.Package) map[string][]*fileImport {
	imports := make(map[string][]*fileImport)

	for _, f := range p.Files {
		if f.Scope == nil {
			continue
		}

		for _, itm := range f.Scope.Nodes {
			imp, _ := itm.(*ast.Import)
			if imp == nil {
				continue
			}

			if f.Imports == nil {
				// 95% of all files have a single import group containing all imports,
				// so this is a pretty solid guess
				f.Imports = make([]*file.Import, 0, len(imp.Imports)+5)
			}

			for _, impItm := range imp.Imports {
				spec, _ := impItm.(*ast.ImportSpec)
				if spec == nil {
					continue
				}

				fImp := &fileImport{file: f, Import: &file.Import{ImportSpec: spec}}
				f.Imports = append(f.Imports, fImp.Import)

				if spec.Path == nil {
					continue
				}

				path := fImp.ImportPath()
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

func checkDuplicateImports(ctx *context, p *file.Package) {
	duplNamespace := make([]*file.Import, 0, 8)
	duplPath := make([]*file.Import, 0, 8)

	for _, f := range p.Files {
		if len(f.Imports) <= 1 {
			continue
		}

		ignore := make(map[*file.Import]struct{}, len(f.Imports))

		for ai, a := range f.Imports[:len(f.Imports)-1] {
			if a.Path == nil {
				continue
			}
			if _, ok := ignore[a]; ok {
				continue
			}

			duplNamespace = duplNamespace[:0]
			duplPath = duplPath[:0]

			for _, b := range f.Imports[ai:] {
				if b.Path == nil {
					continue
				}

				if a.ImportPath() == b.ImportPath() {
					duplPath = append(duplPath, a, b)
					ignore[b] = struct{}{}
				} else if a.Namespace() == b.Namespace() {
					duplNamespace = append(duplNamespace, a, b)
					ignore[b] = struct{}{}
				}
			}

			if len(duplPath) > 0 {
				err := &fileerr.Error{
					Message:         "package imported twice",
					ErrorAnnotation: anno.Node(f, a.ImportSpec, "imported for the first time here"),
					HintAnnotations: make([]fileerr.Annotation, len(duplPath)),
					Suggestions: []fileerr.Suggestion{
						{Suggestion: "only keep one import and remove the others"},
					},
				}
				for i, dupl := range duplPath {
					err.HintAnnotations[i] = anno.Node(f, dupl.ImportSpec, "then again here")
				}
				ctx.err(err)
			}
			if len(duplNamespace) > 0 {
				err := &fileerr.Error{
					Message:         "import collision",
					ErrorAnnotation: anno.Node(f, a.ImportSpec, "namespace `"+a.Namespace()+"` used for the first time here"),
					HintAnnotations: make([]fileerr.Annotation, len(duplNamespace)),
					Suggestions: []fileerr.Suggestion{
						{Suggestion: "use an import alias"},
					},
				}
				for i, dupl := range duplNamespace {
					err.HintAnnotations[i] = anno.Node(f, dupl.ImportSpec, "then again here")
				}
				ctx.err(err)
			}
		}
	}
}

func loadImports(ctx *context, p *file.Package, importer Importer, imports map[string] /*path*/ []*fileImport) {
	var wg conc.WaitGroup
	var mut sync.Mutex

	for _, fImps := range imports {
		ref := fImps[0]
		namespace := ref.Namespace()

		for _, cc := range p.ComponentCalls {
			// Only load the import if we actually need it.
			// Local component calls are linked before loading imports, so that
			// we only need to load a dot import, if it's actually used.
			if cc.Component != nil ||
				(cc.Namespace != nil && cc.Namespace.Ident != namespace) ||
				file.IsExported(cc.Name.Ident) {
				continue
			}

			wg.Go(func() {
				pkg, err := importer(ref.ImportPath())

				mut.Lock()
				defer mut.Unlock()

				if err != nil {
					if lerr := fileerr.As(err); lerr != nil {
						ctx.errs = append(ctx.errs, lerr...)
					} else {
						ctx.err(&fileerr.Error{
							Message:         "failed to load import",
							ErrorAnnotation: anno.Node(ref.file, ref.ImportSpec, err.Error()),
							Cause:           err,
						})
					}
					return
				}

				for _, fImp := range fImps {
					fImp.Package = pkg
				}
			})
			break
		}
	}

	wg.Wait()
}
