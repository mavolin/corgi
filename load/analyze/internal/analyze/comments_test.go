package analyze

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Helper types and functions

type analyzerInfo struct {
	methods map[string]*methodInfo
}

type methodInfo struct {
	analyzerVar       string
	body              *ast.BlockStmt
	checkDependencies []string
	setsFields        []string
	fieldDependencies []string
}

func parseAnalyzerMethods() (*analyzerInfo, error) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		return nil, fmt.Errorf("found no Go files: %w", err)
	}

	analyzer := &analyzerInfo{
		methods: make(map[string]*methodInfo),
	}

	errs := make([]error, 0, 16)
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}

		for _, decl := range f.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			r, _ := utf8.DecodeRuneInString(funcDecl.Name.Name)
			if !unicode.IsUpper(r) { // not exported
				continue
			}

			info := methodInfo{body: funcDecl.Body}

			if funcDecl.Recv == nil || len(funcDecl.Recv.List) != 1 { // not a method
				if funcDecl.Name.Name != "Analyze" {
					continue
				}
				info.analyzerVar = "z"
			} else {
				recv := funcDecl.Recv.List[0]
				if !isAnalyzerReceiver(recv) {
					continue
				}
				info.analyzerVar = recv.Names[0].Name
			}

			if funcDecl.Name.Name != "Analyze" {
				if funcDecl.Doc == nil {
					errs = append(errs, fmt.Errorf("%s: missing doc comment for exported method", funcDecl.Name.Name))
				} else {
					info.checkDependencies, err = extractSection(funcDecl.Doc.Text(), "Depends on Checks")
					if err != nil {
						errs = append(errs, fmt.Errorf("%s: %w", funcDecl.Name.Name, err))
					}

					info.setsFields, err = extractSection(funcDecl.Doc.Text(), "Sets Fields")
					if err != nil {
						errs = append(errs, fmt.Errorf("%s: %w", funcDecl.Name.Name, err))
					}

					info.fieldDependencies, err = extractSection(funcDecl.Doc.Text(), "Depends on Fields")
					if err != nil {
						errs = append(errs, fmt.Errorf("%s: %w", funcDecl.Name.Name, err))
					}
				}
			}

			analyzer.methods[funcDecl.Name.Name] = &info
		}
	}

	return analyzer, errors.Join(errs...)
}

func isAnalyzerReceiver(recv *ast.Field) bool {
	switch expr := recv.Type.(type) {
	case *ast.StarExpr:
		ident, ok := expr.X.(*ast.Ident)
		return ok && ident.Name == "analyzer"
	case *ast.Ident:
		return expr.Name == "analyzer"
	}
	return false
}

func extractSection(docText string, name string) ([]string, error) {
	if strings.Contains(docText, name+": None") {
		return nil, nil
	}

	// Find section through header
	sectionRegexp := regexp.MustCompile(`(?s)` + name + `:\n(.*?)(?:\n\n|\z)`)
	matches := sectionRegexp.FindStringSubmatch(docText)
	if len(matches) < 2 {
		return nil, fmt.Errorf("missing section '%s' in doc comment", name)
	}

	// Extract each item
	list := matches[1]
	listItemRegexp := regexp.MustCompile(`(?m)^\s*-\s+(\S+)`)
	listItemMatches := listItemRegexp.FindAllStringSubmatch(list, -1)

	items := make([]string, len(listItemMatches))
	for i, match := range listItemMatches {
		if len(match) < 2 {
			return nil, fmt.Errorf("malformed list item in section '%s'", name)
		}
		items[i] = match[1]
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no items found in section '%s'", name)
	}

	return items, nil
}

type executionOrder struct {
	name   string
	method *methodInfo
	calls  []*executionOrder
}

func (o *executionOrder) findBetween(name, end string) *executionOrder {
	switch o.name {
	case end:
		return nil
	case name:
		return o
	}

	for _, call := range o.calls {
		if found := call.findBetween(name, end); found != nil {
			return found
		}
	}
	return nil
}

func (o *executionOrder) find(name string) *executionOrder {
	if o.name == name {
		return o
	}
	for _, call := range o.calls {
		if found := call.find(name); found != nil {
			return found
		}
	}
	return nil
}

func (o *executionOrder) contains(name string) bool {
	return o.find(name) != nil
}

func (o *executionOrder) callsBefore(name, before string) bool {
	return o.findBetween(name, before) != nil
}

func getExecutionOrder(analyzer *analyzerInfo) *executionOrder {
	const entrypoint = "Analyze" // the Analyze function (not method) is the entry point
	analyzeMethod := analyzer.methods[entrypoint]
	if analyzeMethod == nil {
		panic("Analyze function not found")
	}

	root := &executionOrder{name: "Analyze", method: analyzeMethod}
	root.calls = getCalls(analyzer, analyzeMethod, root)
	return root
}

// addCalls recursively adds all methods called by the given method to the execution order
func getCalls(analyzer *analyzerInfo, method *methodInfo, root *executionOrder) []*executionOrder {
	callNames := extractAnalyzerCalls(analyzer, method)

	calls := make([]*executionOrder, len(callNames))
	for i, call := range callNames {
		method := analyzer.methods[call]
		if root.contains(call) {
			calls[i] = &executionOrder{name: call, method: method}
			continue
		}

		calls[i] = &executionOrder{name: call, method: method}
		calls[i].calls = getCalls(analyzer, method, root)
	}
	return calls
}

func extractAnalyzerCalls(analyzer *analyzerInfo, method *methodInfo) []string {
	var calls []string
	ast.Inspect(method.body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if receiver, ok := sel.X.(*ast.Ident); ok && receiver.Name == method.analyzerVar {
					callName := sel.Sel.Name
					// Only include methods that are defined on *analyzer
					if analyzer.methods[callName] != nil {
						calls = append(calls, callName)
					}
				}
			}
		}
		return true
	})
	return calls
}
