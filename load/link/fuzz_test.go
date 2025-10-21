package link

import (
	"context"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
	"github.com/mavolin/corgi/v2/std/iters"
)

func FuzzLink(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed int64) {
		ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
		t.Cleanup(cancel)

		r := rand.New(rand.NewPCG(uint64(seed), 0)) //nolint:gosec

		g := GeneratePackageTree(r)

		var opts Options
		if g.Builtin != nil {
			opts.BuiltinPath = g.Builtin.CorgiImportPath
		}
		// opts.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

		needsLinking := make(map[*file.Package]*sync.Once, len(g.pkgs))
		needsLinking[g.Root] = new(sync.Once)
		for _, p := range g.pkgs {
			if chance(r, 0.8) {
				needsLinking[p] = new(sync.Once)
			}
		}

		seen := make(map[file.CorgiImportPath]bool, len(g.pkgs))
		var seenMu sync.Mutex
		opts.Importer = func(ctx context.Context, imp file.CorgiImportPath) (*file.Package, diagnostic.List, error) {
			seenMu.Lock()
			if seen[imp] {
				seenMu.Unlock()
				panic("importer called multiple times for the same path: " + imp)
			}
			seen[imp] = true
			seenMu.Unlock()

			p := g.pkgs[imp]
			if p == nil {
				return nil, nil, nil
			}

			if once := needsLinking[p]; once != nil {
				if ctx.Err() != nil {
					return nil, nil, ctx.Err()
				}

				once.Do(func() {
					d := Link(ctx, p, opts)
					if len(d) > 0 {
						shouldValidDiagnostics(t, d)
					}
				})
			}

			return p, nil, nil
		}

		d := Link(ctx, g.Root, opts)
		if len(d) > 0 {
			shouldValidDiagnostics(t, d)
		}

		d = Link(ctx, g.Root, opts)
		if len(d) > 0 {
			shouldValidDiagnostics(t, d)
		}

		should.NoError(t, ctx.Err()) // timeout
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

// generator

type PackageTree struct {
	r       *rand.Rand
	pkgs    map[file.CorgiImportPath]*file.Package
	Root    *file.Package
	Builtin *file.Package
}

func GeneratePackageTree(r *rand.Rand) *PackageTree {
	g := &PackageTree{
		r:    r,
		pkgs: make(map[file.CorgiImportPath]*file.Package, 64),
	}

	g.Root = createPackage("fuzz/root")
	g.Root.Name = "root"
	g.pkgs[g.Root.CorgiImportPath] = g.Root

	// optional builtin, never imports
	if g.r.IntN(4) == 0 {
		g.genBuiltin()
	}

	g.genPackageTree(g.Root, 0)
	return g
}

// genBuiltin creates a random number of files in p and fills them with
// nodes only.
func (t *PackageTree) genBuiltin() {
	t.Builtin = createPackage("fuzz/builtin")
	t.Builtin.Name = "builtin"
	t.pkgs[t.Builtin.CorgiImportPath] = t.Builtin

	nFiles := t.r.IntN(3) + 1
	for i := range nFiles {
		f := createFile(t.Builtin, file.Name("builtin"+itoa(i)+".corgi"))
		t.genNodesInFile(new(ast.Position), f, nil)
	}
}

func (t *PackageTree) genPackageTree(p *file.Package, depth int) {
	nFiles := t.r.IntN(4)
	if depth == 0 && nFiles == 0 {
		nFiles = 1
	}
	files := make([]*file.File, 0, nFiles)
	for i := range nFiles {
		files = append(files, createFile(p, file.Name("f"+itoa(i)+".corgi")))
	}

	// add imports and nodes for each file independently
	for _, f := range files {
		namespaces := make([]file.Qualifier, 0, 8)
		var pos ast.Position
		t.addImports(&pos, f, depth, &namespaces)
		t.genNodesInFile(&pos, f, namespaces)
	}
	// recurse based on imports from all files
	for _, f := range p.Files {
		for _, imp := range f.Imports {
			child := t.pkgs[imp.CorgiPath]
			if child == nil || child == t.Builtin || child == p {
				continue
			}
			if child.Files != nil {
				continue
			}
			t.genPackageTree(child, depth+1)
		}
	}
}

func (t *PackageTree) addImports(pos *ast.Position, f *file.File, depth int, namespaces *[]file.Qualifier) {
	attempts := 1 + t.r.IntN(5)
	for i := range attempts {
		// the deeper we go, the less imports
		if chance(t.r, 0.2*float64(depth)) {
			continue
		}

		// import an already existing package
		// if enough packages import each other, we get the chance of cycles
		var child *file.Package
		if chance(t.r, 0.5) && len(t.pkgs) > 1 {
			pool := make([]*file.Package, 0, len(t.pkgs))
			for _, cand := range iters.OrderedByKey(t.pkgs) {
				pool = append(pool, cand)
			}
			if len(pool) > 0 {
				child = pick(t.r, pool...)
			}
		}
		if child == nil {
			impPath := file.PackagePath(string(f.PathInModule()) + "/imp" + itoa(i)) // guaranteed unique
			child = createPackage(impPath)
			child.Name = t.pickAlias()
			t.pkgs[child.CorgiImportPath] = child
		}

		alias := t.pickAlias()
		imp := createImport(f, pos, alias, child.CorgiImportPath)
		appendNamespace(namespaces, imp, child)
		if chance(t.r, 0.01) { // duplicate import
			createImport(f, pos, ".", child.CorgiImportPath)
		}
	}
}

func (t *PackageTree) genNodesInFile(pos *ast.Position, f *file.File, aliases []file.Qualifier) {
	localNames := make([]file.Identifier, 0, 16)
	n := 3 + t.r.IntN(30)
	for range n {
		switch t.r.IntN(8) {
		case 0:
			name := t.pickIdentifier()
			comp := createComponent(f, pos, name)
			localNames = append(localNames, name)
			for range t.r.IntN(4) {
				createParameter(comp, pos, t.pickIdentifier())
			}
		case 1:
			createElementSpec(f, pos, "", t.pickElem(), t.pickElemType())
		case 2:
			prefix := file.CanonicalAttributeName(pick(t.r, "hx-", "data-", ""))
			name := t.pickAttr()
			createBasicAttributeSpec(f, pos, prefix, name, nil, t.pickAttrType())
		case 3:
			var q file.Qualifier
			if len(aliases) > 0 && chance(t.r, 0.5) {
				q = pick(t.r, aliases...)
			}
			var name file.Identifier
			if q == "" && len(localNames) > 0 && chance(t.r, 0.5) {
				name = pick(t.r, localNames...)
			} else {
				name = file.Identifier(pick(t.r, "A", "B", "C"))
			}
			cc := createComponentCall(f, pos, q, name)
			for range t.r.IntN(4) {
				createArgument(cc, pos, t.pickIdentifier())
			}
		case 4:
			var q file.Qualifier
			if len(aliases) > 0 && chance(t.r, 0.33) {
				q = pick(t.r, aliases...)
			}
			createElementReference(f, pos, q, t.pickElem())
		case 5:
			var q file.Qualifier
			if len(aliases) > 0 && chance(t.r, 0.33) {
				q = pick(t.r, aliases...)
			}
			createAttributeReference(f, pos, q, t.pickAttr())
		}
	}
}

func (t *PackageTree) pickAlias() file.Qualifier {
	// Bias: most imports unaliased.
	switch {
	case chance(t.r, 0.85):
		return ""
	case chance(t.r, 0.15):
		return "."
	case chance(t.r, 0.9):
		return file.Qualifier("imp" + itoa(t.r.IntN(10)))
	default:
		return "__corgi_disallowed"
	}
}

var names = func() []file.Identifier {
	qualifier := make([]file.Identifier, 0, 26)
	for c := 'A'; c <= 'Z'; c++ {
		qualifier = append(qualifier, file.Identifier(c))
	}
	for c := 'a'; c <= 'z'; c++ {
		qualifier = append(qualifier, file.Identifier(c))
	}
	qualifier = append(qualifier, "Biscuit", "Muffin", "Waffles", "Toast", "Pancake")
	qualifier = append(qualifier, "biscuit", "muffin", "waffles", "toast", "pancake")
	return qualifier
}()

func (t *PackageTree) pickIdentifier() file.Identifier {
	return pick(t.r, names...)
}

func (t *PackageTree) pickElem() string {
	return pick(t.r, "div", "span", "button", "p", "a", "custom-elem")
}

func (t *PackageTree) pickAttr() string {
	return pick(t.r, "id", "class", "title", "hx-get", "data-id")
}

var allAttrTypes = func() []attrtype.Type {
	ts := make([]attrtype.Type, 1, len(attrtype.All)+1)
	ts[0] = attrtype.Unknown
	ts = append(ts, attrtype.All[:]...)
	return ts
}()

func (t *PackageTree) pickAttrType() attrtype.Type { return pick(t.r, allAttrTypes...) }

var allElemTypes = func() []elemtype.Type {
	ts := make([]elemtype.Type, 1, len(elemtype.All)+1)
	ts[0] = elemtype.Unknown
	ts = append(ts, elemtype.All[:]...)
	return ts
}()

func (t *PackageTree) pickElemType() elemtype.Type { return pick(t.r, allElemTypes...) }

// ============================================================================
// Helpers
// ======================================================================================

func itoa(i int) string { return string('0' + rune(i%10)) }

// chance returns true with probability p in [0,1].
func chance(r *rand.Rand, p float64) bool { return r.Float64() < p }

// pick selects a random value from xs using r. xs must be non-empty.
func pick[T any](r *rand.Rand, xs ...T) T { return xs[r.IntN(len(xs))] }

func appendNamespace(aliases *[]file.Qualifier, imp *file.Import, child *file.Package) {
	switch {
	case imp.Alias == ".":
		*aliases = append(*aliases, "")
	case imp.Alias != "":
		*aliases = append(*aliases, imp.Alias)
	default:
		*aliases = append(*aliases, child.Name)
	}
}
