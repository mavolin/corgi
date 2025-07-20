package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestStatement(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, ParsedStatement(), testParsedStatement)
	testutil.AssertAlsoFulfils(t, Statement(Regular), func(t *testing.T, f parser.Func[*ast.Statement]) {
		testSimpleStatement(t, func(p *parser.Parser) (*ast.SimpleStatement, *diagnostic.Diagnostic) {
			s, err := f(p)
			if err != nil {
				return nil, err
			}

			ss := &ast.SimpleStatement{Nodes: make(ast.Code, len(s.Nodes))}
			for i, sn := range s.Nodes {
				ss.Nodes[i] = sn.(ast.CodeNode)
			}
			ss.Parsed = s.Parsed.(ast.ParsedSimpleStatement)

			return ss, nil
		})
	})
}

func TestParsedStatement(t *testing.T) {
	t.Parallel()
	testParsedStatement(t, ParsedStatement())
}

func testParsedStatement(t *testing.T, f parser.Func[*ast.Statement]) {
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testReturn))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testBreak))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testContinue))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testFallthrough))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testDefer))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testIncDec))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testLabel))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testZeroCoalescingAssignment))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testConstDeclaration))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testVarDeclaration))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testShortVarDeclaration))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testLabel))
	testutil.AssertAlsoFulfils(t, f, parsedStatementAsStatement(testAssignment))
}

func parsedStatementAsStatement[PS ast.ParsedStatement](subTest func(*testing.T, parser.Func[PS])) func(*testing.T, parser.Func[*ast.Statement]) {
	return func(t *testing.T, f parser.Func[*ast.Statement]) {
		subTest(t, func(p *parser.Parser) (PS, *diagnostic.Diagnostic) {
			var zero PS

			s, err := f(p)
			if err != nil {
				return zero, err
			}

			if s.Parsed == nil {
				return zero, &diagnostic.Diagnostic{
					Message: "missing parsed statement",
				}
			}

			return s.Parsed.(PS), nil
		})
	}
}

func TestSimpleStatement(t *testing.T) {
	t.Parallel()
	testSimpleStatement(t, SimpleStatement(Regular))
}

func testSimpleStatement(t *testing.T, f parser.Func[*ast.SimpleStatement]) {
	testutil.AssertAlsoFulfils(t, f, testParsedSimpleStatement)
}

func TestParsedSimpleStatement(t *testing.T) {
	t.Parallel()
	testParsedSimpleStatement(t, ParsedSimpleStatement())
}

func testParsedSimpleStatement(t *testing.T, f parser.Func[*ast.SimpleStatement]) {
	testutil.AssertAlsoFulfils(t, f, parsedSimpleStatementAsSimpleStatement(testIncDec))
	testutil.AssertAlsoFulfils(t, f, parsedSimpleStatementAsSimpleStatement(testZeroCoalescingAssignment))
	testutil.AssertAlsoFulfils(t, f, parsedSimpleStatementAsSimpleStatement(testShortVarDeclaration))
	testutil.AssertAlsoFulfils(t, f, parsedSimpleStatementAsSimpleStatement(testAssignment))
}

func parsedSimpleStatementAsSimpleStatement[PS ast.ParsedSimpleStatement](subTest func(*testing.T, parser.Func[PS])) func(*testing.T, parser.Func[*ast.SimpleStatement]) {
	return func(t *testing.T, f parser.Func[*ast.SimpleStatement]) {
		subTest(t, func(p *parser.Parser) (PS, *diagnostic.Diagnostic) {
			var zero PS

			s, err := f(p)
			if err != nil {
				return zero, err
			}

			if s.Parsed == nil {
				return zero, &diagnostic.Diagnostic{
					Message: "missing parsed statement",
				}
			}

			return s.Parsed.(PS), nil
		})
	}
}

func TestReturn(t *testing.T) {
	t.Parallel()
	testReturn(t, Return())
}

func testReturn(t *testing.T, f parser.Func[*ast.Return]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.Return
	}{
		{
			name: "no error",
			in:   "return",
			expect: &ast.Return{
				Return: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "with error",
			in:   "return err",
			expect: &ast.Return{
				Return: &ast.Position{Line: 1, Col: 1},
				Error: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{
							Code:     "err",
							Position: &ast.Position{Line: 1, Col: 8},
						},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestBreak(t *testing.T) {
	t.Parallel()
	testBreak(t, Break())
}

func testBreak(t *testing.T, f parser.Func[*ast.Break]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.Break
	}{
		{
			name: "no label",
			in:   "break",
			expect: &ast.Break{
				Break: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "with label",
			in:   "break myLabel",
			expect: &ast.Break{
				Break: &ast.Position{Line: 1, Col: 1},
				Label: &ast.Identifier{
					Name:     "myLabel",
					Position: &ast.Position{Line: 1, Col: 7},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestContinue(t *testing.T) {
	t.Parallel()
	testContinue(t, Continue())
}

func testContinue(t *testing.T, f parser.Func[*ast.Continue]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.Continue
	}{
		{
			name: "no label",
			in:   "continue",
			expect: &ast.Continue{
				Continue: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "with label",
			in:   "continue myLabel",
			expect: &ast.Continue{
				Continue: &ast.Position{Line: 1, Col: 1},
				Label: &ast.Identifier{
					Name:     "myLabel",
					Position: &ast.Position{Line: 1, Col: 10},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestFallthrough(t *testing.T) {
	t.Parallel()
	testFallthrough(t, Fallthrough())
}

func testFallthrough(t *testing.T, f parser.Func[*ast.Fallthrough]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.Fallthrough
	}{
		{
			name: "no label",
			in:   "fallthrough",
			expect: &ast.Fallthrough{
				Fallthrough: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "with label",
			in:   "fallthrough myLabel",
			expect: &ast.Fallthrough{
				Fallthrough: &ast.Position{Line: 1, Col: 1},
				Label: &ast.Identifier{
					Name:     "myLabel",
					Position: &ast.Position{Line: 1, Col: 13},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestDefer(t *testing.T) {
	t.Parallel()
	testDefer(t, Defer())
}

func testDefer(t *testing.T, f parser.Func[*ast.Defer]) {
	in := "defer foo(bar, baz)"
	expect := &ast.Defer{
		Defer: &ast.Position{Line: 1, Col: 1},
		Expression: &ast.Expression{
			Nodes: ast.Code{
				&ast.GoCode{
					Code:     "foo(bar, baz)",
					Position: &ast.Position{Line: 1, Col: 7},
				},
			},
		},
	}

	actual := parsesCodeNodeFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestZeroCoalescingAssignment(t *testing.T) {
	t.Parallel()
	testZeroCoalescingAssignment(t, ZeroCoalescingAssignment())
}

func testZeroCoalescingAssignment(t *testing.T, f parser.Func[*ast.ZeroCoalescingAssignment]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.ZeroCoalescingAssignment
	}{
		{
			name: "assignment",
			in:   "foo = bar?",
			expect: &ast.ZeroCoalescingAssignment{
				ValueExpression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				EqualSign: &ast.Position{Line: 1, Col: 5},
				Expression: &ast.ZeroCoalescing{
					Root: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "bar", Position: &ast.Position{Line: 1, Col: 7}},
						},
					},
					CheckRoot: &ast.Position{Line: 1, Col: 10},
				},
			},
		}, {
			name: "declaration",
			in:   "foo := bar?",
			expect: &ast.ZeroCoalescingAssignment{
				ValueExpression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				Colon:     &ast.Position{Line: 1, Col: 5},
				EqualSign: &ast.Position{Line: 1, Col: 6},
				Expression: &ast.ZeroCoalescing{
					Root: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "bar", Position: &ast.Position{Line: 1, Col: 8}},
						},
					},
					CheckRoot: &ast.Position{Line: 1, Col: 11},
				},
			},
		}, {
			name: "with ok var",
			in:   "foo, ok = bar?",
			expect: &ast.ZeroCoalescingAssignment{
				ValueExpression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				VarComma: &ast.Position{Line: 1, Col: 4},
				OkExpression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "ok", Position: &ast.Position{Line: 1, Col: 6}},
					},
				},
				EqualSign: &ast.Position{Line: 1, Col: 9},
				Expression: &ast.ZeroCoalescing{
					DerefCount: 0,
					Root: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "bar", Position: &ast.Position{Line: 1, Col: 11}},
						},
					},
					CheckRoot: &ast.Position{Line: 1, Col: 14},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestIncDec(t *testing.T) {
	t.Parallel()
	testIncDec(t, IncDec())
}

func testIncDec(t *testing.T, f parser.Func[*ast.IncDec]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.IncDec
	}{
		{
			name: "increment",
			in:   "foo++",
			expect: &ast.IncDec{
				Expression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				IncrPos: &ast.Position{Line: 1, Col: 4},
			},
		}, {
			name: "decrement",
			in:   "foo--",
			expect: &ast.IncDec{
				Expression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				DecrPos: &ast.Position{Line: 1, Col: 4},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestLabel(t *testing.T) {
	t.Parallel()
	testLabel(t, Label())
}

func testLabel(t *testing.T, f parser.Func[*ast.Label]) {
	in := "myLabel:"
	expect := &ast.Label{
		Name:  &ast.Identifier{Name: "myLabel", Position: &ast.Position{Line: 1, Col: 1}},
		Colon: &ast.Position{Line: 1, Col: 8},
	}

	actual := parsesCodeNodeFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestConstDeclaration(t *testing.T) {
	t.Parallel()
	testConstDeclaration(t, ConstDeclaration())
}

func testConstDeclaration(t *testing.T, f parser.Func[*ast.ConstDeclaration]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.ConstDeclaration
	}{
		{
			name: "no type",
			in:   "const foo = 42",
			expect: &ast.ConstDeclaration{
				Const: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.ConstSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 7}},
						},
						EqualSign: &ast.Position{Line: 1, Col: 11},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 13}},
								},
							},
						},
					},
				},
			},
		}, {
			name: "with type",
			in:   "const foo int = 42",
			expect: &ast.ConstDeclaration{
				Const: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.ConstSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 7}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 11}},
							},
							Type:  "int",
							From:  ast.Position{Line: 1, Col: 11},
							Until: ast.Position{Line: 1, Col: 14},
						},
						EqualSign: &ast.Position{Line: 1, Col: 15},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 17}},
								},
							},
						},
					},
				},
			},
		}, {
			name: "multiple names, multiple values",
			in:   "const foo, bar = 42, 43",
			expect: &ast.ConstDeclaration{
				Const: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.ConstSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 7}},
							{Name: "bar", Position: &ast.Position{Line: 1, Col: 12}},
						},
						EqualSign: &ast.Position{Line: 1, Col: 16},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 18}},
								},
							}, {
								Nodes: ast.Code{
									&ast.GoCode{Code: "43", Position: &ast.Position{Line: 1, Col: 22}},
								},
							},
						},
					},
				},
			},
		}, {
			name: "multiple names, single value",
			in:   "const foo, bar = baz()",
			expect: &ast.ConstDeclaration{
				Const: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.ConstSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 7}},
							{Name: "bar", Position: &ast.Position{Line: 1, Col: 12}},
						},
						EqualSign: &ast.Position{Line: 1, Col: 16},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "baz()", Position: &ast.Position{Line: 1, Col: 18}},
								},
							},
						},
					},
				},
			},
		}, {
			name: "multiple specs",
			in: "const (\n" +
				"\tfoo = 42\n" +
				"\tbar = 43" +
				")",
			expect: &ast.ConstDeclaration{
				Const:  &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 7},
				Specs: []*ast.ConstSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 2, Col: 2}},
						},
						EqualSign: &ast.Position{Line: 2, Col: 6},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 2, Col: 8}},
								},
							},
						},
					}, {
						Names: []*ast.Identifier{
							{Name: "bar", Position: &ast.Position{Line: 3, Col: 2}},
						},
						EqualSign: &ast.Position{Line: 3, Col: 6},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "43", Position: &ast.Position{Line: 3, Col: 8}},
								},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 3, Col: 10},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestVarDeclaration(t *testing.T) {
	t.Parallel()
	testVarDeclaration(t, VarDeclaration())
}

func testVarDeclaration(t *testing.T, f parser.Func[*ast.VarDeclaration]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.VarDeclaration
	}{
		{
			name: "no type",
			in:   "var foo = 42",
			expect: &ast.VarDeclaration{
				Var: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.VarSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 5}},
						},
						EqualSign: &ast.Position{Line: 1, Col: 9},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 11}},
								},
							},
						},
					},
				},
			},
		}, {
			name: "with type",
			in:   "var foo int = 42",
			expect: &ast.VarDeclaration{
				Var: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.VarSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 5}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 9}},
							},
							Type:  "int",
							From:  ast.Position{Line: 1, Col: 9},
							Until: ast.Position{Line: 1, Col: 12},
						},
						EqualSign: &ast.Position{Line: 1, Col: 13},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 15}},
								},
							},
						},
					},
				},
			},
		}, {
			name: "only type",
			in:   "var foo int",
			expect: &ast.VarDeclaration{
				Var: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.VarSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 5}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 9}},
							},
							Type:  "int",
							From:  ast.Position{Line: 1, Col: 9},
							Until: ast.Position{Line: 1, Col: 12},
						},
					},
				},
			},
		}, {
			name: "multiple names, multiple values",
			in:   "var foo, bar = 42, 43",
			expect: &ast.VarDeclaration{
				Var: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.VarSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 5}},
							{Name: "bar", Position: &ast.Position{Line: 1, Col: 10}},
						},
						EqualSign: &ast.Position{Line: 1, Col: 14},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 16}},
								},
							}, {
								Nodes: ast.Code{
									&ast.GoCode{Code: "43", Position: &ast.Position{Line: 1, Col: 20}},
								},
							},
						},
					},
				},
			},
		}, {
			name: "multiple names, single value",
			in:   "var foo, bar = baz()",
			expect: &ast.VarDeclaration{
				Var: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.VarSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 5}},
							{Name: "bar", Position: &ast.Position{Line: 1, Col: 10}},
						},
						EqualSign: &ast.Position{Line: 1, Col: 14},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "baz()", Position: &ast.Position{Line: 1, Col: 16}},
								},
							},
						},
					},
				},
			},
		}, {
			name: "multiple specs",
			in: "var (\n" +
				"\tfoo = 42\n" +
				"\tbar = 43" +
				")",
			expect: &ast.VarDeclaration{
				Var:    &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 5},
				Specs: []*ast.VarSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 2, Col: 2}},
						},
						EqualSign: &ast.Position{Line: 2, Col: 6},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 2, Col: 8}},
								},
							},
						},
					}, {
						Names: []*ast.Identifier{
							{Name: "bar", Position: &ast.Position{Line: 3, Col: 2}},
						},
						EqualSign: &ast.Position{Line: 3, Col: 6},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "43", Position: &ast.Position{Line: 3, Col: 8}},
								},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 3, Col: 10},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestShortVarDeclaration(t *testing.T) {
	t.Parallel()
	testShortVarDeclaration(t, ShortVarDeclaration())
}

func testShortVarDeclaration(t *testing.T, f parser.Func[*ast.ShortVarDeclaration]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.ShortVarDeclaration
	}{
		{
			name: "single",
			in:   "foo := 42",
			expect: &ast.ShortVarDeclaration{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				},
				ColonEqualSign: &ast.Position{Line: 1, Col: 5},
				Values: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 8}},
						},
					},
				},
			},
		}, {
			name: "multiple",
			in:   "foo, bar := 42, 43",
			expect: &ast.ShortVarDeclaration{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					{Name: "bar", Position: &ast.Position{Line: 1, Col: 6}},
				},
				ColonEqualSign: &ast.Position{Line: 1, Col: 10},
				Values: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 13}},
						},
					}, {
						Nodes: ast.Code{
							&ast.GoCode{Code: "43", Position: &ast.Position{Line: 1, Col: 17}},
						},
					},
				},
			},
		}, {
			name: "multiple, single value",
			in:   "foo, bar := baz()",
			expect: &ast.ShortVarDeclaration{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					{Name: "bar", Position: &ast.Position{Line: 1, Col: 6}},
				},
				ColonEqualSign: &ast.Position{Line: 1, Col: 10},
				Values: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "baz()", Position: &ast.Position{Line: 1, Col: 13}},
						},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestAssignment(t *testing.T) {
	t.Parallel()
	testAssignment(t, Assignment())
}

func testAssignment(t *testing.T, f parser.Func[*ast.Assignment]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.Assignment
	}{
		{
			name: "single",
			in:   "foo = 42",
			expect: &ast.Assignment{
				LHS: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
						},
					},
				},
				OperatorPosition: &ast.Position{Line: 1, Col: 5},
				RHS: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 7}},
						},
					},
				},
			},
		}, {
			name: "multiple",
			in:   "foo, bar = 42, 43",
			expect: &ast.Assignment{
				LHS: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
						},
					}, {
						Nodes: ast.Code{
							&ast.GoCode{Code: "bar", Position: &ast.Position{Line: 1, Col: 6}},
						},
					},
				},
				OperatorPosition: &ast.Position{Line: 1, Col: 10},
				RHS: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 12}},
						},
					}, {
						Nodes: ast.Code{
							&ast.GoCode{Code: "43", Position: &ast.Position{Line: 1, Col: 16}},
						},
					},
				},
			},
		}, {
			name: "multiple, single value",
			in:   "foo, bar = baz()",
			expect: &ast.Assignment{
				LHS: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
						},
					}, {
						Nodes: ast.Code{
							&ast.GoCode{Code: "bar", Position: &ast.Position{Line: 1, Col: 6}},
						},
					},
				},
				OperatorPosition: &ast.Position{Line: 1, Col: 10},
				RHS: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "baz()", Position: &ast.Position{Line: 1, Col: 12}},
						},
					},
				},
			},
		}, {
			name: "special operator",
			in:   "foo += 42",
			expect: &ast.Assignment{
				LHS: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 1}},
						},
					},
				},
				OperatorPosition: &ast.Position{Line: 1, Col: 5},
				SpecialOperator:  "+",
				RHS: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 8}},
						},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}
