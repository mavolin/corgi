package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestType(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		t.Run("type lit", func(t *testing.T) {
			t.Parallel()
			testTypeLit(t, Type())
		})
		t.Run("named type", func(t *testing.T) {
			t.Parallel()

			in := "foo.Bar"
			want := &ast.Type{
				Type: in,
				Parsed: &ast.NamedType{
					Name: &ast.QualifiedIdentifier{
						Package: &ast.Identifier{
							Name:     "foo",
							Position: &ast.Position{Line: 1, Col: 1},
						},
						Dot: &ast.Position{Line: 1, Col: 4},
						Name: &ast.Identifier{
							Name:     "Bar",
							Position: &ast.Position{Line: 1, Col: 5},
						},
					},
				},
				From:  ast.Position{Line: 1, Col: 1},
				Until: parsetest.CalcEndPos(in),
			}

			got := parsetest.ParsesFully(t, in, Type())
			should.Equal(t, want, got)
		})
		t.Run("paren type", func(t *testing.T) {
			t.Parallel()

			in := "(foo.Bar)"
			want := &ast.Type{
				Type:  in,
				From:  ast.Position{Line: 1, Col: 1},
				Until: parsetest.CalcEndPos(in),
			}

			got := parsetest.ParsesFully(t, in, Type())
			should.Equal(t, want, got)
		})
	})

	t.Run("restore", func(t *testing.T) {
		t.Parallel()

		t.Run("paren type", func(t *testing.T) {
			t.Parallel()

			restoreCases := []string{
				"(foo",
			}

			for _, in := range restoreCases {
				t.Run(in, func(t *testing.T) {
					t.Parallel()

					want := &ast.Type{
						Type:  in,
						From:  ast.Position{Line: 1, Col: 1},
						Until: parsetest.CalcEndPos(in),
					}

					p := parsetest.NewParser(t, in+" 123")
					got := parsetest.AssertMatchesButError(t, p, Type())
					wantLine, wantCol, wantIndex := parsetest.CalcEnd(1, 1, 0, in)
					parsetest.AssertPosition(t, p, wantLine, wantCol, wantIndex)
					should.Equal(t, want, got)
				})
			}
		})
	})
}

func TestTypeLit(t *testing.T) {
	t.Parallel()
	testTypeLit(t, TypeLit())
}

func testTypeLit(t *testing.T, f parser.Func[*ast.Type]) {
	t.Run("array type", func(t *testing.T) {
		t.Parallel()
		testArrayType(t, f)
	})

	t.Run("struct type", func(t *testing.T) {
		t.Parallel()
		testStructType(t, f)
	})

	t.Run("pointer type", func(t *testing.T) {
		t.Parallel()
		testPointerType(t, f)
	})

	t.Run("function type", func(t *testing.T) {
		t.Parallel()
		testFunctionType(t, f)
	})

	t.Run("interface type", func(t *testing.T) {
		t.Parallel()
		testInterfaceType(t, f)
	})

	t.Run("slice type", func(t *testing.T) {
		t.Parallel()
		testSliceType(t, f)
	})

	t.Run("map type", func(t *testing.T) {
		t.Parallel()
		testMapType(t, f)
	})

	t.Run("channel type", func(t *testing.T) {
		t.Parallel()
		testChannelType(t, f)
	})
}

func TestArrayType(t *testing.T) {
	t.Parallel()
	testArrayType(t, ArrayType())
}

func testArrayType(t *testing.T, f parser.Func[*ast.Type]) {
	successCases := []string{
		"[3]int", "[2+3]string", "[2][3]int",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, in := range successCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				got := parsetest.ParsesFully(t, in, f)
				should.Equal(t, want, got)
			})
		}
	})

	recoverCases := []string{
		"[3", "[3]",
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, in := range recoverCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				p := parsetest.NewParser(t, in+" ; aa")
				got := parsetest.AssertMatchesButError(t, p, f)
				wantLine, wantCol, wantIndex := parsetest.CalcEnd(1, 1, 0, in)
				parsetest.AssertPosition(t, p, wantLine, wantCol, wantIndex)
				should.Equal(t, want, got)
			})
		}
	})
}

func TestStructType(t *testing.T) {
	t.Parallel()
	testStructType(t, StructType())
}

func testStructType(t *testing.T, f parser.Func[*ast.Type]) {
	successCases := []string{
		"struct{}", "struct{foo int; bar string}",
		"struct{\nfoo int\nbar string\n}",
	}

	t.Run("success", func(t *testing.T) {
		for _, in := range successCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				got := parsetest.ParsesFully(t, in, f)
				should.Equal(t, want, got)
			})
		}
	})

	recoverCases := []string{
		"struct", "struct{", "struct{foo", "struct{\nfoo int",
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, in := range recoverCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				p := parsetest.NewParser(t, in+" ")
				got := parsetest.AssertMatchesButError(t, p, f)
				wantLine, wantCol, wantIndex := parsetest.CalcEnd(1, 1, 0, in)
				parsetest.AssertPosition(t, p, wantLine, wantCol, wantIndex)
				should.Equal(t, want, got)
			})
		}
	})
}

func TestPointerType(t *testing.T) {
	t.Parallel()
	testPointerType(t, PointerType())
}

func testPointerType(t *testing.T, f parser.Func[*ast.Type]) {
	successCases := []string{
		"*int", "*string", "*[]int", "*[3]int",
	}

	for _, in := range successCases {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			want := &ast.Type{
				Type:  in,
				From:  ast.Position{Line: 1, Col: 1},
				Until: parsetest.CalcEndPos(in),
			}

			got := parsetest.ParsesFully(t, in, f)
			should.Equal(t, want, got)
		})
	}
}

func TestFunctionType(t *testing.T) {
	t.Parallel()
	testFunctionType(t, FunctionType())
}

func testFunctionType(t *testing.T, f parser.Func[*ast.Type]) {
	successCases := []string{
		"func()", "func(int) string", "func(foo int, bar string) (int, string)",
		"func(foo, bar int, baz string)", "func() (foo, bar int, baz string)",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, in := range successCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				got := parsetest.ParsesFully(t, in, f)
				should.Equal(t, want, got)
			})
		}
	})

	recoverCases := []string{
		"func",
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, in := range recoverCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				p := parsetest.NewParser(t, in+" 123")
				got := parsetest.AssertMatchesButError(t, p, f)
				wantLine, wantCol, wantIndex := parsetest.CalcEnd(1, 1, 0, in)
				parsetest.AssertPosition(t, p, wantLine, wantCol, wantIndex)
				should.Equal(t, want, got)
			})
		}
	})
}

func TestInterfaceType(t *testing.T) {
	t.Parallel()
	testInterfaceType(t, InterfaceType())
}

func testInterfaceType(t *testing.T, f parser.Func[*ast.Type]) {
	successCases := []string{
		"interface{}", "interface{foo(); bar() int}",
		"interface{\nfoo()\nbar() int\n}",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, in := range successCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				got := parsetest.ParsesFully(t, in, f)
				should.Equal(t, want, got)
			})
		}
	})

	recoverCases := []string{
		"interface", "interface{", "interface{foo", "interface{foo()",
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, in := range recoverCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				p := parsetest.NewParser(t, in+" ")
				got := parsetest.AssertMatchesButError(t, p, f)
				wantLine, wantCol, wantIndex := parsetest.CalcEnd(1, 1, 0, in)
				parsetest.AssertPosition(t, p, wantLine, wantCol, wantIndex)
				should.Equal(t, want, got)
			})
		}
	})
}

func TestSliceType(t *testing.T) {
	t.Parallel()
	testSliceType(t, SliceType())
}

func testSliceType(t *testing.T, f parser.Func[*ast.Type]) {
	successCases := []string{
		"[]int", "[]string", "[][]int",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, in := range successCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				got := parsetest.ParsesFully(t, in, f)
				should.Equal(t, want, got)
			})
		}
	})

	recoverCases := []string{
		"[]",
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, in := range recoverCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				p := parsetest.NewParser(t, in+" 123")
				got := parsetest.AssertMatchesButError(t, p, f)
				wantLine, wantCol, wantIndex := parsetest.CalcEnd(1, 1, 0, in)
				parsetest.AssertPosition(t, p, wantLine, wantCol, wantIndex)
				should.Equal(t, want, got)
			})
		}
	})
}

func TestMapType(t *testing.T) {
	t.Parallel()
	testMapType(t, MapType())
}

func testMapType(t *testing.T, f parser.Func[*ast.Type]) {
	successCases := []string{
		"map[int]string", "map[string]int", "map[int]map[string]string",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, in := range successCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				got := parsetest.ParsesFully(t, in, f)
				should.Equal(t, want, got)
			})
		}
	})

	recoverCases := []string{
		"map", "map int", "map[int", "map[int]",
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, in := range recoverCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				p := parsetest.NewParser(t, in+" 123")
				got := parsetest.AssertMatchesButError(t, p, f)
				wantLine, wantCol, wantIndex := parsetest.CalcEnd(1, 1, 0, in)
				parsetest.AssertPosition(t, p, wantLine, wantCol, wantIndex)
				should.Equal(t, want, got)
			})
		}
	})
}

func TestChannelType(t *testing.T) {
	t.Parallel()
	testChannelType(t, ChannelType())
}

func testChannelType(t *testing.T, f parser.Func[*ast.Type]) {
	successCases := []string{
		"chan int", "chan string", "chan chan int",
		"chan<- int", "<-chan string", "<-chan<- chan int",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, in := range successCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				got := parsetest.ParsesFully(t, in, f)
				should.Equal(t, want, got)
			})
		}
	})

	recoverCases := []string{
		"chan", "<-", "<-chan", "chan<-", "<-chan<-",
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, in := range recoverCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				want := &ast.Type{
					Type:  in,
					From:  ast.Position{Line: 1, Col: 1},
					Until: parsetest.CalcEndPos(in),
				}

				p := parsetest.NewParser(t, in+" 123")
				got := parsetest.AssertMatchesButError(t, p, f)
				wantLine, wantCol, wantIndex := parsetest.CalcEnd(1, 1, 0, in)
				parsetest.AssertPosition(t, p, wantLine, wantCol, wantIndex)
				should.Equal(t, want, got)
			})
		}
	})
}

func TestNamedType(t *testing.T) {
	t.Parallel()
	testNamedType(t, NamedType())
}

func testNamedType(t *testing.T, f parser.Func[*ast.NamedType]) {
	tests := []struct {
		name string
		in   string
		want *ast.NamedType
	}{
		{
			name: "int",
			in:   "int",
			want: &ast.NamedType{
				Name: &ast.Identifier{
					Name:     "int",
					Position: &ast.Position{Line: 1, Col: 1},
				},
			},
		}, {
			name: "custom",
			in:   "foo",
			want: &ast.NamedType{
				Name: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
			},
		}, {
			name: "qualified",
			in:   "foo.Bar",
			want: &ast.NamedType{
				Name: &ast.QualifiedIdentifier{
					Package: &ast.Identifier{
						Name:     "foo",
						Position: &ast.Position{Line: 1, Col: 1},
					},
					Dot: &ast.Position{Line: 1, Col: 4},
					Name: &ast.Identifier{
						Name:     "Bar",
						Position: &ast.Position{Line: 1, Col: 5},
					},
				},
			},
		}, {
			name: "type args",
			in:   "foo.Bar[int, foobar]",
			want: &ast.NamedType{
				Name: &ast.QualifiedIdentifier{
					Package: &ast.Identifier{
						Name:     "foo",
						Position: &ast.Position{Line: 1, Col: 1},
					},
					Dot: &ast.Position{Line: 1, Col: 4},
					Name: &ast.Identifier{
						Name:     "Bar",
						Position: &ast.Position{Line: 1, Col: 5},
					},
				},
				TypeArgs: &ast.TypeArguments{
					LBracket: &ast.Position{Line: 1, Col: 8},
					Types: []*ast.Type{
						{
							Type: "int",
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{
									Name:     "int",
									Position: &ast.Position{Line: 1, Col: 9},
								},
							},
							From:  ast.Position{Line: 1, Col: 9},
							Until: ast.Position{Line: 1, Col: 12},
						}, {
							Type: "foobar",
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{
									Name:     "foobar",
									Position: &ast.Position{Line: 1, Col: 14},
								},
							},
							From:  ast.Position{Line: 1, Col: 14},
							Until: ast.Position{Line: 1, Col: 20},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 20},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, f)
			should.Equal(t, c.want, got)
		})
	}
}

func TestTypeParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.TypeParameters
	}{
		{
			name: "empty",
			in:   "[]",
			want: &ast.TypeParameters{
				LBracket: &ast.Position{Line: 1, Col: 1},
				RBracket: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "single",
			in:   "[T int]",
			want: &ast.TypeParameters{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Params: []*ast.TypeParameter{
					{
						Names: []*ast.Identifier{
							{Name: "T", Position: &ast.Position{Line: 1, Col: 2}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 4}},
							},
							Type:  "int",
							From:  ast.Position{Line: 1, Col: 4},
							Until: ast.Position{Line: 1, Col: 7},
						},
					},
				},
				RBracket: &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "single",
			in:   "[K int, V string]",
			want: &ast.TypeParameters{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Params: []*ast.TypeParameter{
					{
						Names: []*ast.Identifier{
							{Name: "K", Position: &ast.Position{Line: 1, Col: 2}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 4}},
							},
							Type:  "int",
							From:  ast.Position{Line: 1, Col: 4},
							Until: ast.Position{Line: 1, Col: 7},
						},
					}, {
						Names: []*ast.Identifier{
							{Name: "V", Position: &ast.Position{Line: 1, Col: 9}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "string", Position: &ast.Position{Line: 1, Col: 11}},
							},
							Type:  "string",
							From:  ast.Position{Line: 1, Col: 11},
							Until: ast.Position{Line: 1, Col: 17},
						},
					},
				},
				RBracket: &ast.Position{Line: 1, Col: 17},
			},
		}, {
			name: "comma",
			in:   "[T int,]",
			want: &ast.TypeParameters{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Params: []*ast.TypeParameter{
					{
						Names: []*ast.Identifier{
							{Name: "T", Position: &ast.Position{Line: 1, Col: 2}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 4}},
							},
							Type:  "int",
							From:  ast.Position{Line: 1, Col: 4},
							Until: ast.Position{Line: 1, Col: 7},
						},
					},
				},
				RBracket: &ast.Position{Line: 1, Col: 8},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, TypeParameters())
			should.Equal(t, c.want, got)
		})
	}
}

func TestTypeParameterDecl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.TypeParameter
	}{
		{
			name: "single",
			in:   "T int",
			want: &ast.TypeParameter{
				Names: []*ast.Identifier{
					{Name: "T", Position: &ast.Position{Line: 1, Col: 1}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 3}},
					},
					Type:  "int",
					From:  ast.Position{Line: 1, Col: 3},
					Until: ast.Position{Line: 1, Col: 6},
				},
			},
		}, {
			name: "multiple",
			in:   "K, V string",
			want: &ast.TypeParameter{
				Names: []*ast.Identifier{
					{Name: "K", Position: &ast.Position{Line: 1, Col: 1}},
					{Name: "V", Position: &ast.Position{Line: 1, Col: 4}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Identifier{Name: "string", Position: &ast.Position{Line: 1, Col: 6}},
					},
					Type:  "string",
					From:  ast.Position{Line: 1, Col: 6},
					Until: ast.Position{Line: 1, Col: 12},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, TypeParameterDecl())
			should.Equal(t, c.want, got)
		})
	}
}
