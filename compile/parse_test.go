package main

import (
	"testing"
)

func TestParsers(t *testing.T) {
	testCases := []struct {
		name       string
		input      Input
		wanted     Node
		wantedRest Input
		parse      Parser
	}{
		{
			name:   "term-ident",
			input:  "bar",
			wanted: TermIdent("bar"),
			parse:  ParseTerm,
		},
		{
			name:   "term-int",
			input:  "132",
			wanted: TermInt(132),
			parse:  ParseTerm,
		},
		{
			name:   "term-string",
			input:  `"foo"`,
			wanted: TermString("foo"),
			parse:  ParseTerm,
		},
		{
			name:   "term-paren-group",
			input:  "(foo)",
			wanted: TermExpr(NewExpr(TermIdent("foo"))),
			parse:  ParseTerm,
		},
		{
			name:   "expr-ident",
			input:  "_abc123",
			wanted: NewExpr(TermIdent("_abc123")),
			parse:  ParseExpr,
		},
		{
			name:  "expr",
			input: `foo 1 2 "abc"`,
			wanted: NewExpr(
				TermIdent("foo"),
				TermInt(1),
				TermInt(2),
				TermString("abc"),
			),
			parse: ParseExpr,
		},
		{
			name:  "expr-w-paren-group",
			input: "foo (bar baz) qux",
			wanted: NewExpr(
				TermIdent("foo"),
				TermExpr(NewExpr(TermIdent("bar"), TermIdent("baz"))),
				TermIdent("qux"),
			),
			parse: ParseExpr,
		},
		{
			name:   "binding-decl",
			input:  "let x = 42",
			wanted: BindingDecl{"x", NewExpr(TermInt(42))},
			parse:  ParseBindingDecl,
		},
		{
			name:   "stmt-binding",
			input:  "let x = 42;",
			wanted: StmtBinding(BindingDecl{"x", NewExpr(TermInt(42))}),
			parse:  ParseStmt,
		},
		{
			name:  "stmt-expr",
			input: "foo 42 i;",
			wanted: StmtExpr(NewExpr(
				TermIdent("foo"),
				TermInt(42),
				TermIdent("i"),
			)),
			parse: ParseStmt,
		},
		{
			name:  "block",
			input: "{ foo; bar; baz }",
			wanted: Block{
				Stmts: []Stmt{
					StmtExpr(NewExpr(TermIdent("foo"))),
					StmtExpr(NewExpr(TermIdent("bar"))),
				},
				Expr: &Expr{Head: TermIdent("baz")},
			},
			parse: ParseBlock,
		},
		{
			name:  "term-block",
			input: "{ foo; bar; baz }",
			wanted: TermBlock(Block{
				Stmts: []Stmt{
					StmtExpr(NewExpr(TermIdent("foo"))),
					StmtExpr(NewExpr(TermIdent("bar"))),
				},
				Expr: &Expr{Head: TermIdent("baz")},
			}),
			parse: ParseTerm,
		},
		{
			name:  "lambda-simple",
			input: "|x| x",
			wanted: Lambda{
				Args: []ArgSpec{{Name: "x"}},
				Body: NewExpr(TermIdent("x")),
			},
			parse: ParseLambda,
		},
		{
			name:  "lambda-multi-args",
			input: "|x, y| add x y",
			wanted: Lambda{
				Args: []ArgSpec{{Name: "x"}, {Name: "y"}},
				Body: NewExpr(
					TermIdent("add"),
					TermIdent("x"),
					TermIdent("y"),
				),
			},
			parse: ParseLambda,
		},
		{
			name:  "lambda-w-return-type",
			input: "|x, y| -> int 5",
			wanted: Lambda{
				Args: []ArgSpec{{Name: "x"}, {Name: "y"}},
				Ret:  &TypeExpr{Head: "int"},
				Body: NewExpr(TermInt(5)),
			},
			parse: ParseLambda,
		},
		{
			name: "lambda-w-block-expr",
			input: `|x, y| -> int {
				let z = add x y;
				add z z
			}`,
			wanted: Lambda{
				Args: []ArgSpec{{Name: "x"}, {Name: "y"}},
				Ret:  &TypeExpr{Head: "int"},
				Body: NewExpr(TermBlock(Block{
					Stmts: []Stmt{StmtBinding(BindingDecl{
						Binding: "z",
						Expr: NewExpr(
							TermIdent("add"),
							TermIdent("x"),
							TermIdent("y"),
						),
					})},
					Expr: NewExpr(
						TermIdent("add"),
						TermIdent("z"),
						TermIdent("z"),
					).Ptr(),
				})),
			},
			parse: ParseLambda,
		},
		{
			name:  "lambda-w-typed-arg",
			input: "|x: int| -> int x",
			wanted: Lambda{
				Args: []ArgSpec{{Name: "x", Type: &TypeExpr{Head: "int"}}},
				Ret:  &TypeExpr{Head: "int"},
				Body: NewExpr(TermIdent("x")),
			},
			parse: ParseLambda,
		},
		{
			name:  "lambda-w-multiple-typed-args",
			input: "|x: int, y: int| -> int add x y",
			wanted: Lambda{
				Args: []ArgSpec{
					{Name: "x", Type: &TypeExpr{Head: "int"}},
					{Name: "y", Type: &TypeExpr{Head: "int"}},
				},
				Ret: &TypeExpr{Head: "int"},
				Body: NewExpr(
					TermIdent("add"),
					TermIdent("x"),
					TermIdent("y"),
				),
			},
			parse: ParseLambda,
		},
		{
			name:  "term-lambda",
			input: "|x| x",
			wanted: TermLambda(Lambda{
				Args: []ArgSpec{{Name: "x"}},
				Body: NewExpr(TermIdent("x")),
			}),
			parse: ParseTerm,
		},
		{
			name:  "expr-w-lambda",
			input: "(|x| x) 4",
			wanted: NewExpr(
				TermExpr(NewExpr(TermLambda(Lambda{
					Args: []ArgSpec{{Name: "x"}},
					Body: NewExpr(TermIdent("x")),
				}))),
				TermInt(4),
			),
			parse: ParseExpr,
		},
		{
			name:   "file-wo-imports-or-decls",
			input:  "package main\n",
			wanted: File{Package: "main"},
			parse:  ParseFile,
		},
		{
			name:  "file-w-decl",
			input: "package main\nlet x = 4;",
			wanted: File{
				Package: "main",
				Decls: []BindingDecl{{
					Binding: "x",
					Expr:    NewExpr(TermInt(4)),
				}},
			},
			parse: ParseFile,
		},
		{
			name:   "file-w-single-import",
			input:  "package main\nimport (\"foo\")\n",
			wanted: File{Package: "main", Imports: []string{"foo"}},
			parse:  ParseFile,
		},
		{
			name: "file-w-multi-imports",
			input: `package main

					import (
						"foo"
						"bar"
					)

					let x = 4;
					let y = "lol";`,
			wanted: File{
				"main",
				[]string{"foo", "bar"},
				[]BindingDecl{
					{"x", NewExpr(TermInt(4))},
					{"y", NewExpr(TermString("lol"))},
				},
			},
			parse: ParseFile,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rslt := testCase.parse(testCase.input)
			if rslt.IsErr() {
				t.Fatal("Unexpected error:", rslt.Err)
			}
			if rslt.Rest != testCase.wantedRest {
				t.Fatalf(
					"Wanted remainder: '%s', got: '%s'",
					testCase.wantedRest.Sample(15),
					rslt.Rest.Sample(15),
				)
			}
			if !rslt.Node.Equal(testCase.wanted) {
				t.Fatalf("Wanted:\n%s\n\nGot:\n%s", testCase.wanted, rslt.Node)
			}
		})
	}
}
