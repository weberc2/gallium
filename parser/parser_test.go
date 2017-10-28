package parser

import (
	"reflect"
	"testing"

	"github.com/weberc2/gallium/ast"
	"github.com/weberc2/gallium/combinator"

	"github.com/kr/pretty"
)

func TestParser(t *testing.T) {
	testCases := []struct {
		Name        string
		Input       combinator.Input
		WantedRest  combinator.Input
		WantedValue interface{}
		WantedErr   bool
		Parser      combinator.Parser
	}{
		{
			Name:        "ws",
			Input:       " \t\n",
			WantedValue: " \t\n",
			Parser:      combinator.WS,
		},
		{
			Name:        "can-ws",
			Input:       " \t\n",
			WantedValue: " \t\n",
			Parser:      combinator.CanWS,
		},
		// {
		// 	Name:        "func-spec-simple",
		// 	Input:       "fn()",
		// 	WantedValue: ast.FuncSpec{},
		// 	Parser:      FuncSpec,
		// },
		// {
		// 	Name:        "func-spec-no-args-w-ret",
		// 	Input:       "fn() -> int",
		// 	WantedValue: ast.FuncSpec{Ret: ast.TypeRef{Name: "int"}},
		// 	Parser:      FuncSpec,
		// },
		// {
		// 	Name:  "func-spec-one-arg-no-ret",
		// 	Input: "fn(i int)",
		// 	WantedValue: ast.FuncSpec{
		// 		Arg: ast.ArgSpec{"i", ast.TypeRef{Name: "int"}},
		// 	},
		// 	Parser: FuncSpec,
		// },
		// {
		// 	Name:  "func-spec-multi-args-no-ret",
		// 	Input: "fn(i int, j int)",
		// 	WantedValue: ast.FuncSpec{Args: []ast.ArgSpec{
		// 		{"i", ast.TypeRef{Name: "int"}},
		// 		{"j", ast.TypeRef{Name: "int"}},
		// 	}},
		// 	Parser: FuncSpec,
		// },
		// {
		// 	Name:  "func-spec-multi-args-w-ret",
		// 	Input: "fn(i int, j int) -> bool",
		// 	WantedValue: ast.FuncSpec{
		// 		Args: []ast.ArgSpec{
		// 			{"i", ast.TypeRef{Name: "int"}},
		// 			{"j", ast.TypeRef{Name: "int"}},
		// 		},
		// 		Ret: ast.TypeRef{Name: "bool"},
		// 	},
		// 	Parser: FuncSpec,
		// },
		// {
		// 	Name:  "func-spec-w-untyped-args",
		// 	Input: "fn(i, j)",
		// 	WantedValue: ast.FuncSpec{Args: []ast.ArgSpec{
		// 		{Name: "i"},
		// 		{Name: "j"},
		// 	}},
		// 	Parser: FuncSpec,
		// },
		{
			Name:  "type-decl-simple",
			Input: "type foo = int",
			WantedValue: ast.TypeDecl{
				Name: "foo",
				Type: ast.TypeRef{Name: "int"},
			},
			Parser: TypeDecl,
		},
		{
			Name:  "type-decl-generic",
			Input: "type foo a b = bar a b",
			WantedValue: ast.TypeDecl{
				Name: "foo",
				Type: ast.TypeRef{
					Name: "bar",
					Arg:  ast.TypeRef{Name: "a", Arg: ast.TypeRef{Name: "b"}},
				},
				Args: []ast.TypeVar{"a", "b"},
			},
			Parser: TypeDecl,
		},
		{
			Name:  "type-decl-generic-tuple",
			Input: "type foo a b = (int, a, b)",
			WantedValue: ast.TypeDecl{
				Name: "foo",
				Type: ast.TupleSpec{
					ast.TypeRef{Name: "int"},
					ast.TypeRef{Name: "a"},
					ast.TypeRef{Name: "b"},
				},
				Args: []ast.TypeVar{"a", "b"},
			},
			Parser: TypeDecl,
		},
		// {
		// 	Name:  "type-decl-simple-function",
		// 	Input: "type Parser = fn(input Input) -> Result",
		// 	WantedValue: ast.TypeDecl{
		// 		Name: "Parser",
		// 		Type: ast.FuncSpec{
		// 			Args: []ast.ArgSpec{
		// 				{Name: "input", Type: ast.TypeRef{Name: "Input"}},
		// 			},
		// 			Ret: ast.TypeRef{Name: "Result"},
		// 		},
		// 	},
		// 	Parser: TypeDecl,
		// },
		{
			Name:        "string-lit-empty",
			Input:       `""`,
			WantedValue: ast.StringLit(""),
			Parser:      StringLit,
		},
		{
			Name:        "string-lit-non-empty",
			Input:       `"abc"`,
			WantedValue: ast.StringLit("abc"),
			Parser:      StringLit,
		},
		{
			Name:        "int-lit",
			Input:       "10",
			WantedValue: ast.IntLit(10),
			Parser:      IntLit,
		},
		{
			Name:        "ident-one-char",
			Input:       "f",
			WantedValue: ast.Ident("f"),
			Parser:      Ident,
		},
		{
			Name:        "ident-one-char-underscore",
			Input:       "_",
			WantedValue: ast.Ident("_"),
			Parser:      Ident,
		},
		{
			Name:        "ident-many-chars",
			Input:       "_foo123",
			WantedValue: ast.Ident("_foo123"),
			Parser:      Ident,
		},
		{
			Name:        "tuple-lit-empty",
			Input:       "()",
			WantedValue: ast.TupleLit{},
			Parser:      TupleLit,
		},
		{
			Name:  "let-decl-no-type",
			Input: "let x = 42",
			WantedValue: ast.LetDecl{
				ast.Ident("x"),
				ast.Expr{Node: ast.IntLit(42)},
			},
			Parser: LetDecl,
		},
		{
			Name:        "block-empty",
			Input:       "{}",
			WantedValue: ast.Block{},
			Parser:      Block,
		},
		{
			Name:        "block-lone-expr",
			Input:       "{ 42 }",
			WantedValue: ast.Block{Expr: ast.Expr{Node: ast.IntLit(42)}},
			Parser:      Block,
		},
		{
			Name:        "block-nested",
			Input:       "{ {} }",
			WantedValue: ast.Block{Expr: ast.Expr{Node: ast.Block{}}},
			Parser:      Block,
		},
		{
			Name:  "block-w-let-stmt",
			Input: "{ let x = 42; }",
			WantedValue: ast.Block{Stmts: []ast.Stmt{ast.LetDecl{
				ast.Ident("x"),
				ast.Expr{Node: ast.IntLit(42)},
			}}},
			Parser: Block,
		},
		{
			Name:  "call",
			Input: "foo bar",
			WantedValue: ast.Call{
				Fn:  ast.Expr{Node: ast.Ident("foo")},
				Arg: ast.Expr{Node: ast.Ident("bar")},
			},
			Parser: Call,
		},
		{
			Name:  "call-multi-arg",
			Input: "add 1 2",
			WantedValue: ast.Call{
				Fn: ast.Expr{Node: ast.Call{
					Fn:  ast.Expr{Node: ast.Ident("add")},
					Arg: ast.Expr{Node: ast.IntLit(1)},
				}},
				Arg: ast.Expr{Node: ast.IntLit(2)},
			},
			Parser: Call,
		},
		{
			Name:  "call-multi-arg-first-arg-ident",
			Input: "add x y",
			WantedValue: ast.Call{
				Fn: ast.Expr{Node: ast.Call{
					Fn:  ast.Expr{Node: ast.Ident("add")},
					Arg: ast.Expr{Node: ast.Ident("x")},
				}},
				Arg: ast.Expr{Node: ast.Ident("y")},
			},
			Parser: Call,
		},
		{
			Name:  "func-lit-int-body",
			Input: "_ -> 4",
			WantedValue: ast.FuncLit{
				Arg:  "_",
				Body: ast.Expr{Node: ast.IntLit(4)},
			},
			Parser: FuncLit,
		},
		{
			Name:  "func-lit-tuple-body",
			Input: `_ -> (a, 4, "")`,
			WantedValue: ast.FuncLit{
				Arg: "_",
				Body: ast.Expr{Node: ast.TupleLit{
					ast.Expr{Node: ast.Ident("a")},
					ast.Expr{Node: ast.IntLit(4)},
					ast.Expr{Node: ast.StringLit("")},
				}},
			},
			Parser: FuncLit,
		},
		{
			Name:  "func-lit-block-body",
			Input: "x -> { 42 }",
			WantedValue: ast.FuncLit{
				Arg: ast.Ident("x"),
				Body: ast.Expr{Node: ast.Block{Expr: ast.Expr{
					Node: ast.IntLit(42),
				}}},
			},
			Parser: FuncLit,
		},
		{
			Name:  "tuple-lit-multi-elt",
			Input: "(a, b, c)",
			WantedValue: ast.TupleLit{
				ast.Expr{Node: ast.Ident("a")},
				ast.Expr{Node: ast.Ident("b")},
				ast.Expr{Node: ast.Ident("c")},
			},
			Parser: TupleLit,
		},
		{
			Name:  "tuple-lit-many-simple-exprs",
			Input: `(1, "a", foo)`,
			WantedValue: ast.TupleLit{
				ast.Expr{Node: ast.IntLit(1)},
				ast.Expr{Node: ast.StringLit("a")},
				ast.Expr{Node: ast.Ident("foo")},
			},
			Parser: TupleLit,
		},
		{
			Name:  "tuple-lit-nested",
			Input: "((foo, bar))",
			WantedValue: ast.TupleLit{ast.Expr{Node: ast.TupleLit{
				ast.Expr{Node: ast.Ident("foo")},
				ast.Expr{Node: ast.Ident("bar")},
			}}},
			Parser: TupleLit,
		},
		{
			Name:        "expr-int-lit",
			Input:       "25",
			WantedValue: ast.Expr{Node: ast.IntLit(25)},
			Parser:      Expr,
		},
		{
			Name:        "expr-string-lit",
			Input:       `"abcd"`,
			WantedValue: ast.Expr{Node: ast.StringLit("abcd")},
			Parser:      Expr,
		},
		{
			Name:  "expr-tuple-lit",
			Input: "(foo, bar)",
			WantedValue: ast.Expr{Node: ast.TupleLit{
				ast.Expr{Node: ast.Ident("foo")},
				ast.Expr{Node: ast.Ident("bar")},
			}},
			Parser: Expr,
		},
		{
			Name:  "expr-block",
			Input: "{ 42 }",
			WantedValue: ast.Expr{Node: ast.Block{Expr: ast.Expr{
				Node: ast.IntLit(42),
			}}},
			Parser: Expr,
		},
		{
			Name:  "expr-call",
			Input: "foo 42",
			WantedValue: ast.Expr{Node: ast.Call{
				ast.Expr{Node: ast.Ident("foo")},
				ast.Expr{Node: ast.IntLit(42)},
			}},
			Parser: Expr,
		},
		{
			Name:  "expr-func-lit",
			Input: "a -> addOne a",
			WantedValue: ast.Expr{Node: ast.FuncLit{
				ast.Ident("a"),
				ast.Expr{Node: ast.Call{
					ast.Expr{Node: ast.Ident("addOne")},
					ast.Expr{Node: ast.Ident("a")},
				}},
			}},
			Parser: Expr,
		},
		{
			Name:  "expr-parens-group",
			Input: "print (add 1 2)",
			WantedValue: ast.Expr{Node: ast.Call{
				Arg: ast.Expr{Node: ast.Call{
					Fn: ast.Expr{Node: ast.Call{
						Fn:  ast.Expr{Node: ast.Ident("add")},
						Arg: ast.Expr{Node: ast.IntLit(1)},
					}},
					Arg: ast.Expr{Node: ast.IntLit(2)},
				}},
				Fn: ast.Expr{Node: ast.Ident("print")},
			}},
			Parser: Expr,
		},
		{
			Name:  "decl-type-decl",
			Input: "type foo = int",
			WantedValue: ast.TypeDecl{
				Name: "foo",
				Type: ast.TypeRef{Name: "int"},
			},
			Parser: Decl,
		},
		{
			Name:  "decl-let-decl",
			Input: "let x = 0",
			WantedValue: ast.LetDecl{
				ast.Ident("x"),
				ast.Expr{Node: ast.IntLit(0)},
			},
			Parser: Decl,
		},
		{
			Name:  "stmt-decl",
			Input: "let x = 0;",
			WantedValue: ast.LetDecl{
				ast.Ident("x"),
				ast.Expr{Node: ast.IntLit(0)},
			},
			Parser: Stmt,
		},
		{
			Name:  "stmt-expr",
			Input: `println "Hello, world!";`,
			WantedValue: ast.Expr{Node: ast.Call{
				ast.Expr{Node: ast.Ident("println")},
				ast.Expr{Node: ast.StringLit("Hello, world!")},
			}},
			Parser: Stmt,
		},
		{
			Name:        "file-empty",
			Input:       "package main",
			WantedValue: ast.File{Package: "main"},
			Parser:      File,
		},
		{
			Name:        "file-empty-trailing-whitespace",
			Input:       "package main\n",
			WantedValue: ast.File{Package: "main"},
			Parser:      File,
		},
		{
			Name: "file-w-lone-type-decl",
			Input: `package main

					type foo = int;`,
			WantedValue: ast.File{
				Package: "main",
				Stmts: []ast.Stmt{ast.TypeDecl{
					Name: "foo",
					Type: ast.TypeRef{Name: "int"},
				}},
			},
			Parser: File,
		},
		{
			Name: "file-w-many-type-decls",
			Input: `package main

					type x = foo;

					type y = bar;`,
			WantedValue: ast.File{
				Package: "main",
				Stmts: []ast.Stmt{
					ast.TypeDecl{Name: "x", Type: ast.TypeRef{Name: "foo"}},
					ast.TypeDecl{Name: "y", Type: ast.TypeRef{Name: "bar"}},
				},
			},
			Parser: File,
		},
		{
			Name: "file-w-type-and-let-decls",
			Input: `package main

					type X = Foo;

					let main = _ -> { println "Hello, world"; };
					`,
			WantedValue: ast.File{
				Package: "main",
				Stmts: []ast.Stmt{
					ast.TypeDecl{Name: "X", Type: ast.TypeRef{Name: "Foo"}},
					ast.LetDecl{
						ast.Ident("main"),
						ast.Expr{Node: ast.FuncLit{
							Arg: "_",
							Body: ast.Expr{Node: ast.Block{
								Stmts: []ast.Stmt{ast.Expr{Node: ast.Call{
									ast.Expr{Node: ast.Ident("println")},
									ast.Expr{
										Node: ast.StringLit("Hello, world"),
									},
								}}},
							}},
						}},
					},
				},
			},
			Parser: File,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := testCase.Parser(testCase.Input)
			if testCase.WantedErr && result.Err == nil {
				t.Fatal("Wanted an error but didn't get any")
			}
			if !testCase.WantedErr && result.Err != nil {
				t.Fatal("Unexpected error:", result)
			}

			if testCase.WantedRest != result.Rest {
				t.Fatalf(
					"Wanted REST: %#v; got REST: %#v",
					testCase.WantedRest,
					result.Rest,
				)
			}

			if wanted, ok := testCase.WantedValue.(ast.Node); ok {
				if got, ok := result.Value.(ast.Node); ok {
					if wanted.EqualNode(got) {
						return
					}
				}
			} else {
				if reflect.DeepEqual(result.Value, testCase.WantedValue) {
					return
				}
			}

			t.Fatalf(
				"Wanted:\n%# v\n\nGot:\n%# v\n",
				pretty.Formatter(testCase.WantedValue),
				pretty.Formatter(result.Value),
			)
		})
	}
}
