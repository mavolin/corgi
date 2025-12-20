package analyze

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/internal/meta"
	"golang.org/x/tools/go/packages"
)

// TestGetterUsage ensures that if a getter exists in assertions.go for a given
// type member (field or method), other files in this package do not read that
// member directly.
// Only the function that assigns to that member is allowed to read or write it
// directly.
func TestGetterUsage(t *testing.T) {
	requirers, err := loadRequirers()
	if err != nil {
		t.Fatalf("loadRequirers: %v", err)
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedCompiledGoFiles | packages.NeedTypes | packages.NeedTypesInfo,
		Dir: ".",
	}, meta.Module+"/load/analyze/internal/analyze")
	if err != nil {
		t.Fatalf("packages.Load: %v", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		t.Fatalf("failed to load package")
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	pkg := pkgs[0]

	var errs []string

	// Walk all non-test, non-assertions files looking for direct reads of
	// banned members.
	for i, f := range pkg.Syntax {
		filename := pkg.CompiledGoFiles[i]
		if shouldSkip(filename) {
			continue
		}
		parents := buildParentMap(f)
		funcAssigned := collectAssignedKeysPerFunc(requirers, pkg, f, parents)

		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ban := isBanned(requirers, pkg, sel, nil)
			if ban == "" {
				return true
			}

			if isOnAssignmentLHS(parents, sel) {
				return true
			}
			if fn := enclosingFunc(parents, sel); fn != nil {
				if assigned := funcAssigned[fn]; assigned != nil && assigned[ban] {
					return true
				}
			}

			pos := pkg.Fset.Position(sel.Sel.Pos())
			errs = append(errs, pos.String()+": direct access to "+ban+"; use the getter instead")
			return true
		})
	}

	if len(errs) > 0 {
		for _, e := range errs {
			t.Error(e)
		}
		t.Fatalf("found %d direct access(es) to members that have getters", len(errs))
	}
}

func isBanned(gs []requireableGroup, pkg *packages.Package, expr ast.Expr, sels []string) string {
	selExpr, _ := expr.(*ast.SelectorExpr)
	if selExpr == nil {
		return ""
	}
	sels = append([]string{selExpr.Sel.Name}, sels...)

	if ban := isBanned(gs, pkg, selExpr.X, sels); ban != "" {
		return ban
	}

	xType := pkg.TypesInfo.TypeOf(selExpr.X)
	for {
		if ptr, _ := xType.(*types.Pointer); ptr != nil {
			xType = ptr.Elem()
			continue
		}
		break
	}
	xTypeNamed, _ := xType.(*types.Named)
	if xTypeNamed == nil {
		return ""
	}
	xTypeObj := xTypeNamed.Obj()
	if xTypeObj.Pkg().Name() != "file" {
		return ""
	}

	if isBannedSymbol(gs, xTypeObj.Name(), sels...) {
		return xTypeObj.Name() + "." + strings.Join(sels, ".")
	}
	return ""
}

func isBannedSymbol(gs []requireableGroup, typ string, chain ...string) bool {
	if len(chain) == 0 {
		return false
	}
	for _, g := range gs {
		if g.Name == typ {
			return isBannedSymbolGroup(g, chain...)
		}
	}
	return false
}

func isBannedSymbolGroup(g requireableGroup, chain ...string) bool {
	if len(chain) == 0 {
		return false
	}
	for _, s := range g.Symbols {
		if s.Name == chain[0] {
			return len(chain) == 1
		}
	}
	for _, f := range g.Fields {
		if f.Name == chain[0] {
			if len(chain) == 1 {
				return false
			}
			return isBannedSymbolGroup(f, chain[1:]...)
		}
	}
	return false
}

// buildParentMap builds a parent relation map for all nodes in the file.
func buildParentMap(f *ast.File) map[ast.Node]ast.Node {
	parents := make(map[ast.Node]ast.Node)
	var stack []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return true
		}
		if len(stack) > 0 {
			parents[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		return true
	})
	return parents
}

// collectAssignedKeysPerFunc collects, for each function literal/decl,
// the set of keys that are assigned at least once on the LHS inside it.
func collectAssignedKeysPerFunc(requirers []requireableGroup, pkg *packages.Package, f *ast.File, parents map[ast.Node]ast.Node) map[ast.Node]map[string]bool {
	out := make(map[ast.Node]map[string]bool)

	ast.Inspect(f, func(n ast.Node) bool {
		// Case 1: direct assignment to a field on LHS (x.Field = ...).
		if as, ok := n.(*ast.AssignStmt); ok {
			if as.Tok == token.ASSIGN {
				// Determine the enclosing function for this assignment.
				fn := enclosingFunc(parents, n)
				if fn == nil {
					return true
				}
				for _, lhs := range as.Lhs {
					key := isBanned(requirers, pkg, lhs, nil)
					if key != "" {
						recordAssignedKey(out, parents, fn, key)
					}
				}
			}
			return true
		}

		// Case 2: calls on a banned field (x.Field.Set...(...)). Treat as an assignment
		// to allow reads of that member within the same function.
		if call, ok := n.(*ast.CallExpr); ok {
			// Find enclosing function for this call.
			fn := enclosingFunc(parents, n)
			if fn == nil {
				return true
			}
			// We are interested in method calls where the receiver is a selector to a banned field.
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				// Only consider setter-like methods (name starts with "Set").
				if !strings.HasPrefix(sel.Sel.Name, "Set") {
					return true
				}
				if ban := isBanned(requirers, pkg, sel.X, nil); ban != "" {
					recordAssignedKey(out, parents, fn, ban)
				}
			}
			return true
		}

		return true
	})

	return out
}

// recordAssignedKey records a banned key as assigned within the given function
// and propagates this information up to enclosing functions.
func recordAssignedKey(out map[ast.Node]map[string]bool, parents map[ast.Node]ast.Node, fn ast.Node, key string) {
	if _, ok := out[fn]; !ok {
		out[fn] = make(map[string]bool)
	}
	out[fn][key] = true
	for p := parents[fn]; p != nil; p = parents[p] {
		switch p.(type) {
		case *ast.FuncDecl, *ast.FuncLit:
			if _, ok := out[p]; !ok {
				out[p] = make(map[string]bool)
			}
			out[p][key] = true
		}
	}
}

// enclosingFunc finds the nearest enclosing function literal or declaration for node n.
func enclosingFunc(parents map[ast.Node]ast.Node, n ast.Node) ast.Node {
	for p := parents[n]; p != nil; p = parents[p] {
		switch p.(type) {
		case *ast.FuncDecl, *ast.FuncLit:
			return p
		}
	}
	return nil
}

// isOnAssignmentLHS reports whether sel is on the left-hand side of an
// assignment statement. Only plain assignments ("=") allow direct access.
func isOnAssignmentLHS(parents map[ast.Node]ast.Node, sel *ast.SelectorExpr) bool {
	// Walk up if the selector is nested in indexing or further selection on LHS
	// (e.g., a.Field[i] = ... or a.Field.Sub = ...). Those still count as LHS
	// if ultimately under AssignStmt LHS.
	curr := ast.Node(sel)
	lhsIndex := -1
	for {
		pp := parents[curr]
		switch x := pp.(type) {
		case *ast.IndexExpr, *ast.SliceExpr, *ast.SelectorExpr, *ast.StarExpr, *ast.ParenExpr:
			curr = pp
			continue
		case *ast.AssignStmt:
			// Find if curr is one of the LHS expressions.
			if x.Tok != token.ASSIGN {
				return false
			}
			for i, l := range x.Lhs {
				if l == curr {
					lhsIndex = i
					break
				}
			}
			return lhsIndex >= 0
		default:
			return false
		}
	}
}

// shouldSkip tells whether a file should be skipped (requirer.go and *_test.go files).
func shouldSkip(filename string) bool {
	base := filepath.Base(filename)
	return base == "requirer.go" || base == "assertions.go" || strings.HasSuffix(base, "_test.go")
}

type (
	requireableGroup struct {
		Name    string
		Symbols []requireableSymbol
		Fields  []requireableGroup
	}

	requireableSymbol struct {
		Name   string
		Type   string
		Method bool
		Params [][2]string
	}
)

func loadRequirers() ([]requireableGroup, error) {
	pkgs, _ := packages.Load(&packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedName,
	}, ".", meta.Module+"/file")

	analyzePkg := pkgs[slices.IndexFunc(pkgs, func(p *packages.Package) bool {
		return p.Name == "analyze"
	})]
	filePkg := pkgs[slices.IndexFunc(pkgs, func(p *packages.Package) bool {
		return p.Name == "file"
	})]

	return loadRequirersFromFile(analyzePkg, filePkg), nil
}

func loadRequirersFromFile(analyzePkg, filePkg *packages.Package) []requireableGroup {
	groups := make([]requireableGroup, 0)

	requirerObj := analyzePkg.Types.Scope().Lookup("requirer")
	requirerType := requirerObj.Type().(*types.Named)

	for fn := range requirerType.Methods() {
		split := strings.Split(fn.Name(), "_")
		symbol := split[len(split)-1]

		var group *requireableGroup

		chain := split[:len(split)-1]
		fields := &groups
		for _, name := range chain {
			i := slices.IndexFunc(*fields, func(g requireableGroup) bool {
				return g.Name == name
			})
			if i == -1 {
				newGroup := requireableGroup{Name: name, Fields: make([]requireableGroup, 0)}
				*fields = append(*fields, newGroup)
				group = &(*fields)[len(*fields)-1]
			} else {
				group = &(*fields)[i]
			}
			fields = &group.Fields
		}

		var symObj types.Object
		symObj = fn.Type().(*types.Signature).Params().At(1)
		for _, field := range split[1:] {
			symObj, _, _ = types.LookupFieldOrMethod(symObj.Type(), false, filePkg.Types, field)
		}

		rs := requireableSymbol{
			Name: symbol,
		}

		switch symObj := symObj.(type) {
		case *types.Var:
			rs.Type = typeString(symObj.Type())
		case *types.Func:
			signature := symObj.Type().(*types.Signature)
			rs.Type = typeString(signature.Results())
			rs.Method = true
			rs.Params = make([][2]string, 0, signature.Params().Len())
			for p := range signature.Params().Variables() {
				rs.Params = append(rs.Params, [2]string{p.Name(), typeString(p.Type())})
			}
		}

		group.Symbols = append(group.Symbols, rs)
	}

	return groups
}

func typeString(t types.Type) string {
	return types.TypeString(t, (*types.Package).Name)
}
