package analyze

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// TestGetterUsage ensures that if a getter exists in assertions.go for a given
// type member (field or method), other files in this package do not read that
// member directly.
// Only the function that assigns to that member is allowed to read or write it
// directly.
func TestGetterUsage(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  ".",
	}
	pkgs, err := packages.Load(cfg, ".")
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

	// Find assertions.go and build the banned set from getters.
	var assertionsFile *ast.File
	for i, f := range pkg.Syntax {
		if filepath.Base(pkg.CompiledGoFiles[i]) == "assertions.go" {
			assertionsFile = f
			break
		}
	}
	if assertionsFile == nil {
		t.Fatalf("assertions.go not found")
	}

	banned := make(map[string]struct{})

	// Build a set of fully-qualified type+member pairs that have a getter in assertions.go.
	// The getter naming convention is: <entity>_<Member>(x *SomeType, ...)
	// We skip require* helpers.
	for _, decl := range assertionsFile.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name == nil {
			continue
		}
		// Only consider analyzer methods.
		if len(fn.Recv.List) != 1 {
			continue
		}
		// Skip require* helpers.
		if strings.HasPrefix(fn.Name.Name, "require") {
			continue
		}
		// Must follow <entity>_<Member> naming.
		name := fn.Name.Name
		us := strings.IndexByte(name, '_')
		if us <= 0 || us == len(name)-1 {
			continue
		}
		member := name[us+1:]
		if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
			continue
		}
		// First parameter type is the target type. Resolve fully-qualified type key.
		typeKey := typeKeyOfTypeExpr(pkg, fn.Type.Params.List[0].Type)
		if typeKey == "" {
			continue
		}
		banned[typeKey+"|"+member] = struct{}{}
	}

	fset := pkg.Fset
	var errs []string

	// Walk all non-test, non-assertions files looking for direct reads of
	// banned members.
	for i, f := range pkg.Syntax {
		filename := pkg.CompiledGoFiles[i]
		if isSkippable(filename) {
			continue
		}
		// Build parent map for the file.
		parents := buildParentMap(f)

		// Pre-compute, for each function (FuncDecl/FuncLit), which banned keys
		// are assigned at least once in that function.
		funcAssigned := collectAssignedBannedKeysPerFunc(pkg, f, parents, banned)

		// Now scan selectors and flag violations, with the allowances.
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			// Extract underlying fully-qualified type key of X for matching.
			typeKey := typeKeyOfExpr(pkg, sel.X)
			if typeKey == "" {
				return true
			}
			key := typeKey + "|" + sel.Sel.Name
			if _, ok := banned[key]; !ok {
				return true
			}
			// Allow if selector is on LHS of plain assignment.
			if isOnAssignmentLHS(parents, sel) {
				return true
			}
			// Allow if inside a function that assigns to this key, directly or in a nested closure.
			if fn := enclosingFunc(parents, sel); fn != nil {
				if assigned := funcAssigned[fn]; assigned != nil && assigned[key] {
					return true
				}
			}

			// Otherwise, it's a disallowed read (field access or method call).
			pos := fset.Position(sel.Sel.Pos())
			errs = append(errs, pos.String()+": direct access to "+key+"; use the getter instead")
			return true
		})
	}

	if len(errs) > 0 {
		for _, e := range errs {
			t.Error(e)
		}
		t.Fatalf("found %d direct accesses to members that have getters", len(errs))
	}
}

// typeKeyOfTypeExpr resolves a fully-qualified type key (pkgPath|TypeName) from a type expression.
func typeKeyOfTypeExpr(pkg *packages.Package, e ast.Expr) string {
	t := typeOfExpr(pkg, e)
	return typeKeyFromType(t)
}

// typeKeyOfExpr resolves a fully-qualified type key (pkgPath|TypeName) from a value expression.
func typeKeyOfExpr(pkg *packages.Package, e ast.Expr) string {
	tv, ok := pkg.TypesInfo.Types[e]
	if !ok || tv.Type == nil {
		return ""
	}
	return typeKeyFromType(tv.Type)
}

// typeKeyFromType returns the fully-qualified type key for the base named type (after stripping pointers).
func typeKeyFromType(t types.Type) string {
	if t == nil {
		return ""
	}
	if n := namedBaseType(t); n != nil && n.Obj() != nil && n.Obj().Pkg() != nil {
		return n.Obj().Pkg().Path() + "|" + n.Obj().Name()
	}
	return ""
}

// namedBaseType strips pointers and returns the underlying named type, if any.
func namedBaseType(t types.Type) *types.Named {
	for {
		switch p := t.(type) {
		case *types.Pointer:
			t = p.Elem()
		default:
			if n, ok := t.(*types.Named); ok {
				return n
			}
			return nil
		}
	}
}

// typeOfExpr returns the go/types.Type associated with the given type expression.
func typeOfExpr(pkg *packages.Package, e ast.Expr) types.Type {
	// For type expressions, TypesInfo.Types holds the type of the expression.
	if tv, ok := pkg.TypesInfo.Types[e]; ok {
		return tv.Type
	}
	// If not directly found (rare), try to look through star and selector.
	switch v := e.(type) {
	case *ast.StarExpr:
		return typeOfExpr(pkg, v.X)
	case *ast.SelectorExpr:
		if obj := pkg.TypesInfo.Uses[v.Sel]; obj != nil {
			return obj.Type()
		}
	}
	return nil
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

// collectAssignedBannedKeysPerFunc collects, for each function literal/decl,
// the set of banned keys that are assigned at least once on the LHS inside it.
func collectAssignedBannedKeysPerFunc(pkg *packages.Package, f *ast.File, parents map[ast.Node]ast.Node, banned map[string]struct{}) map[ast.Node]map[string]bool {
	out := make(map[ast.Node]map[string]bool)

	ast.Inspect(f, func(n ast.Node) bool {
		// Case 1: direct assignment to a banned field on LHS (x.Field = ...).
		if as, ok := n.(*ast.AssignStmt); ok {
			if as.Tok == token.ASSIGN {
				// Determine the enclosing function for this assignment.
				fn := enclosingFunc(parents, n)
				if fn == nil {
					return true
				}
				for _, lhs := range as.Lhs {
					// Unwrap parentheses etc.
					l := lhs
					for {
						switch v := l.(type) {
						case *ast.ParenExpr:
							l = v.X
						default:
							goto doneUnwrap
						}
					}
				doneUnwrap:
					sel, ok := l.(*ast.SelectorExpr)
					if !ok {
						continue
					}
					typeKey := typeKeyOfExpr(pkg, sel.X)
					if typeKey == "" {
						continue
					}
					key := typeKey + "|" + sel.Sel.Name
					if _, ok := banned[key]; ok {
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
				if recvSel, ok := sel.X.(*ast.SelectorExpr); ok {
					typeKey := typeKeyOfExpr(pkg, recvSel.X)
					if typeKey != "" {
						key := typeKey + "|" + recvSel.Sel.Name
						if _, ok := banned[key]; ok {
							recordAssignedKey(out, parents, fn, key)
						}
					}
				}
			}
			return true
		}

		return true
	})

	return out
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

// isSkippable tells whether a file should be skipped (assertions.go and *_test.go files).
func isSkippable(filename string) bool {
	base := filepath.Base(filename)
	return base == "assertions.go" || strings.HasSuffix(base, "_test.go")
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
