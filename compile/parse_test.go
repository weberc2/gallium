package main

import (
	"fmt"
	"testing"
)

func TestParseFile(t *testing.T) {
	input := `package main

	fn double(i: int) -> int {
		let two = 2;
		mul(i, two)
	}

	fn main() {
		println(double(1));
	}`

	wanted := File{
		Package: "main",
		Decls: []FuncDecl{{
			Name: "double",
			Args: []ArgDecl{{Name: "i", Type: "int"}},
			Ret:  "int",
			Body: ExprBlockExpr(BlockExpr{
				Stmts: []Stmt{StmtLetStmt(LetStmt{
					Binding: "two",
					Expr:    ExprInt(2),
				})},
				Expr: ExprCallExpr(&CallExpr{
					Callable:  ExprIdent("mul"),
					Arguments: []Expr{ExprIdent("i"), ExprIdent("two")},
				}).Ptr(),
			}),
		}, {
			Name: "main",
			Body: ExprBlockExpr(BlockExpr{
				Stmts: []Stmt{
					StmtExpr(ExprCallExpr(&CallExpr{
						Callable: ExprIdent("println"),
						Arguments: []Expr{
							ExprCallExpr(&CallExpr{
								Callable:  ExprIdent("double"),
								Arguments: []Expr{ExprInt(1)},
							}),
						},
					})),
				},
			}),
		}},
	}

	file, _, err := parseFile(NewStringInput(input))
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if !file.Equal(wanted) {
		t.Fatalf("Wanted:\n%s\n\nGot:\n%s", wanted.Pretty(0), file.Pretty(0))
	}
}

func TestParseFuncDecl(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wanted    FuncDecl
		wantedErr error
		skip      bool
	}{{
		name:   "no-args_no-ret_int-expr-body",
		input:  "fn thirtyThree() 33",
		wanted: FuncDecl{Name: "thirtyThree", Body: ExprInt(33)},
	}, {
		name:  "no-args_explicit-ret_int-expr-body",
		input: "fn thirtyThree() -> int 33",
		wanted: FuncDecl{
			Name: "thirtyThree",
			Body: ExprInt(33),
			Ret:  Type("int"),
		},
	}, {
		name:  "one-arg_explicit-ret_ident-expr-body",
		input: "fn identity(i:int) -> int i",
		wanted: FuncDecl{
			Name: "identity",
			Args: []ArgDecl{{Name: "i", Type: Type("int")}},
			Body: ExprIdent("i"),
			Ret:  Type("int"),
		},
	}, {
		name:   "empty-body",
		input:  "fn empty() {}",
		wanted: FuncDecl{Name: "empty", Body: ExprBlockExpr(BlockExpr{})},
	}, {
		name:  "complex-body",
		input: "fn complex() -> int { let x = pow(4, 2); div(x, 2) }",
		wanted: FuncDecl{
			Name: "complex",
			Body: ExprBlockExpr(BlockExpr{
				Stmts: []Stmt{
					StmtLetStmt(LetStmt{
						Binding: "x",
						Expr: ExprCallExpr(&CallExpr{
							Callable:  ExprIdent("pow"),
							Arguments: []Expr{ExprInt(4), ExprInt(2)},
						}),
					}),
				},
				Expr: ExprCallExpr(&CallExpr{
					Callable:  ExprIdent("div"),
					Arguments: []Expr{ExprIdent("x"), ExprInt(2)},
				}).Ptr(),
			}),
			Ret: Type("int"),
		},
	}}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			decl, _, err := parseFuncDecl(NewStringInput(testCase.input))
			if err != nil {
				t.Fatal("Unexpected error:", err)
			}

			if !testCase.wanted.Equal(decl) {
				t.Fatalf(
					"Wanted:\n%s\n\nGot:\n%s",
					testCase.wanted.Pretty(0),
					decl.Pretty(0),
				)
			}
		})
	}
}

func TestParseExpr(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wanted    Expr
		wantedErr error
		skip      bool
	}{{
		name:   "ident",
		input:  "foo",
		wanted: ExprIdent("foo"),
	}, {
		name:   "str",
		input:  `"some string"`,
		wanted: ExprStr("some string"),
	}, {
		name:   "int",
		input:  "123",
		wanted: ExprInt(123),
	}, {
		name:  "bin expr eq",
		input: "a==b",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprIdent("a"),
			Operator: "==",
			Right:    ExprIdent("b"),
		}),
	}, {
		name:  "bin expr ne",
		input: "1!=2",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "!=",
			Right:    ExprInt(2),
		}),
	}, {
		name:  "bin expr gt",
		input: `"abc">"def"`,
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprStr("abc"),
			Operator: ">",
			Right:    ExprStr("def"),
		}),
	}, {
		name:  "bin expr lt",
		input: `"abc"<"def"`,
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprStr("abc"),
			Operator: "<",
			Right:    ExprStr("def"),
		}),
	}, {
		name:  "bin expr ge",
		input: "1>=2",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: ">=",
			Right:    ExprInt(2),
		}),
	}, {
		name:  "bin expr le",
		input: "1<=2",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "<=",
			Right:    ExprInt(2),
		}),
	}, {
		name:  "bin expr plus",
		input: "1+2",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "+",
			Right:    ExprInt(2),
		}),
	}, {
		name:  "bin expr minus",
		input: "1-2",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "-",
			Right:    ExprInt(2),
		}),
	}, {
		name:  "bin expr multiply",
		input: "1*2",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "*",
			Right:    ExprInt(2),
		}),
	}, {
		name:  "bin expr divide",
		input: "1/2",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "/",
			Right:    ExprInt(2),
		}),
	}, {
		name:  "bin expr modulo",
		input: "1%2",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "%",
			Right:    ExprInt(2),
		}),
	}, {
		name:  "bin expr order of operations - multiplication > addition",
		input: "1+2*3",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "+",
			Right: ExprBinExpr(&BinExpr{
				Left:     ExprInt(2),
				Operator: "*",
				Right:    ExprInt(3),
			}),
		}),
	}, {
		name:  "bin expr order of operations - division > addition",
		input: "1+2/3",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "+",
			Right: ExprBinExpr(&BinExpr{
				Left:     ExprInt(2),
				Operator: "/",
				Right:    ExprInt(3),
			}),
		}),
	}, {
		name:  "bin expr order of operations - multiplication > subtraction",
		input: "1-2*3",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "-",
			Right: ExprBinExpr(&BinExpr{
				Left:     ExprInt(2),
				Operator: "*",
				Right:    ExprInt(3),
			}),
		}),
	}, {
		name:  "bin expr order of operations - division > subtraction",
		input: "1-2/3",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "-",
			Right: ExprBinExpr(&BinExpr{
				Left:     ExprInt(2),
				Operator: "/",
				Right:    ExprInt(3),
			}),
		}),
	}, {
		name:  "bin expr recursive",
		skip:  true,
		input: "1+1+1",
		wanted: ExprBinExpr(&BinExpr{
			Left: ExprBinExpr(&BinExpr{
				Left:     ExprInt(1),
				Operator: "+",
				Right:    ExprInt(1),
			}),
			Operator: "+",
			Right:    ExprInt(1),
		}),
	}, {
		name:  "bin expr with unary expr left operand",
		input: "-1+1",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprUnExpr(&UnExpr{Operator: '-', Operand: ExprInt(1)}),
			Operator: "+",
			Right:    ExprInt(1),
		}),
	}, {
		name:  "bin expr with unary expr right operand",
		input: "1+-1",
		wanted: ExprBinExpr(&BinExpr{
			Left:     ExprInt(1),
			Operator: "+",
			Right: ExprUnExpr(&UnExpr{
				Operator: '-',
				Operand:  ExprInt(1),
			}),
		}),
	}, {
		name:   "unary expr bang",
		input:  "!1",
		wanted: ExprUnExpr(&UnExpr{Operator: '!', Operand: ExprInt(1)}),
	}, {
		name:   "unary expr star",
		input:  "*foo",
		wanted: ExprUnExpr(&UnExpr{Operator: '*', Operand: ExprIdent("foo")}),
	}, {
		name:   "unary expr negate",
		input:  "-4",
		wanted: ExprUnExpr(&UnExpr{Operator: '-', Operand: ExprInt(4)}),
	}, {
		name:   "empty block expr",
		input:  "{}",
		wanted: ExprBlockExpr(BlockExpr{}),
	}, {
		name:   "simple block expr",
		input:  "{ foo }",
		wanted: ExprBlockExpr(BlockExpr{Expr: ExprIdent("foo").Ptr()}),
	}, {
		name:   "empty block expr",
		input:  "{}",
		wanted: ExprBlockExpr(BlockExpr{}),
	}, {
		name:  "block expr w only let statement",
		input: "{let foo = bar;}",
		wanted: ExprBlockExpr(BlockExpr{
			Stmts: []Stmt{
				StmtLetStmt(LetStmt{Binding: "foo", Expr: ExprIdent("bar")}),
			},
		}),
	}, {
		name:  "block expr w only expr statement",
		input: "{bar;}",
		wanted: ExprBlockExpr(
			BlockExpr{Stmts: []Stmt{StmtExpr(ExprIdent("bar"))}},
		),
	}, {
		name:  "block expr with multiple statements",
		input: "{let foo = bar; bar;}",
		wanted: ExprBlockExpr(BlockExpr{Stmts: []Stmt{
			StmtLetStmt(LetStmt{Binding: "foo", Expr: ExprIdent("bar")}),
			StmtExpr(ExprIdent("bar")),
		}}),
	}, {
		name:  "block expr w statements and expr",
		input: "{let foo = bar; foo}",
		wanted: ExprBlockExpr(BlockExpr{
			Stmts: []Stmt{
				StmtLetStmt(LetStmt{Binding: "foo", Expr: ExprIdent("bar")}),
			},
			Expr: ExprIdent("foo").Ptr(),
		}),
	}, {
		name:   "call expr no args",
		input:  "foo()",
		wanted: ExprCallExpr(&CallExpr{Callable: ExprIdent("foo")}),
	}, {
		name:  "call expr one arg",
		input: "foo(bar)",
		wanted: ExprCallExpr(&CallExpr{
			ExprIdent("foo"),
			[]Expr{ExprIdent("bar")},
		}),
	}, {
		name:  "call expr multi args",
		input: "foo(bar, \"abc\", 123)",
		wanted: ExprCallExpr(&CallExpr{
			ExprIdent("foo"),
			[]Expr{ExprIdent("bar"), ExprStr("abc"), ExprInt(123)},
		}),
	}, {
		name:  "call expr w non-ident callable",
		input: "getGetCallable()()(bar)",
		wanted: ExprCallExpr(&CallExpr{
			ExprCallExpr(&CallExpr{
				Callable: ExprCallExpr(&CallExpr{
					Callable: ExprIdent("getGetCallable"),
				}),
			}),
			[]Expr{ExprIdent("bar")},
		}),
	}, {
		name:  "call expr w call expr args",
		input: "foo(bar(baz()))",
		wanted: ExprCallExpr(&CallExpr{
			ExprIdent("foo"),
			[]Expr{
				ExprCallExpr(&CallExpr{
					ExprIdent("bar"),
					[]Expr{
						ExprCallExpr(&CallExpr{Callable: ExprIdent("baz")}),
					},
				}),
			},
		}),
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.skip {
				fmt.Println("SKIP:", testCase.name)
				t.SkipNow()
			}
			expr, _, err := parseExpr(NewStringInput(testCase.input))
			if err != testCase.wantedErr {
				t.Fatalf(
					"%s: Wanted error '%v'; got '%v'",
					testCase.name,
					testCase.wantedErr,
					err,
				)
			}
			if !expr.Equal(testCase.wanted) {
				t.Fatalf(
					"%s\nWanted expr:\n%s\n\nGot expr:\n%s",
					testCase.name,
					testCase.wanted.Pretty(0),
					expr.Pretty(0),
				)
			}
		})
	}
}
