package infer

import (
	"testing"

	"github.com/weberc2/gallium/ast"

	"github.com/kr/pretty"
)

func TestInfer(t *testing.T) {
	testCases := []struct {
		Name      string
		Env       Environment
		Input     ast.Expr
		Wanted    ast.Expr
		WantedErr bool
		Skip      bool
	}{
		{
			Name:   "simple-int-lit",
			Env:    Environment{},
			Input:  ast.Expr{Node: ast.IntLit(0)},
			Wanted: ast.Expr{Type: ast.Primitive("int"), Node: ast.IntLit(0)},
		},
		{
			Name:  "simple-string-lit",
			Env:   Environment{},
			Input: ast.Expr{Node: ast.StringLit("")},
			Wanted: ast.Expr{
				Type: ast.Primitive("string"),
				Node: ast.StringLit(""),
			},
		},
		{
			Name:  "string-ident-simple",
			Env:   Environment{"foo": ast.Primitive("string")},
			Input: ast.Expr{Node: ast.Ident("foo")},
			Wanted: ast.Expr{
				Type: ast.Primitive("string"),
				Node: ast.Ident("foo"),
			},
		},
		{
			Name:   "tuple-lit-simple",
			Env:    Environment{},
			Input:  ast.Expr{Node: ast.TupleLit{}},
			Wanted: ast.Expr{Type: ast.TupleSpec{}, Node: ast.TupleLit{}},
		},
		{
			Name:  "tuple-lit-simple-one-elt",
			Env:   Environment{},
			Input: ast.Expr{Node: ast.TupleLit{ast.Expr{Node: ast.IntLit(0)}}},
			Wanted: ast.Expr{
				Type: ast.TupleSpec{ast.Primitive("int")},
				Node: ast.TupleLit{ast.Expr{
					Node: ast.IntLit(0),
					Type: ast.Primitive("int"),
				}},
			},
		},
		{
			Name: "tuple-lit-simple-multi-elt",
			Env:  Environment{},
			Input: ast.Expr{Node: ast.TupleLit{
				ast.Expr{Node: ast.IntLit(0)},
				ast.Expr{Node: ast.StringLit("")},
			}},
			Wanted: ast.Expr{
				Type: ast.TupleSpec{
					ast.Primitive("int"),
					ast.Primitive("string"),
				},
				Node: ast.TupleLit{
					ast.Expr{Node: ast.IntLit(0), Type: ast.Primitive("int")},
					ast.Expr{
						Node: ast.StringLit(""),
						Type: ast.Primitive("string"),
					},
				},
			},
		},
		{
			// x => 4
			Name: "func-lit-constant",
			Env:  Environment{},
			Input: ast.Expr{Node: ast.FuncLit{
				Arg:  "x",
				Body: ast.Expr{Node: ast.IntLit(4)},
			}},
			Wanted: ast.Expr{
				Node: ast.FuncLit{
					Arg: "x",
					Body: ast.Expr{
						Node: ast.IntLit(4),
						Type: ast.Primitive("int"),
					},
				},
				Type: ast.FuncSpec{
					Arg: ast.TypeVar("a"),
					Ret: ast.Primitive("int"),
				},
			},
		},
		{
			// x => x
			// TODO: The remaining typevar in `Wanted` should probably be 'a',
			// but the functionally important thing is that the typevar is
			// consistent.
			Name: "func-lit-identity",
			Env:  Environment{},
			Input: ast.Expr{Node: ast.FuncLit{
				Arg:  "x",
				Body: ast.Expr{Node: ast.Ident("x")},
			}},
			Wanted: ast.Expr{
				Node: ast.FuncLit{
					Arg: "x",
					Body: ast.Expr{
						Node: ast.Ident("x"),
						Type: ast.TypeVar("b"),
					},
				},
				Type: ast.FuncSpec{
					Arg: ast.TypeVar("b"),
					Ret: ast.TypeVar("b"),
				},
			},
		},
		{
			Name: "block-just-expr",
			Env:  Environment{},
			Input: ast.Expr{Node: ast.Block{
				Expr: ast.Expr{Node: ast.IntLit(1)},
			}},
			Wanted: ast.Expr{
				Type: ast.Primitive("int"),
				Node: ast.Block{
					Expr: ast.Expr{
						Type: ast.Primitive("int"),
						Node: ast.IntLit(1),
					},
				},
			},
		},
		{
			Name: "block-w-let-decl",
			Env:  Environment{},
			Input: ast.Expr{Node: ast.Block{
				Stmts: []ast.Stmt{
					ast.LetDecl{
						Ident:   "x",
						Binding: ast.Expr{Node: ast.StringLit("foo")},
					},
				},
				Expr: ast.Expr{Node: ast.Ident("x")},
			}},
			Wanted: ast.Expr{
				Type: ast.Primitive("string"),
				Node: ast.Block{
					Stmts: []ast.Stmt{
						ast.LetDecl{
							Ident:   "x",
							Binding: ast.Expr{Node: ast.StringLit("foo")},
						},
					},
					Expr: ast.Expr{
						Type: ast.Primitive("string"),
						Node: ast.Ident("x"),
					},
				},
			},
		},
		{
			Name: "block-w-dependent-let-decls",
			Env: Environment{
				ast.Ident("add"): ast.FuncSpec{
					Arg: ast.Primitive("int"),
					Ret: ast.FuncSpec{
						Arg: ast.Primitive("int"),
						Ret: ast.Primitive("int"),
					},
				},
			},
			Input: ast.Expr{Node: ast.Block{
				Stmts: []ast.Stmt{
					ast.LetDecl{
						Ident: ast.Ident("y"),
						Binding: ast.Expr{Node: ast.Call{
							Fn: ast.Expr{Node: ast.Call{
								Fn:  ast.Expr{Node: ast.Ident("add")},
								Arg: ast.Expr{Node: ast.IntLit(1)},
							}},
							Arg: ast.Expr{Node: ast.IntLit(1)},
						}},
					},
				},
				Expr: ast.Expr{Node: ast.Ident("y")},
			}},
			Wanted: ast.Expr{
				Type: ast.Primitive("int"),
				Node: ast.Block{
					Stmts: []ast.Stmt{
						ast.LetDecl{
							Ident: ast.Ident("y"),
							Binding: ast.Expr{Node: ast.Call{
								Fn: ast.Expr{Node: ast.Call{
									Fn:  ast.Expr{Node: ast.Ident("add")},
									Arg: ast.Expr{Node: ast.IntLit(1)},
								}},
								Arg: ast.Expr{Node: ast.IntLit(1)},
							}},
						},
					},
					Expr: ast.Expr{
						Type: ast.Primitive("int"),
						Node: ast.Ident("y"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.Skip {
				t.SkipNow()
			}
			got, err := Infer(testCase.Env, testCase.Input)
			if err != nil {
				if !testCase.WantedErr {
					t.Fatal("Unexpected error:", err)
				}
				return
			}
			if !got.Equal(testCase.Wanted) {
				t.Fatalf(
					"WANTED:\n%# v\n\nGOT:\n%# v\n",
					pretty.Formatter(testCase.Wanted),
					pretty.Formatter(got),
				)
			}
		})
	}
}
